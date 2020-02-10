package pep

import (
	"context"
	"testing"
	"time"
)

const testTimeout = 10 * time.Second

func TestStreamConnConnectOk(t *testing.T) {
	pdpServer := startTestPDPServer(allPermitPolicy, 5555, t)
	defer func() {
		if logs := pdpServer.Stop(); len(logs) > 0 {
			t.Logf("server logs:\n%s", logs)
		}
	}()

	statesCh := make(chan int, 0)
	testConn := newStreamConn(context.Background(), "127.0.0.1:5555", 3, nil,
		func(address string, state int, err error) {
			statesCh <- state
		},
	)

	go testConn.connect()

	testTimer := time.NewTimer(testTimeout)

wait:
	for {
		select {
		case st := <-statesCh:
			if st == StreamingConnectionFailure {
				t.Error("streamConn.connect() unexpectedly failed")
				break wait
			} else if st == StreamingConnectionEstablished {
				break wait
			}
		case <-testTimer.C:
			t.Error("streamConn.connect() test timed out")
			break wait
		}
	}
	testConn.closeConn()
}

func TestStreamConnConnectError(t *testing.T) {
	statesCh := make(chan int, 0)
	testConn := newStreamConn(context.Background(), "127.34.56.78:9", 3, nil,
		func(address string, state int, err error) {
			statesCh <- state
		},
	)

	go testConn.connect()

	testTimer := time.NewTimer(testTimeout)

wait:
	for {
		select {
		case st := <-statesCh:
			if st == StreamingConnectionFailure {
				break wait
			} else if st == StreamingConnectionEstablished {
				t.Error("streamConn.connect() unexpectedly connected")
				break wait
			}
		case <-testTimer.C:
			t.Error("streamConn.connect() test timed out")
			break wait
		}
	}

	testConn.closeConn()

	testTimer.Reset(testTimeout)
	cbTimer := time.NewTimer(time.Second)

wait2:
	for {
		select {
		case <-statesCh:
			cbTimer.Reset(time.Second)
		case <-cbTimer.C:
			// no connection notifications got in last second, propably connection loop exited
			break wait2
		case <-testTimer.C:
			// connection notifications didn't stop, propably connection looped forever
			t.Error("test timed out, propably streamConn.connect() looped forever")
			break wait2
		}
	}
}
