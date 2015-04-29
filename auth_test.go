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
	err := zmq.AuthStart()
	if checkErr(err) {
		return
	}
	defer zmq.AuthStop()

	zmq.AuthSetMetadataHandler(
		func(version, request_id, domain, address, identity, mechanism string, credentials ...string) (metadata map[string]string) {
			return map[string]string{
				"Identity": identity,
				"User-Id":  "anonymous",
				"Hello":    "World!",
				"Foo":      "Bar",
			}
		})

	zmq.AuthAllow("domain1", "127.0.0.1")

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
	zmq.AuthCurveAdd("domain1", client_public)

	//  Create and bind server socket
	server, err := zmq.NewSocket(zmq.DEALER)
	if checkErr(err) {
		return
	}
	defer server.Close()
	server.SetIdentity("Server1")
	server.ServerAuthCurve("domain1", server_secret)
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
	server.SetIdentity("Client1")
	client.ClientAuthCurve(server_public, client_public, client_secret)
	err = client.Connect("tcp://127.0.0.1:9000")
	if checkErr(err) {
		return
	}

	//  Send a message from client to server
	_, err = client.SendMessage("Greetings", "Earthlings!")
	if checkErr(err) {
		return
	}

	// Receive message and metadata on the server
	keys := []string{"Identity", "User-Id", "Socket-Type", "Hello", "Foo", "Fuz"}
	message, metadata, err := server.RecvMessageWithMetadata(0, keys...)
	if checkErr(err) {
		return
	}
	fmt.Println(message)
	if _, minor, _ := zmq.Version(); minor < 1 {
		// Metadata requires at least ZeroMQ version 4.1
		fmt.Println(`Identity: "Server1" true`)
		fmt.Println(`User-Id: "anonymous" true`)
		fmt.Println(`Socket-Type: "DEALER" true`)
		fmt.Println(`Hello: "World!" true`)
		fmt.Println(`Foo: "Bar" true`)
		fmt.Println(`Fuz: "" false`)
	} else {
		for _, key := range keys {
			value, ok := metadata[key]
			fmt.Printf("%v: %q %v\n", key, value, ok)
		}
	}
	// Output:
	// [Greetings Earthlings!]
	// Identity: "Server1" true
	// User-Id: "anonymous" true
	// Socket-Type: "DEALER" true
	// Hello: "World!" true
	// Foo: "Bar" true
	// Fuz: "" false
}
