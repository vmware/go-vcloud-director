//go:build unit || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func retryTestClient(url string, timeoutSeconds int) *Client {
	retryInitialBackoff = time.Millisecond
	retryMaxBackoff = 5 * time.Millisecond
	c := &Client{MaxRetryTimeout: timeoutSeconds}
	c.Http = http.Client{Transport: newRetryTransport(http.DefaultTransport, c)}
	return c
}

func doRetryTest(t *testing.T, c *Client, method, url, body string) (*http.Response, string) {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, r)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.Http.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp, string(out)
}

func Test_retryTransport429ThenOk(t *testing.T) {
	var calls, bodies int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if string(b) == "payload" {
			atomic.AddInt32(&bodies, 1)
		}
		if atomic.AddInt32(&calls, 1) < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	resp, out := doRetryTest(t, retryTestClient(srv.URL, 5), http.MethodPost, srv.URL, "payload")
	if resp.StatusCode != 200 || out != "ok" || calls != 3 {
		t.Fatalf("status %d body %q calls %d", resp.StatusCode, out, calls)
	}
	if bodies != 3 {
		t.Fatalf("body must be resent on every attempt, got %d", bodies)
	}
}

func Test_retryTransportRateLimit500(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Client 'cloudadmin' exceeded request rate of 100 per second"))
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	resp, out := doRetryTest(t, retryTestClient(srv.URL, 5), http.MethodGet, srv.URL, "")
	if resp.StatusCode != 200 || out != "ok" || calls != 2 {
		t.Fatalf("status %d body %q calls %d", resp.StatusCode, out, calls)
	}
}

func Test_retryTransportOther500NotRetriedAndBodyKept(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("real failure"))
	}))
	defer srv.Close()

	resp, out := doRetryTest(t, retryTestClient(srv.URL, 5), http.MethodGet, srv.URL, "")
	if resp.StatusCode != 500 || out != "real failure" || calls != 1 {
		t.Fatalf("status %d body %q calls %d", resp.StatusCode, out, calls)
	}
}

func Test_retryTransportTimeoutExpires(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	start := time.Now()
	resp, _ := doRetryTest(t, retryTestClient(srv.URL, 1), http.MethodGet, srv.URL, "")
	if resp.StatusCode != 503 || calls < 2 || time.Since(start) > 3*time.Second {
		t.Fatalf("status %d calls %d elapsed %s", resp.StatusCode, calls, time.Since(start))
	}
}

func Test_retryTransportDisabled(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	resp, _ := doRetryTest(t, retryTestClient(srv.URL, 0), http.MethodGet, srv.URL, "")
	if resp.StatusCode != 429 || calls != 1 {
		t.Fatalf("status %d calls %d", resp.StatusCode, calls)
	}
}

func Test_retryTransportNetworkErrors(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			// Drop the connection without answering
			conn, _, _ := w.(http.Hijacker).Hijack()
			_ = conn.Close()
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	// GET is retried after the connection is dropped
	_, out := doRetryTest(t, retryTestClient(srv.URL, 5), http.MethodGet, srv.URL, "")
	if out != "ok" || calls != 2 {
		t.Fatalf("body %q calls %d", out, calls)
	}

	// POST is not retried after the connection is dropped
	atomic.StoreInt32(&calls, 0)
	c := retryTestClient(srv.URL, 5)
	req, _ := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader("x"))
	if resp, err := c.Http.Do(req); err == nil {
		_ = resp.Body.Close()
		t.Fatal("expected error for POST with dropped connection")
	}
	if calls != 1 {
		t.Fatalf("POST must not be retried, calls %d", calls)
	}
}

func Test_parseRetryAfter(t *testing.T) {
	if parseRetryAfter("3") != 3*time.Second || parseRetryAfter("") != 0 || parseRetryAfter("junk") != 0 {
		t.Fatal("unexpected Retry-After parsing")
	}
	if d := parseRetryAfter(time.Now().Add(10 * time.Second).UTC().Format(http.TimeFormat)); d <= 0 || d > 10*time.Second {
		t.Fatalf("unexpected date parsing: %s", d)
	}
}

func Test_retryTransportMaxAttempts(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := retryTestClient(srv.URL, 30)
	c.MaxRetryAttempts = 3
	resp, _ := doRetryTest(t, c, http.MethodGet, srv.URL, "")
	if resp.StatusCode != 429 || calls != 3 {
		t.Fatalf("status %d calls %d", resp.StatusCode, calls)
	}

	atomic.StoreInt32(&calls, 0)
	c.MaxRetryAttempts = 1
	doRetryTest(t, c, http.MethodGet, srv.URL, "")
	if calls != 1 {
		t.Fatalf("MaxRetryAttempts=1 must not retry, calls %d", calls)
	}
}

func Test_WithMaxRetryAttempts(t *testing.T) {
	c := &VCDClient{}
	if err := WithMaxRetryAttempts(-1)(c); err == nil {
		t.Fatal("expected error for negative attempts")
	}
	if err := WithMaxRetryAttempts(4)(c); err != nil || c.Client.MaxRetryAttempts != 4 {
		t.Fatalf("unexpected result: %v, %d", err, c.Client.MaxRetryAttempts)
	}
}
