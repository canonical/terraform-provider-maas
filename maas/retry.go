package maas

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"syscall"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/juju/gomaasapi/v2"
)

// isTransient reports whether err is a network-level transient failure that is
// safe to retry. HTTP/API responses and ordinary client errors fall through.
func isTransient(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := gomaasapi.GetServerError(err); ok {
		return false
	}

	var serverErr gomaasapi.ServerError
	if errors.As(err, &serverErr) {
		return false
	}

	if errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, syscall.ECONNRESET) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	return false
}

// retryOnTransient retries fn up to maxAttempts on transient errors using the
// Terraform SDK retry helper. Non-transient errors and successes return
// immediately.
func retryOnTransient(ctx context.Context, fn func() error, maxAttempts int) error {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	attempts := 0
	timeout := time.Duration(maxAttempts) * 600 * time.Millisecond

	return retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		attempts++

		err := fn()
		if err == nil {
			return nil
		}

		if !isTransient(err) {
			return retry.NonRetryableError(err)
		}

		if attempts >= maxAttempts {
			return retry.NonRetryableError(err)
		}

		log.Printf("[WARN] transient MAAS error (attempt %d/%d): %v; retrying",
			attempts, maxAttempts, err)

		return retry.RetryableError(err)
	})
}
