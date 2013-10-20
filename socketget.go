package zmq4

/*
#include <zmq.h>
#include <stdint.h>
*/
import "C"

import (
	"strings"
	"time"
	"unsafe"
)

func (soc *Socket) getString(opt C.int, bufsize int) (string, error) {
	value := make([]byte, bufsize)
	size := C.size_t(bufsize)
	if i, err := C.zmq_getsockopt(soc.soc, opt, unsafe.Pointer(&value[0]), &size); i != 0 {
		return "", errget(err)
	}
	return strings.TrimRight(string(value[:int(size)]), "\x00"), nil
}

func (soc *Socket) getStringRaw(opt C.int, bufsize int) (string, error) {
	value := make([]byte, bufsize)
	size := C.size_t(bufsize)
	if i, err := C.zmq_getsockopt(soc.soc, opt, unsafe.Pointer(&value[0]), &size); i != 0 {
		return "", errget(err)
	}
	return string(value[:int(size)]), nil
}

func (soc *Socket) getInt(opt C.int) (int, error) {
	value := C.int(0)
	size := C.size_t(unsafe.Sizeof(value))
	if i, err := C.zmq_getsockopt(soc.soc, opt, unsafe.Pointer(&value), &size); i != 0 {
		return 0, errget(err)
	}
	return int(value), nil
}

func (soc *Socket) getInt64(opt C.int) (int64, error) {
	value := C.int64_t(0)
	size := C.size_t(unsafe.Sizeof(value))
	if i, err := C.zmq_getsockopt(soc.soc, opt, unsafe.Pointer(&value), &size); i != 0 {
		return 0, errget(err)
	}
	return int64(value), nil
}

func (soc *Socket) getUInt64(opt C.int) (uint64, error) {
	value := C.uint64_t(0)
	size := C.size_t(unsafe.Sizeof(value))
	if i, err := C.zmq_getsockopt(soc.soc, opt, unsafe.Pointer(&value), &size); i != 0 {
		return 0, errget(err)
	}
	return uint64(value), nil
}

// ZMQ_TYPE: Retrieve socket type
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc3
func (soc *Socket) GetType() (Type, error) {
	v, err := soc.getInt(C.ZMQ_TYPE)
	return Type(v), err
}

// ZMQ_RCVMORE: More message data parts to follow
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc4
func (soc *Socket) GetRcvmore() (bool, error) {
	v, err := soc.getInt(C.ZMQ_RCVMORE)
	return v != 0, err
}

// ZMQ_SNDHWM: Retrieves high water mark for outbound messages
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc5
func (soc *Socket) GetSndhwm() (int, error) {
	return soc.getInt(C.ZMQ_SNDHWM)
}

// ZMQ_RCVHWM: Retrieve high water mark for inbound messages
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc6
func (soc *Socket) GetRcvhwm() (int, error) {
	return soc.getInt(C.ZMQ_RCVHWM)
}

// ZMQ_AFFINITY: Retrieve I/O thread affinity
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc7
func (soc *Socket) GetAffinity() (uint64, error) {
	return soc.getUInt64(C.ZMQ_AFFINITY)
}

// ZMQ_IDENTITY: Retrieve socket identity
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc8
func (soc *Socket) GetIdentity() (string, error) {
	return soc.getString(C.ZMQ_IDENTITY, 256)
}

// ZMQ_RATE: Retrieve multicast data rate
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc9
func (soc *Socket) GetRate() (int, error) {
	return soc.getInt(C.ZMQ_RATE)
}

// ZMQ_RECOVERY_IVL: Get multicast recovery interval
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc10
func (soc *Socket) GetRecoveryIvl() (time.Duration, error) {
	v, err := soc.getInt(C.ZMQ_RECOVERY_IVL)
	return time.Duration(v) * time.Millisecond, err
}

// ZMQ_SNDBUF: Retrieve kernel transmit buffer size
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc11
func (soc *Socket) GetSndbuf() (int, error) {
	return soc.getInt(C.ZMQ_SNDBUF)
}

