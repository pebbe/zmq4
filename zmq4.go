// A Go interface to ZeroMQ (zmq, 0mq) version 4.
//
// For ZeroMQ version 3, see: http://github.com/pebbe/zmq3
//
// For ZeroMQ version 2, see: http://github.com/pebbe/zmq2
//
// http://www.zeromq.org/
package zmq4

/*
#cgo !windows pkg-config: libzmq
#cgo windows CFLAGS: -I/usr/local/include
#cgo windows LDFLAGS: -L/usr/local/lib -lzmq
#include <zmq.h>
#include <zmq_utils.h>
#include <stdlib.h>
#include <string.h>

void *my_memcpy(void *dest, const void *src, size_t n) {
	return memcpy(dest, src, n);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

var (
	ctx unsafe.Pointer
)

func init() {
	var err error
	ctx, err = C.zmq_ctx_new()
	if ctx == nil {
		panic("Init of ZeroMQ context failed: " + errget(err).Error())
	}
}

//. Util

func errget(err error) error {
	errno, ok := err.(syscall.Errno)
	if ok && errno >= C.ZMQ_HAUSNUMERO {
		return errors.New(C.GoString(C.zmq_strerror(C.int(errno))))
	}
	return err
}

// Report 0MQ library version.
func Version() (major, minor, patch int) {
	var maj, min, pat C.int
	C.zmq_version(&maj, &min, &pat)
	return int(maj), int(min), int(pat)
}

// Get 0MQ error message string.
func Error(e int) string {
	return C.GoString(C.zmq_strerror(C.int(e)))
}

//. Context

const (
	MaxSocketsDflt = int(C.ZMQ_MAX_SOCKETS_DFLT)
	IoThreadsDflt  = int(C.ZMQ_IO_THREADS_DFLT)
)

func getOption(o C.int) (int, error) {
	nc, err := C.zmq_ctx_get(ctx, o)
	n := int(nc)
	if n < 0 {
		return n, errget(err)
	}
	return n, nil
}

// Returns the size of the 0MQ thread pool.
func GetIoThreads() (int, error) {
	return getOption(C.ZMQ_IO_THREADS)
}

// Returns the maximum number of sockets allowed.
func GetMaxSockets() (int, error) {
	return getOption(C.ZMQ_MAX_SOCKETS)
}

// Returns the IPv6 option
func GetIpv6() (bool, error) {
	i, e := getOption(C.ZMQ_IPV6)
	if i == 0 {
		return false, e
	}
	return true, e
}

func setOption(o C.int, n int) error {
	i, err := C.zmq_ctx_set(ctx, o, C.int(n))
	if int(i) != 0 {
		return errget(err)
	}
	return nil
}

/*
Specifies the size of the 0MQ thread pool to handle I/O operations. If
your application is using only the inproc transport for messaging you
may set this to zero, otherwise set it to at least one. This option only
applies before creating any sockets.

Default value   1
*/
func SetIoThreads(n int) error {
	return setOption(C.ZMQ_IO_THREADS, n)
}

/*
Sets the maximum number of sockets allowed.

Default value   1024
*/
func SetMaxSockets(n int) error {
	return setOption(C.ZMQ_MAX_SOCKETS, n)
}

/*
Sets the IPv6 value for all sockets created in the context from this point onwards.
A value of true means IPv6 is enabled, while false means the socket will use only IPv4.
When IPv6 is enabled, a socket will connect to, or accept connections from, both IPv4 and IPv6 hosts.

Default value	false
*/
func SetIpv6(i bool) error {
	n := 0
	if i {
		n = 1
	}
	return setOption(C.ZMQ_IPV6, n)
}

/*
Terminates the context.

For linger behavior, see: http://api.zeromq.org/4-0:zmq-ctx-term
*/
func Term() error {
	n, err := C.zmq_ctx_term(ctx)
	if n != 0 {
		return errget(err)
	}
	return nil
}

//. Sockets

// Specifies the type of a socket, used by NewSocket()
type Type int

const (
	// Constants for NewSocket()
	// See: http://api.zeromq.org/4-0:zmq-socket#toc3
	REQ    = Type(C.ZMQ_REQ)
	REP    = Type(C.ZMQ_REP)
	DEALER = Type(C.ZMQ_DEALER)
	ROUTER = Type(C.ZMQ_ROUTER)
	PUB    = Type(C.ZMQ_PUB)
	SUB    = Type(C.ZMQ_SUB)
	XPUB   = Type(C.ZMQ_XPUB)
	XSUB   = Type(C.ZMQ_XSUB)
	PUSH   = Type(C.ZMQ_PUSH)
	PULL   = Type(C.ZMQ_PULL)
	PAIR   = Type(C.ZMQ_PAIR)
	STREAM = Type(C.ZMQ_STREAM)
)

