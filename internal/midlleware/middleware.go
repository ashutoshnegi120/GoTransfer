package midlleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ashutos120/go_transfer/internal/jwt"
	"github.com/ashutos120/go_transfer/internal/utile"
)

func JWTMiddleware(secretKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		secretKey := utile.GetConfig(r).SecretKey

		// Validate token
		claim, err := jwt.ValidateJWT(tokenString, secretKey)
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}
		// You can attach claims to context if needed (optional)
		ctx := context.WithValue(r.Context(), utile.ClaimsContextKey, claim.UserID)

		// Call actual handler
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
