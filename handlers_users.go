package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/LunarDrift/warehouse-api/internal/auth"
	"github.com/LunarDrift/warehouse-api/internal/database"
	"github.com/google/uuid"
)

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var params userParams
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

func (s *Server) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	var params userParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	dbUser, err := s.dbQueries.GetUserFromName(r.Context(), params.UserName)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect username or password", err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, dbUser.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect username or password", err)
		return
	}

	accessToken, err := auth.MakeJWT(dbUser.ID, s.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not make JWT", err)
		return
	}

	refreshToken := auth.MakeRefreshToken()

	_, err = s.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    dbUser.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not save refresh token", err)
		return
	}

	type loginResponse struct {
		User         User   `json:"user"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	loginResp := loginResponse{
		User: User{
			ID:        dbUser.ID,
			UserName:  dbUser.Username,
			Role:      dbUser.Role,
			CreatedAt: dbUser.CreatedAt,
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
	}
	respondWithJSON(w, http.StatusOK, loginResp)
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	requestingUser, ok := r.Context().Value(userKey).(database.User)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var params userParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userIDStr := r.PathValue("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// allow if admin or changing own password
	if requestingUser.Role != "admin" && requestingUser.ID != userID {
		respondWithError(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not hash password", err)
		return
	}
	user, err := s.dbQueries.ResetUserPassword(r.Context(), database.ResetUserPasswordParams{
		ID:             userID,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not change password", err)
		return
	}
	respondWithJSON(w, http.StatusOK, User{
		ID:        user.ID,
		UserName:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

func (s *Server) handleUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	requestingUser, ok := r.Context().Value(userKey).(database.User)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var params struct {
		Role string `json:"role"`
	}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: could not find role", err)
		return
	}

	userIDStr := r.PathValue("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	if requestingUser.Role != "admin" {
		respondWithError(w, http.StatusForbidden, "Forbidden", nil)
	}

	user, err := s.dbQueries.UpdateUserRole(r.Context(), database.UpdateUserRoleParams{
		ID:   userID,
		Role: params.Role,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update user role", err)
		return
	}
	respondWithJSON(w, http.StatusOK, User{
		ID:        user.ID,
		UserName:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

func (s *Server) handleRefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: could not find token", err)
		return
	}

	user, err := s.dbQueries.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, s.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: could not validate token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}

func (s *Server) handleRevokeAccessToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: could not find token", err)
		return
	}
	_, err = s.dbQueries.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not revoke session", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.dbQueries.GetAllUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch users", err)
		return
	}

	result := make([]User, len(users))
	for i, u := range users {
		result[i] = User{
			ID:        u.ID,
			UserName:  u.Username,
			Role:      u.Role,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		}
	}
	respondWithJSON(w, http.StatusOK, result)
}
