package zmq4_test

import (
	zmq "github.com/pebbe/zmq4"

	"fmt"
	"time"
)

func Example_auth() {

	time.Sleep(100 * time.Millisecond)

	//  Start authentication engine
	zmq.AuthStart()
	defer zmq.AuthStop()

	zmq.AuthAllow("127.0.0.1")

	zmq.AuthSetMetaHandler(func(version, request_id, domain, address, identity, mechanism string) (user_id string, metadata []byte) {
		b, _ := zmq.AuthMetaBlobs(
			[2]string{"Hello", "World!"},
			[2]string{"Foo", "Bar"},
		)
		return "anonymous", b
	})

	//  We need two certificates, one for the client and one for
	//  the server. The client must know the server's public key
	//  to make a CURVE connection.
	client_public, client_secret, err := zmq.NewCurveKeypair()
	if checkErr(err) {
		return
	}
	server_public, server_secret, err := zmq.NewCurveKeypair()
	if checkErr(err) {
		return
	}

	//  Tell authenticator to use this public client key
	zmq.AuthCurveAdd("global", client_public)

	//  Create and bind server socket
	server, err := zmq.NewSocket(zmq.DEALER)
	if checkErr(err) {
		return
	}
	defer server.Close()
	server.ServerAuthCurve("global", server_secret)
	server.Bind("tcp://*:9000")

	//  Create and connect client socket
	client, err := zmq.NewSocket(zmq.DEALER)
	if checkErr(err) {
		return
	}
	defer client.Close()
	client.ClientAuthCurve(server_public, client_public, client_secret)
	client.Connect("tcp://127.0.0.1:9000")

	//  Send a single message from client to server
	_, err = client.Send("Greatings!", 0)
	if checkErr(err) {
		return
	}

	// Receive message on the server
	message, metadata, err := server.RecvWithMetadata(0, "User-Id", "Socket-Type", "Hello", "Foo", "Fuz")
	if checkErr(err) {
		return
	}
	if _, minor, _ := zmq.Version(); minor < 1 {
		fmt.Println("[{anonymous <nil>} {DEALER <nil>} {World! <nil>} {Bar <nil>} { invalid argument}]")
	} else {
		fmt.Println(metadata)
	}
	fmt.Println(message)

	zmq.AuthStop()

	// Output:
	// [{anonymous <nil>} {DEALER <nil>} {World! <nil>} {Bar <nil>} { invalid argument}]
	// Greatings!
}
