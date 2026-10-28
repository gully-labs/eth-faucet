package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

// ---- Gully ID: who is claiming? -------------------------------------------------------------------------------

// GullyUser is the part of the Gully ID profile the faucet uses.
type GullyUser struct {
	Sub     string  `json:"sub"`
	Client  *string `json:"-"`
	Wallets []struct {
		ChainID *int64 `json:"chainId"`
		Address string `json:"address"`
		Kind    string `json:"kind"`
	} `json:"wallets"`
	// Optional developer markers (only read when the developer tier trusts the profile, -devtier.profile)
	Developer *bool    `json:"developer,omitempty"`
	Roles     []string `json:"roles,omitempty"`
}

// IsDeveloper: the profile says "developer": true, or lists "developer" in its roles.
func (u *GullyUser) IsDeveloper() bool {
	if u.Developer != nil && *u.Developer {
		return true
	}
	for _, r := range u.Roles {
		if strings.EqualFold(strings.TrimSpace(r), "developer") {
			return true
		}
	}
	return false
}

// SmartAccounts returns the user's Gully smart accounts (kernel), lowercased.
func (u *GullyUser) SmartAccounts() map[string]bool {
	out := map[string]bool{}
	for _, w := range u.Wallets {
		if strings.HasPrefix(w.Kind, "kernel") {
			out[strings.ToLower(w.Address)] = true
		}
	}
	return out
}

type GullyIDVerifier interface {
	Verify(ctx context.Context, accessToken string) (*GullyUser, error)
}

var errNotSignedIn = errors.New("not signed in")

// httpGullyID asks `<issuer>/v1/profile` with the session's access token (docs: gully-id/docs/first-party-apps.md,
// "Backends: who is this user?"). OIDC tokens carry `client`, which must be our own client id so a token minted for
// another Gully app cannot be replayed here; gid_ passkey session tokens have client = null and are accepted.
type httpGullyID struct {
	issuer, clientID string
	hc               *http.Client
	mu               sync.Mutex
	cache            map[string]cachedUser
}

type cachedUser struct {
	u   *GullyUser
	exp time.Time
}

func NewGullyIDVerifier(issuer, clientID string) GullyIDVerifier {
	return &httpGullyID{issuer: strings.TrimRight(issuer, "/"), clientID: clientID, hc: &http.Client{Timeout: 6 * time.Second}, cache: map[string]cachedUser{}}
}

func (g *httpGullyID) Verify(ctx context.Context, token string) (*GullyUser, error) {
	token = strings.TrimSpace(token)
	if token == "" || len(token) > 4096 {
		return nil, errNotSignedIn
	}
	sum := sha256.Sum256([]byte(token))
	key := hex.EncodeToString(sum[:])
	g.mu.Lock()
	if c, ok := g.cache[key]; ok && time.Now().Before(c.exp) {
		g.mu.Unlock()
		return c.u, nil
	}
	g.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, "GET", g.issuer+"/v1/profile", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	res, err := g.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gully id unreachable: %w", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return nil, errNotSignedIn
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gully id: HTTP %d", res.StatusCode)
	}
	var raw struct {
		GullyUser
		Client *string `json:"client"`
	}
	if err := json.Unmarshal(body, &raw); err != nil || raw.Sub == "" {
		return nil, fmt.Errorf("gully id: bad profile")
	}
	if raw.Client != nil && *raw.Client != g.clientID {
		log.WithField("client", *raw.Client).Warn("Gully ID token issued to another client refused")
		return nil, errNotSignedIn
	}
	u := raw.GullyUser
	u.Client = raw.Client
	g.mu.Lock()
	if len(g.cache) > 5000 {
		g.cache = map[string]cachedUser{}
	}
	g.cache[key] = cachedUser{u: &u, exp: time.Now().Add(5 * time.Minute)}
	g.mu.Unlock()
	return &u, nil
}

// ---- eligibility: does the recipient look like a real user? ---------------------------------------------------

// ActivityChecker says whether an address has any history (sent a transaction, or is a deployed contract) on one of
// the configured chains. Every recipient of the 2026-10-27 drain had nonce 0 on Gully Chain AND Arbitrum Sepolia.
type ActivityChecker interface {
	HasActivity(ctx context.Context, address string) (bool, error)
}

type rpcActivity struct {
	urls []string
}

func NewActivityChecker(urls []string) ActivityChecker {
	return &rpcActivity{urls: urls}
}

func (a *rpcActivity) HasActivity(ctx context.Context, address string) (bool, error) {
	addr := common.HexToAddress(address)
	var lastErr error
	for _, u := range a.urls {
		c, err := ethclient.DialContext(ctx, u)
		if err != nil {
			lastErr = err
			continue
		}
		cctx, cancel := context.WithTimeout(ctx, 4*time.Second)
		n, err := c.NonceAt(cctx, addr, nil)
		if err == nil && n > 0 {
			cancel()
			c.Close()
			return true, nil
		}
		if err != nil {
			lastErr = err
		}
		code, err := c.CodeAt(cctx, addr, nil)
		cancel()
		c.Close()
		if err == nil && len(code) > 0 {
			return true, nil
		}
		if err != nil {
			lastErr = err
		}
	}
	if lastErr != nil {
		// some chain could not be checked and none showed activity: let the caller decide (fail closed)
		return false, lastErr
	}
	return false, nil
}

// ---- alerts ----------------------------------------------------------------------------------------------------

type Alerter interface {
	Alert(text string)
}

// telegramAlerter posts to a Telegram chat and/or a JSON webhook ({"text": ...}); with neither set it only logs.
type telegramAlerter struct {
	botToken, chatID, webhook string
	hc                        *http.Client
}

func NewAlerter(botToken, chatID, webhook string) Alerter {
	return &telegramAlerter{botToken: strings.TrimSpace(botToken), chatID: strings.TrimSpace(chatID), webhook: strings.TrimSpace(webhook), hc: &http.Client{Timeout: 8 * time.Second}}
}

func (t *telegramAlerter) Alert(text string) {
	log.WithField("alert", text).Error("FAUCET ALERT")
	go func() {
		if t.botToken != "" && t.chatID != "" {
			b, _ := json.Marshal(map[string]string{"chat_id": t.chatID, "text": text})
			if res, err := t.hc.Post("https://api.telegram.org/bot"+t.botToken+"/sendMessage", "application/json", strings.NewReader(string(b))); err != nil {
				log.Warn("faucet alert: telegram send failed")
			} else {
				res.Body.Close()
			}
		}
		if t.webhook != "" {
			b, _ := json.Marshal(map[string]string{"text": text, "source": "gully-faucet"})
			if res, err := t.hc.Post(t.webhook, "application/json", strings.NewReader(string(b))); err != nil {
				log.Warn("faucet alert: webhook failed")
			} else {
				res.Body.Close()
			}
		}
	}()
}
