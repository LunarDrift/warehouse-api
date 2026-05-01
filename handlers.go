package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/LunarDrift/warehouse-api/internal/auth"
	"github.com/LunarDrift/warehouse-api/internal/database"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, `{"status": "ok"}`)
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var params struct {
		UserName string `json:"username"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not hash password", err)
		return
	}

	dbUser, err := s.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Username:       params.UserName,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		// not best solution. would be better to use pq's error types to check specific postgres error codes
		if strings.Contains(err.Error(), "unique constraint") {
			respondWithError(w, http.StatusBadRequest, "Username already taken", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not create user", err)
		return
	}

	// main.User so we're not sending hashed password in response
	user := User{
		ID:        dbUser.ID,
		UserName:  dbUser.Username,
		Role:      dbUser.Role,
		CreatedAt: dbUser.CreatedAt,
	}
	respondWithJSON(w, http.StatusCreated, user)
}
