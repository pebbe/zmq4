package zmq4_test

import (
	zmq "github.com/pebbe/zmq4"

	"errors"
	"fmt"
	"runtime"
	"strconv"
	"testing"
	"time"
)

var (
	errerr = errors.New("error")
	err32  = errors.New("rc != 32")
)

func TestVersion(t *testing.T) {
	major, _, _ := zmq.Version()
	if major != 4 {
		t.Errorf("Expected major version 4, got %d", major)
	}
}

func TestMultipleContexts(t *testing.T) {

	chQuit := make(chan interface{})
	chErr := make(chan error, 2)
	needQuit := false
	var sock1, sock2, serv1, serv2 *zmq.Socket
	var serv_ctx1, serv_ctx2, ctx1, ctx2 *zmq.Context
	var err error

	defer func() {
		if needQuit {
			chQuit <- true
			chQuit <- true
			<-chErr
			<-chErr
		}
		if sock1 != nil {
			sock1.Close()
		}
		if sock2 != nil {
			sock2.Close()
		}
		if serv1 != nil {
			serv1.Close()
		}
		if serv2 != nil {
			serv2.Close()
		}
		if serv_ctx1 != nil {
			serv_ctx1.Term()
		}
		if serv_ctx2 != nil {
			serv_ctx2.Term()
		}
		if ctx1 != nil {
			ctx1.Term()
		}
		if ctx2 != nil {
			ctx2.Term()
		}
	}()

	addr1 := "tcp://127.0.0.1:9997"
	addr2 := "tcp://127.0.0.1:9998"

	serv_ctx1, err = zmq.NewContext()
	if err != nil {
		t.Fatal("NewContext:", err)
	}
	serv1, err = serv_ctx1.NewSocket(zmq.REP)
	if err != nil {
		t.Fatal("NewSocket:", err)
	}
	err = serv1.Bind(addr1)
	if err != nil {
		t.Fatal("Bind:", err)
	}

	serv_ctx2, err = zmq.NewContext()
	if err != nil {
		t.Fatal("NewContext:", err)
	}
	serv2, err = serv_ctx2.NewSocket(zmq.REP)
	if err != nil {
		t.Fatal("NewSocket:", err)
	}
	err = serv2.Bind(addr2)
	if err != nil {
		t.Fatal("Bind:", err)
	}

	new_service := func(sock *zmq.Socket, addr string) {
		socket_handler := func(state zmq.State) error {
			msg, err := sock.RecvMessage(0)
			if err != nil {
				return err
			}
			_, err = sock.SendMessage(addr, msg)
			return err
		}
		quit_handler := func(interface{}) error {
			return errors.New("quit")
		}

		reactor := zmq.NewReactor()
		reactor.AddSocket(sock, zmq.POLLIN, socket_handler)
		reactor.AddChannel(chQuit, 1, quit_handler)
		err = reactor.Run(100 * time.Millisecond)
		chErr <- err
	}

	go new_service(serv1, addr1)
	go new_service(serv2, addr2)
	needQuit = true

	time.Sleep(time.Second)

	// default context

	sock1, err = zmq.NewSocket(zmq.REQ)
	if err != nil {
		t.Fatal("NewSocket:", err)
	}
	sock2, err = zmq.NewSocket(zmq.REQ)
	if err != nil {
		t.Fatal("NewSocket:", err)
	}
	err = sock1.Connect(addr1)
	if err != nil {
		t.Fatal("sock1.Connect:", err)
	}
	err = sock2.Connect(addr2)
	if err != nil {
		t.Fatal("sock2.Connect:", err)
	}
	_, err = sock1.SendMessage(addr1)
	if err != nil {
		t.Fatal("sock1.SendMessage:", err)
	}
	_, err = sock2.SendMessage(addr2)
	if err != nil {
		t.Fatal("sock2.SendMessage:", err)
	}
	msg, err := sock1.RecvMessage(0)
	expected := []string{addr1, addr1}
	if err != nil || !arrayEqual(msg, expected) {
		t.Errorf("sock1.RecvMessage: expected %v %v, got %v %v", nil, expected, err, msg)
	}
	msg, err = sock2.RecvMessage(0)
	expected = []string{addr2, addr2}
	if err != nil || !arrayEqual(msg, expected) {
		t.Errorf("sock2.RecvMessage: expected %v %v, got %v %v", nil, expected, err, msg)
	}
	err = sock1.Close()
	sock1 = nil
	if err != nil {
		t.Fatal("sock1.Close:", err)
	}
	err = sock2.Close()
	sock2 = nil
	if err != nil {
		t.Fatal("sock2.Close:", err)
	}

	// non-default contexts

	ctx1, err = zmq.NewContext()
	if err != nil {
		t.Fatal("NewContext:", err)
	}
	ctx2, err = zmq.NewContext()
	if err != nil {
		t.Fatal("NewContext:", err)
	}
	sock1, err = ctx1.NewSocket(zmq.REQ)
	if err != nil {
		t.Fatal("ctx1.NewSocket:", err)
	}
	sock2, err = ctx2.NewSocket(zmq.REQ)
	if err != nil {
		t.Fatal("ctx2.NewSocket:", err)
	}
	err = sock1.Connect(addr1)
	if err != nil {
		t.Fatal("sock1.Connect:", err)
	}
	err = sock2.Connect(addr2)
	if err != nil {
		t.Fatal("sock2.Connect:", err)
	}
	_, err = sock1.SendMessage(addr1)
	if err != nil {
		t.Fatal("sock1.SendMessage:", err)
	}
	_, err = sock2.SendMessage(addr2)
	if err != nil {
		t.Fatal("sock2.SendMessage:", err)
	}
	msg, err = sock1.RecvMessage(0)
	expected = []string{addr1, addr1}
	if err != nil || !arrayEqual(msg, expected) {
		t.Errorf("sock1.RecvMessage: expected %v %v, got %v %v", nil, expected, err, msg)
	}
	msg, err = sock2.RecvMessage(0)
	expected = []string{addr2, addr2}
	if err != nil || !arrayEqual(msg, expected) {
		t.Errorf("sock2.RecvMessage: expected %v %v, got %v %v", nil, expected, err, msg)
	}
	err = sock1.Close()
	sock1 = nil
	if err != nil {
		t.Fatal("sock1.Close:", err)
	}
	err = sock2.Close()
	sock2 = nil
	if err != nil {
		t.Fatal("sock2.Close:", err)
	}

	err = ctx1.Term()
	ctx1 = nil
	if err != nil {
		t.Fatal("ctx1.Term", nil)
	}
	err = ctx2.Term()
	ctx1 = nil
	if err != nil {
		t.Fatal("ctx2.Term", nil)
	}

	needQuit = false
	for i := 0; i < 2; i++ {
		// close(chQuit) doesn't work because the reactor removes closed channels, instead of acting on them
		chQuit <- true
		err = <-chErr
		if err.Error() != "quit" {
			t.Errorf("Expected error value quit, got %v", err)
		}
	}
}

