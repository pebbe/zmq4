package zmq4

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type ChanAdapterTestCase struct {
	// Begin of test case parameters
	socketAddr string
	rxChanSize int
	txChanSize int
	identity   []byte
	// End of test case parameters
	bindSocket     *Socket
	bindAdapter    *ChanAdapter
	cancel         func()
	connectSocket  *Socket
	connectAdapter *ChanAdapter
}

func (tc *ChanAdapterTestCase) setUp(t *testing.T, connectType, bindType Type) {
	tc.bindSocket, _ = NewSocket(bindType)
	tc.bindSocket.Bind(tc.socketAddr)
	tc.bindAdapter = NewChanAdapter(tc.bindSocket, tc.rxChanSize, tc.txChanSize)

	tc.connectSocket, _ = NewSocket(connectType)
	if tc.identity != nil {
		tc.connectSocket.SetIdentity(string(tc.identity[:]))
	}
	tc.connectSocket.Connect(tc.socketAddr)
	tc.connectAdapter = NewChanAdapter(tc.connectSocket, tc.rxChanSize, tc.txChanSize)

	ctx, cancel := context.WithCancel(context.Background())
	tc.cancel = cancel
	tc.bindAdapter.Start(ctx)
	tc.connectAdapter.Start(ctx)

	t.Cleanup(func() {
		if tc.cancel != nil {
			tc.cancel()
		}
		if tc.bindAdapter != nil {
			tc.bindAdapter.Close()
		}
		if tc.connectAdapter != nil {
			tc.connectAdapter.Close()
		}
		if tc.bindSocket != nil {
			tc.bindSocket.SetLinger(0)
			tc.bindSocket.Close()
		}
		if tc.connectSocket != nil {
			tc.connectSocket.SetLinger(0)
			tc.connectSocket.Close()
		}
	})
}

