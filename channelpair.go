package zmq4

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

var uniqueIndex uint64 // counter for allocation of internal pairs

type ChannelPair struct {
	zSock       *Socket        // ZMQ socket
	zTx         []*Socket      // Pair of sockets for internal TX buffering
	zControl    []*Socket      // Pair of sockets for internal control channel
	wg          sync.WaitGroup // Waitgroup for synchronizing Close()
	txChan      chan [][]byte  // Transmit channel; user writes to this
	rxChan      chan [][]byte  // Receive channel; user reads from this
	errorChan   chan error     // Error channel, to propagate errors back to the user
	controlChan chan bool      // Internal control channel
}

const (
	ZC_IN  = iota
	ZC_OUT = iota
)

// func barrier()
//
// Magic voodoo to provide a 'complete memory barrier' as seemingly required
// to pass zmq sockets between threads

func barrier() {
	var mutex sync.Mutex
	mutex.Lock()
	mutex.Unlock()
}

// func getUniqueId() uint64
//
// returns a unique ID

func getUniqueId() uint64 {
	return atomic.AddUint64(&uniqueIndex, 1)
}

// func (cp *ChannelPair) runChannels()
//
// This function:
// 1. Reads the TX channel, and places the output into the zTx[ZC_IN] pipe pair; and
// 2. Reads the Control channel, and places the output into the zControl[ZC_IN] pipe pair

func (cp *ChannelPair) runChannels() {
	defer func() {
		cp.zTx[ZC_IN].Close()
		cp.zControl[ZC_IN].Close()
		cp.wg.Done()
	}()
	for {
		select {
		case msg, ok := <-cp.txChan:
			if !ok {
				// it's closed - this should never happen
				cp.errorChan <- errors.New("ZMQ tx channel unexpectedly closed")
				// it's closed - this should not ever happen
				return
			} else {
				if _, err := cp.zTx[ZC_IN].SendMessage(msg); err != nil {
					cp.errorChan <- err
					return
				}
			}
		case control, ok := <-cp.controlChan:
			if !ok {
				cp.errorChan <- errors.New("ZMQ control channel unexpectedly closed")
				// it's closed - this should not ever happen
				return
			} else {
				// If it's come externally, send a control message; ignore errors
				if control {
					cp.zControl[ZC_IN].SendMessage("")
				}
				return
			}
		}
	}
}

// func (cp *ChannelPair) runSockets()
//
// This function
// 1. Reads the main socket, and places the output into the rx channel
// 2. Reads the zTx[ZC_OUT] pipe pair and ...
// 3. Puts the output into the main socket
// 4. Reads the zControl[ZC_OUT] pipe pair

func (cp *ChannelPair) runSockets() {
	defer func() {
		cp.zTx[ZC_OUT].Close()
		cp.zControl[ZC_OUT].Close()
		cp.zSock.Close()
		cp.wg.Done()
	}()
	var toXmit [][]byte = nil
	poller := NewPoller()
	idxSock := poller.Add(cp.zSock, 0)
	idxTxOut := poller.Add(cp.zTx[ZC_OUT], 0)
	idxControlOut := poller.Add(cp.zControl[ZC_OUT], POLLIN)

	for {
		var zSockflags State = 0
		if len(cp.rxChan) < cap(cp.rxChan) || cap(cp.rxChan) == 0 {
			zSockflags |= POLLIN
		}
		var txsockflags State = 0
		// only if we have something to transmit are we interested in polling for output availability
		// else we just poll the input socket
		if toXmit == nil {
			txsockflags |= POLLIN
		} else {
			zSockflags |= POLLOUT
		}
		poller.Update(idxSock, zSockflags)
		poller.Update(idxTxOut, txsockflags)
		if sockets, err := poller.PollAll(-1); err != nil {
			cp.errorChan <- err
			cp.controlChan <- false
			return
		} else {
			if sockets[idxSock].Events&POLLIN != 0 {
				// we have received something on the main socket
				// we need to send it to the RX channel
				if parts, err := cp.zSock.RecvMessageBytes(0); err != nil {
					cp.errorChan <- err
					cp.controlChan <- false
					return
				} else {
					cp.rxChan <- parts
				}
			}
			if sockets[idxSock].Events&POLLOUT != 0 && toXmit != nil {
				// we are ready to send something on the main socket
				if _, err := cp.zSock.SendMessage(toXmit); err != nil {
					cp.errorChan <- err
					cp.controlChan <- false
					return
				} else {
					toXmit = nil
				}
			}
			if sockets[idxTxOut].Events&POLLIN != 0 && toXmit == nil {
				// we have something on the input socket, put it in xmit
				var err error
				toXmit, err = cp.zTx[ZC_OUT].RecvMessageBytes(0)
				if err != nil {
					cp.errorChan <- err
					cp.controlChan <- false
					return
				}
			}
			if sockets[idxControlOut].Events&POLLIN != 0 {
				// Something has arrived on the control channel
				// ignore errors
				_, _ = cp.zControl[ZC_OUT].RecvMessageBytes(0)
				// No need to signal the other end as we know it is already exiting
				// what we need to do is ensure any transmitted stuff is sent.

				// This is more tricky than you might think. The data could be
				// in ToXmit, in the TX socket pair, or in the TX channel.

				// block in these cases for as long as the linger value
				// FIXME: Ideally we'd block in TOTAL for the linger time,
				// rather than on each send for the linger time.
				if linger, err := cp.zSock.GetLinger(); err == nil {
					cp.zSock.SetSndtimeo(linger)
				}
				if toXmit != nil {
					if _, err := cp.zSock.SendMessage(toXmit); err != nil {
						cp.errorChan <- err
						return
					}
				} else {
					toXmit = nil
				}

				poller.Update(idxControlOut, 0)
				poller.Update(idxSock, 0)
				poller.Update(idxTxOut, POLLIN)
				for {
					if sockets, err := poller.PollAll(0); err != nil {
						cp.errorChan <- err
						return
					} else if sockets[idxTxOut].Events&POLLIN != 0 && toXmit == nil {
						// we have something on the input socket, put it in xmit
						var err error
						toXmit, err = cp.zTx[ZC_OUT].RecvMessageBytes(0)
						if err != nil {
							cp.errorChan <- err
							return
						}
						if _, err := cp.zSock.SendMessage(toXmit); err != nil {
							cp.errorChan <- err
							return
						}
					} else {
						break
					}
				}

				// Now read the TX channel until it is empty
				done := false
				for !done {
					select {
					case msg, ok := <-cp.txChan:
						if ok {
							if _, err := cp.zSock.SendMessage(msg); err != nil {
								cp.errorChan <- err
								return
							}
						} else {
							cp.errorChan <- errors.New("ZMQ tx channel unexpectedly closed")
							return
						}
					default:
						done = true
					}
				}
				return
			}
		}
	}
}