/*
Socket type as string.
*/
func (t Type) String() string {
	switch t {
	case REQ:
		return "REQ"
	case REP:
		return "REP"
	case DEALER:
		return "DEALER"
	case ROUTER:
		return "ROUTER"
	case PUB:
		return "PUB"
	case SUB:
		return "SUB"
	case XPUB:
		return "XPUB"
	case XSUB:
		return "XSUB"
	case PUSH:
		return "PUSH"
	case PULL:
		return "PULL"
	case PAIR:
		return "PAIR"
	case STREAM:
		return "STREAM"
	}
	return "<INVALID>"
}

// Used by  (*Socket)Send() and (*Socket)Recv()
type Flag int

const (
	// Flags for (*Socket)Send(), (*Socket)Recv()
	// For Send, see: http://api.zeromq.org/4-0:zmq-send#toc2
	// For Recv, see: http://api.zeromq.org/4-0:zmq-msg-recv#toc2
	DONTWAIT = Flag(C.ZMQ_DONTWAIT)
	SNDMORE  = Flag(C.ZMQ_SNDMORE)
)

/*
Socket flag as string.
*/
func (f Flag) String() string {
	ff := make([]string, 0)
	if f&DONTWAIT != 0 {
		ff = append(ff, "DONTWAIT")
	}
	if f&SNDMORE != 0 {
		ff = append(ff, "SNDMORE")
	}
	if len(ff) == 0 {
		return "<NONE>"
	}
	return strings.Join(ff, "|")
}

// Used by (*Socket)Monitor() and (*Socket)RecvEvent()
type Event int

const (
	// Flags for (*Socket)Monitor() and (*Socket)RecvEvent()
	// See: http://api.zeromq.org/4-0:zmq-socket-monitor#toc3
	EVENT_ALL             = Event(C.ZMQ_EVENT_ALL)
	EVENT_CONNECTED       = Event(C.ZMQ_EVENT_CONNECTED)
	EVENT_CONNECT_DELAYED = Event(C.ZMQ_EVENT_CONNECT_DELAYED)
	EVENT_CONNECT_RETRIED = Event(C.ZMQ_EVENT_CONNECT_RETRIED)
	EVENT_LISTENING       = Event(C.ZMQ_EVENT_LISTENING)
	EVENT_BIND_FAILED     = Event(C.ZMQ_EVENT_BIND_FAILED)
	EVENT_ACCEPTED        = Event(C.ZMQ_EVENT_ACCEPTED)
	EVENT_ACCEPT_FAILED   = Event(C.ZMQ_EVENT_ACCEPT_FAILED)
	EVENT_CLOSED          = Event(C.ZMQ_EVENT_CLOSED)
	EVENT_CLOSE_FAILED    = Event(C.ZMQ_EVENT_CLOSE_FAILED)
	EVENT_DISCONNECTED    = Event(C.ZMQ_EVENT_DISCONNECTED)
)

/*
Socket event as string.
*/
func (e Event) String() string {
	if e == EVENT_ALL {
		return "EVENT_ALL"
	}
	ee := make([]string, 0)
	if e&EVENT_CONNECTED != 0 {
		ee = append(ee, "EVENT_CONNECTED")
	}
	if e&EVENT_CONNECT_DELAYED != 0 {
		ee = append(ee, "EVENT_CONNECT_DELAYED")
	}
	if e&EVENT_CONNECT_RETRIED != 0 {
		ee = append(ee, "EVENT_CONNECT_RETRIED")
	}
	if e&EVENT_LISTENING != 0 {
		ee = append(ee, "EVENT_LISTENING")
	}
	if e&EVENT_BIND_FAILED != 0 {
		ee = append(ee, "EVENT_BIND_FAILED")
	}
	if e&EVENT_ACCEPTED != 0 {
		ee = append(ee, "EVENT_ACCEPTED")
	}
	if e&EVENT_ACCEPT_FAILED != 0 {
		ee = append(ee, "EVENT_ACCEPT_FAILED")
	}
	if e&EVENT_CLOSED != 0 {
		ee = append(ee, "EVENT_CLOSED")
	}
	if e&EVENT_CLOSE_FAILED != 0 {
		ee = append(ee, "EVENT_CLOSE_FAILED")
	}
	if e&EVENT_DISCONNECTED != 0 {
		ee = append(ee, "EVENT_DISCONNECTED")
	}
	if len(ee) == 0 {
		return "<NONE>"
	}
	return strings.Join(ee, "|")
}

// Used by (soc *Socket)GetEvents()
type State int

const (
	// Flags for (*Socket)GetEvents()
	// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc25
	POLLIN  = State(C.ZMQ_POLLIN)
	POLLOUT = State(C.ZMQ_POLLOUT)
)

