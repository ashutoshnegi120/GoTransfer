package utile

import (
	"net/http"

	"github.com/ashutos120/go_transfer/internal/config"
)

type contextKey string

const ConfigKey = contextKey("config")
const ClaimsContextKey = contextKey("jwtClaims")

func GetConfig(r *http.Request) *config.Config {
	return r.Context().Value(ConfigKey).(*config.Config)
}

func GetJWTClaim(r *http.Request) string {
	val := r.Context().Value(ClaimsContextKey)
	if claims, ok := val.(string); ok {
		return claims
	}
	return ""
}