// func (cp *ChannelPair) Close() error
//
// Close a ChannelPair. This will kill the internal goroutines, and close
// the main ZMQ Socket. It will also close the error channel, so a select() on
// it will return 'ok' as false. If an error is produced either during the close
// or has been produced prior to the close, it will be returned.

func (cp *ChannelPair) Close() error {
	cp.controlChan <- true
	cp.wg.Wait()
	var err error = nil
	select {
	case err = <-cp.errorChan:
	default:
	}

	close(cp.txChan)
	close(cp.rxChan)
	close(cp.errorChan)
	close(cp.controlChan)
	return err
}

// func (cp *ChannelPair) TxChan() chan<- [][]byte
//
// Get the TxChannel as a write only channel

func (cp *ChannelPair) TxChan() chan<- [][]byte {
	return cp.txChan
}

// func (cp *ChannelPair) RxChan() <-chan [][]byte
//
// Get the RxChannel as a read only channel

func (cp *ChannelPair) RxChan() <-chan [][]byte {
	return cp.rxChan
}

// func (cp *ChannelPair) Errors() <-chan error
//
// Get the Errors channel as a read only channel

func (cp *ChannelPair) Errors() <-chan error {
	return cp.errorChan
}

// func closePair(sockets []*Socket)
//
// Close a socket pair

func closePair(sockets []*Socket) {
	for i, cp := range sockets {
		if cp != nil {
			cp.Close()
			sockets[i] = nil
		}
	}
}

// func newPair(c *Context) (sockets []*Socket, err error)
//
// Create a new socket pair

func newPair(c *Context) (sockets []*Socket, err error) {
	sockets = make([]*Socket, 2)
	addr := fmt.Sprintf("inproc://_channelpair_internal-%d", getUniqueId())
	if sockets[ZC_IN], err = c.NewSocket(PAIR); err != nil {
		goto Error
	}
	if err = sockets[ZC_IN].Bind(addr); err != nil {
		goto Error
	}
	if sockets[ZC_OUT], err = c.NewSocket(PAIR); err != nil {
		goto Error
	}
	if err = sockets[ZC_OUT].Connect(addr); err != nil {
		goto Error
	}
	return

Error:
	closePair(sockets)
	return
}

// func NewChannelPair(zSock *Socket, txbuf int, rxbuf int) (*ChannelPair, error)
//
// Produce a new ChannelPair. Pass a zmq Socket, plus the buffering parameters for the channels.
//
// If this call succeeds (err == nil), then a ChannelPair is returned, and control of your Socket is passed
// irrevocably to this routine. You should forget you ever had the socket. Do not attempt to use it in any way,
// as its manipulation is now the responsibility of goroutines launched by this routine. Closing the ChannelPair
// will also close your Socket.
//
// If this routine errors, it is the caller's responsibility to close the Socket
//
// The buffering parameters control the maximum amount of buffered data, in and out. An extra message may
// be buffered under some circumstances for internal reasons.

func NewChannelPair(zSock *Socket, txbuf int, rxbuf int) (*ChannelPair, error) {
	cp := &ChannelPair{
		zSock: zSock,
	}

	zmqContext, err := zSock.Context()
	if err != nil {
		return nil, err
	}

	if cp.zControl, err = newPair(zmqContext); err != nil {
		return nil, err
	}

	if cp.zTx, err = newPair(zmqContext); err != nil {
		closePair(cp.zControl)
		return nil, err
	}

	// as we should never read or send to these sockets unless they are ready
	// we set the timeout to 0 so a write or read in any other circumstance
	// returns a immediate error
	cp.zSock.SetRcvtimeo(0)
	cp.zSock.SetSndtimeo(0)
	for i := ZC_IN; i <= ZC_OUT; i++ {
		cp.zTx[i].SetRcvtimeo(0)
		cp.zTx[i].SetSndtimeo(0)
		cp.zControl[i].SetRcvtimeo(0)
		cp.zControl[i].SetSndtimeo(0)
	}

	cp.txChan = make(chan [][]byte, txbuf)
	cp.rxChan = make(chan [][]byte, rxbuf)
	cp.errorChan = make(chan error, 2)
	cp.controlChan = make(chan bool, 2)

	barrier()
	cp.wg.Add(2)
	go cp.runSockets()
	go cp.runChannels()
	return cp, nil
}
