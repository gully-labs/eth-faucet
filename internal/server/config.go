package server

import "fmt"

type Config struct {
	network         string
	symbol          string
	httpPort        int
	interval        int
	payout          float64
	proxyCount      int
	hcaptchaSiteKey string
	hcaptchaSecret  string
	gullyIDClient   string // empty = no "Sign in with Gully" button (see SetGullyID)
	gullyIDIssuer   string
	protection      Protection
	token           *TokenDrip // nil = ETH only (see SetTokenDrip)
}

func NewConfig(network, symbol string, httpPort, interval, proxyCount int, payout float64, hcaptchaSiteKey, hcaptchaSecret string) *Config {
	return &Config{
		network:         network,
		symbol:          symbol,
		httpPort:        httpPort,
		interval:        interval,
		payout:          payout,
		proxyCount:      proxyCount,
		hcaptchaSiteKey: hcaptchaSiteKey,
		hcaptchaSecret:  hcaptchaSecret,
	}
}

// Protection settings (2026-10-27 drain hardening). Set with SetProtection; the zero value is the old behaviour minus
// the broken IP logic: no sign-in, no eligibility, no budget, memory-only state.
type Protection struct {
	GullyIDRequired bool            // claims need a "Sign in with Gully" access token (Authorization: Bearer)
	GullyID         GullyIDVerifier // built from the client id + issuer when nil and GullyIDRequired
	Activity        ActivityChecker // non-nil = fresh recipients (no history on these chains) are refused
	HourCapEth      float64         // 0 = no hourly budget
	DayCapEth       float64         // 0 = no daily budget
	StateFile       string          // "" = memory only
	Deny            map[string]bool
	TrustedProxies  string // CIDRs; "" = loopback + private ranges
	CloudflareCIDRs string // CIDRs; "" = built-in Cloudflare list
	Alert           Alerter
}

func (c *Config) SetProtection(p Protection) error {
	if p.GullyIDRequired && p.GullyID == nil {
		if c.gullyIDClient == "" {
			return fmt.Errorf("-gullyid.required needs -gullyid.client (GULLY_ID_CLIENT_ID)")
		}
		p.GullyID = NewGullyIDVerifier(c.gullyIDIssuer, c.gullyIDClient)
	}
	if _, err := NewIPResolver(p.TrustedProxies, p.CloudflareCIDRs); err != nil {
		return err
	}
	c.protection = p
	return nil
}
