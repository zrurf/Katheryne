package botproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// NewProxy returns an http.Handler that proxies requests to the given bot API host.
// It strips the "/api/v1" prefix so that gateway requests like
//
//	/api/v1/bot/developer/v1/bots
//
// become
//
//	/bot/developer/v1/bots
//
// on the target bot-api server.
func NewHandler(botAPIHost string) http.Handler {
	target, _ := url.Parse("http://" + botAPIHost)
	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/bot/") || r.URL.Path == "/api/v1/bot" {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api/v1")
			proxy.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})
}
