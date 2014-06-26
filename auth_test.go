package zmq4_test

import (
	zmq "github.com/pebbe/zmq4"

	"fmt"
	"log"
)

func ExampleAuthStart() {

	checkErr := func(err error) bool {
		if err == nil {
			return false
		}
		log.Println(err)
		return true
	}

	zmq.AuthSetVerbose(false)

	//  Start authentication engine
	zmq.AuthStart()
	defer zmq.AuthStop()

	zmq.AuthAllow("127.0.0.1")

	zmq.AuthSetMetaHandler(func(version, request_id, domain, address, identity, mechanism string) (metadata map[string]string) {
		return map[string]string{
			"User-Id": "anonymous",
			"Hello":   "World!",
			"Foo":     "Bar",
		}
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
	err = server.Bind("tcp://*:9000")
	if checkErr(err) {
		return
	}

	//  Create and connect client socket
	client, err := zmq.NewSocket(zmq.DEALER)
	if checkErr(err) {
		return
	}
	defer client.Close()
	client.ClientAuthCurve(server_public, client_public, client_secret)
	err = client.Connect("tcp://127.0.0.1:9000")
	if checkErr(err) {
		return
	}

	//  Send a message from client to server
	_, err = client.SendMessage("Greatings", "Earthlings!")
	if checkErr(err) {
		return
	}

	// Receive message and metadata on the server
	message, metadata, err := server.RecvMessageWithMetadata(0, "User-Id", "Socket-Type", "Hello", "Foo", "Fuz")
	if checkErr(err) {
		return
	}
	fmt.Println(message)
	if _, minor, _ := zmq.Version(); minor < 1 {
		// Metadata requires at least ZeroMQ version 4.1
		fmt.Println("[{anonymous <nil>} {DEALER <nil>} {World! <nil>} {Bar <nil>} { invalid argument}]")
	} else {
		fmt.Println(metadata)
	}

	zmq.AuthStop()

	// Output:
	// [Greatings Earthlings!]
	// [{anonymous <nil>} {DEALER <nil>} {World! <nil>} {Bar <nil>} { invalid argument}]
}
