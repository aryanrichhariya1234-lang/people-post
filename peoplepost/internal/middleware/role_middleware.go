package middleware

import (
	"net/http"
)

func Authorize(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value("userRole").(string)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			allowed := false
			for _, role := range roles {
				if role == userRole {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "Not authorized", http.StatusForbidden)
				return
			}

			next(w, r)
		}
	}
}