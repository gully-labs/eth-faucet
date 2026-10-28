package server

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// DefaultGullyIDIssuer is the Gully ID server the page loads its browser SDK from (`<issuer>/sdk.js`).
const DefaultGullyIDIssuer = "https://gullyid.xyz"

var gullyIDClientRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

// SetGullyID turns on the "Sign in with Gully" button on the page. An empty clientID leaves it off (the default).
//
// It is identity only: the page uses the signed-in user's main wallet to pre-fill the address field and shows the
// .gully name. The server never sees or trusts a Gully ID token; /api/claim keeps the same hCaptcha and rate limits.
func (c *Config) SetGullyID(clientID, issuer string) error {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		c.gullyIDClient, c.gullyIDIssuer = "", ""
		return nil
	}
	if !gullyIDClientRe.MatchString(clientID) {
		return fmt.Errorf("invalid Gully ID client id %q (lowercase letters, digits and dashes)", clientID)
	}
	normalized, err := normalizeGullyIDIssuer(issuer)
	if err != nil {
		return err
	}
	c.gullyIDClient, c.gullyIDIssuer = clientID, normalized
	return nil
}

// normalizeGullyIDIssuer accepts a bare origin (https, or http only for localhost) and returns it without a trailing
// slash. Empty means DefaultGullyIDIssuer.
func normalizeGullyIDIssuer(issuer string) (string, error) {
	issuer = strings.TrimSpace(issuer)
	if issuer == "" {
		return DefaultGullyIDIssuer, nil
	}
	u, err := url.Parse(issuer)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("invalid Gully ID issuer %q", issuer)
	}
	host := u.Hostname()
	local := host == "localhost" || host == "127.0.0.1"
	if u.Scheme != "https" && !(u.Scheme == "http" && local) {
		return "", fmt.Errorf("invalid Gully ID issuer %q: must use https", issuer)
	}
	if (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return "", fmt.Errorf("invalid Gully ID issuer %q: must be an origin (scheme://host[:port])", issuer)
	}
	return u.Scheme + "://" + u.Host, nil
}
