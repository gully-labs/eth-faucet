package server

import (
	"bufio"
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/chainflag/eth-faucet/internal/chain"
)

// DevTier is the developer tier (2026-10-28): a larger drip and a shorter cooldown for signed-in Gully ID users who are
// verified developers, so teams building on Gully Chain can deploy and test without asking for ETH by hand.
//
// Who qualifies (any one, always with a valid Gully ID session; hCaptcha still runs first):
//   - the Gully ID subject is on the allowlist (Subs),
//   - one of the user's Gully ID wallets, or the recipient address, is on the wallet allowlist (Wallets),
//   - TrustProfile is on and the Gully ID profile marks the user a developer ("developer": true, or "developer" in
//     "roles").
//
// A developer claim skips the fresh-address check (developers deploy from new keys) and the per-IP window (offices and
// CI share IPs), keeps the per-account and per-address windows (Window, its own keys), and counts toward the global
// hourly/daily budget plus the developer tier's own daily budget (DayCap).
type DevTier struct {
	Payout       float64       // ETH per developer claim; 0 = tier off
	Window       time.Duration // cooldown per Gully ID account and per recipient address
	DayCap       float64       // ETH per rolling 24 h across all developer claims; 0 = only the global caps
	Subs         map[string]bool
	Wallets      map[string]bool // lowercased 0x addresses
	TrustProfile bool
}

func (d *DevTier) enabled() bool { return d != nil && d.Payout > 0 }

// Validate checks the tier against the base config: it needs Gully ID and at least one way to qualify.
func (d *DevTier) Validate(baseWindow time.Duration, gullyIDClient string) error {
	if !d.enabled() {
		return nil
	}
	if gullyIDClient == "" {
		return fmt.Errorf("developer tier needs Gully ID: set -gullyid.client (GULLY_ID_CLIENT_ID)")
	}
	if len(d.Subs) == 0 && len(d.Wallets) == 0 && !d.TrustProfile {
		return fmt.Errorf("developer tier is on but nobody can qualify: set FAUCET_DEV_SUBS, FAUCET_DEV_WALLETS, FAUCET_DEV_ALLOWLIST_FILE or FAUCET_DEV_TRUST_PROFILE")
	}
	if d.Window <= 0 {
		return fmt.Errorf("developer tier window must be > 0 minutes")
	}
	if baseWindow > 0 && d.Window > baseWindow {
		return fmt.Errorf("developer tier window (%s) is longer than the normal window (%s)", d.Window, baseWindow)
	}
	return nil
}

// Match says whether a signed-in user claiming to addr gets the developer tier, and why (for the log).
func (d *DevTier) Match(u *GullyUser, addr string) (bool, string) {
	if !d.enabled() || u == nil || u.Sub == "" {
		return false, ""
	}
	if d.Subs[u.Sub] {
		return true, "allowlisted account"
	}
	for _, w := range u.Wallets {
		if d.Wallets[strings.ToLower(w.Address)] {
			return true, "allowlisted wallet"
		}
	}
	if d.Wallets[strings.ToLower(addr)] {
		return true, "allowlisted recipient"
	}
	if d.TrustProfile && u.IsDeveloper() {
		return true, "Gully ID developer"
	}
	return false, ""
}

// ParseDevList splits a comma, space or newline separated list into Gully ID subjects and 0x addresses ('#' starts a
// comment). Addresses are lowercased; subjects are kept as they are.
func ParseDevList(s string, subs, wallets map[string]bool) {
	for _, line := range strings.Split(s, "\n") {
		line = strings.SplitN(line, "#", 2)[0]
		for _, f := range strings.FieldsFunc(line, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\r' }) {
			if len(f) == 42 && strings.HasPrefix(strings.ToLower(f), "0x") {
				wallets[strings.ToLower(f)] = true
			} else if f != "" {
				subs[f] = true
			}
		}
	}
}

// LoadDevAllowlist reads one Gully ID subject or 0x address per line into subs/wallets. A missing file is empty.
func LoadDevAllowlist(path string, subs, wallets map[string]bool) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		ParseDevList(sc.Text(), subs, wallets)
	}
	return sc.Err()
}

func (c *Config) SetDevTier(d *DevTier) error {
	if !d.enabled() {
		c.devTier = nil
		return nil
	}
	if err := d.Validate(time.Duration(c.interval)*time.Minute, c.gullyIDClient); err != nil {
		return err
	}
	c.devTier = d
	return nil
}

// ---- the claim's tier and amount, from the guard to the claim handler ----

type claimTierKey struct{}

type claimTier struct {
	name   string // "standard" or "developer"
	amount *big.Int
}

func withClaimTier(ctx context.Context, name string, amount *big.Int) context.Context {
	return context.WithValue(ctx, claimTierKey{}, claimTier{name: name, amount: amount})
}

// tierFrom returns the guard's tier and amount for this request, or the standard payout when no guard set one.
func tierFrom(ctx context.Context, payoutEth float64) claimTier {
	if t, ok := ctx.Value(claimTierKey{}).(claimTier); ok && t.amount != nil {
		return t
	}
	return claimTier{name: "standard", amount: chain.EtherToWei(payoutEth)}
}

type devTierInfo struct {
	Payout      string  `json:"payout"`
	WindowHours float64 `json:"window_hours"`
}
