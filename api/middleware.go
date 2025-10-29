package api

import (
	"net/http"

	stationHttpUtils "github.com/massalabs/station/pkg/http"
	"github.com/massalabs/station/pkg/logger"
)

func OriginRestrictMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := stationHttpUtils.GetRequestOrigin(r)
		hostname := stationHttpUtils.ExtractHostname(origin)

		for _, allowedDomain := range allowedDomains() {
			if hostname == allowedDomain {
				next.ServeHTTP(w, r)
				return
			}
		}
		logger.Warnf("Origin %s not allowed", origin)
		http.Error(w, "Origin not allowed", http.StatusForbidden)
	})
}

func allowedDomains() []string {
	return []string{"station.massa", "localhost", "127.0.0.1"}
}
