package utile

import (
	"database/sql"
	"net/http"

	"github.com/ashutos120/go_transfer/internal/config"
)

type contextKey string

const ConfigKey = contextKey("config")
const DbKey = contextKey("db")
const ClaimsContextKey = contextKey("jwtClaims")

func GetConfig(r *http.Request) *config.Config {
	return r.Context().Value(ConfigKey).(*config.Config)
}

func GetDB(r *http.Request) *sql.DB {
	return r.Context().Value(DbKey).(*sql.DB)
}
func GetJWTClaim(r *http.Request) string {
	val := r.Context().Value(ClaimsContextKey)
	if claims, ok := val.(string); ok {
		return claims
	}
	return ""
}
