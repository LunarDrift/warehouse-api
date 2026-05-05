package main

import (
	"context"
	"net/http"

	"github.com/LunarDrift/warehouse-api/internal/auth"
	"github.com/LunarDrift/warehouse-api/internal/database"
)

// custom key type to avoid context key collisions
type contextKey string

const userKey contextKey = "user"

func (s *Server) middlewareAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Missing or invalid token", err)
			return
		}
		userID, err := auth.ValidateJWT(token, s.jwtSecret)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
			return
		}
		// fetch full user so role is available downstream
		user, err := s.dbQueries.GetUserFromID(r.Context(), userID)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "User not found", err)
			return
		}
		// attach user to context
		ctx := context.WithValue(r.Context(), userKey, user)
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) middlewareRequireRole(role Role, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// fetch user from context
		user, ok := r.Context().Value(userKey).(database.User)
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
			return
		}
		if Role(user.Role) != role {
			respondWithError(w, http.StatusForbidden, "Forbidden", nil)
			return
		}
		next(w, r)
	}
}
