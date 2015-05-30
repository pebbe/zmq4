package zmq4_test

import (
	zmq "github.com/pebbe/zmq4"

	"errors"
	"fmt"
	"runtime"
	"strconv"
	"time"
)

func Example_test_version() {
	major, _, _ := zmq.Version()
	fmt.Println("Version:", major)
	// Output:
	// Version: 4
}

func Example_multiple_contexts() {
	chQuit := make(chan interface{})
	chReactor := make(chan bool)

	addr1 := "tcp://127.0.0.1:9997"
	addr2 := "tcp://127.0.0.1:9998"

	serv_ctx1, err := zmq.NewContext()
	if checkErr(err) {
		return
	}
	serv1, err := serv_ctx1.NewSocket(zmq.REP)
	if checkErr(err) {
		return
	}
	err = serv1.Bind(addr1)
	if checkErr(err) {
		return
	}
	defer func() {
		serv1.Close()
		serv_ctx1.Term()
	}()

	serv_ctx2, err := zmq.NewContext()
	if checkErr(err) {
		return
	}
	serv2, err := serv_ctx2.NewSocket(zmq.REP)
	if checkErr(err) {
		return
	}
	err = serv2.Bind(addr2)
	if checkErr(err) {
		return
	}
	defer func() {
		serv2.Close()
		serv_ctx2.Term()
	}()

	new_service := func(sock *zmq.Socket, addr string) {
		socket_handler := func(state zmq.State) error {
			msg, err := sock.RecvMessage(0)
			if checkErr(err) {
				return err
			}
			_, err = sock.SendMessage(addr, msg)
			if checkErr(err) {
				return err
			}
			return nil
		}
		quit_handler := func(interface{}) error {
			return errors.New("quit")
		}

		defer func() {
			chReactor <- true
		}()

		reactor := zmq.NewReactor()
		reactor.AddSocket(sock, zmq.POLLIN, socket_handler)
		reactor.AddChannel(chQuit, 1, quit_handler)
		err = reactor.Run(100 * time.Millisecond)
		fmt.Println(err)
	}

	go new_service(serv1, addr1)
	go new_service(serv2, addr2)

	time.Sleep(time.Second)

	// default context

	sock1, err := zmq.NewSocket(zmq.REQ)
	if checkErr(err) {
		return
	}
	sock2, err := zmq.NewSocket(zmq.REQ)
	if checkErr(err) {
		return
	}
	err = sock1.Connect(addr1)
	if checkErr(err) {
		return
	}
	err = sock2.Connect(addr2)
	if checkErr(err) {
		return
	}
	_, err = sock1.SendMessage(addr1)
	if checkErr(err) {
		return
	}
	_, err = sock2.SendMessage(addr2)
	if checkErr(err) {
		return
	}
	msg, err := sock1.RecvMessage(0)
	fmt.Println(err, msg)
	msg, err = sock2.RecvMessage(0)
	fmt.Println(err, msg)
	err = sock1.Close()
	if checkErr(err) {
		return
	}
	err = sock2.Close()
	if checkErr(err) {
		return
	}

	// non-default contexts

	ctx1, err := zmq.NewContext()
	if checkErr(err) {
		return
	}
	ctx2, err := zmq.NewContext()
	if checkErr(err) {
		return
	}
	sock1, err = ctx1.NewSocket(zmq.REQ)
	if checkErr(err) {
		return
	}
	sock2, err = ctx2.NewSocket(zmq.REQ)
	if checkErr(err) {
		return
	}
	err = sock1.Connect(addr1)
	if checkErr(err) {
		return
	}
	err = sock2.Connect(addr2)
	if checkErr(err) {
		return
	}
	_, err = sock1.SendMessage(addr1)
	if checkErr(err) {
		return
	}
	_, err = sock2.SendMessage(addr2)
	if checkErr(err) {
		return
	}
	msg, err = sock1.RecvMessage(0)
	fmt.Println(err, msg)
	msg, err = sock2.RecvMessage(0)
	fmt.Println(err, msg)
	err = sock1.Close()
	if checkErr(err) {
		return
	}
	err = sock2.Close()
	if checkErr(err) {
		return
	}

	err = ctx1.Term()
	if checkErr(err) {
		return
	}
	err = ctx2.Term()
	if checkErr(err) {
		return
	}

	// close(chQuit) doesn't work because the reactor removes closed channels, instead of acting on them
	chQuit <- true
	<-chReactor
	chQuit <- true
	<-chReactor

	fmt.Println("Done")
	// Output:
	// <nil> [tcp://127.0.0.1:9997 tcp://127.0.0.1:9997]
	// <nil> [tcp://127.0.0.1:9998 tcp://127.0.0.1:9998]
	// <nil> [tcp://127.0.0.1:9997 tcp://127.0.0.1:9997]
	// <nil> [tcp://127.0.0.1:9998 tcp://127.0.0.1:9998]
	// quit
	// quit
	// Done
}

