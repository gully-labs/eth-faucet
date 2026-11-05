package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni/v3"

	"github.com/chainflag/eth-faucet/internal/chain"
)

// Guard is the claim gate, in order: denylist, Gully ID sign-in, the developer tier match, the 24 h windows (address,
// client IP, Gully ID account; developer tier: account and address on its own shorter window), the global hourly/daily
// budget (plus the developer budget), then recipient eligibility (not for the developer tier). Windows and payouts live in a Store that is
// persisted to disk, so restarts do not reset them. A send that fails rolls its reservation back.
//
// Sign-in has three modes (owner 2026-11-05: "i should be able to send test eth to any wallet not just gullyid"):
//   - required (gid): no valid session, no claim. A signed-in user is paid to their own Gully smart account or to an
//     address with history.
//   - optional (optGID, the live setting): without a session any address with history is paid (the eligibility check is
//     what stopped the 2026-10-27 farm, so it stays for anonymous claims). A valid session adds the per-account window
//     and may pay ANY address, a brand-new one included. A missing, expired or invalid session, or a Gully ID outage,
//     never fails the claim: it is an anonymous claim.
//   - off (neither): no sessions at all.
type Guard struct {
	store    *Store
	ip       *IPResolver
	window   time.Duration
	payout   *big.Int
	symbol   string
	deny     map[string]bool
	gid      GullyIDVerifier // nil = sign-in not required
	optGID   GullyIDVerifier // sign-in optional: checks the session a claim carries (nil = sessions are ignored)
	activity ActivityChecker // nil = no eligibility check
	hourCap  *big.Int        // nil = no cap
	dayCap   *big.Int
	dev      *DevTier        // nil = no developer tier
	devPay   *big.Int
	devCap   *big.Int // nil = no developer budget of its own
	alert    Alerter
	lastWarn int64 // unix seconds of the last budget alert
	seq      int64
}

type GuardOptions struct {
	Window   time.Duration
	Payout   *big.Int
	Symbol   string
	Deny     map[string]bool
	GullyID  GullyIDVerifier
	Activity ActivityChecker
	HourCap  *big.Int
	DayCap   *big.Int
	Alert    Alerter
	DevTier  *DevTier
	// OptionalGID verifies the session of a claim when GullyID is nil (sign-in optional): the per-account window, any
	// recipient for a verified user, and the developer tier. Ignored when GullyID is set.
	OptionalGID GullyIDVerifier
}

func NewGuard(store *Store, ip *IPResolver, o GuardOptions) *Guard {
	if o.Alert == nil {
		o.Alert = NewAlerter("", "", "")
	}
	if o.Deny == nil {
		o.Deny = map[string]bool{}
	}
	g := &Guard{store: store, ip: ip, window: o.Window, payout: o.Payout, symbol: o.Symbol, deny: o.Deny, gid: o.GullyID,
		activity: o.Activity, hourCap: o.HourCap, dayCap: o.DayCap, alert: o.Alert}
	if o.GullyID == nil {
		g.optGID = o.OptionalGID
	}
	if o.DevTier.enabled() {
		g.dev, g.devPay = o.DevTier, chain.EtherToWei(o.DevTier.Payout)
		if o.DevTier.DayCap > 0 {
			g.devCap = chain.EtherToWei(o.DevTier.DayCap)
		}
	}
	return g
}

// LoadDenylist reads one address per line ('#' comments allowed). A missing file is an empty list.
func LoadDenylist(path string) (map[string]bool, error) {
	out := map[string]bool{}
	if path == "" {
		return out, nil
	}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return out, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(strings.SplitN(sc.Text(), "#", 2)[0])
		if len(line) == 42 && strings.HasPrefix(line, "0x") {
			out[strings.ToLower(line)] = true
		}
	}
	return out, sc.Err()
}

