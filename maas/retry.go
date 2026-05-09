package maas

import (
	"log"
	"strings"
	"time"
)

// isTransient reports whether err is a network-level transient failure that is
// safe to retry. Conservative: only RST / EOF / I/O timeout patterns observed
// in practice. Real client errors (4xx, 5xx response bodies, schema errors)
// MUST fall through.
func isTransient(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()

	return strings.Contains(msg, "connection reset by peer") ||
		strings.Contains(msg, "EOF") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "TLS handshake timeout")
}

// retryOnTransient retries fn up to maxAttempts on transient errors using
// exponential backoff capped at 30s. Non-transient errors and successes return
// immediately.
func retryOnTransient(fn func() error, maxAttempts int) error {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var err error

	backoff := time.Second

	for i := 0; i < maxAttempts; i++ {
		err = fn()
		if err == nil || !isTransient(err) {
			return err
		}

		log.Printf("[WARN] transient MAAS error (attempt %d/%d): %v; retrying in %s",
			i+1, maxAttempts, err, backoff)

		time.Sleep(backoff)

		if backoff < 30*time.Second {
			backoff *= 2
		}
	}

	return err
}