func Example_test_abstract_ipc() {

	addr := "ipc://@/tmp/tester"

	// This is Linux only
	if runtime.GOOS != "linux" {
		fmt.Printf("%q\n", addr)
		fmt.Println("Done")
		return
	}

	sb, err := zmq.NewSocket(zmq.PAIR)
	if checkErr(err) {
		return
	}

	err = sb.Bind(addr)
	if checkErr(err) {
		return
	}

	endpoint, err := sb.GetLastEndpoint()
	if checkErr(err) {
		return
	}
	fmt.Printf("%q\n", endpoint)

	sc, err := zmq.NewSocket(zmq.PAIR)
	if checkErr(err) {
		return
	}
	err = sc.Connect(addr)
	if checkErr(err) {
		return
	}

	bounce(sb, sc)

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
	// "ipc://@/tmp/tester"
	// Done
}

func Example_test_conflate() {

	bind_to := "tcp://127.0.0.1:5555"

	err := zmq.SetIoThreads(1)
	if checkErr(err) {
		return
	}

	s_in, err := zmq.NewSocket(zmq.PULL)
	if checkErr(err) {
		return
	}

	err = s_in.SetConflate(true)
	if checkErr(err) {
		return
	}

	err = s_in.Bind(bind_to)
	if checkErr(err) {
		return
	}

	s_out, err := zmq.NewSocket(zmq.PUSH)
	if checkErr(err) {
		return
	}

	err = s_out.Connect(bind_to)
	if checkErr(err) {
		return
	}

	message_count := 20

	for j := 0; j < message_count; j++ {
		_, err = s_out.Send(fmt.Sprint(j), 0)
		if checkErr(err) {
			return
		}
	}

	time.Sleep(time.Second)

	payload_recved, err := s_in.Recv(0)
	if checkErr(err) {
		return
	}
	i, err := strconv.Atoi(payload_recved)
	if checkErr(err) {
		return
	}
	if i != message_count-1 {
		checkErr(errors.New("payload_recved != message_count - 1"))
		return
	}

	err = s_in.Close()
	if checkErr(err) {
		return
	}

	err = s_out.Close()
	if checkErr(err) {
		return
	}

	fmt.Println("Done")
	// Output:
	// Done
}

func Example_test_connect_resolve() {

	sock, err := zmq.NewSocket(zmq.PUB)
	if checkErr(err) {
		return
	}

	err = sock.Connect("tcp://localhost:1234")
	checkErr(err)

	err = sock.Connect("tcp://localhost:invalid")
	fmt.Println(err)

	err = sock.Connect("tcp://in val id:1234")
	fmt.Println(err)

	err = sock.Connect("invalid://localhost:1234")
	fmt.Println(err)

	err = sock.Close()
	checkErr(err)

	fmt.Println("Done")
	// Output:
	// invalid argument
	// invalid argument
	// protocol not supported
	// Done
}

/*

func Example_test_ctx_destroy() {

	fmt.Println("Done")
	// Output:
	// Done
}

*/

