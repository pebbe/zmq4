package zmq4_test

import (
	zmq "github.com/pebbe/zmq4"

	"testing"
)

func TestRemoteEndpoint(t *testing.T) {

	if _, minor, _ := zmq.Version(); minor < 1 {
		t.Skip("RemoteEndpoint not avalable in ZeroMQ versions prior to 4.1.0")
	}

	addr := "tcp://127.0.0.1:9560"
	peer := "127.0.0.1"

	rep, err := zmq.NewSocket(zmq.REP)
	if err != nil {
		t.Fatal("NewSocket:", err)
	}
	req, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		rep.Close()
		t.Fatal("NewSocket:", err)
	}

	if err = rep.Bind(addr); err != nil {
		rep.Close()
		req.Close()
		t.Fatal("rep.Bind:", err)
	}
	if err = req.Connect(addr); err != nil {
		rep.Close()
		req.Close()
		t.Fatal("req.Connect:", err)
	}

	tmp := "test"
	if _, err = req.Send(tmp, 0); err != nil {
		rep.Close()
		req.Close()
		t.Fatal("req.Send:", err)
	}

	// get message with peer address (remote endpoint)
	msg, props, err := rep.RecvWithMetadata(0, "Peer-Address")
	if err != nil {
		rep.Close()
		req.Close()
		t.Fatal("rep.RecvWithMetadata:", err)
		return
	}
	if msg != tmp {
		t.Errorf("rep.RecvWithMetadata: expected %q, got %q", tmp, msg)
	}

	if p := props["Peer-Address"]; p != peer {
		t.Errorf("rep.RecvWithMetadata: expected Peer-Address == %q, got %q", peer, p)
	}

	if err = rep.Close(); err != nil {
		req.Close()
		t.Fatal("rep.Close:", err)
	}

	if err = req.Close(); err != nil {
		t.Fatal("req.Close:", err)
	}
}