func bearer(r *http.Request) string {
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(h) > 7 && strings.EqualFold(h[:7], "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

func humanWait(d time.Duration) string {
	if d >= time.Hour {
		h := int(d / time.Hour)
		m := int((d % time.Hour) / time.Minute)
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if d >= time.Minute {
		return fmt.Sprintf("%dm", int(d/time.Minute)+1)
	}
	return fmt.Sprintf("%ds", int(d/time.Second)+1)
}

func weiToEth(v *big.Int) string {
	f, _ := new(big.Float).Quo(new(big.Float).SetInt(v), big.NewFloat(1e18)).Float64()
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func (g *Guard) fail(w http.ResponseWriter, status int, msg string) {
	renderJSON(w, claimResponse{Message: msg}, status)
}

// budgetLocked returns a refusal message if one more payout of `amount` would break a cap (dev: also the developer budget).
func (g *Guard) budgetLocked(amount *big.Int, dev bool) string {
	check := func(capv *big.Int, window time.Duration, label string) string {
		if capv == nil || capv.Sign() <= 0 {
			return ""
		}
		spent := g.store.spentLocked(window)
		if new(big.Int).Add(spent, amount).Cmp(capv) <= 0 {
			return ""
		}
		free := g.store.oldestInLocked(window).Add(window).Sub(g.store.now())
		now := g.store.now().Unix()
		if now-atomic.LoadInt64(&g.lastWarn) > 3600 {
			atomic.StoreInt64(&g.lastWarn, now)
			g.alert.Alert(fmt.Sprintf("Gully faucet: %s budget of %s %s reached (%s paid). Claims paused until it frees up; check faucet.gullychain.xyz logs for farming.", label, weiToEth(capv), g.symbol, weiToEth(spent)))
		}
		return fmt.Sprintf("The faucet has handed out its %s budget. Please try again in %s.", label, humanWait(free))
	}
	if m := check(g.hourCap, time.Hour, "hourly"); m != "" {
		return m
	}
	if m := check(g.dayCap, 24*time.Hour, "daily"); m != "" {
		return m
	}
	if dev && g.devCap != nil && new(big.Int).Add(g.store.devSpentLocked(24*time.Hour), amount).Cmp(g.devCap) > 0 {
		now := g.store.now().Unix()
		if now-atomic.LoadInt64(&g.lastWarn) > 3600 {
			atomic.StoreInt64(&g.lastWarn, now)
			g.alert.Alert(fmt.Sprintf("Gully faucet: developer tier daily budget of %s %s reached.", weiToEth(g.devCap), g.symbol))
		}
		return "The developer tier has handed out its daily budget. Please try again later, or email team@gullylabs.xyz."
	}
	return ""
}

// Refusals for a recipient with no history. "no history" stays in every variant: the page keys on it.
const (
	msgFreshRequired = "This address has no history yet. Send the ETH to your Gully wallet (the smart account of your Gully ID), or to an address that has made at least one transaction on Gully Chain or Arbitrum Sepolia."
	msgFreshOptional = "This address has no history yet. Sign in with Gully ID to fund a brand-new wallet, or use a wallet that has made at least one transaction on Gully Chain or Arbitrum Sepolia."
	msgFreshGIDDown  = "This address has no history yet, and we could not check your Gully ID sign-in right now. Try again in a minute, or use a wallet that has made at least one transaction on Gully Chain or Arbitrum Sepolia."
	msgFreshNoSignIn = "This address has no history yet. Use a wallet that has made at least one transaction on Gully Chain or Arbitrum Sepolia."
)

func (g *Guard) freshAddressMessage(gidDown bool) string {
	switch {
	case g.gid != nil:
		return msgFreshRequired
	case g.optGID != nil && gidDown:
		return msgFreshGIDDown
	case g.optGID != nil:
		return msgFreshOptional
	}
	return msgFreshNoSignIn
}

func (g *Guard) limitedLocked(keys map[string]string) string {
	for _, k := range []string{"gid", "addr", "ip"} {
		key, ok := keys[k]
		if !ok {
			continue
		}
		if d := g.store.remainingLocked(key); d > 0 {
			switch k {
			case "gid":
				if strings.HasPrefix(key, "dev:") {
					return fmt.Sprintf("Developer tier: your Gully ID account already claimed. Next claim in %s.", humanWait(d))
				}
				return fmt.Sprintf("Your Gully ID account already claimed today. Next claim in %s.", humanWait(d))
			case "addr":
				return fmt.Sprintf("This address already received testnet ETH today. Next claim in %s.", humanWait(d))
			default:
				return fmt.Sprintf("A claim was already made from your network today. Next claim in %s.", humanWait(d))
			}
		}
	}
	return ""
}

func (g *Guard) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	address, err := readAddress(r)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			g.fail(w, mr.status, mr.message)
		} else {
			g.fail(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
		return
	}
	addr := strings.ToLower(address)
	clientIP := g.ip.ClientIP(r)
	logf := log.WithFields(log.Fields{"address": address, "clientIP": clientIP, "ua": r.UserAgent()})

	if g.deny[addr] {
		logf.Warn("claim refused: denylisted address")
		g.fail(w, http.StatusForbidden, "This address is not eligible for the faucet (it is linked to automated farming). If that is a mistake, ask in the GullyChain Telegram or Discord.")
		return
	}

	keys := map[string]string{"addr": "addr:" + addr, "ip": "ip:" + ipBucket(clientIP)}
	var user *GullyUser
	gidDown := false // optional mode: a session was sent but Gully ID could not check it
	if g.gid != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 7*time.Second)
		user, err = g.gid.Verify(ctx, bearer(r))
		cancel()
		if errors.Is(err, errNotSignedIn) {
			logf.Info("claim refused: not signed in")
			g.fail(w, http.StatusUnauthorized, "Sign in with Gully to claim testnet ETH. It is free and takes a few seconds (passkey, wallet or email).")
			return
		}
		if err != nil {
			logf.WithError(err).Warn("claim refused: Gully ID check failed")
			g.fail(w, http.StatusServiceUnavailable, "We could not check your Gully ID sign-in right now. Please try again in a minute.")
			return
		}
		keys["gid"] = "gid:" + user.Sub
		logf = logf.WithField("gid", user.Sub)
	} else if g.optGID != nil && bearer(r) != "" {
		// Sign-in is optional: a valid session adds the per-account window, lifts the fresh-address check below and can
		// qualify for the developer tier. An expired or invalid one, or a Gully ID outage, must not fail the claim: it
		// goes on as an anonymous claim (address + IP windows, eligibility check).
		ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
		u, verr := g.optGID.Verify(ctx, bearer(r))
		cancel()
		switch {
		case verr == nil && u != nil && u.Sub != "":
			user = u
			keys["gid"] = "gid:" + u.Sub
			logf = logf.WithField("gid", u.Sub)
		case verr != nil && !errors.Is(verr, errNotSignedIn):
			gidDown = true
			logf.WithError(verr).Warn("optional Gully ID session could not be checked: claim treated as anonymous")
		}
	}
	// a verified session while sign-in is optional: the account is known and rate limited, so any recipient is fine
	anyRecipient := g.gid == nil && user != nil

	payout, window, tier := g.payout, g.window, "standard"
	if ok, why := g.dev.Match(user, addr); ok {
		payout, window, tier = g.devPay, g.dev.Window, "developer"
		// own keys, no IP key: offices and CI share IPs, and the account is known
		keys = map[string]string{"gid": "dev:gid:" + user.Sub, "addr": "dev:addr:" + addr}
		logf = logf.WithFields(log.Fields{"gid": user.Sub, "tier": "developer", "why": why})
	}
	dev := tier == "developer"

	if window > 0 {
		g.store.mu.Lock()
		msg := g.limitedLocked(keys)
		g.store.mu.Unlock()
		if msg != "" {
			logf.Info("claim refused: rate limit")
			g.fail(w, http.StatusTooManyRequests, msg)
			return
		}
	}
	g.store.mu.Lock()
	msg := g.budgetLocked(payout, dev)
	g.store.mu.Unlock()
	if msg != "" {
		logf.Warn("claim refused: budget")
		g.fail(w, http.StatusServiceUnavailable, msg)
		return
	}

	if g.activity != nil && !dev && !anyRecipient && !(user != nil && user.SmartAccounts()[addr]) {
		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		ok, err := g.activity.HasActivity(ctx, address)
		cancel()
		if !ok && err != nil {
			logf.WithError(err).Warn("claim refused: eligibility check failed")
			g.fail(w, http.StatusServiceUnavailable, "We could not check this address right now. Please try again in a minute.")
			return
		}
		if !ok {
			logf.Info("claim refused: fresh address")
			g.fail(w, http.StatusForbidden, g.freshAddressMessage(gidDown))
			return
		}
	}

	// reserve atomically (re-check: another request may have won the race while we were checking)
	g.store.mu.Lock()
	if window > 0 {
		if msg := g.limitedLocked(keys); msg != "" {
			g.store.mu.Unlock()
			g.fail(w, http.StatusTooManyRequests, msg)
			return
		}
	}
	if msg := g.budgetLocked(payout, dev); msg != "" {
		g.store.mu.Unlock()
		g.fail(w, http.StatusServiceUnavailable, msg)
		return
	}
	ref := strconv.FormatInt(atomic.AddInt64(&g.seq, 1), 10) + "-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	exp := g.store.now().Add(window).Unix()
	if window > 0 {
		for _, k := range keys {
			g.store.Limits[k] = exp
		}
	}
	rec := payoutRecord{At: g.store.now().Unix(), Wei: payout.String(), Ref: ref}
	if dev {
		rec.Tier = "dev"
	}
	g.store.Payouts = append(g.store.Payouts, rec)
	g.store.saveLocked()
	g.store.mu.Unlock()

	next.ServeHTTP(w, r.WithContext(withClaimTier(r.Context(), tier, payout)))
	status := http.StatusOK
	if nw, ok := w.(negroni.ResponseWriter); ok {
		status = nw.Status()
	}
	if status != http.StatusOK {
		g.store.mu.Lock()
		for _, k := range keys {
			delete(g.store.Limits, k)
		}
		for i, p := range g.store.Payouts {
			if p.Ref == ref {
				g.store.Payouts = append(g.store.Payouts[:i], g.store.Payouts[i+1:]...)
				break
			}
		}
		g.store.saveLocked()
		g.store.mu.Unlock()
		return
	}
	logf.Info("Request succeeded, rate limit applied")
}
