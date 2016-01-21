package zmq4

import (
	"fmt"
)

/*
Send multi-part message on socket.

Any `[]string' or `[][]byte' is split into separate `string's or `[]byte's

Any other part that isn't a `string' or `[]byte' is converted
to `string' with `fmt.Sprintf("%v", part)'.

Returns total bytes sent.
*/
func (soc *Socket) SendMessage(parts ...interface{}) (total int, err error) {
	return soc.sendMessage(0, parts...)
}

/*
Like SendMessage(), but adding the DONTWAIT flag.
*/
func (soc *Socket) SendMessageDontwait(parts ...interface{}) (total int, err error) {
	return soc.sendMessage(DONTWAIT, parts...)
}

func (soc *Socket) sendMessage(dontwait Flag, parts ...interface{}) (total int, err error) {
	// TODO: make this faster
	// Done, ;)
	var partial int
	last0 := len(parts) - 1
	opt := SNDMORE | dontwait
	for i0, p0 := range parts {
		switch t0 := p0.(type) {
		case []string:
			last1 := len(t0) - 1
			if last1 < 0 && i0 == last0 {
				// A bug in the program?, kept for compatibility of the
				// previous version.
				// The program has sent an empty slice as last
				// argument. Force to send the message with no SNDMORE.
				// I'm (gallir) not sure if must be sent also a zero sized byte
				// array for intermediate empty slices
				partial, err = soc.sendSinglePart(dontwait, []byte{})
				total += partial
			} else {
				for i1, p1 := range t0 {
					if i0 == last0 && i1 == last1 {
						opt = dontwait
					}
					partial, err = soc.sendSinglePart(opt, p1)
					if err != nil {
						break // Don't continue
					}
					total += partial
				}
			}
			if err != nil {
				return -1, err
			}

		case [][]byte:
			last1 := len(t0) - 1
			if last1 < 0 && i0 == last0 {
				// A bug in the program, see above comment..
				partial, err = soc.sendSinglePart(dontwait, []byte{})
				total += partial
			} else {
				for i1, p1 := range t0 {
					if i0 >= last0 && i1 >= last1 {
						opt = dontwait
					}
					partial, err = soc.sendSinglePart(opt, p1)
					if err != nil {
						break // Don't continue
					}
					total += partial
				}
			}
			if err != nil {
				return -1, err
			}
		default:
			if i0 == last0 {
				opt = dontwait
			}
			switch t := p0.(type) {
			case string:
				partial, err = soc.sendSinglePart(opt, t)
			case []byte:
				partial, err = soc.sendSinglePart(opt, t)
			default:
				partial, err = soc.sendSinglePart(opt, fmt.Sprintf("%v", t))
			}
			if err != nil {
				return -1, err
			}
			total += partial
		}
	}
	return
}

func (soc *Socket) sendSinglePart(opt Flag, part interface{}) (total int, err error) {
	switch t := part.(type) {
	case string:
		total, err = soc.Send(t, opt)
	case []byte:
		total, err = soc.SendBytes(t, opt)
	default:
		total, err = soc.Send(fmt.Sprintf("%v", t), opt)
	}
	if err != nil {
		return -1, err
	}
	return
}

/*
Receive parts as message from socket.

Returns last non-nil error code.
*/
func (soc *Socket) RecvMessage(flags Flag) (msg []string, err error) {
	msg = make([]string, 0)
	for {
		s, e := soc.Recv(flags)
		if e == nil {
			msg = append(msg, s)
		} else {
			return msg[0:0], e
		}
		more, e := soc.GetRcvmore()
		if e == nil {
			if !more {
				break
			}
		} else {
			return msg[0:0], e
		}
	}
	return
}

/*
Receive parts as message from socket.

Returns last non-nil error code.
*/
func (soc *Socket) RecvMessageBytes(flags Flag) (msg [][]byte, err error) {
	msg = make([][]byte, 0)
	for {
		b, e := soc.RecvBytes(flags)
		if e == nil {
			msg = append(msg, b)
		} else {
			return msg[0:0], e
		}
		more, e := soc.GetRcvmore()
		if e == nil {
			if !more {
				break
			}
		} else {
			return msg[0:0], e
		}
	}
	return
}

/*
Receive parts as message from socket, including metadata.

Metadata is picked from the first message part.

For details about metadata, see RecvWithMetadata().

Returns last non-nil error code.
*/
func (soc *Socket) RecvMessageWithMetadata(flags Flag, properties ...string) (msg []string, metadata map[string]string, err error) {
	b, p, err := soc.RecvMessageBytesWithMetadata(flags, properties...)
	m := make([]string, len(b))
	for i, bt := range b {
		m[i] = string(bt)
	}
	return m, p, err
}

/*
Receive parts as message from socket, including metadata.

Metadata is picked from the first message part.

For details about metadata, see RecvBytesWithMetadata().

Returns last non-nil error code.
*/
func (soc *Socket) RecvMessageBytesWithMetadata(flags Flag, properties ...string) (msg [][]byte, metadata map[string]string, err error) {
	bb := make([][]byte, 0)
	b, p, err := soc.RecvBytesWithMetadata(flags, properties...)
	if err != nil {
		return bb, p, err
	}
	for {
		bb = append(bb, b)

		var more bool
		more, err = soc.GetRcvmore()
		if err != nil || !more {
			break
		}
		b, err = soc.RecvBytes(flags)
		if err != nil {
			break
		}
	}
	return bb, p, err
}