// ZMQ_RCVBUF: Retrieve kernel receive buffer size
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc12
func (soc *Socket) GetRcvbuf() (int, error) {
	return soc.getInt(C.ZMQ_RCVBUF)
}

// ZMQ_LINGER: Retrieve linger period for socket shutdown
//
// Returns time.Duration(-1) for infinite
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc13
func (soc *Socket) GetLinger() (time.Duration, error) {
	v, err := soc.getInt(C.ZMQ_LINGER)
	if v < 0 {
		return time.Duration(-1), err
	}
	return time.Duration(v) * time.Millisecond, err
}

// ZMQ_RECONNECT_IVL: Retrieve reconnection interval
//
// Returns time.Duration(-1) for no reconnection
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc14
func (soc *Socket) GetReconnectIvl() (time.Duration, error) {
	v, err := soc.getInt(C.ZMQ_RECONNECT_IVL)
	if v < 0 {
		return time.Duration(-1), err
	}
	return time.Duration(v) * time.Millisecond, err
}

// ZMQ_RECONNECT_IVL_MAX: Retrieve maximum reconnection interval
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc15
func (soc *Socket) GetReconnectIvlMax() (time.Duration, error) {
	v, err := soc.getInt(C.ZMQ_RECONNECT_IVL_MAX)
	return time.Duration(v) * time.Millisecond, err
}

// ZMQ_BACKLOG: Retrieve maximum length of the queue of outstanding connections
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc16
func (soc *Socket) GetBacklog() (int, error) {
	return soc.getInt(C.ZMQ_BACKLOG)
}

// ZMQ_MAXMSGSIZE: Maximum acceptable inbound message size
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc17
func (soc *Socket) GetMaxmsgsize() (int64, error) {
	return soc.getInt64(C.ZMQ_MAXMSGSIZE)
}

// ZMQ_MULTICAST_HOPS: Maximum network hops for multicast packets
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc18
func (soc *Socket) GetMulticastHops() (int, error) {
	return soc.getInt(C.ZMQ_MULTICAST_HOPS)
}

// ZMQ_RCVTIMEO: Maximum time before a socket operation returns with EAGAIN
//
// Returns time.Duration(-1) for infinite
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc19
func (soc *Socket) GetRcvtimeo() (time.Duration, error) {
	v, err := soc.getInt(C.ZMQ_RCVTIMEO)
	if v < 0 {
		return time.Duration(-1), err
	}
	return time.Duration(v) * time.Millisecond, err
}

// ZMQ_SNDTIMEO: Maximum time before a socket operation returns with EAGAIN
//
// Returns time.Duration(-1) for infinite
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc20
func (soc *Socket) GetSndtimeo() (time.Duration, error) {
	v, err := soc.getInt(C.ZMQ_SNDTIMEO)
	if v < 0 {
		return time.Duration(-1), err
	}
	return time.Duration(v) * time.Millisecond, err
}

// ZMQ_IPV6: Retrieve IPv6 socket status
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc21
func (soc *Socket) GetIpv6() (bool, error) {
	v, err := soc.getInt(C.ZMQ_IPV6)
	return v != 0, err
}

// ZMQ_IMMEDIATE: Retrieve attach-on-connect value
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc23
func (soc *Socket) GetImmediate() (bool, error) {
	v, err := soc.getInt(C.ZMQ_IMMEDIATE)
	return v != 0, err
}

// ZMQ_FD: Retrieve file descriptor associated with the socket
// see socketget_unix.go and socketget_windows.go

// ZMQ_EVENTS: Retrieve socket event state
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc25
func (soc *Socket) GetEvents() (State, error) {
	v, err :=  soc.getInt(C.ZMQ_EVENTS)
	return State(v), err
}

// ZMQ_LAST_ENDPOINT: Retrieve the last endpoint set
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc26
func (soc *Socket) GetLastEndpoint() (string, error) {
	return soc.getString(C.ZMQ_LAST_ENDPOINT, 1024)
}

// ZMQ_TCP_KEEPALIVE: Override SO_KEEPALIVE socket option
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc27
func (soc *Socket) GetTcpKeepalive() (int, error) {
	return soc.getInt(C.ZMQ_TCP_KEEPALIVE)
}