/*
Socket state as string.
*/
func (s State) String() string {
	ss := make([]string, 0)
	if s&POLLIN != 0 {
		ss = append(ss, "POLLIN")
	}
	if s&POLLOUT != 0 {
		ss = append(ss, "POLLOUT")
	}
	if len(ss) == 0 {
		return "<NONE>"
	}
	return strings.Join(ss, "|")
}

// Specifies the security mechanism, used by (*Socket)GetMechanism()
type Mechanism int

const (
	// Constants for (*Socket)GetMechanism()
	// See: http://api.zeromq.org/4-0:zmq-getsockopt#toc31
	NULL  = Mechanism(C.ZMQ_NULL)
	PLAIN = Mechanism(C.ZMQ_PLAIN)
	CURVE = Mechanism(C.ZMQ_CURVE)
)

/*
Security mechanism as string.
*/
func (m Mechanism) String() string {
	switch m {
	case NULL:
		return "NULL"
	case PLAIN:
		return "PLAIN"
	case CURVE:
		return "CURVE"
	}
	return "<INVALID>"
}

/*
Socket functions starting with `Set` or `Get` are used for setting and
getting socket options.
*/
type Socket struct {
	soc unsafe.Pointer
}

/*
Socket as string.
*/
func (soc Socket) String() string {
	t, _ := soc.GetType()
	i, err := soc.GetIdentity()
	if err == nil && i != "" {
		return fmt.Sprintf("Socket(%v,%q)", t, i)
	}
	return fmt.Sprintf("Socket(%v,%p)", t, soc.soc)
}

/*
Create 0MQ socket.

WARNING:
The Socket is not thread safe. This means that you cannot access the same Socket
from different goroutines without using something like a mutex.

For a description of socket types, see: http://api.zeromq.org/4-0:zmq-socket#toc3
*/
func NewSocket(t Type) (soc *Socket, err error) {
	soc = &Socket{}
	s, e := C.zmq_socket(ctx, C.int(t))
	if s == nil {
		err = errget(e)
	} else {
		soc.soc = s
		runtime.SetFinalizer(soc, (*Socket).Close)
	}
	return
}

// If not called explicitly, the socket will be closed on garbage collection
func (soc *Socket) Close() error {
	if i, err := C.zmq_close(soc.soc); int(i) != 0 {
		return errget(err)
	}
	soc.soc = unsafe.Pointer(nil)
	return nil
}

/*
Accept incoming connections on a socket.

For a description of endpoint, see: http://api.zeromq.org/4-0:zmq-bind#toc2
*/
func (soc *Socket) Bind(endpoint string) error {
	s := C.CString(endpoint)
	defer C.free(unsafe.Pointer(s))
	if i, err := C.zmq_bind(soc.soc, s); int(i) != 0 {
		return errget(err)
	}
	return nil
}

/*
Stop accepting connections on a socket.

For a description of endpoint, see: http://api.zeromq.org/4-0:zmq-bind#toc2
*/
func (soc *Socket) Unbind(endpoint string) error {
	s := C.CString(endpoint)
	defer C.free(unsafe.Pointer(s))
	if i, err := C.zmq_unbind(soc.soc, s); int(i) != 0 {
		return errget(err)
	}
	return nil
}

/*
Create outgoing connection from socket.

For a description of endpoint, see: http://api.zeromq.org/4-0:zmq-connect#toc2
*/
func (soc *Socket) Connect(endpoint string) error {
	s := C.CString(endpoint)
	defer C.free(unsafe.Pointer(s))
	if i, err := C.zmq_connect(soc.soc, s); int(i) != 0 {
		return errget(err)
	}
	return nil
}

/*
Disconnect a socket.

For a description of endpoint, see: http://api.zeromq.org/4-0:zmq-connect#toc2
*/
func (soc *Socket) Disconnect(endpoint string) error {
	s := C.CString(endpoint)
	defer C.free(unsafe.Pointer(s))
	if i, err := C.zmq_disconnect(soc.soc, s); int(i) != 0 {
		return errget(err)
	}
	return nil
}

/*
Receive a message part from a socket.

For a description of flags, see: http://api.zeromq.org/4-0:zmq-msg-recv#toc2
*/
func (soc *Socket) Recv(flags Flag) (string, error) {
	b, err := soc.RecvBytes(flags)
	return string(b), err
}

/*
Receive a message part from a socket.

For a description of flags, see: http://api.zeromq.org/4-0:zmq-msg-recv#toc2
*/
func (soc *Socket) RecvBytes(flags Flag) ([]byte, error) {
	var msg C.zmq_msg_t
	if i, err := C.zmq_msg_init(&msg); i != 0 {
		return []byte{}, errget(err)
	}
	defer C.zmq_msg_close(&msg)

	size, err := C.zmq_msg_recv(&msg, soc.soc, C.int(flags))
	if size < 0 {
		return []byte{}, errget(err)
	}
	if size == 0 {
		return []byte{}, nil
	}
	data := make([]byte, int(size))
	C.my_memcpy(unsafe.Pointer(&data[0]), C.zmq_msg_data(&msg), C.size_t(size))
	return data, nil
}

