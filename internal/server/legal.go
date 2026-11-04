package server

import "net/http"

// Gully Chain has one Terms of Service and one Privacy Policy, on gullychain.xyz; they cover the faucet.
// /terms and /privacy (with or without a trailing slash) answer 301 to them, ahead of the static app.
var legalPages = map[string]string{
	"/terms":   "https://gullychain.xyz/terms",
	"/privacy": "https://gullychain.xyz/privacy",
}

func registerLegalRedirects(router *http.ServeMux) {
	for p, to := range legalPages {
		h := http.RedirectHandler(to, http.StatusMovedPermanently)
		router.Handle(p, h)
		router.Handle(p+"/", h)
	}
}
