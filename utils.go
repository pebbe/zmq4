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

	// flatten first, just in case the last part may be an empty []string or [][]byte
	pp := make([]interface{}, 0)
	for _, p := range parts {
		switch t := p.(type) {
		case []string:
			for _, s := range t {
				pp = append(pp, s)
			}
		case [][]byte:
			for _, b := range t {
				pp = append(pp, b)
			}
		default:
			pp = append(pp, t)
		}
	}

	n := len(pp)
	if n == 0 {
		return
	}
	opt := SNDMORE | dontwait
	for i, p := range pp {
		if i == n-1 {
			opt = dontwait
		}
		switch t := p.(type) {
		case string:
			j, e := soc.Send(t, opt)
			if e == nil {
				total += j
			} else {
				return -1, e
			}
		case []byte:
			j, e := soc.SendBytes(t, opt)
			if e == nil {
				total += j
			} else {
				return -1, e
			}
		default:
			j, e := soc.Send(fmt.Sprintf("%v", t), opt)
			if e == nil {
				total += j
			} else {
				return -1, e
			}
		}
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