/*
Send a message part on a socket.

For a description of flags, see: http://api.zeromq.org/4-0:zmq-send#toc2
*/
func (soc *Socket) Send(data string, flags Flag) (int, error) {
	return soc.SendBytes([]byte(data), flags)
}

/*
Send a message part on a socket.

For a description of flags, see: http://api.zeromq.org/4-0:zmq-send#toc2
*/
func (soc *Socket) SendBytes(data []byte, flags Flag) (int, error) {
	d := data
	if len(data) == 0 {
		d = []byte{0}
	}
	size, err := C.zmq_send(soc.soc, unsafe.Pointer(&d[0]), C.size_t(len(data)), C.int(flags))
	if size < 0 {
		return int(size), errget(err)
	}
	return int(size), nil
}

/*
Register a monitoring callback.

See: http://api.zeromq.org/4-0:zmq-socket-monitor#toc2

Example:

    package main

    import (
        zmq "github.com/pebbe/zmq4"
        "log"
        "time"
    )

    func rep_socket_monitor(addr string) {
        s, err := zmq.NewSocket(zmq.PAIR)
        if err != nil {
            log.Fatalln(err)
        }
        err = s.Connect(addr)
        if err != nil {
            log.Fatalln(err)
        }
        for {
            a, b, c, err := s.RecvEvent(0)
            if err != nil {
                log.Println(err)
                break
            }
            log.Println(a, b, c)
        }
        s.Close()
    }

    func main() {

        // REP socket
        rep, err := zmq.NewSocket(zmq.REP)
        if err != nil {
            log.Fatalln(err)
        }

        // REP socket monitor, all events
        err = rep.Monitor("inproc://monitor.rep", zmq.EVENT_ALL)
        if err != nil {
            log.Fatalln(err)
        }
        go rep_socket_monitor("inproc://monitor.rep")

        // Generate an event
        rep.Bind("tcp://*:5555")
        if err != nil {
            log.Fatalln(err)
        }

        // Allow some time for event detection
        time.Sleep(time.Second)

        rep.Close()
        zmq.Term()
    }
*/
func (soc *Socket) Monitor(addr string, events Event) error {
	s := C.CString(addr)
	defer C.free(unsafe.Pointer(s))
	if i, err := C.zmq_socket_monitor(soc.soc, s, C.int(events)); i != 0 {
		return errget(err)
	}
	return nil
}

/*
Start built-in ØMQ proxy

See: http://api.zeromq.org/4-0:zmq-proxy#toc2
*/
func Proxy(frontend, backend, capture *Socket) error {
	var capt unsafe.Pointer
	if capture != nil {
		capt = capture.soc
	}
	_, err := C.zmq_proxy(frontend.soc, backend.soc, capt)
	return errget(err)
}

//. CURVE

/*
Encode a binary key as Z85 printable text

See: http://api.zeromq.org/4-0:zmq-z85-encode
*/
func Z85encode(data string) string {
	l1 := len(data)
	if l1%4 != 0 {
		panic("Z85encode: Length of data not a multiple of 4")
	}
	d := []byte(data)

	l2 := 5 * l1 / 4
	dest := make([]byte, l2+1)

	C.zmq_z85_encode((*C.char)(unsafe.Pointer(&dest[0])), (*C.uint8_t)(&d[0]), C.size_t(l1))

	return string(dest[:l2])
}

/*
Decode a binary key from Z85 printable text

See: http://api.zeromq.org/4-0:zmq-z85-decode
*/
func Z85decode(s string) string {
	l1 := len(s)
	if l1%5 != 0 {
		panic("Z85decode: Length of Z85 string not a multiple of 5")
	}
	l2 := 4 * l1 / 5
	dest := make([]byte, l2)
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	C.zmq_z85_decode((*C.uint8_t)(&dest[0]), cs)
	return string(dest)
}

/*
Generate a new CURVE keypair

See: http://api.zeromq.org/4-0:zmq-curve-keypair
*/
func NewCurveKeypair() (z85_public_key, z85_secret_key string, err error) {
	var pubkey, seckey [41]byte
	if i, err := C.zmq_curve_keypair((*C.char)(unsafe.Pointer(&pubkey[0])), (*C.char)(unsafe.Pointer(&seckey[0]))); i != 0 {
		return "", "", errget(err)
	}
	return string(pubkey[:40]), string(seckey[:40]), nil
}
