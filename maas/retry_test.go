package maas

import (
	"errors"
	"testing"
)

func TestIsTransient(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"plain", errors.New("boom"), false},
		{"rst", errors.New("read tcp 1.2.3.4:80: connection reset by peer"), true},
		{"eof", errors.New("Post \"http://maas/MAAS/api/2.0/machines/\": EOF"), true},
		{"timeout", errors.New("dial tcp 1.2.3.4:80: i/o timeout"), true},
		{"tls", errors.New("TLS handshake timeout"), true},
		{"server-400", errors.New("ServerError: 400 Bad Request"), false},
		{"server-500", errors.New("ServerError: 500 Internal Server Error"), false},
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
	t.Run("success-first-attempt", func(t *testing.T) {
		calls := 0

		err := retryOnTransient(func() error {
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
		boom := errors.New("ServerError: 400 Bad Request")

		err := retryOnTransient(func() error {
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

		err := retryOnTransient(func() error {
			calls++
			if calls < 3 {
				return errors.New("connection reset by peer")
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
		rst := errors.New("connection reset by peer")

		err := retryOnTransient(func() error {
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
