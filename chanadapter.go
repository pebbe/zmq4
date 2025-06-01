package zmq4

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
)

// ChanAdapter provides a Go channel-based interface for ZeroMQ sockets.
// It bridges the gap between ZeroMQ's message-passing model and Go's channel-based
// concurrency model by creating receive (Rx) and transmit (Tx) channels that can be
// used in standard Go select statements and channel operations.
//
// The adapter runs two internal goroutines: one for routing incoming messages
// from the ZMQ socket to the Rx channel, and another for sending messages
// from the Tx channel to the ZMQ socket.
type ChanAdapter struct {
	socket         *Socket
	pairAddr       string
	rxChan         chan [][]byte
	txChan         chan [][]byte
	sendBufferChan chan [][]byte
	closeOnce      sync.Once
	closeChan      func()
	ctxCancel      func()
	wg             sync.WaitGroup
}

var (
	chanAdapterUniqueID atomic.Uint64
	chanAdapterOpClose  = []byte{0x0} // the adapter is being closed
	chanAdapterOpSend   = []byte{0x1} // send a message to the ZMQ socket
)

// NewChanAdapter creates a new ChanAdapter for the given ZMQ socket.
// The rxChanSize parameter specifies the buffer size for the receive channel,
// and txChanSize specifies the buffer size for the transmit channel.
//
// The adapter must be started with Start() and should be closed with Close()
// when no longer needed to ensure proper cleanup of resources.
//
// Example:
//
//	socket, err := zmq4.NewSocket(zmq4.REQ)
//	if err != nil {
//		log.Fatal(err)
//	}
//	socket.Connect("<socket-address>")
//
//	adapter := zmq4.NewChanAdapter(socket, 100, 100)
//	defer adapter.Close()
//
//	// Start the adapter with a context for cancellation
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	adapter.Start(ctx)
//
//	msg := [][]byte{[]byte("Hello"), []byte("World")}
//	adapter.TxChan() <- msg
//
//	rxChan := adapter.RxChan()
//
//	select {
//	case msg, ok := <-rxChan:
//		// Handle incoming message
//	case <-time.After(time.Second):
//		// Handle timeout
//	}
func NewChanAdapter(socket *Socket, rxChanSize int, txChanSize int) *ChanAdapter {
	pairAddr := fmt.Sprintf("inproc://zmq-chan-transceiver-%d", chanAdapterUniqueID.Add(1))
	rxChan := make(chan [][]byte, rxChanSize)
	txChan := make(chan [][]byte, txChanSize)
	sendBufferChan := make(chan [][]byte) // unbuffered channel

	closeChan := func() {
		close(rxChan)
		close(txChan)
		close(sendBufferChan)
	}

	return &ChanAdapter{
		socket:         socket,
		pairAddr:       pairAddr,
		rxChan:         rxChan,
		txChan:         txChan,
		sendBufferChan: sendBufferChan,
		closeChan:      closeChan,
	}
}

// RxChan returns a receive-only channel for incoming messages from the ZMQ socket.
// Messages received on the underlying ZMQ socket will be forwarded to this channel
// as byte slice arrays, where each element represents a message part in multi-part messages.
//
// The channel will be closed when the adapter is shut down.
func (t *ChanAdapter) RxChan() <-chan [][]byte {
	return t.rxChan
}

// TxChan returns a send-only channel for outgoing messages to the ZMQ socket.
// Messages sent to this channel will be forwarded to the underlying ZMQ socket.
// Each message should be provided as a byte slice array, where each element
// represents a message part for multi-part messages.
//
// The channel will be closed when the adapter is shut down.
func (t *ChanAdapter) TxChan() chan<- [][]byte {
	return t.txChan
}

