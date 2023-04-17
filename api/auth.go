package api

import (
	"crypto/rand"
	"fmt"
	"net/http"
)

func generateSalt() ([]byte, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %v", err)
	}

	return bytes, nil
}

func (s *Server) Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := s.Sessions.GetString(r.Context(), "userID")
		if userID == "" {
			// User is not logged in, return an error response or redirect to the login page
			http.Error(w, "You must be logged in to perform this action", http.StatusForbidden)
			return
		}
		// User is logged in, continue to the next handler
		next.ServeHTTP(w, r)
	})
}
