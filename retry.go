package zmq4

import (
	"syscall"
)

// IsRetryError returns true if err indicates an interruption (as
// opposed to a failure) in a zmq_recv* or zmq_send* call; otherwise, it
// returns false.
//
// The errors considered to not represent failure cases are EAGAIN and
// EINTR.
func IsRetryError(err error) bool {
	if err == syscall.EAGAIN {
		return true
	}
	if err == syscall.EINTR {
		return true
	}
	return false
}

// RetryRecv retries a Recv call until successful or a non-retryable
// error code is returned.
func (soc *Socket) RetryRecv(flags Flag) (string, error) {
	var err error
	err = syscall.EAGAIN
	var data string
	for IsRetryError(err) {
		data, err = soc.Recv(flags)
	}
	return data, err
}

// RetryRecvMessage retries a RecvMessage call until successful or a
// non-retryable error code is returned.
func (soc *Socket) RetryRecvMessage(flags Flag) ([]string, error) {
	var err error
	err = syscall.EAGAIN
	var data []string
	for IsRetryError(err) {
		data, err = soc.RecvMessage(flags)
	}
	return data, err
}

// RetryRecvBytes retries a RecvBytes call until successful or a
// non-retryable error code is returned.
func (soc *Socket) RetryRecvBytes(flags Flag) ([]byte, error) {
	var err error
	err = syscall.EAGAIN
	var data []byte
	for IsRetryError(err) {
		data, err = soc.RecvBytes(flags)
	}
	return data, err
}
// RetryRecvMessageBytes retries a RecvMessageBytes call until
// successful or a non-retryable error code is returned.
func (soc *Socket) RetryRecvMessageBytes(flags Flag) ([][]byte, error) {
	var err error
	err = syscall.EAGAIN
	var data [][]byte
	for IsRetryError(err) {
		data, err = soc.RecvMessageBytes(flags)
	}
	return data, err
}

// RetrySend retries a Send call until successful or a non-retryable
// error code is returned.
func (soc *Socket) RetrySend(data string, flags Flag) (int, error) {
	var err error
	err = syscall.EAGAIN
	var written int
	for IsRetryError(err) {
		written, err = soc.Send(data, flags)
	}
	return written, err
}

// RetrySendMessage retries a SendMessage call until successful or a
// non-retryable error code is returned.
func (soc *Socket) RetrySendMessage(parts ...interface{}) (int, error) {
	var err error
	err = syscall.EAGAIN
	var written int
	for IsRetryError(err) {
		written, err = soc.SendMessage(parts...)
	}
	return written, err
}

// RetrySendBytes retries a SendBytes call until successful or a
// non-retryable error code is returned.
func (soc *Socket) RetrySendBytes(data []byte, flags Flag) (int, error) {
	var err error
	err = syscall.EAGAIN
	var written int
	for IsRetryError(err) {
		written, err = soc.SendBytes(data, flags)
	}
	return written, err
}
