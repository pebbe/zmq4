package zmq4

/*
#include <zmq.h>
*/
import "C"

import (
	"fmt"
	"time"
)

// Return type for (*Poller)Poll
type Polled struct {
	Socket *Socket // socket with matched event(s)
	Events State   // actual matched event(s)
}

type Poller struct {
	items []C.zmq_pollitem_t
	socks []*Socket
	size  int
}

// Create a new Poller
func NewPoller() *Poller {
	return &Poller{
		items: make([]C.zmq_pollitem_t, 0),
		socks: make([]*Socket, 0),
		size:  0}
}

// Add items to the poller
//
// Events is a bitwise OR of zmq.POLLIN and zmq.POLLOUT
func (p *Poller) Add(soc *Socket, events State) {
	var item C.zmq_pollitem_t
	item.socket = soc.soc
	item.fd = 0
	item.events = C.short(events)
	p.items = append(p.items, item)
	p.socks = append(p.socks, soc)
	p.size += 1
}

/*
Input/output multiplexing

If timeout < 0, wait forever until a matching event is detected

Only sockets with matching socket events are returned in the list.

Example:

    poller := zmq.NewPoller()
    poller.Add(socket0, zmq.POLLIN)
    poller.Add(socket1, zmq.POLLIN)
    //  Process messages from both sockets
    for {
        sockets, _ := poller.Poll(-1)
        for _, socket := range sockets {
            switch s := socket.Socket; s {
            case socket0:
                msg, _ := s.Recv(0)
                //  Process msg
            case socket1:
                msg, _ := s.Recv(0)
                //  Process msg
            }
        }
    }
*/
func (p *Poller) Poll(timeout time.Duration) ([]Polled, error) {
	return p.poll(timeout, false)
}

/*
This is like (*Poller)Poll, but it returns a list of all sockets,
in the same order as they were added to the poller,
not just those sockets that had an event.

For each socket in the list, you have to check the Events field
to see if there was actually an event.

When error is not nil, the return list contains no sockets.
*/
func (p *Poller) PollAll(timeout time.Duration) ([]Polled, error) {
	return p.poll(timeout, true)
}

func (p *Poller) poll(timeout time.Duration, all bool) ([]Polled, error) {
	lst := make([]Polled, 0, p.size)

	for _, soc := range p.socks {
		if !soc.opened {
			return lst, ErrorSocketClosed
		}
	}

	t := timeout
	if t > 0 {
		t = t / time.Millisecond
	}
	if t < 0 {
		t = -1
	}
	rv, err := C.zmq_poll(&p.items[0], C.int(len(p.items)), C.long(t))
	if rv < 0 {
		return lst, errget(err)
	}
	for i, it := range p.items {
		if all || it.events&it.revents != 0 {
			lst = append(lst, Polled{p.socks[i], State(it.revents)})
		}
	}
	return lst, nil
}

// Poller as string.
func (p *Poller) String() string {
	str := make([]string, 0)
	for i, poll := range p.items {
		str = append(str, fmt.Sprintf("%v%v", p.socks[i], State(poll.events)))
	}
	return fmt.Sprint("Poller", str)
}