// ZMQ_TCP_KEEPALIVE_IDLE: Override TCP_KEEPCNT(or TCP_KEEPALIVE on some OS)
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc28
func (soc *Socket) GetTcpKeepaliveIdle() (int, error) {
	return soc.getInt(C.ZMQ_TCP_KEEPALIVE_IDLE)
}

// ZMQ_TCP_KEEPALIVE_CNT: Override TCP_KEEPCNT socket option
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc29
func (soc *Socket) GetTcpKeepaliveCnt() (int, error) {
	return soc.getInt(C.ZMQ_TCP_KEEPALIVE_CNT)
}

// ZMQ_TCP_KEEPALIVE_INTVL: Override TCP_KEEPINTVL socket option
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc30
func (soc *Socket) GetTcpKeepaliveIntvl() (int, error) {
	return soc.getInt(C.ZMQ_TCP_KEEPALIVE_INTVL)
}

// ZMQ_MECHANISM: Retrieve current security mechanism
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc31
func (soc *Socket) GetMechanism() (Mechanism, error) {
	v, err := soc.getInt(C.ZMQ_MECHANISM)
	return Mechanism(v), err
}

// ZMQ_PLAIN_SERVER: Retrieve current PLAIN server role
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc32
func (soc *Socket) GetPlainServer() (int, error) {
	return soc.getInt(C.ZMQ_PLAIN_SERVER)
}

// ZMQ_PLAIN_USERNAME: Retrieve current PLAIN username
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc33
func (soc *Socket) GetPlainUsername() (string, error) {
	s, err := soc.getString(C.ZMQ_PLAIN_USERNAME, 1024)
	if n := len(s); n > 0 && s[n - 1] == 0 {
		s = s[:n - 1]
	}
	return s, err
}

// ZMQ_PLAIN_PASSWORD: Retrieve current password
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc34
func (soc *Socket) GetPlainPassword() (string, error) {
	s, err := soc.getString(C.ZMQ_PLAIN_PASSWORD, 1024)
	if n := len(s); n > 0 && s[n - 1] == 0 {
		s = s[:n - 1]
	}
	return s, err
}

// ZMQ_CURVE_PUBLICKEY: Retrieve current CURVE public key
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc35
func (soc *Socket) GetCurvePublickeyRaw() (string, error) {
	return soc.getStringRaw(C.ZMQ_CURVE_PUBLICKEY, 32)
}

// ZMQ_CURVE_PUBLICKEY: Retrieve current CURVE public key
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc35
func (soc *Socket) GetCurvePublickeykeyZ85() (string, error) {
	return soc.getString(C.ZMQ_CURVE_PUBLICKEY, 41)
}

// ZMQ_CURVE_SECRETKEY: Retrieve current CURVE secret key
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc36
func (soc *Socket) GetCurveSecretkeyRaw() (string, error) {
	return soc.getStringRaw(C.ZMQ_CURVE_SECRETKEY, 32)
}

// ZMQ_CURVE_SECRETKEY: Retrieve current CURVE secret key
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc36
func (soc *Socket) GetCurveSecretkeyZ85() (string, error) {
	return soc.getString(C.ZMQ_CURVE_SECRETKEY, 41)
}

// ZMQ_CURVE_SERVERKEY: Retrieve current CURVE server key
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc37
func (soc *Socket) GetCurveServerkeyRaw() (string, error) {
	return soc.getStringRaw(C.ZMQ_CURVE_SERVERKEY, 32)
}

// ZMQ_CURVE_SERVERKEY: Retrieve current CURVE server key
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc37
func (soc *Socket) GetCurveServerkeyZ85() (string, error) {
	return soc.getString(C.ZMQ_CURVE_SERVERKEY, 41)
}

// ZMQ_ZAP_DOMAIN: Retrieve RFC 27 authentication domain
//
// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc38
func (soc *Socket) GetZapDomain() (string, error) {
	return soc.getString(C.ZMQ_ZAP_DOMAIN, 1024)
}