// Start begins the adapter's operation by launching two background goroutines:
// one for routing messages from the ZMQ socket to the receive channel, and
// another for routing messages from the transmit channel to the ZMQ socket.
//
// The provided context can be used to cancel the adapter's operation.
// When the context is cancelled or Close() is called, both goroutines will
// be shut down gracefully.
//
// Start should only be called once per adapter instance.
func (t *ChanAdapter) Start(ctx context.Context) {
	childCtx, cancel := context.WithCancel(ctx)
	t.ctxCancel = cancel

	t.wg.Add(2)
	go t.runRouterLoop(childCtx)
	go t.runSenderLoop(childCtx)
}

// runRouterLoop handles incoming messages from the ZMQ socket and internal
// coordination messages from the sender loop. It forwards socket messages
// to the receive channel and processes send requests from the sender loop.
func (t *ChanAdapter) runRouterLoop(_ context.Context) {
	var err error
	defer t.wg.Done()
	defer t.ctxCancel()

	pair, err := NewSocket(PAIR)
	if err != nil {
		log.Println("E: failed to create PAIR socket")
		return
	}
	defer pair.Close()

	err = pair.Bind(t.pairAddr)
	if err != nil {
		log.Println("E: failed to connect PAIR socket")
		return
	}

	poller := NewPoller()
	poller.Add(t.socket, POLLIN)
	poller.Add(pair, POLLIN)

	for {
		events, err := poller.PollAll(-1)
		if err != nil {
			log.Println(err)
			return
		}

		if len(events) == 0 {
			log.Println("I: socket timeout")
			continue
		}

		for _, item := range events {
			switch s := item.Socket; s {
			case pair:
				if item.Events&POLLIN != 0 {
					opCode, err := pair.RecvBytes(0)
					if err != nil {
						log.Println(err)
						continue
					}

					if bytes.Equal(opCode, chanAdapterOpClose) {
						return
					} else if bytes.Equal(opCode, chanAdapterOpSend) {
						msg, ok := <-t.sendBufferChan
						if !ok {
							log.Println("E: sendBufferChan is closed")
							return
						}
						// send the message payload to the ZMQ socket
						_, err = t.socket.SendMessage(msg)
						if err != nil {
							log.Println("E: failed to send message to socket. ", err)
							continue
						}
					}
				}
			case t.socket:
				if item.Events&POLLIN != 0 {
					// process incoming messages
					msg, err := t.socket.RecvMessageBytes(0)
					if err != nil {
						log.Println(err)
						continue
					}
					// forward the message to rxChan
					t.rxChan <- msg
				}
			}

		}
	}

}

// runSenderLoop handles outgoing messages from the transmit channel.
// It coordinates with the router loop via an internal PAIR socket to
// ensure thread-safe message transmission to the ZMQ socket.
func (t *ChanAdapter) runSenderLoop(ctx context.Context) {
	var err error
	defer t.wg.Done()
	defer t.ctxCancel()

	pair, err := NewSocket(PAIR)
	if err != nil {
		log.Println("E: failed to create PAIR socket")
		return
	}
	defer pair.Close()

	err = pair.Connect(t.pairAddr)
	if err != nil {
		log.Println("E: failed to connect PAIR socket")
		return
	}

	for {
		select {
		case <-ctx.Done():
			// ensure the router loop is closed gracefully
			_, err = pair.SendBytes(chanAdapterOpClose, 0)
			if err != nil {
				log.Println("E: failed to send close message")
			}
			return
		case msg, ok := <-t.txChan:
			if !ok {
				log.Println("E: txChan is closed")
				return
			}
			_, err = pair.SendBytes(chanAdapterOpSend, 0)
			if err != nil {
				log.Println(err)
				continue
			}
			// block until the message payload is recieved by Router loop
			t.sendBufferChan <- msg
		}
	}
}

// Close gracefully shuts down the adapter by cancelling the context,
// waiting for all goroutines to complete, and closing all channels.
// It's safe to call Close multiple times.
//
// After Close is called, the receive and transmit channels will be closed
// and no further message processing will occur.
func (t *ChanAdapter) Close() {
	if t.ctxCancel != nil {
		t.ctxCancel()
	}
	t.wg.Wait()
	if t.closeChan != nil {
		t.closeOnce.Do(t.closeChan)
	}
}
