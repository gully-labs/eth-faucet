package chain

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// DefaultProviders is the ordered public RPC list of Gully Chain Testnet (chain 485590), used when no provider is
// configured. First healthy URL wins. rpc.gullychain.xyz is deliberately absent: it becomes the mainnet (48559) RPC on
// cutover day, and a testnet faucet must never fall back to it.
var DefaultProviders = []string{
	"https://testnet-rpc.gullychain.xyz",
	"https://rpc.gullyrpc.xyz/gullychain/evm/485590",
	"https://rpc2.gullychain.xyz",
}

const (
	// DefaultAttemptTimeout bounds one request to one URL before the next URL is tried.
	DefaultAttemptTimeout = 8 * time.Second
	// DefaultCooldown is how long a failed URL is skipped before it is tried again (simple circuit breaker).
	DefaultCooldown = 30 * time.Second
)

// ParseProviders splits a comma-separated provider list (flag or WEB3_PROVIDER env), trimming blanks. An empty value
// gives DefaultProviders.
func ParseProviders(value string) []string {
	var urls []string
	for _, part := range strings.Split(value, ",") {
		if u := strings.TrimSpace(part); u != "" {
			urls = append(urls, u)
		}
	}
	if len(urls) == 0 {
		return append([]string(nil), DefaultProviders...)
	}
	return urls
}

// FallbackTransport is an http.RoundTripper that sends each JSON-RPC request to the first healthy URL of an ordered
// list and moves on to the next one after a network error, a timeout, HTTP 5xx or 429, or a JSON-RPC rate-limit
// error. A URL that failed is skipped for Cooldown, then tried again, so the preferred URL comes back on its own.
// Attempts are strictly sequential: a request (for example eth_sendRawTransaction) is never sent to two URLs at once.
type FallbackTransport struct {
	URLs           []string
	AttemptTimeout time.Duration
	Cooldown       time.Duration
	Base           http.RoundTripper

	mu        sync.Mutex
	downUntil map[int]time.Time
	now       func() time.Time
}

// NewFallbackTransport returns a FallbackTransport over urls with the default timeout and cooldown.
func NewFallbackTransport(urls []string) *FallbackTransport {
	return &FallbackTransport{URLs: urls, AttemptTimeout: DefaultAttemptTimeout, Cooldown: DefaultCooldown}
}

func (t *FallbackTransport) clock() time.Time {
	if t.now != nil {
		return t.now()
	}
	return time.Now()
}

// order returns the URL indexes to try: healthy ones in list order, then the ones still cooling down (so a request is
// never refused just because every URL failed recently).
func (t *FallbackTransport) order() []int {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.clock()
	healthy := make([]int, 0, len(t.URLs))
	var cooling []int
	for i := range t.URLs {
		if until, ok := t.downUntil[i]; ok && now.Before(until) {
			cooling = append(cooling, i)
			continue
		}
		healthy = append(healthy, i)
	}
	return append(healthy, cooling...)
}

func (t *FallbackTransport) mark(i int, ok bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if ok {
		delete(t.downUntil, i)
		return
	}
	if t.downUntil == nil {
		t.downUntil = map[int]time.Time{}
	}
	cooldown := t.Cooldown
	if cooldown <= 0 {
		cooldown = DefaultCooldown
	}
	t.downUntil[i] = t.clock().Add(cooldown)
}

// RoundTrip implements http.RoundTripper. The request URL is ignored: each attempt goes to one of t.URLs.
func (t *FallbackTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(t.URLs) == 0 {
		return nil, errors.New("rpc fallback: no provider URLs")
	}
	var body []byte
	if req.Body != nil {
		b, err := ioutil.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return nil, err
		}
		body = b
	}
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	timeout := t.AttemptTimeout
	if timeout <= 0 {
		timeout = DefaultAttemptTimeout
	}

	var lastErr error
	for _, i := range t.order() {
		if err := req.Context().Err(); err != nil {
			return nil, err
		}
		resp, err := t.attempt(req, base, t.URLs[i], body, timeout)
		if err == nil {
			t.mark(i, true)
			return resp, nil
		}
		t.mark(i, false)
		lastErr = err
		log.WithFields(log.Fields{"rpc": t.URLs[i], "error": err}).Warn("rpc provider failed, trying the next one")
	}
	return nil, fmt.Errorf("rpc fallback: all %d providers failed, last error: %w", len(t.URLs), lastErr)
}

func (t *FallbackTransport) attempt(req *http.Request, base http.RoundTripper, url string, body []byte, timeout time.Duration) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(req.Context(), timeout)
	defer cancel()
	out, err := http.NewRequestWithContext(ctx, req.Method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	for k, v := range req.Header {
		out.Header[k] = append([]string(nil), v...)
	}
	resp, err := base.RoundTrip(out)
	if err != nil {
		return nil, err
	}
	// Read the whole body inside the attempt deadline, so a stalled body counts as a failed attempt too.
	respBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	if isRateLimited(respBody) {
		return nil, errors.New("json-rpc rate limited")
	}
	resp.Body = io.NopCloser(bytes.NewReader(respBody))
	resp.ContentLength = int64(len(respBody))
	return resp, nil
}

type rpcErrorEnvelope struct {
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// isRateLimited reports a JSON-RPC rate-limit error (code -32005 or 429, or a "rate limit" / "too many requests"
// message) in a single or batch response. Other JSON-RPC errors are real answers and are returned to the caller.
func isRateLimited(body []byte) bool {
	trimmed := bytes.TrimSpace(body)
	var envelopes []rpcErrorEnvelope
	if len(trimmed) > 0 && trimmed[0] == '[' {
		if json.Unmarshal(trimmed, &envelopes) != nil {
			return false
		}
	} else {
		var one rpcErrorEnvelope
		if json.Unmarshal(trimmed, &one) != nil {
			return false
		}
		envelopes = []rpcErrorEnvelope{one}
	}
	for _, e := range envelopes {
		if e.Error == nil {
			continue
		}
		msg := strings.ToLower(e.Error.Message)
		if e.Error.Code == -32005 || e.Error.Code == 429 ||
			strings.Contains(msg, "rate limit") || strings.Contains(msg, "rate-limit") || strings.Contains(msg, "too many requests") {
			return true
		}
	}
	return false
}