func TestAbstractIpc(t *testing.T) {

	var sb, sc *zmq.Socket
	defer func() {
		if sb != nil {
			sb.Close()
		}
		if sc != nil {
			sc.Close()
		}
	}()

	addr := "ipc://@/tmp/tester"

	// This is Linux only
	if runtime.GOOS != "linux" {
		t.Skip("Only on Linux")
	}

	sb, err := zmq.NewSocket(zmq.PAIR)
	if err != nil {
		t.Fatal("NewSocket:", err)
	}

	err = sb.Bind(addr)
	if err != nil {
		t.Fatal("sb.Bind:", err)
	}

	endpoint, err := sb.GetLastEndpoint()
	expected := "ipc://@/tmp/tester"
	if endpoint != expected || err != nil {
		t.Fatalf("sb.GetLastEndpoint: expected 'nil' %q, got '%v' %q", expected, err, endpoint)
		return
	}

	sc, err = zmq.NewSocket(zmq.PAIR)
	if err != nil {
		t.Fatal("NewSocket:", err)
	}
	err = sc.Connect(addr)
	if err != nil {
		t.Fatal("sc.Bind:", err)
	}

	resp, err := bouncee(sb, sc)
	if err != nil {
		t.Error(resp+":", err)
	}

	err = sc.Close()
	sc = nil
	if err != nil {
		t.Fatal("sc.Close:", err)
	}

	err = sb.Close()
	sb = nil
	if err != nil {
		t.Fatal("sb.Close:", err)
	}
}

func TestConflate(t *testing.T) {

	var s_in, s_out *zmq.Socket

	defer func() {
		if s_in != nil {
			s_in.Close()
		}
		if s_out != nil {
			s_out.Close()
		}
	}()

	bind_to := "tcp://127.0.0.1:5555"

	err := zmq.SetIoThreads(1)
	if err != nil {
		t.Fatal("SetIoThreads(1):", err)
	}

	s_in, err = zmq.NewSocket(zmq.PULL)
	if err != nil {
		t.Fatal("NewSocket 1:", err)
	}

	err = s_in.SetConflate(true)
	if err != nil {
		t.Fatal("SetConflate(true):", err)
	}

	err = s_in.Bind(bind_to)
	if err != nil {
		t.Fatal("s_in.Bind:", err)
	}

	s_out, err = zmq.NewSocket(zmq.PUSH)
	if err != nil {
		t.Fatal("NewSocket 2:", err)
	}

	err = s_out.Connect(bind_to)
	if err != nil {
		t.Fatal("s_out.Connect:", err)
	}

	message_count := 20

	for j := 0; j < message_count; j++ {
		_, err = s_out.Send(fmt.Sprint(j), 0)
		if err != nil {
			t.Fatalf("s_out.Send %d: %v", j, err)
		}
	}

	time.Sleep(time.Second)

	payload_recved, err := s_in.Recv(0)
	if err != nil {
		t.Error("s_in.Recv:", err)
	} else {
		i, err := strconv.Atoi(payload_recved)
		if err != nil {
			t.Error("strconv.Atoi:", err)
		}
		if i != message_count-1 {
			t.Error("payload_recved != message_count - 1")
		}
	}

	err = s_in.Close()
	s_in = nil
	if err != nil {
		t.Error("s_in.Close:", err)
	}

	err = s_out.Close()
	if err != nil {
		t.Error("s_out.Close:", err)
	}
}

func TestConnectResolve(t *testing.T) {

	sock, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		t.Fatal("NewSocket:", err)
	}

	err = sock.Connect("tcp://localhost:1234")
	if err != nil {
		t.Error("sock.Connect:", err)
	}

	fails := []string{
		"tcp://localhost:invalid",
		"tcp://in val id:1234",
		"invalid://localhost:1234",
	}
	for _, fail := range fails {
		if err = sock.Connect(fail); err == nil {
			t.Errorf("Connect %s, expected fail, got success", fail)
		}
	}

	if err = sock.Close(); err != nil {
		t.Error("sock.Close:", err)
	}
}

/*

func Example_test_ctx_destroy() {

	fmt.Println("Done")
	// Output:
	// Done
}

*/

func TestCtxOptions(t *testing.T) {

	type Result struct {
		value interface{}
		err   error
	}

	i, err := zmq.GetMaxSockets()
	if err != nil {
		t.Error("GetMaxSockets:", err)
	}
	if i != zmq.MaxSocketsDflt {
		t.Errorf("MaxSockets != MaxSocketsDflt: %d != %d", i, zmq.MaxSocketsDflt)
	}

	i, err = zmq.GetIoThreads()
	if err != nil {
		t.Error("GetIoThreads:", err)
	}
	if i != zmq.IoThreadsDflt {
		t.Errorf("IoThreads != IoThreadsDflt: %d != %d", i, zmq.IoThreadsDflt)
	}

	b, err := zmq.GetIpv6()
	if b != false || err != nil {
		t.Errorf("GetIpv6 1: expected false <nil>, got %v %v", b, err)
	}

	zmq.SetIpv6(true)
	b, err = zmq.GetIpv6()
	if b != true || err != nil {
		t.Errorf("GetIpv6 2: expected true <nil>, got %v %v", b, err)
	}

	router, _ := zmq.NewSocket(zmq.ROUTER)
	b, err = router.GetIpv6()
	if b != true || err != nil {
		t.Errorf("GetIpv6 3: expected true <nil>, got %v %v", b, err)
	}

	router.Close()

	zmq.SetIpv6(false)
}

