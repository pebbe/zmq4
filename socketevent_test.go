package zmq4_test

import (
	zmq "github.com/pebbe/zmq4"

	"fmt"
	"time"
)

func rep_socket_monitor(addr string) {
	s, err := zmq.NewSocket(zmq.PAIR)
	if checkErr(err) {
		return
	}
	defer s.Close()
	err = s.Connect(addr)
	if checkErr(err) {
		return
	}
	for {
		a, b, _, err := s.RecvEvent(0)
		if checkErr(err) {
			break
		}
		fmt.Println(a, b)
		if a == zmq.EVENT_CLOSED {
			break
		}
	}
}

func Example_socket_event() {

	// REP socket
	rep, err := zmq.NewSocket(zmq.REP)
	if checkErr(err) {
		return
	}

	// REP socket monitor, all events
	err = rep.Monitor("inproc://monitor.rep", zmq.EVENT_ALL)
	if checkErr(err) {
		rep.Close()
		return
	}
	go rep_socket_monitor("inproc://monitor.rep")
	time.Sleep(time.Second)

	// Generate an event
	rep.Bind("tcp://*:9689")
	if checkErr(err) {
		rep.Close()
		return
	}

	rep.Close()

	// Allow some time for event detection
	time.Sleep(time.Second)

	fmt.Println("Done")
	// Output:
	// EVENT_LISTENING tcp://0.0.0.0:9689
	// EVENT_CLOSED tcp://0.0.0.0:9689
	// Done
}
