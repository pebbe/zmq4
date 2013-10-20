package zmq4_test

import (
	zmq "github.com/pebbe/zmq4"

	"errors"
	"fmt"
	"runtime"
	"strconv"
	"time"
)

func Example_test_abstract_ipc() {
	sb, err := zmq.NewSocket(zmq.PAIR)
	if checkErr(err) {
		return
	}

	err = sb.Bind("ipc://@/tmp/tester")
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
	err = sc.Connect("ipc://@/tmp/tester")
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

	// Output:
	// "ipc://@/tmp/tester"
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
	if i != message_count - 1 {
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

	// Output:
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

	// Output:
	// invalid argument
	// invalid argument
	// protocol not supported
}


func Example_test_ctx_destroy() {


	// Output:
}


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

	// Output:
	// true <nil>
	// true <nil>
	// false <nil>
	// true <nil>
	// true <nil>
	// <nil>
}


func Example_test_disconnect_inproc() {


	// Output:
}


func Example_test_fork() {


	// Output:
}


func Example_test_hwm() {


	// Output:
}


func Example_test_immediate() {


	// Output:
}


func Example_test_inproc_connect() {


	// Output:
}


func Example_test_invalid_rep() {


	// Output:
}


func Example_test_iov() {


	// Output:
}


func Example_test_issue_566() {


	// Output:
}


func Example_test_last_endpoint() {


	// Output:
}


func Example_test_linger() {


	// Output:
}


func Example_test_monitor() {


	// Output:
}


func Example_test_msg_flags() {


	// Output:
}


func Example_test_pair_inproc() {


	// Output:
}


func Example_test_pair_ipc() {


	// Output:
}


func Example_test_pair_tcp() {


	// Output:
}


func Example_test_probe_router() {


	// Output:
}


func Example_test_req_correlate() {


	// Output:
}


func Example_test_req_relaxed() {


	// Output:
}


func Example_test_reqrep_device() {


	// Output:
}


func Example_test_reqrep_inproc() {


	// Output:
}


func Example_test_reqrep_ipc() {


	// Output:
}


func Example_test_reqrep_tcp() {


	// Output:
}


func Example_test_router_mandatory() {


	// Output:
}


func Example_test_security_curve() {


	// Output:
}


func Example_test_security_null() {


	// Output:
}


func Example_test_security_plain() {


	// Output:
}


func Example_test_shutdown_stress() {


	// Output:
}


func Example_test_spec_dealer() {


	// Output:
}


func Example_test_spec_pushpull() {


	// Output:
}


func Example_test_spec_rep() {


	// Output:
}


func Example_test_spec_req() {


	// Output:
}


func Example_test_spec_router() {


	// Output:
}


func Example_test_stream() {


	// Output:
}


func Example_test_sub_forward() {


	// Output:
}


func Example_test_system() {


	// Output:
}


func Example_test_term_endpoint() {


	// Output:
}


func Example_test_timeo() {


	// Output:
}



func bounce(server, client *zmq.Socket) {

    content := "12345678ABCDEFGH12345678abcdefgh"

    //  Send message from client to server
	rc, err := client.Send(content, zmq.SNDMORE)
	if checkErr(err) {
		return
	}
	if rc != 32 {
		checkErr(errors.New("rc != 32"))
	}

	rc, err = client.Send(content, 0)
	if checkErr(err) {
		return
	}
	if rc != 32 {
		checkErr(errors.New("rc != 32"))
	}

    //  Receive message at server side
	msg, err := server.Recv(0)
	if checkErr(err) {
		return
	}

    //  Check that message is still the same
	if msg != content {
		checkErr(errors.New(fmt.Sprintf("%q != %q", msg, content)))
	}

	rcvmore, err := server.GetRcvmore()
	if checkErr(err) {
		return
	}
	if ! rcvmore {
		checkErr(errors.New(fmt.Sprint("rcvmore ==", rcvmore)))
		return
	}

    //  Receive message at server side
	msg, err = server.Recv(0)
	if checkErr(err) {
		return
	}

    //  Check that message is still the same
	if msg != content {
		checkErr(errors.New(fmt.Sprintf("%q != %q", msg, content)))
	}

	rcvmore, err = server.GetRcvmore()
	if checkErr(err) {
		return
	}
	if rcvmore {
		checkErr(errors.New(fmt.Sprint("rcvmore == ", rcvmore)))
		return
	}


	// The same, from server back to client

    //  Send message from server to client
	rc, err = server.Send(content, zmq.SNDMORE)
	if checkErr(err) {
		return
	}
	if rc != 32 {
		checkErr(errors.New("rc != 32"))
	}

	rc, err = server.Send(content, 0)
	if checkErr(err) {
		return
	}
	if rc != 32 {
		checkErr(errors.New("rc != 32"))
	}

    //  Receive message at client side
	msg, err = client.Recv(0)
	if checkErr(err) {
		return
	}

    //  Check that message is still the same
	if msg != content {
		checkErr(errors.New(fmt.Sprintf("%q != %q", msg, content)))
	}

	rcvmore, err = client.GetRcvmore()
	if checkErr(err) {
		return
	}
	if ! rcvmore {
		checkErr(errors.New(fmt.Sprint("rcvmore ==", rcvmore)))
		return
	}

    //  Receive message at client side
	msg, err = client.Recv(0)
	if checkErr(err) {
		return
	}

    //  Check that message is still the same
	if msg != content {
		checkErr(errors.New(fmt.Sprintf("%q != %q", msg, content)))
	}

	rcvmore, err = client.GetRcvmore()
	if checkErr(err) {
		return
	}
	if rcvmore {
		checkErr(errors.New(fmt.Sprint("rcvmore == ", rcvmore)))
		return
	}


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