func Example_test_disconnect_inproc() {

	publicationsReceived := 0
	isSubscribed := false

	pubSocket, err := zmq.NewSocket(zmq.XPUB)
	if checkErr(err) {
		return
	}
	subSocket, err := zmq.NewSocket(zmq.SUB)
	if checkErr(err) {
		return
	}
	err = subSocket.SetSubscribe("foo")
	if checkErr(err) {
		return
	}

	err = pubSocket.Bind("inproc://someInProcDescriptor")
	if checkErr(err) {
		return
	}

	iteration := 0

	poller := zmq.NewPoller()
	poller.Add(subSocket, zmq.POLLIN) // read publications
	poller.Add(pubSocket, zmq.POLLIN) // read subscriptions
	for {
		sockets, err := poller.Poll(100 * time.Millisecond)
		if checkErr(err) {
			break //  Interrupted
		}

		for _, socket := range sockets {
			if socket.Socket == pubSocket {
				for {
					buffer, err := pubSocket.Recv(0)
					if checkErr(err) {
						return
					}
					fmt.Printf("pubSocket: %q\n", buffer)

					if buffer[0] == 0 {
						fmt.Println("pubSocket, isSubscribed == true:", isSubscribed == true)
						isSubscribed = false
					} else {
						fmt.Println("pubSocket, isSubscribed == false:", isSubscribed == false)
						isSubscribed = true
					}

					more, err := pubSocket.GetRcvmore()
					if checkErr(err) {
						return
					}
					if !more {
						break //  Last message part
					}
				}
				break
			}
		}

		for _, socket := range sockets {
			if socket.Socket == subSocket {
				for {
					msg, err := subSocket.Recv(0)
					if checkErr(err) {
						return
					}
					fmt.Printf("subSocket: %q\n", msg)
					more, err := subSocket.GetRcvmore()
					if checkErr(err) {
						return
					}
					if !more {
						publicationsReceived++
						break //  Last message part
					}

				}
				break
			}
		}

		if iteration == 1 {
			err := subSocket.Connect("inproc://someInProcDescriptor")
			checkErr(err)
		}
		if iteration == 4 {
			err := subSocket.Disconnect("inproc://someInProcDescriptor")
			checkErr(err)
		}
		if iteration > 4 && len(sockets) == 0 {
			break
		}

		_, err = pubSocket.Send("foo", zmq.SNDMORE)
		checkErr(err)
		_, err = pubSocket.Send("this is foo!", 0)
		checkErr(err)

		iteration++

	}

	fmt.Println("publicationsReceived == 3:", publicationsReceived == 3)
	fmt.Println("!isSubscribed:", !isSubscribed)

	err = pubSocket.Close()
	checkErr(err)
	err = subSocket.Close()
	checkErr(err)

	fmt.Println("Done")
	// Output:
	// pubSocket: "\x01foo"
	// pubSocket, isSubscribed == false: true
	// subSocket: "foo"
	// subSocket: "this is foo!"
	// subSocket: "foo"
	// subSocket: "this is foo!"
	// subSocket: "foo"
	// subSocket: "this is foo!"
	// pubSocket: "\x00foo"
	// pubSocket, isSubscribed == true: true
	// publicationsReceived == 3: true
	// !isSubscribed: true
	// Done
}

func Example_test_fork() {

	address := "tcp://127.0.0.1:6571"
	NUM_MESSAGES := 5

	//  Create and bind pull socket to receive messages
	pull, err := zmq.NewSocket(zmq.PULL)
	if checkErr(err) {
		return
	}
	err = pull.Bind(address)
	if checkErr(err) {
		return
	}

	done := make(chan bool)

	go func() {
		defer func() { done <- true }()

		//  Create new socket, connect and send some messages

		push, err := zmq.NewSocket(zmq.PUSH)
		if checkErr(err) {
			return
		}
		defer func() {
			err := push.Close()
			checkErr(err)
		}()

		err = push.Connect(address)
		if checkErr(err) {
			return
		}

		for count := 0; count < NUM_MESSAGES; count++ {
			_, err = push.Send("Hello", 0)
			if checkErr(err) {
				return
			}
		}

	}()

	for count := 0; count < NUM_MESSAGES; count++ {
		msg, err := pull.Recv(0)
		fmt.Printf("%q %v\n", msg, err)
	}

	err = pull.Close()
	checkErr(err)

	<-done

	fmt.Println("Done")
	// Output:
	// "Hello" <nil>
	// "Hello" <nil>
	// "Hello" <nil>
	// "Hello" <nil>
	// "Hello" <nil>
	// Done
}

