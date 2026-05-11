package maas

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
	"testing"

	jujuerrors "github.com/juju/errors"
	"github.com/juju/gomaasapi/v2"
)

type timeoutError struct{}

func (timeoutError) Error() string {
	return "request timed out"
}

func (timeoutError) Timeout() bool {
	return true
}

func (timeoutError) Temporary() bool {
	return false
}

func TestIsTransient(t *testing.T) {
	serverEOF := gomaasapi.ServerError{
		StatusCode:  500,
		BodyMessage: "backend returned EOF",
	}
	serverTimeout := gomaasapi.ServerError{
		StatusCode:  504,
		BodyMessage: "dial tcp 1.2.3.4:80: i/o timeout",
	}

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"plain", errors.New("boom"), false},
		{"server-eof-body", serverEOF, false},
		{"std-wrapped-server-eof-body", fmt.Errorf("get machine: %w", serverEOF), false},
		{"wrapped-server-timeout-body", jujuerrors.Trace(serverTimeout), false},
		{"wrapped-eof", fmt.Errorf("read response: %w", io.EOF), true},
		{"wrapped-unexpected-eof", fmt.Errorf("read response: %w", io.ErrUnexpectedEOF), true},
		{"wrapped-econnreset", fmt.Errorf("read tcp: %w", syscall.ECONNRESET), true},
		{"net-timeout", &net.OpError{Op: "read", Net: "tcp", Err: timeoutError{}}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isTransient(tc.err); got != tc.want {
				t.Fatalf("isTransient(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestRetryOnTransient(t *testing.T) {
	ctx := context.Background()

	t.Run("success-first-attempt", func(t *testing.T) {
		calls := 0

		err := retryOnTransient(ctx, func() error {
			calls++

			return nil
		}, 5)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		if calls != 1 {
			t.Fatalf("calls = %d, want 1", calls)
		}
	})

	t.Run("non-transient-no-retry", func(t *testing.T) {
		calls := 0
		boom := errors.New("boom")

		err := retryOnTransient(ctx, func() error {
			calls++

			return boom
		}, 5)
		if !errors.Is(err, boom) {
			t.Fatalf("got %v, want %v", err, boom)
		}

		if calls != 1 {
			t.Fatalf("calls = %d, want 1", calls)
		}
	})

	t.Run("transient-recovers", func(t *testing.T) {
		calls := 0

		err := retryOnTransient(ctx, func() error {
			calls++
			if calls < 3 {
				return fmt.Errorf("read tcp: %w", syscall.ECONNRESET)
			}

			return nil
		}, 5)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		if calls != 3 {
			t.Fatalf("calls = %d, want 3", calls)
		}
	})

	t.Run("transient-exhausts", func(t *testing.T) {
		calls := 0
		rst := fmt.Errorf("read tcp: %w", syscall.ECONNRESET)

		err := retryOnTransient(ctx, func() error {
			calls++

			return rst
		}, 3)
		if !errors.Is(err, rst) {
			t.Fatalf("got %v, want %v", err, rst)
		}

		if calls != 3 {
			t.Fatalf("calls = %d, want 3", calls)
		}
	})
}
