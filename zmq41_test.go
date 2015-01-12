package zmq4_test

import (
	zmq "github.com/pebbe/zmq4"

	"fmt"
	"time"
)

func Example_test_srcfd() {

	if _, minor, _ := zmq.Version(); minor < 1 {
		fmt.Println("127.0.0.1")
		fmt.Println("Done")
		return
	}

	addr := "tcp://127.0.0.1:9560"

	rep, err := zmq.NewSocket(zmq.REP)
	if checkErr(err) {
		return
	}
	req, err := zmq.NewSocket(zmq.REQ)
	if checkErr(err) {
		rep.Close()
		return
	}

	err = rep.Bind(addr)
	if checkErr(err) {
		rep.Close()
		req.Close()
		return
	}
	err = req.Connect(addr)
	if checkErr(err) {
		rep.Close()
		req.Close()
		return
	}

	tmp := "test"
	_, err = req.Send(tmp, 0)
	if checkErr(err) {
		rep.Close()
		req.Close()
		return
	}

	msg, prop, _, err := rep.RecvWithProperties(0, []zmq.Property{zmq.SRCFD}, []string{})
	if checkErr(err) {
		rep.Close()
		req.Close()
		return
	}
	if msg != tmp {
		fmt.Println(tmp, "!=", msg)
	}

	// get the messages source file descriptor
	if prop[0] < 0 {
		fmt.Println("SRCFD < 0")
	}

	// get the remote endpoint
	sockstr, err := zmq.GetPeerAddr(prop[0])
	if checkErr(err) {
		rep.Close()
		req.Close()
		return
	}
	fmt.Println(sockstr)

	err = rep.Close()
	if checkErr(err) {
		req.Close()
		return
	}

	err = req.Close()
	if checkErr(err) {
		return
	}

	// sleep a bit for the socket to be freed
	time.Sleep(30 * time.Millisecond)

	// getting name from closed socket will fail
	_, err = zmq.GetPeerAddr(prop[0])
	if err == nil {
		fmt.Println("GetPeerAddr should fail")
	}

	fmt.Println("Done")
	// Output:
	// 127.0.0.1
	// Done
}