func Example_test_hwm() {

	MAX_SENDS := 10000
	BIND_FIRST := 1
	CONNECT_FIRST := 2

	test_defaults := func() int {

		// Set up bind socket
		bind_socket, err := zmq.NewSocket(zmq.PULL)
		if checkErr(err) {
			return 0
		}
		defer func() {
			err := bind_socket.Close()
			checkErr(err)
		}()

		err = bind_socket.Bind("inproc://a")
		if checkErr(err) {
			return 0
		}

		// Set up connect socket
		connect_socket, err := zmq.NewSocket(zmq.PUSH)
		if checkErr(err) {
			return 0
		}
		defer func() {
			err := connect_socket.Close()
			checkErr(err)
		}()

		err = connect_socket.Connect("inproc://a")
		if checkErr(err) {
			return 0
		}

		// Send until we block
		send_count := 0
		for send_count < MAX_SENDS {
			_, err := connect_socket.Send("", zmq.DONTWAIT)
			if err != nil {
				break
			}
			send_count++
		}

		// Now receive all sent messages
		recv_count := 0
		for {
			_, err := bind_socket.Recv(zmq.DONTWAIT)
			if err != nil {
				break
			}
			recv_count++
		}
		fmt.Println("send_count == recv_count:", send_count == recv_count)

		return send_count
	}

	count_msg := func(send_hwm, recv_hwm, testType int) int {

		var bind_socket, connect_socket *zmq.Socket
		var err error

		if testType == BIND_FIRST {
			// Set up bind socket
			bind_socket, err = zmq.NewSocket(zmq.PULL)
			if checkErr(err) {
				return 0
			}
			defer func() {
				err := bind_socket.Close()
				checkErr(err)
			}()

			err = bind_socket.SetRcvhwm(recv_hwm)
			if checkErr(err) {
				return 0
			}

			err = bind_socket.Bind("inproc://a")
			if checkErr(err) {
				return 0
			}

			// Set up connect socket
			connect_socket, err = zmq.NewSocket(zmq.PUSH)
			if checkErr(err) {
				return 0
			}
			defer func() {
				err := connect_socket.Close()
				checkErr(err)
			}()

			err = connect_socket.SetSndhwm(send_hwm)
			if checkErr(err) {
				return 0
			}

			err = connect_socket.Connect("inproc://a")
			if checkErr(err) {
				return 0
			}
		} else {
			// Set up connect socket
			connect_socket, err = zmq.NewSocket(zmq.PUSH)
			if checkErr(err) {
				return 0
			}
			defer func() {
				err := connect_socket.Close()
				checkErr(err)
			}()

			err = connect_socket.SetSndhwm(send_hwm)
			if checkErr(err) {
				return 0
			}

			err = connect_socket.Connect("inproc://a")
			if checkErr(err) {
				return 0
			}

			// Set up bind socket
			bind_socket, err = zmq.NewSocket(zmq.PULL)
			if checkErr(err) {
				return 0
			}
			defer func() {
				err := bind_socket.Close()
				checkErr(err)
			}()

			err = bind_socket.SetRcvhwm(recv_hwm)
			if checkErr(err) {
				return 0
			}

			err = bind_socket.Bind("inproc://a")
			if checkErr(err) {
				return 0
			}
		}

		// Send until we block
		send_count := 0
		for send_count < MAX_SENDS {
			_, err := connect_socket.Send("", zmq.DONTWAIT)
			if err != nil {
				break
			}
			send_count++
		}

		// Now receive all sent messages
		recv_count := 0
		for {
			_, err := bind_socket.Recv(zmq.DONTWAIT)
			if err != nil {
				break
			}
			recv_count++
		}
		fmt.Println("send_count == recv_count:", send_count == recv_count)

		// Now it should be possible to send one more.
		_, err = connect_socket.Send("", 0)
		if checkErr(err) {
			return 0
		}

		//  Consume the remaining message.
		_, err = bind_socket.Recv(0)
		checkErr(err)

		return send_count
	}

	test_inproc_bind_first := func(send_hwm, recv_hwm int) int {
		return count_msg(send_hwm, recv_hwm, BIND_FIRST)
	}

	test_inproc_connect_first := func(send_hwm, recv_hwm int) int {
		return count_msg(send_hwm, recv_hwm, CONNECT_FIRST)
	}

	test_inproc_connect_and_close_first := func(send_hwm, recv_hwm int) int {

		// Set up connect socket
		connect_socket, err := zmq.NewSocket(zmq.PUSH)
		if checkErr(err) {
			return 0
		}

		err = connect_socket.SetSndhwm(send_hwm)
		if checkErr(err) {
			connect_socket.Close()
			return 0
		}

		err = connect_socket.Connect("inproc://a")
		if checkErr(err) {
			connect_socket.Close()
			return 0
		}

		// Send until we block
		send_count := 0
		for send_count < MAX_SENDS {
			_, err := connect_socket.Send("", zmq.DONTWAIT)
			if err != nil {
				break
			}
			send_count++
		}

		// Close connect
		err = connect_socket.Close()
		if checkErr(err) {
			return 0
		}

		// Set up bind socket
		bind_socket, err := zmq.NewSocket(zmq.PULL)
		if checkErr(err) {
			return 0
		}
		defer func() {
			err := bind_socket.Close()
			checkErr(err)
		}()

		err = bind_socket.SetRcvhwm(recv_hwm)
		if checkErr(err) {
			return 0
		}

		err = bind_socket.Bind("inproc://a")
		if checkErr(err) {
			return 0
		}

		// Now receive all sent messages
		recv_count := 0
		for {
			_, err := bind_socket.Recv(zmq.DONTWAIT)
			if err != nil {
				break
			}
			recv_count++
		}
		fmt.Println("send_count == recv_count:", send_count == recv_count)

		return send_count
	}

	// Default values are 1000 on send and 1000 one receive, so 2000 total
	fmt.Println("Default values")
	count := test_defaults()
	fmt.Println("count:", count)
	time.Sleep(100 * time.Millisecond)

	// Infinite send and receive buffer
	fmt.Println("\nInfinite send and receive")
	count = test_inproc_bind_first(0, 0)
	fmt.Println("count:", count)
	time.Sleep(100 * time.Millisecond)
	count = test_inproc_connect_first(0, 0)
	fmt.Println("count:", count)
	time.Sleep(100 * time.Millisecond)

	// Infinite send buffer
	fmt.Println("\nInfinite send buffer")
	count = test_inproc_bind_first(1, 0)
	fmt.Println("count:", count)
	time.Sleep(100 * time.Millisecond)
	count = test_inproc_connect_first(1, 0)
	fmt.Println("count:", count)
	time.Sleep(100 * time.Millisecond)

	// Infinite receive buffer
	fmt.Println("\nInfinite receive buffer")
	count = test_inproc_bind_first(0, 1)
	fmt.Println("count:", count)
	time.Sleep(100 * time.Millisecond)
	count = test_inproc_connect_first(0, 1)
	fmt.Println("count:", count)
	time.Sleep(100 * time.Millisecond)

	// Send and recv buffers hwm 1, so total that can be queued is 2
	fmt.Println("\nSend and recv buffers hwm 1")
	count = test_inproc_bind_first(1, 1)
	fmt.Println("count:", count)
	time.Sleep(100 * time.Millisecond)
	count = test_inproc_connect_first(1, 1)
	fmt.Println("count:", count)
	time.Sleep(100 * time.Millisecond)

	// Send hwm of 1, send before bind so total that can be queued is 1
	fmt.Println("\nSend hwm of 1, send before bind")
	count = test_inproc_connect_and_close_first(1, 0)
	fmt.Println("count:", count)
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\nDone")
	// Output:
	// Default values
	// send_count == recv_count: true
	// count: 2000
	//
	// Infinite send and receive
	// send_count == recv_count: true
	// count: 10000
	// send_count == recv_count: true
	// count: 10000
	//
	// Infinite send buffer
	// send_count == recv_count: true
	// count: 10000
	// send_count == recv_count: true
	// count: 10000
	//
	// Infinite receive buffer
	// send_count == recv_count: true
	// count: 10000
	// send_count == recv_count: true
	// count: 10000
	//
	// Send and recv buffers hwm 1
	// send_count == recv_count: true
	// count: 2
	// send_count == recv_count: true
	// count: 2
	//
	// Send hwm of 1, send before bind
	// send_count == recv_count: true
	// count: 1
	//
	// Done
}