// Test cases
func TestChanAdapterReq2Rep(t *testing.T) {
	tc := ChanAdapterTestCase{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tc.setUp(t, REQ, REP)

	// Test
	requestMsg := [][]byte{[]byte("Hello, world!")}
	replyMsg := [][]byte{[]byte("Ack")}

	tc.connectAdapter.TxChan() <- requestMsg
	gotBindMsg, ok := <-tc.bindAdapter.RxChan()
	requireTrue(t, ok, "bindAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, requestMsg, gotBindMsg, "bindAdapter received wrong message")

	tc.bindAdapter.TxChan() <- replyMsg
	gotConnectMsg, ok := <-tc.connectAdapter.RxChan()
	requireTrue(t, ok, "connectAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, replyMsg, gotConnectMsg, "connectAdapter received wrong message")
}

func TestChanAdapterReq2Router(t *testing.T) {
	tc := ChanAdapterTestCase{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
		identity:   []byte("client-12345"),
	}
	tc.setUp(t, REQ, ROUTER)

	// Test
	emptyDelimiter := []byte{}
	routeHeader := [][]byte{tc.identity, emptyDelimiter}
	replyMsg := [][]byte{[]byte("Reply")}
	requestMsg := [][]byte{[]byte("Request"), []byte("Hello, world!")}
	recvRequestMsg := append(routeHeader, requestMsg...)
	sendReplyMsg := append(routeHeader, replyMsg...)

	tc.connectAdapter.TxChan() <- requestMsg
	gotBindMsg, ok := <-tc.bindAdapter.RxChan()
	requireTrue(t, ok, "bindAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, recvRequestMsg, gotBindMsg, "bindAdapter received wrong message")

	tc.bindAdapter.TxChan() <- sendReplyMsg
	gotConnectMsg, ok := <-tc.connectAdapter.RxChan()
	requireTrue(t, ok, "connectAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, replyMsg, gotConnectMsg, "connectAdapter received wrong message")
}

func TestChanAdapterDealer2Router(t *testing.T) {
	tc := ChanAdapterTestCase{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
		identity:   []byte("client-12345"),
	}
	tc.setUp(t, DEALER, ROUTER)

	// Test
	emptyDelimiter := []byte{}
	routeHeader := [][]byte{tc.identity, emptyDelimiter}
	dealerHeader := [][]byte{emptyDelimiter}
	replyMsg := [][]byte{[]byte("Reqly")}
	requestMsg := [][]byte{[]byte("Request"), []byte("Hello, world!")}
	sendRequestMsg := append(dealerHeader, requestMsg...)
	recvRequestMsg := append(routeHeader, requestMsg...)
	sendReplyMsg := append(routeHeader, replyMsg...)
	recvReplyMsg := append(dealerHeader, replyMsg...)

	tc.connectAdapter.TxChan() <- sendRequestMsg
	gotBindMsg, ok := <-tc.bindAdapter.RxChan()
	requireTrue(t, ok, "bindAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, recvRequestMsg, gotBindMsg, "bindAdapter received wrong message")

	tc.bindAdapter.TxChan() <- sendReplyMsg
	gotConnectMsg, ok := <-tc.connectAdapter.RxChan()
	requireTrue(t, ok, "connectAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, recvReplyMsg, gotConnectMsg, "connectAdapter received wrong message")
}

func TestChanAdapterReconnectDealer2Router(t *testing.T) {
	tc := ChanAdapterTestCase{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
		identity:   []byte("client-12345"),
	}
	tc.setUp(t, DEALER, ROUTER)

	// Test
	emptyDelimiter := []byte{}
	routeHeader := [][]byte{tc.identity, emptyDelimiter}
	dealerHeader := [][]byte{emptyDelimiter}
	replyMsg := [][]byte{[]byte("Reqly")}
	requestMsg := [][]byte{[]byte("Request"), []byte("Hello, world!")}
	sendDropMsg := append(dealerHeader, replyMsg...)
	sendRequestMsg := append(dealerHeader, requestMsg...)
	recvRequestMsg := append(routeHeader, requestMsg...)
	sendReplyMsg := append(routeHeader, replyMsg...)
	recvReplyMsg := append(dealerHeader, replyMsg...)

	tc.connectAdapter.TxChan() <- sendDropMsg
	tc.connectAdapter.Close()
	tc.connectSocket.SetLinger(0)
	tc.connectSocket.Close()

	time.Sleep(100 * time.Millisecond)
	rxChan := tc.bindAdapter.RxChan()
	// drop all messages in the channel
drainLoop:
	for {
		select {
		case _, ok := <-rxChan:
			if !ok {
				break drainLoop
			}
		default:
			break drainLoop
		}
	}

	var err error
	tc.connectSocket, err = NewSocket(DEALER)
	requireNoError(t, err)
	err = tc.connectSocket.SetIdentity(string(tc.identity))
	requireNoError(t, err)
	err = tc.connectSocket.Connect(tc.socketAddr)
	requireNoError(t, err)
	tc.connectAdapter = NewChanAdapter(tc.connectSocket, tc.rxChanSize, tc.txChanSize)
	tc.connectAdapter.Start(context.Background())
	defer tc.connectAdapter.Close()

	tc.connectAdapter.TxChan() <- sendRequestMsg
	gotBindMsg, ok := <-tc.bindAdapter.RxChan()
	requireTrue(t, ok, "bindAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, recvRequestMsg, gotBindMsg, "bindAdapter received wrong message")

	tc.bindAdapter.TxChan() <- sendReplyMsg
	gotConnectMsg, ok := <-tc.connectAdapter.RxChan()
	requireTrue(t, ok, "connectAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, recvReplyMsg, gotConnectMsg, "connectAdapter received wrong message")
}

func TestChanAdapterPub2Sub(t *testing.T) {
	tc := ChanAdapterTestCase{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tc.setUp(t, SUB, PUB)

	// SUB socket needs to subscribe to receive messages
	tc.connectSocket.SetSubscribe("")

	// Allow some time for subscription to propagate
	time.Sleep(50 * time.Millisecond)

	// Test
	pubMsg := [][]byte{[]byte("Hello, subscribers!")}

	tc.bindAdapter.TxChan() <- pubMsg
	gotSubMsg, ok := <-tc.connectAdapter.RxChan()
	requireTrue(t, ok, "connectAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, pubMsg, gotSubMsg, "SUB socket received wrong message")
}

func TestChanAdapterPush2Pull(t *testing.T) {
	tc := ChanAdapterTestCase{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tc.setUp(t, PUSH, PULL)

	// Test
	pushMsg := [][]byte{[]byte("Work item"), []byte("data")}

	tc.connectAdapter.TxChan() <- pushMsg
	gotPullMsg, ok := <-tc.bindAdapter.RxChan()
	requireTrue(t, ok, "bindAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, pushMsg, gotPullMsg, "PULL socket received wrong message")
}

func TestChanAdapterPair2Pair(t *testing.T) {
	tc := ChanAdapterTestCase{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tc.setUp(t, PAIR, PAIR)

	// Test bidirectional communication
	msg1 := [][]byte{[]byte("Message from connect")}
	msg2 := [][]byte{[]byte("Message from bind")}

	tc.connectAdapter.TxChan() <- msg1
	gotBindMsg, ok := <-tc.bindAdapter.RxChan()
	requireTrue(t, ok, "bindAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, msg1, gotBindMsg, "PAIR bind received wrong message")

	tc.bindAdapter.TxChan() <- msg2
	gotConnectMsg, ok := <-tc.connectAdapter.RxChan()
	requireTrue(t, ok, "connectAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, msg2, gotConnectMsg, "PAIR connect received wrong message")
}

func TestChanAdapterEmptyMessage(t *testing.T) {
	tc := ChanAdapterTestCase{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tc.setUp(t, REQ, REP)

	// Test empty message parts
	emptyMsg := [][]byte{[]byte{}}

	tc.connectAdapter.TxChan() <- emptyMsg
	gotBindMsg, ok := <-tc.bindAdapter.RxChan()
	requireTrue(t, ok, "bindAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, emptyMsg, gotBindMsg, "bindAdapter received wrong empty message")

	tc.bindAdapter.TxChan() <- emptyMsg
	gotConnectMsg, ok := <-tc.connectAdapter.RxChan()
	requireTrue(t, ok, "connectAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, emptyMsg, gotConnectMsg, "connectAdapter received wrong empty message")
}

func TestChanAdapterLargeMessage(t *testing.T) {
	tc := ChanAdapterTestCase{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tc.setUp(t, REQ, REP)

	// Test large message (1MB)
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}
	largeMsg := [][]byte{largeData}

	tc.connectAdapter.TxChan() <- largeMsg
	gotBindMsg, ok := <-tc.bindAdapter.RxChan()
	requireTrue(t, ok, "bindAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, largeMsg, gotBindMsg, "bindAdapter received wrong large message")

	ackMsg := [][]byte{[]byte("ACK")}
	tc.bindAdapter.TxChan() <- ackMsg
	gotConnectMsg, ok := <-tc.connectAdapter.RxChan()
	requireTrue(t, ok, "connectAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, ackMsg, gotConnectMsg, "connectAdapter received wrong ack message")
}

func TestChanAdapterMultipartMessages(t *testing.T) {
	tc := ChanAdapterTestCase{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tc.setUp(t, REQ, REP)

	// Test message with multiple parts
	multipartMsg := [][]byte{
		[]byte("part1"),
		[]byte("part2"),
		[]byte("part3"),
		[]byte("part4"),
	}

	tc.connectAdapter.TxChan() <- multipartMsg
	gotBindMsg, ok := <-tc.bindAdapter.RxChan()
	requireTrue(t, ok, "bindAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, multipartMsg, gotBindMsg, "bindAdapter received wrong multipart message")

	replyMsg := [][]byte{[]byte("reply"), []byte("with"), []byte("multiple"), []byte("parts")}
	tc.bindAdapter.TxChan() <- replyMsg
	gotConnectMsg, ok := <-tc.connectAdapter.RxChan()
	requireTrue(t, ok, "connectAdapter.RxChan() should return a message")
	assertByteArrayEqual(t, replyMsg, gotConnectMsg, "connectAdapter received wrong multipart reply")
}

func TestChanAdapterCloseBeforeStart(t *testing.T) {
	socketAddr := fmt.Sprintf("inproc://test-%d", time.Now().UnixNano())

	bindSocket, _ := NewSocket(REP)
	bindSocket.Bind(socketAddr)
	adapter := NewChanAdapter(bindSocket, 10, 10)

	// Close before starting - should not panic
	adapter.Close()

	// Don't call Start after Close as it may cause issues
	// The test is to verify that Close before Start doesn't cause problems

	// Final cleanup
	bindSocket.SetLinger(0)
	bindSocket.Close()
}

func TestChanAdapterChannelSizes(t *testing.T) {
	// Test with different channel buffer sizes
	testCases := []struct {
		rxSize, txSize int
	}{
		{0, 0},    // Unbuffered
		{1, 1},    // Minimal buffer
		{100, 50}, // Different sizes
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("rx:%d,tx:%d", tc.rxSize, tc.txSize), func(t *testing.T) {
			testCase := ChanAdapterTestCase{
				socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
				rxChanSize: tc.rxSize,
				txChanSize: tc.txSize,
			}
			testCase.setUp(t, REQ, REP)

			// Basic functionality test
			testMsg := [][]byte{[]byte("test message")}
			testCase.connectAdapter.TxChan() <- testMsg
			gotMsg, ok := <-testCase.bindAdapter.RxChan()
			requireTrue(t, ok, "Should receive message")
			assertByteArrayEqual(t, testMsg, gotMsg, "Message should match")
		})
	}
}

// Helper functions to replace testify assertions

func assertByteArrayEqual(t *testing.T, expected, actual [][]byte, msg string) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Errorf("%s: length mismatch: expected %d, got %d", msg, len(expected), len(actual))
		return
	}
	for i := range expected {
		if len(expected[i]) != len(actual[i]) {
			t.Errorf("%s: length mismatch at index %d: expected %d, got %d", msg, i, len(expected[i]), len(actual[i]))
			return
		}
		for j := range expected[i] {
			if expected[i][j] != actual[i][j] {
				t.Errorf("%s: mismatch at index [%d][%d]: expected %d, got %d", msg, i, j, expected[i][j], actual[i][j])
				return
			}
		}
	}
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func requireTrue(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Fatalf("condition not met: %s", msg)
	}
}

func TestChanAdapterConcurrentAccess(t *testing.T) {
	tc := ChanAdapterTestCase{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 100,
		txChanSize: 100,
	}
	tc.setUp(t, DEALER, ROUTER)

	const numGoroutines = 10
	const messagesPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // senders + receivers

	// Multiple goroutines sending messages
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				msg := [][]byte{[]byte{}, []byte(fmt.Sprintf("msg-%d-%d", id, j))}
				tc.connectAdapter.TxChan() <- msg
			}
		}(i)
	}

	// Multiple goroutines receiving messages
	received := make(chan [][]byte, numGoroutines*messagesPerGoroutine)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				select {
				case msg := <-tc.bindAdapter.RxChan():
					received <- msg
				case <-time.After(5 * time.Second):
					t.Errorf("Timeout waiting for message")
					return
				}
			}
		}()
	}

	wg.Wait()
	close(received)

	// Verify we received all messages
	count := 0
	for range received {
		count++
	}

	if count != numGoroutines*messagesPerGoroutine {
		t.Errorf("Expected %d messages, got %d", numGoroutines*messagesPerGoroutine, count)
	}
}

func TestChanAdapterHighVolumeMessages(t *testing.T) {
	tc := ChanAdapterTestCase{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 1000,
		txChanSize: 1000,
	}
	tc.setUp(t, PUSH, PULL)

	const numMessages = 1000
	done := make(chan bool)

	// Receiver goroutine
	go func() {
		for i := 0; i < numMessages; i++ {
			select {
			case msg := <-tc.bindAdapter.RxChan():
				expected := fmt.Sprintf("message-%d", i)
				if string(msg[0]) != expected {
					t.Errorf("Expected %s, got %s", expected, string(msg[0]))
				}
			case <-time.After(5 * time.Second):
				t.Errorf("Timeout waiting for message %d", i)
				return
			}
		}
		done <- true
	}()

	// Send messages rapidly
	for i := 0; i < numMessages; i++ {
		msg := [][]byte{[]byte(fmt.Sprintf("message-%d", i))}
		tc.connectAdapter.TxChan() <- msg
	}

	// Wait for completion
	select {
	case <-done:
		// Success
	case <-time.After(10 * time.Second):
		t.Error("Test timed out")
	}
}

func TestChanAdapterBackpressure(t *testing.T) {
	// Small buffer to test backpressure
	tc := ChanAdapterTestCase{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 2, // Very small buffer
		txChanSize: 2,
	}
	// Use PUSH-PULL instead of REQ-REP to avoid strict alternating requirements
	tc.setUp(t, PUSH, PULL)

	// Send multiple messages without reading them
	// This should test backpressure handling
	sent := 0
sendLoop:
	for i := 0; i < 5; i++ {
		msg := [][]byte{[]byte(fmt.Sprintf("msg-%d", i))}
		select {
		case tc.connectAdapter.TxChan() <- msg:
			sent++
		case <-time.After(100 * time.Millisecond):
			// This is expected due to backpressure
			break sendLoop
		}
	}

	// Now start reading messages
	for i := 0; i < sent; i++ {
		select {
		case <-tc.bindAdapter.RxChan():
			// Good, received a message
		case <-time.After(1 * time.Second):
			t.Errorf("Timeout receiving message %d", i)
		}
	}
}

func TestChanAdapterErrorRecovery(t *testing.T) {
	tc := ChanAdapterTestCase{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tc.setUp(t, REQ, REP)

	// Send a normal message first
	normalMsg := [][]byte{[]byte("normal")}
	tc.connectAdapter.TxChan() <- normalMsg

	_, ok := <-tc.bindAdapter.RxChan()
	requireTrue(t, ok, "Should receive normal message")

	// Force close the underlying socket to simulate an error
	tc.connectSocket.Close()

	// Give some time for the error to propagate
	time.Sleep(100 * time.Millisecond)

	// The adapter should handle this gracefully with no panic or deadlock
}

func TestChanAdapterMessageOrder(t *testing.T) {
	tc := ChanAdapterTestCase{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 100,
		txChanSize: 100,
	}
	tc.setUp(t, PUSH, PULL)

	const numMessages = 100

	// Send messages in order
	go func() {
		for i := 0; i < numMessages; i++ {
			msg := [][]byte{[]byte(fmt.Sprintf("%d", i))}
			tc.connectAdapter.TxChan() <- msg
		}
	}()

	// Receive messages and verify order
	for i := 0; i < numMessages; i++ {
		select {
		case msg := <-tc.bindAdapter.RxChan():
			expected := fmt.Sprintf("%d", i)
			if string(msg[0]) != expected {
				t.Errorf("Message order violated: expected %s, got %s", expected, string(msg[0]))
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("Timeout waiting for message %d", i)
		}
	}
}

func TestChanAdapterZeroByteMessage(t *testing.T) {
	tc := ChanAdapterTestCase{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tc.setUp(t, REQ, REP)

	// Send message with zero-length parts
	zeroMsg := [][]byte{[]byte{}, []byte{}, []byte("actual-data")}
	tc.connectAdapter.TxChan() <- zeroMsg

	gotMsg, ok := <-tc.bindAdapter.RxChan()
	requireTrue(t, ok, "Should receive zero-byte message")
	assertByteArrayEqual(t, zeroMsg, gotMsg, "Zero-byte message should be preserved")
}
