package zmq4_test

import (
	zmq "github.com/pebbe/zmq4"

	"fmt"
)

func Example_test_remote_endpoint() {

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

	// get message with peer address (remote endpoint)
	msg, props, err := rep.RecvWithMetadata(0, "Peer-Address")
	if checkErr(err) {
		rep.Close()
		req.Close()
		return
	}
	if msg != tmp {
		fmt.Println(tmp, "!=", msg)
	}

	fmt.Println(props["Peer-Address"])

	err = rep.Close()
	if checkErr(err) {
		req.Close()
		return
	}

	err = req.Close()
	if checkErr(err) {
		return
	}

	fmt.Println("Done")
	// Output:
	// 127.0.0.1
	// Done
}