func Example_test_ctx_options() {

	i, err := zmq.GetMaxSockets()
	fmt.Println(i == zmq.MaxSocketsDflt, err)
	i, err = zmq.GetIoThreads()
	fmt.Println(i == zmq.IoThreadsDflt, err)
	b, err := zmq.GetIpv6()
	fmt.Println(b, err)

	zmq.SetIpv6(true)
	b, err = zmq.GetIpv6()
	fmt.Println(b, err)

	router, _ := zmq.NewSocket(zmq.ROUTER)
	b, err = router.GetIpv6()
	fmt.Println(b, err)

	fmt.Println(router.Close())

	zmq.SetIpv6(false)

	fmt.Println("Done")
	// Output:
	// true <nil>
	// true <nil>
	// false <nil>
	// true <nil>
	// true <nil>
	// <nil>
	// Done
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
				fmt.Println("Send:", err)
				break
			}
			send_count++
		}

		// Now receive all sent messages
		recv_count := 0
		for {
			_, err := bind_socket.Recv(zmq.DONTWAIT)
			if err != nil {
				fmt.Println("Recv:", err)
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
				fmt.Println("Send:", err)
				break
			}
			send_count++
		}

		// Now receive all sent messages
		recv_count := 0
		for {
			_, err := bind_socket.Recv(zmq.DONTWAIT)
			if err != nil {
				fmt.Println("Recv:", err)
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
				fmt.Println("Send:", err)
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
				fmt.Println("Recv:", err)
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
	// Send: resource temporarily unavailable
	// Recv: resource temporarily unavailable
	// send_count == recv_count: true
	// count: 2000
	//
	// Infinite send and receive
	// Recv: resource temporarily unavailable
	// send_count == recv_count: true
	// count: 10000
	// Recv: resource temporarily unavailable
	// send_count == recv_count: true
	// count: 10000
	//
	// Infinite send buffer
	// Recv: resource temporarily unavailable
	// send_count == recv_count: true
	// count: 10000
	// Recv: resource temporarily unavailable
	// send_count == recv_count: true
	// count: 10000
	//
	// Infinite receive buffer
	// Recv: resource temporarily unavailable
	// send_count == recv_count: true
	// count: 10000
	// Recv: resource temporarily unavailable
	// send_count == recv_count: true
	// count: 10000
	//
	// Send and recv buffers hwm 1
	// Send: resource temporarily unavailable
	// Recv: resource temporarily unavailable
	// send_count == recv_count: true
	// count: 2
	// Send: resource temporarily unavailable
	// Recv: resource temporarily unavailable
	// send_count == recv_count: true
	// count: 2
	//
	// Send hwm of 1, send before bind
	// Send: resource temporarily unavailable
	// Recv: resource temporarily unavailable
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

	bounce(sb, sc)

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

	bounce(sb, sc)

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
	bounce(server, client)
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
	bounce(server, client)
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
	bounce(server, client)
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
	bounce(server, client)
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
	bounce(server, client)
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
	// resource temporarily unavailable
	// resource temporarily unavailable
	// resource temporarily unavailable
	// resource temporarily unavailable
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
	bounce(server, client)
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
	bounce(server, client)
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
	bounce(server, client)
	server.Unbind("tcp://127.0.0.1:9688")
	client.Disconnect("tcp://127.0.0.1:9688")

	err = client.Close()
	checkErr(err)
	err = server.Close()
	checkErr(err)

	fmt.Println("Done")
	// Output:
	// resource temporarily unavailable
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
	bounce(server, client)
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
	bounce(server, client)
	client.SetLinger(0)
	err = client.Close()
	if checkErr(err) {
		return
	}

	err = server.Close()
	checkErr(err)

	fmt.Println("Done")
	// Output:
	// resource temporarily unavailable
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

func bounce(server, client *zmq.Socket) {

	content := "12345678ABCDEFGH12345678abcdefgh"

	//  Send message from client to server
	rc, err := client.Send(content, zmq.SNDMORE|zmq.DONTWAIT)
	if checkErr0(err) {
		return
	}
	if rc != 32 {
		checkErr0(errors.New("rc != 32"))
	}

	rc, err = client.Send(content, zmq.DONTWAIT)
	if checkErr0(err) {
		return
	}
	if rc != 32 {
		checkErr0(errors.New("rc != 32"))
	}

	//  Receive message at server side
	msg, err := server.Recv(0)
	if checkErr0(err) {
		return
	}

	//  Check that message is still the same
	if msg != content {
		checkErr0(errors.New(fmt.Sprintf("%q != %q", msg, content)))
	}

	rcvmore, err := server.GetRcvmore()
	if checkErr0(err) {
		return
	}
	if !rcvmore {
		checkErr0(errors.New(fmt.Sprint("rcvmore ==", rcvmore)))
		return
	}

	//  Receive message at server side
	msg, err = server.Recv(0)
	if checkErr0(err) {
		return
	}

	//  Check that message is still the same
	if msg != content {
		checkErr0(errors.New(fmt.Sprintf("%q != %q", msg, content)))
	}

	rcvmore, err = server.GetRcvmore()
	if checkErr0(err) {
		return
	}
	if rcvmore {
		checkErr0(errors.New(fmt.Sprint("rcvmore == ", rcvmore)))
		return
	}

	// The same, from server back to client

	//  Send message from server to client
	rc, err = server.Send(content, zmq.SNDMORE)
	if checkErr0(err) {
		return
	}
	if rc != 32 {
		checkErr0(errors.New("rc != 32"))
	}

	rc, err = server.Send(content, 0)
	if checkErr0(err) {
		return
	}
	if rc != 32 {
		checkErr0(errors.New("rc != 32"))
	}

	//  Receive message at client side
	msg, err = client.Recv(0)
	if checkErr0(err) {
		return
	}

	//  Check that message is still the same
	if msg != content {
		checkErr0(errors.New(fmt.Sprintf("%q != %q", msg, content)))
	}

	rcvmore, err = client.GetRcvmore()
	if checkErr0(err) {
		return
	}
	if !rcvmore {
		checkErr0(errors.New(fmt.Sprint("rcvmore ==", rcvmore)))
		return
	}

	//  Receive message at client side
	msg, err = client.Recv(0)
	if checkErr0(err) {
		return
	}

	//  Check that message is still the same
	if msg != content {
		checkErr0(errors.New(fmt.Sprintf("%q != %q", msg, content)))
	}

	rcvmore, err = client.GetRcvmore()
	if checkErr0(err) {
		return
	}
	if rcvmore {
		checkErr0(errors.New(fmt.Sprint("rcvmore == ", rcvmore)))
		return
	}

}

func checkErr0(err error) bool {
	if err != nil {
		fmt.Println(err)
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
