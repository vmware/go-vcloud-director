// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"bytes"
	"context"
	"crypto/x509"
	"errors"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/vmware/go-vcloud-director/v3/util"
)

const (
	// retryRateLimitMessage is part of the body that VCD returns (with HTTP status 500) when the
	// backing network provider throttles a client
	retryRateLimitMessage = "exceeded request rate"
	// retryMaxBodyBytes limits how much of a 500 response body is inspected
	retryMaxBodyBytes = 64 * 1024
)

// Backoff settings. They are variables so that unit tests can shorten them.
var (
	retryInitialBackoff = 500 * time.Millisecond
	retryMaxBackoff     = 10 * time.Second
)

// retryTransport is an http.RoundTripper that transparently retries transient failures until
// Client.MaxRetryTimeout (seconds) is exhausted or, when set, Client.MaxRetryAttempts total
// attempts were made. A MaxRetryTimeout of 0 disables retries.
//
// Retried conditions:
//   - HTTP 429 and 503 (Retry-After is honoured) for every method
//   - HTTP 500 with a "request rate exceeded" message for every method
//   - network errors (connection reset, EOF, timeouts) for every method except POST, because a
//     POST may have been applied by the server before the connection failed
type retryTransport struct {
	next   http.RoundTripper
	client *Client
}

// newRetryTransport wraps a transport with retry logic driven by client.MaxRetryTimeout.
// A transport that is already wrapped is returned as is.
func newRetryTransport(next http.RoundTripper, client *Client) http.RoundTripper {
	if _, ok := next.(*retryTransport); ok {
		return next
	}
	if next == nil {
		next = http.DefaultTransport
	}
	return &retryTransport{next: next, client: client}
}

// RoundTrip implements http.RoundTripper
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	budget := time.Duration(t.client.MaxRetryTimeout) * time.Second
	// A body that cannot be replayed cannot be retried
	replayable := req.Body == nil || req.Body == http.NoBody || req.GetBody != nil
	if budget <= 0 || !replayable {
		return t.next.RoundTrip(req)
	}

	deadline := time.Now().Add(budget)
	for attempt := 0; ; attempt++ {
		attemptReq := req
		if attempt > 0 && req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			attemptReq = req.Clone(req.Context())
			attemptReq.Body = body
		}

		resp, err := t.next.RoundTrip(attemptReq)

		var wait time.Duration
		retry := false
		if err != nil {
			retry = isRetryableNetworkError(req.Method, err)
			wait = backoff(attempt)
		} else {
			var retryAfter time.Duration
			retry, retryAfter = isRetryableResponse(resp)
			wait = retryAfter
			if wait <= 0 {
				wait = backoff(attempt)
			}
		}

		attemptsExhausted := t.client.MaxRetryAttempts > 0 && attempt+1 >= t.client.MaxRetryAttempts
		if !retry || attemptsExhausted || time.Now().Add(wait).After(deadline) {
			return resp, err
		}

		util.Logger.Printf("[DEBUG - retryTransport] %s %s: retrying in %s (attempt %d, error: %v, response: %s)",
			req.Method, req.URL.Path, wait, attempt+1, err, statusOrNone(resp))
		if resp != nil {
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, retryMaxBodyBytes))
			_ = resp.Body.Close()
		}

		if err := sleepWithContext(req.Context(), wait); err != nil {
			return nil, err
		}
	}
}

func statusOrNone(resp *http.Response) string {
	if resp == nil {
		return "none"
	}
	return resp.Status
}

// isRetryableResponse reports whether the response is a transient failure and the delay
// requested by the server, if any. The body of a 500 response is restored after being inspected.
func isRetryableResponse(resp *http.Response) (bool, time.Duration) {
	switch resp.StatusCode {
	case http.StatusTooManyRequests, http.StatusServiceUnavailable:
		return true, parseRetryAfter(resp.Header.Get("Retry-After"))
	case http.StatusInternalServerError:
		if resp.Body == nil {
			return false, 0
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, retryMaxBodyBytes))
		// Put back what was read, followed by whatever is left in the original body
		resp.Body = struct {
			io.Reader
			io.Closer
		}{io.MultiReader(bytes.NewReader(body), resp.Body), resp.Body}
		if err != nil {
			return false, 0
		}
		return strings.Contains(strings.ToLower(string(body)), retryRateLimitMessage), 0
	}
	return false, 0
}

// isRetryableNetworkError reports whether a transport error is transient. POST is never retried
// because the server may have processed it already.
func isRetryableNetworkError(method string, err error) bool {
	if method == http.MethodPost {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var certErr x509.UnknownAuthorityError
	var hostErr x509.HostnameError
	var invalidErr x509.CertificateInvalidError
	if errors.As(err, &certErr) || errors.As(err, &hostErr) || errors.As(err, &invalidErr) {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNABORTED) ||
		errors.Is(err, syscall.EPIPE) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr)
}

// parseRetryAfter parses a Retry-After header (seconds or HTTP date). It returns 0 when absent or invalid.
func parseRetryAfter(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	if date, err := http.ParseTime(value); err == nil {
		if wait := time.Until(date); wait > 0 {
			return wait
		}
	}
	return 0
}

// backoff returns an exponential delay with jitter (50% to 100% of the computed value)
func backoff(attempt int) time.Duration {
	delay := retryInitialBackoff
	for i := 0; i < attempt && delay < retryMaxBackoff; i++ {
		delay *= 2
	}
	if delay > retryMaxBackoff {
		delay = retryMaxBackoff
	}
	// #nosec G404 -- jitter does not need a cryptographically secure source
	return delay/2 + time.Duration(rand.Int63n(int64(delay/2)+1))
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
