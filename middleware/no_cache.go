package middleware

import (
	"net/http"
	"strings"
)

var Cacheables = []string{}

func NoCache(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		cacheable := false
		for _, c := range Cacheables {
			if strings.HasPrefix(r.URL.Path, c) {
				cacheable = true
				break
			}
		}
		if !cacheable {
			w.Header().Add("Cache-Control", "no-cache")
			w.Header().Add("Cache-Control", "no-store")
			w.Header().Set("Pragma", "no-cache")
		}
	})
}
