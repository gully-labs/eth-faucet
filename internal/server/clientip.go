package server

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

// Real client IP behind Cloudflare -> Caddy -> faucet.
//
// The 2026-10-27 drain showed the old X-Forwarded-For counting returned Cloudflare edge IPs (172.68.x, 104.23.x,
// 162.158.x, ...): every request got a different "client", so the per-IP limit did nothing. The rule now:
//
//  1. peer = the TCP peer. If it is a trusted proxy (loopback / private ranges: the edge Caddy container reaches the
//     faucet over the Docker bridge), peer = the right-most X-Forwarded-For entry, which is what Caddy itself saw.
//  2. If that peer is a Cloudflare edge address and CF-Connecting-IP is a valid IP, the client is CF-Connecting-IP.
//     Cloudflare overwrites that header, so it can only be forged by someone who bypasses Cloudflare, and then the
//     peer is not a Cloudflare address and the header is ignored.
//  3. Otherwise the client is the peer.
//
// IPv6 clients are bucketed per /64 (one home or VPS usually gets a whole /64).

// cloudflareCIDRs is https://www.cloudflare.com/ips-v4 and ips-v6 (2026-10). Override with -proxy.cloudflare.
var cloudflareCIDRs = []string{
	"173.245.48.0/20", "103.21.244.0/22", "103.22.200.0/22", "103.31.4.0/22", "141.101.64.0/18", "108.162.192.0/18",
	"190.93.240.0/20", "188.114.96.0/20", "197.234.240.0/22", "198.41.128.0/17", "162.158.0.0/15", "104.16.0.0/13",
	"104.24.0.0/14", "172.64.0.0/13", "131.0.72.0/22",
	"2400:cb00::/32", "2606:4700::/32", "2803:f800::/32", "2405:b500::/32", "2405:8100::/32", "2a06:98c0::/29",
	"2c0f:f248::/32",
}

var defaultTrustedProxies = []string{"127.0.0.0/8", "::1/128", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}

type IPResolver struct {
	trusted    []*net.IPNet
	cloudflare []*net.IPNet
}

func parseCIDRs(list []string) ([]*net.IPNet, error) {
	var out []*net.IPNet
	for _, s := range list {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", s, err)
		}
		out = append(out, n)
	}
	return out, nil
}

// NewIPResolver takes comma-separated CIDR lists; empty means the defaults above.
func NewIPResolver(trusted, cloudflare string) (*IPResolver, error) {
	tl, cl := defaultTrustedProxies, cloudflareCIDRs
	if strings.TrimSpace(trusted) != "" {
		tl = strings.Split(trusted, ",")
	}
	if strings.TrimSpace(cloudflare) != "" {
		cl = strings.Split(cloudflare, ",")
	}
	t, err := parseCIDRs(tl)
	if err != nil {
		return nil, err
	}
	c, err := parseCIDRs(cl)
	if err != nil {
		return nil, err
	}
	return &IPResolver{trusted: t, cloudflare: c}, nil
}

func inAny(ip net.IP, nets []*net.IPNet) bool {
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// ClientIP returns the client IP as a string (see the rules above).
func (r *IPResolver) ClientIP(req *http.Request) string {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
	}
	peer := net.ParseIP(host)
	if peer != nil && inAny(peer, r.trusted) {
		if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if p := net.ParseIP(strings.TrimSpace(parts[len(parts)-1])); p != nil {
				peer = p
			}
		}
	}
	if peer != nil && inAny(peer, r.cloudflare) {
		if cf := net.ParseIP(strings.TrimSpace(req.Header.Get("CF-Connecting-IP"))); cf != nil {
			return cf.String()
		}
	}
	if peer == nil {
		return host
	}
	return peer.String()
}

// ipBucket is the rate-limit key for an IP: the address itself for IPv4, its /64 for IPv6.
func ipBucket(ip string) string {
	p := net.ParseIP(ip)
	if p == nil || p.To4() != nil {
		return ip
	}
	return p.Mask(net.CIDRMask(64, 128)).String() + "/64"
}