/*

func Example_test_immediate() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_inproc_connect() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_invalid_rep() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_iov() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_issue_566() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_last_endpoint() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_linger() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_monitor() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_msg_flags() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_pair_inproc() {

	fmt.Println("Done")
	// Output:
	// Done
}

*/

func Example_test_pair_ipc() {

	sb, err := zmq.NewSocket(zmq.PAIR)
	if checkErr(err) {
		return
	}

	err = sb.Bind("ipc:///tmp/tester")
	if checkErr(err) {
		return
	}

	sc, err := zmq.NewSocket(zmq.PAIR)
	if checkErr(err) {
		return
	}

	err = sc.Connect("ipc:///tmp/tester")
	if checkErr(err) {
		return
	}

	bounce(sb, sc, false)

	err = sc.Close()
	if checkErr(err) {
		return
	}

	err = sb.Close()
	if checkErr(err) {
		return
	}

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_pair_tcp() {

	sb, err := zmq.NewSocket(zmq.PAIR)
	if checkErr(err) {
		return
	}

	err = sb.Bind("tcp://127.0.0.1:9736")
	if checkErr(err) {
		return
	}

	sc, err := zmq.NewSocket(zmq.PAIR)
	if checkErr(err) {
		return
	}

	err = sc.Connect("tcp://127.0.0.1:9736")
	if checkErr(err) {
		return
	}

	bounce(sb, sc, false)

	err = sc.Close()
	if checkErr(err) {
		return
	}

	err = sb.Close()
	if checkErr(err) {
		return
	}

	fmt.Println("Done")
	// Output:
	// Done
}

/*

func Example_test_probe_router() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_req_correlate() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_req_relaxed() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_reqrep_device() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_reqrep_inproc() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_reqrep_ipc() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_reqrep_tcp() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_router_mandatory() {

	fmt.Println("Done")
	// Output:
	// Done
}

*/

func Example_test_security_curve() {

	time.Sleep(100 * time.Millisecond)

	//  Generate new keypairs for this test
	client_public, client_secret, err := zmq.NewCurveKeypair()
	if checkErr(err) {
		return
	}
	server_public, server_secret, err := zmq.NewCurveKeypair()
	if checkErr(err) {
		return
	}

	handler, err := zmq.NewSocket(zmq.REP)
	if checkErr(err) {
		return
	}
	err = handler.Bind("inproc://zeromq.zap.01")
	if checkErr(err) {
		return
	}

	doHandler := func(state zmq.State) error {
		msg, err := handler.RecvMessage(0)
		if err != nil {
			return err //  Terminating
		}
		version := msg[0]
		sequence := msg[1]
		// domain := msg[2]
		// address := msg[3]
		identity := msg[4]
		mechanism := msg[5]
		client_key := msg[6]
		client_key_text := zmq.Z85encode(client_key)

		if version != "1.0" {
			return errors.New("version != 1.0")
		}
		if mechanism != "CURVE" {
			return errors.New("mechanism != CURVE")
		}
		if identity != "IDENT" {
			return errors.New("identity != IDENT")
		}

		if client_key_text == client_public {
			handler.SendMessage(version, sequence, "200", "OK", "anonymous", "")
		} else {
			handler.SendMessage(version, sequence, "400", "Invalid client public key", "", "")
		}
		return nil
	}

	doQuit := func(i interface{}) error {
		err := handler.Close()
		checkErr(err)
		fmt.Println("Handler closed")
		return errors.New("Quit")
	}
	quit := make(chan interface{})

	reactor := zmq.NewReactor()
	reactor.AddSocket(handler, zmq.POLLIN, doHandler)
	reactor.AddChannel(quit, 0, doQuit)
	go func() {
		reactor.Run(100 * time.Millisecond)
		fmt.Println("Reactor finished")
		quit <- true
	}()
	defer func() {
		quit <- true
		<-quit
		close(quit)
	}()

	//  Server socket will accept connections
	server, err := zmq.NewSocket(zmq.DEALER)
	if checkErr(err) {
		return
	}
	err = server.SetCurveServer(1)
	if checkErr(err) {
		return
	}
	err = server.SetCurveSecretkey(server_secret)
	if checkErr(err) {
		return
	}
	err = server.SetIdentity("IDENT")
	if checkErr(err) {
		return
	}
	server.Bind("tcp://127.0.0.1:9998")
	if checkErr(err) {
		return
	}

	err = server.SetRcvtimeo(time.Second)
	if checkErr(err) {
		return
	}

	//  Check CURVE security with valid credentials
	client, err := zmq.NewSocket(zmq.DEALER)
	if checkErr(err) {
		return
	}
	err = client.SetCurveServerkey(server_public)
	if checkErr(err) {
		return
	}
	err = client.SetCurvePublickey(client_public)
	if checkErr(err) {
		return
	}
	err = client.SetCurveSecretkey(client_secret)
	if checkErr(err) {
		return
	}
	err = client.Connect("tcp://127.0.0.1:9998")
	if checkErr(err) {
		return
	}
	bounce(server, client, false)
	err = client.Close()
	if checkErr(err) {
		return
	}

	time.Sleep(100 * time.Millisecond)

	//  Check CURVE security with a garbage server key
	//  This will be caught by the curve_server class, not passed to ZAP
	garbage_key := "0000111122223333444455556666777788889999"
	client, err = zmq.NewSocket(zmq.DEALER)
	if checkErr(err) {
		return
	}
	err = client.SetCurveServerkey(garbage_key)
	if checkErr(err) {
		return
	}
	err = client.SetCurvePublickey(client_public)
	if checkErr(err) {
		return
	}
	err = client.SetCurveSecretkey(client_secret)
	if checkErr(err) {
		return
	}
	err = client.Connect("tcp://127.0.0.1:9998")
	if checkErr(err) {
		return
	}
	err = client.SetRcvtimeo(time.Second)
	if checkErr(err) {
		return
	}
	bounce(server, client, true)
	client.SetLinger(0)
	err = client.Close()
	if checkErr(err) {
		return
	}

	time.Sleep(100 * time.Millisecond)

	//  Check CURVE security with a garbage client secret key
	//  This will be caught by the curve_server class, not passed to ZAP
	client, err = zmq.NewSocket(zmq.DEALER)
	if checkErr(err) {
		return
	}
	err = client.SetCurveServerkey(server_public)
	if checkErr(err) {
		return
	}
	err = client.SetCurvePublickey(garbage_key)
	if checkErr(err) {
		return
	}
	err = client.SetCurveSecretkey(client_secret)
	if checkErr(err) {
		return
	}
	err = client.Connect("tcp://127.0.0.1:9998")
	if checkErr(err) {
		return
	}
	err = client.SetRcvtimeo(time.Second)
	if checkErr(err) {
		return
	}
	bounce(server, client, true)
	client.SetLinger(0)
	err = client.Close()
	if checkErr(err) {
		return
	}

	time.Sleep(100 * time.Millisecond)

	//  Check CURVE security with a garbage client secret key
	//  This will be caught by the curve_server class, not passed to ZAP
	client, err = zmq.NewSocket(zmq.DEALER)
	if checkErr(err) {
		return
	}
	err = client.SetCurveServerkey(server_public)
	if checkErr(err) {
		return
	}
	err = client.SetCurvePublickey(client_public)
	if checkErr(err) {
		return
	}
	err = client.SetCurveSecretkey(garbage_key)
	if checkErr(err) {
		return
	}
	err = client.Connect("tcp://127.0.0.1:9998")
	if checkErr(err) {
		return
	}
	err = client.SetRcvtimeo(time.Second)
	if checkErr(err) {
		return
	}
	bounce(server, client, true)
	client.SetLinger(0)
	err = client.Close()
	if checkErr(err) {
		return
	}

	time.Sleep(100 * time.Millisecond)

	//  Check CURVE security with bogus client credentials
	//  This must be caught by the ZAP handler

	bogus_public, bogus_secret, _ := zmq.NewCurveKeypair()
	client, err = zmq.NewSocket(zmq.DEALER)
	if checkErr(err) {
		return
	}
	err = client.SetCurveServerkey(server_public)
	if checkErr(err) {
		return
	}
	err = client.SetCurvePublickey(bogus_public)
	if checkErr(err) {
		return
	}
	err = client.SetCurveSecretkey(bogus_secret)
	if checkErr(err) {
		return
	}
	err = client.Connect("tcp://127.0.0.1:9998")
	if checkErr(err) {
		return
	}
	err = client.SetRcvtimeo(time.Second)
	if checkErr(err) {
		return
	}
	bounce(server, client, true)
	client.SetLinger(0)
	err = client.Close()
	if checkErr(err) {
		return
	}

	//  Shutdown
	err = server.Close()
	checkErr(err)

	fmt.Println("Done")
	// Output:
	// 5 error
	// 5 error
	// 5 error
	// 5 error
	// Done
	// Handler closed
	// Reactor finished
}

func Example_test_security_null() {

	time.Sleep(100 * time.Millisecond)

	handler, err := zmq.NewSocket(zmq.REP)
	if checkErr(err) {
		return
	}
	err = handler.Bind("inproc://zeromq.zap.01")
	if checkErr(err) {
		return
	}

	doHandler := func(state zmq.State) error {
		msg, err := handler.RecvMessage(0)
		if err != nil {
			return err //  Terminating
		}
		version := msg[0]
		sequence := msg[1]
		domain := msg[2]
		// address := msg[3]
		// identity := msg[4]
		mechanism := msg[5]

		if version != "1.0" {
			return errors.New("version != 1.0")
		}
		if mechanism != "NULL" {
			return errors.New("mechanism != NULL")
		}

		if domain == "TEST" {
			handler.SendMessage(version, sequence, "200", "OK", "anonymous", "")
		} else {
			handler.SendMessage(version, sequence, "400", "BAD DOMAIN", "", "")
		}
		return nil
	}

	doQuit := func(i interface{}) error {
		err := handler.Close()
		checkErr(err)
		fmt.Println("Handler closed")
		return errors.New("Quit")
	}
	quit := make(chan interface{})

	reactor := zmq.NewReactor()
	reactor.AddSocket(handler, zmq.POLLIN, doHandler)
	reactor.AddChannel(quit, 0, doQuit)
	go func() {
		reactor.Run(100 * time.Millisecond)
		fmt.Println("Reactor finished")
		quit <- true
	}()
	defer func() {
		quit <- true
		<-quit
		close(quit)
	}()

	//  We bounce between a binding server and a connecting client
	server, err := zmq.NewSocket(zmq.DEALER)
	if checkErr(err) {
		return
	}
	client, err := zmq.NewSocket(zmq.DEALER)
	if checkErr(err) {
		return
	}

	//  We first test client/server with no ZAP domain
	//  Libzmq does not call our ZAP handler, the connect must succeed
	err = server.Bind("tcp://127.0.0.1:9683")
	if checkErr(err) {
		return
	}
	err = client.Connect("tcp://127.0.0.1:9683")
	if checkErr(err) {
		return
	}
	bounce(server, client, false)
	server.Unbind("tcp://127.0.0.1:9683")
	client.Disconnect("tcp://127.0.0.1:9683")

	//  Now define a ZAP domain for the server; this enables
	//  authentication. We're using the wrong domain so this test
	//  must fail.
	err = server.SetZapDomain("WRONG")
	if checkErr(err) {
		return
	}
	err = server.Bind("tcp://127.0.0.1:9687")
	if checkErr(err) {
		return
	}
	err = client.Connect("tcp://127.0.0.1:9687")
	if checkErr(err) {
		return
	}
	err = client.SetRcvtimeo(time.Second)
	if checkErr(err) {
		return
	}
	err = server.SetRcvtimeo(time.Second)
	if checkErr(err) {
		return
	}
	bounce(server, client, true)
	server.Unbind("tcp://127.0.0.1:9687")
	client.Disconnect("tcp://127.0.0.1:9687")

	//  Now use the right domain, the test must pass
	err = server.SetZapDomain("TEST")
	if checkErr(err) {
		return
	}
	err = server.Bind("tcp://127.0.0.1:9688")
	if checkErr(err) {
		return
	}
	err = client.Connect("tcp://127.0.0.1:9688")
	if checkErr(err) {
		return
	}
	bounce(server, client, false)
	server.Unbind("tcp://127.0.0.1:9688")
	client.Disconnect("tcp://127.0.0.1:9688")

	err = client.Close()
	checkErr(err)
	err = server.Close()
	checkErr(err)

	fmt.Println("Done")
	// Output:
	// 5 error
	// Done
	// Handler closed
	// Reactor finished
}

func Example_test_security_plain() {

	time.Sleep(100 * time.Millisecond)

	handler, err := zmq.NewSocket(zmq.REP)
	if checkErr(err) {
		return
	}
	err = handler.Bind("inproc://zeromq.zap.01")
	if checkErr(err) {
		return
	}

	doHandler := func(state zmq.State) error {
		msg, err := handler.RecvMessage(0)
		if err != nil {
			return err //  Terminating
		}
		version := msg[0]
		sequence := msg[1]
		// domain := msg[2]
		// address := msg[3]
		identity := msg[4]
		mechanism := msg[5]
		username := msg[6]
		password := msg[7]

		if version != "1.0" {
			return errors.New("version != 1.0")
		}
		if mechanism != "PLAIN" {
			return errors.New("mechanism != PLAIN")
		}
		if identity != "IDENT" {
			return errors.New("identity != IDENT")
		}

		if username == "admin" && password == "password" {
			handler.SendMessage(version, sequence, "200", "OK", "anonymous", "")
		} else {
			handler.SendMessage(version, sequence, "400", "Invalid username or password", "", "")
		}
		return nil
	}

	doQuit := func(i interface{}) error {
		err := handler.Close()
		checkErr(err)
		fmt.Println("Handler closed")
		return errors.New("Quit")
	}
	quit := make(chan interface{})

	reactor := zmq.NewReactor()
	reactor.AddSocket(handler, zmq.POLLIN, doHandler)
	reactor.AddChannel(quit, 0, doQuit)
	go func() {
		reactor.Run(100 * time.Millisecond)
		fmt.Println("Reactor finished")
		quit <- true
	}()
	defer func() {
		quit <- true
		<-quit
		close(quit)
	}()

	//  Server socket will accept connections
	server, err := zmq.NewSocket(zmq.DEALER)
	if checkErr(err) {
		return
	}
	err = server.SetIdentity("IDENT")
	if checkErr(err) {
		return
	}
	err = server.SetPlainServer(1)
	if checkErr(err) {
		return
	}
	err = server.Bind("tcp://127.0.0.1:9998")
	if checkErr(err) {
		return
	}

	//  Check PLAIN security with correct username/password
	client, err := zmq.NewSocket(zmq.DEALER)
	if checkErr(err) {
		return
	}
	err = client.SetPlainUsername("admin")
	if checkErr(err) {
		return
	}
	err = client.SetPlainPassword("password")
	if checkErr(err) {
		return
	}
	err = client.Connect("tcp://127.0.0.1:9998")
	if checkErr(err) {
		return
	}
	bounce(server, client, false)
	err = client.Close()
	if checkErr(err) {
		return
	}

	//  Check PLAIN security with badly configured client (as_server)
	//  This will be caught by the plain_server class, not passed to ZAP
	client, err = zmq.NewSocket(zmq.DEALER)
	if checkErr(err) {
		return
	}
	client.SetPlainServer(1)
	if checkErr(err) {
		return
	}
	err = client.Connect("tcp://127.0.0.1:9998")
	if checkErr(err) {
		return
	}
	err = client.SetRcvtimeo(time.Second)
	if checkErr(err) {
		return
	}
	err = server.SetRcvtimeo(time.Second)
	if checkErr(err) {
		return
	}
	bounce(server, client, true)
	client.SetLinger(0)
	err = client.Close()
	if checkErr(err) {
		return
	}

	err = server.Close()
	checkErr(err)

	fmt.Println("Done")
	// Output:
	// 5 error
	// Done
	// Handler closed
	// Reactor finished

}

/*

func Example_test_shutdown_stress() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_spec_dealer() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_spec_pushpull() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_spec_rep() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_spec_req() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_spec_router() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_stream() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_sub_forward() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_system() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_term_endpoint() {

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_timeo() {

	fmt.Println("Done")
	// Output:
	// Done
}

*/

func bounce(server, client *zmq.Socket, willfail bool) {

	content := "12345678ABCDEFGH12345678abcdefgh"

	//  Send message from client to server
	rc, err := client.Send(content, zmq.SNDMORE|zmq.DONTWAIT)
	if checkErr0(err, 1) {
		return
	}
	if rc != 32 {
		checkErr0(errors.New("rc != 32"), 2)
	}

	rc, err = client.Send(content, zmq.DONTWAIT)
	if checkErr0(err, 3) {
		return
	}
	if rc != 32 {
		checkErr0(errors.New("rc != 32"), 4)
	}

	//  Receive message at server side
	msg, err := server.Recv(0)
	if checkErr0(e(err, willfail), 5) {
		return
	}

	//  Check that message is still the same
	if msg != content {
		checkErr0(errors.New(fmt.Sprintf("%q != %q", msg, content)), 6)
	}

	rcvmore, err := server.GetRcvmore()
	if checkErr0(err, 7) {
		return
	}
	if !rcvmore {
		checkErr0(errors.New(fmt.Sprint("rcvmore ==", rcvmore)), 8)
		return
	}

	//  Receive message at server side
	msg, err = server.Recv(0)
	if checkErr0(err, 9) {
		return
	}

	//  Check that message is still the same
	if msg != content {
		checkErr0(errors.New(fmt.Sprintf("%q != %q", msg, content)), 10)
	}

	rcvmore, err = server.GetRcvmore()
	if checkErr0(err, 11) {
		return
	}
	if rcvmore {
		checkErr0(errors.New(fmt.Sprint("rcvmore == ", rcvmore)), 12)
		return
	}

	// The same, from server back to client

	//  Send message from server to client
	rc, err = server.Send(content, zmq.SNDMORE)
	if checkErr0(err, 13) {
		return
	}
	if rc != 32 {
		checkErr0(errors.New("rc != 32"), 14)
	}

	rc, err = server.Send(content, 0)
	if checkErr0(err, 15) {
		return
	}
	if rc != 32 {
		checkErr0(errors.New("rc != 32"), 16)
	}

	//  Receive message at client side
	msg, err = client.Recv(0)
	if checkErr0(err, 17) {
		return
	}

	//  Check that message is still the same
	if msg != content {
		checkErr0(errors.New(fmt.Sprintf("%q != %q", msg, content)), 18)
	}

	rcvmore, err = client.GetRcvmore()
	if checkErr0(err, 19) {
		return
	}
	if !rcvmore {
		checkErr0(errors.New(fmt.Sprint("rcvmore ==", rcvmore)), 20)
		return
	}

	//  Receive message at client side
	msg, err = client.Recv(0)
	if checkErr0(err, 21) {
		return
	}

	//  Check that message is still the same
	if msg != content {
		checkErr0(errors.New(fmt.Sprintf("%q != %q", msg, content)), 22)
	}

	rcvmore, err = client.GetRcvmore()
	if checkErr0(err, 23) {
		return
	}
	if rcvmore {
		checkErr0(errors.New(fmt.Sprint("rcvmore == ", rcvmore)), 24)
		return
	}

}

func bouncee(server, client *zmq.Socket) (msg string, err error) {

	content := "12345678ABCDEFGH12345678abcdefgh"

	//  Send message from client to server
	rc, err := client.Send(content, zmq.SNDMORE|zmq.DONTWAIT)
	if err != nil {
		return "client.Send SNDMORE|DONTWAIT", err
	}
	if rc != 32 {
		return "client.Send SNDMORE|DONTWAIT", err32
	}

	rc, err = client.Send(content, zmq.DONTWAIT)
	if err != nil {
		return "client.Send DONTWAIT", err
	}
	if rc != 32 {
		return "client.Send DONTWAIT", err32
	}

	//  Receive message at server side
	msg, err = server.Recv(0)
	if err != nil {
		return "server.Recv 1", err
	}

	//  Check that message is still the same
	if msg != content {
		return "server.Recv 1", errors.New(fmt.Sprintf("%q != %q", msg, content))
	}

	rcvmore, err := server.GetRcvmore()
	if err != nil {
		return "server.GetRcvmore 1", err
	}
	if !rcvmore {
		return "server.GetRcvmore 1", errors.New(fmt.Sprint("rcvmore ==", rcvmore))
	}

	//  Receive message at server side
	msg, err = server.Recv(0)
	if err != nil {
		return "server.Recv 2", err
	}

	//  Check that message is still the same
	if msg != content {
		return "server.Recv 2", errors.New(fmt.Sprintf("%q != %q", msg, content))
	}

	rcvmore, err = server.GetRcvmore()
	if err != nil {
		return "server.GetRcvmore 2", err
	}
	if rcvmore {
		return "server.GetRcvmore 2", errors.New(fmt.Sprint("rcvmore == ", rcvmore))
	}

	// The same, from server back to client

	//  Send message from server to client
	rc, err = server.Send(content, zmq.SNDMORE)
	if err != nil {
		return "server.Send SNDMORE", err
	}
	if rc != 32 {
		return "server.Send SNDMORE", err32
	}

	rc, err = server.Send(content, 0)
	if err != nil {
		return "server.Send 0", err
	}
	if rc != 32 {
		return "server.Send 0", err32
	}

	//  Receive message at client side
	msg, err = client.Recv(0)
	if err != nil {
		return "client.Recv 1", err
	}

	//  Check that message is still the same
	if msg != content {
		return "client.Recv 1", errors.New(fmt.Sprintf("%q != %q", msg, content))
	}

	rcvmore, err = client.GetRcvmore()
	if err != nil {
		return "client.GetRcvmore 1", err
	}
	if !rcvmore {
		return "client.GetRcvmore 1", errors.New(fmt.Sprint("rcvmore ==", rcvmore))
	}

	//  Receive message at client side
	msg, err = client.Recv(0)
	if err != nil {
		return "client.Recv 2", err
	}

	//  Check that message is still the same
	if msg != content {
		return "client.Recv 2", errors.New(fmt.Sprintf("%q != %q", msg, content))
	}

	rcvmore, err = client.GetRcvmore()
	if err != nil {
		return "client.GetRcvmore 2", err
	}
	if rcvmore {
		return "client.GetRcvmore 2", errors.New(fmt.Sprint("rcvmore == ", rcvmore))
	}
	return "OK", nil
}

func checkErr0(err error, num int) bool {
	if err != nil {
		fmt.Println(num, err)
		return true
	}
	return false
}

func checkErr(err error) bool {
	if err != nil {
		_, filename, lineno, ok := runtime.Caller(1)
		if ok {
			fmt.Printf("%v:%v: %v\n", filename, lineno, err)
		} else {
			fmt.Println(err)
		}
		return true
	}
	return false
}

func e(err error, willfail bool) error {
	if err == nil || willfail == false {
		return err
	}
	return errerr
}

func arrayEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
