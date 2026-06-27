package middleware

import (
	"net/http"

	utils "github.com/Clint-Mathews/EchoGate/internal/config"
)

const XInternalTokenKey string = "x-api-key"

func ApiKeyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(XInternalTokenKey) != utils.GetXInternalToken() {
			http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
