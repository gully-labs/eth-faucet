package server

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Store keeps the rate-limit windows and the payout log, and writes them to a JSON file after every change so a
// restart (deploys restart the container several times a day) no longer resets every limit. Empty path = memory only.
type Store struct {
	mu      sync.Mutex
	path    string
	now     func() time.Time
	Limits  map[string]int64 `json:"limits"`  // key -> unix expiry
	Payouts []payoutRecord   `json:"payouts"` // last 24 h
	// token drip (glUSD): its own payout log (last 24 h) and the lifetime total, both in token base units
	TokenPayouts []payoutRecord `json:"token_payouts,omitempty"`
	TokenTotal   string         `json:"token_total,omitempty"`
}

type payoutRecord struct {
	At  int64  `json:"t"`
	Wei string `json:"wei"`
	Ref string `json:"ref"` // reservation id, so a failed send can be rolled back
	// Tier is "dev" for developer-tier payouts (counted against the developer daily budget), empty otherwise
	Tier string `json:"tier,omitempty"`
}

func NewStore(path string) *Store {
	s := &Store{path: path, now: time.Now, Limits: map[string]int64{}}
	if path == "" {
		return s
	}
	b, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(b, s); err != nil {
			log.WithError(err).Warn("faucet state file unreadable, starting empty")
		}
	} else if !os.IsNotExist(err) {
		log.WithError(err).Warn("faucet state file unreadable, starting empty")
	}
	if s.Limits == nil {
		s.Limits = map[string]int64{}
	}
	s.mu.Lock()
	s.pruneLocked()
	s.mu.Unlock()
	log.WithFields(log.Fields{"path": path, "limits": len(s.Limits), "payouts": len(s.Payouts)}).Info("faucet state loaded")
	return s
}

func (s *Store) pruneLocked() {
	now := s.now().Unix()
	for k, exp := range s.Limits {
		if exp <= now {
			delete(s.Limits, k)
		}
	}
	keep := s.Payouts[:0]
	for _, p := range s.Payouts {
		if p.At > now-86400 {
			keep = append(keep, p)
		}
	}
	s.Payouts = keep
	tkeep := s.TokenPayouts[:0]
	for _, p := range s.TokenPayouts {
		if p.At > now-86400 {
			tkeep = append(tkeep, p)
		}
	}
	s.TokenPayouts = tkeep
}

func (s *Store) saveLocked() {
	if s.path == "" {
		return
	}
	b, err := json.Marshal(s)
	if err != nil {
		log.WithError(err).Error("faucet state: marshal failed")
		return
	}
	tmp := s.path + ".tmp"
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err == nil {
		if err = os.WriteFile(tmp, b, 0o600); err == nil {
			err = os.Rename(tmp, s.path)
		}
		if err == nil {
			return
		}
		log.WithError(err).Error("faucet state: write failed")
	}
}

// remainingLocked returns how long key is still limited (0 = free).
func (s *Store) remainingLocked(key string) time.Duration {
	exp, ok := s.Limits[key]
	if !ok {
		return 0
	}
	d := time.Unix(exp, 0).Sub(s.now())
	if d <= 0 {
		delete(s.Limits, key)
		return 0
	}
	return d
}

// spentLocked sums the payouts in the last `window`.
func (s *Store) spentLocked(window time.Duration) *big.Int {
	return sumSince(s.Payouts, s.now().Add(-window).Unix())
}

// devSpentLocked sums the developer-tier payouts in the last `window`.
func (s *Store) devSpentLocked(window time.Duration) *big.Int {
	from := s.now().Add(-window).Unix()
	sum := new(big.Int)
	for _, p := range s.Payouts {
		if p.Tier == "dev" && p.At > from {
			if v, ok := new(big.Int).SetString(p.Wei, 10); ok {
				sum.Add(sum, v)
			}
		}
	}
	return sum
}

// tokenSpentLocked sums the token payouts in the last `window`.
func (s *Store) tokenSpentLocked(window time.Duration) *big.Int {
	return sumSince(s.TokenPayouts, s.now().Add(-window).Unix())
}

// tokenTotalLocked is the lifetime token total (base units).
func (s *Store) tokenTotalLocked() *big.Int {
	if v, ok := new(big.Int).SetString(s.TokenTotal, 10); ok {
		return v
	}
	return new(big.Int)
}

func sumSince(records []payoutRecord, from int64) *big.Int {
	sum := new(big.Int)
	for _, p := range records {
		if p.At > from {
			if v, ok := new(big.Int).SetString(p.Wei, 10); ok {
				sum.Add(sum, v)
			}
		}
	}
	return sum
}

// oldestInLocked is the timestamp of the oldest payout inside the window (when budget frees up again).
func (s *Store) oldestInLocked(window time.Duration) time.Time {
	from := s.now().Add(-window).Unix()
	for _, p := range s.Payouts {
		if p.At > from {
			return time.Unix(p.At, 0)
		}
	}
	return s.now()
}
