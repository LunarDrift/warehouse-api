package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/LunarDrift/warehouse-api/internal/database"
)

type Server struct {
	mux       *http.ServeMux
	sqlDB     *sql.DB
	dbQueries *database.Queries
	jwtSecret string
}

func NewServer(db *sql.DB, dbQueries *database.Queries, jwtSecret string) *Server {
	srv := &Server{
		mux:       http.NewServeMux(),
		sqlDB:     db,
		dbQueries: dbQueries,
		jwtSecret: jwtSecret,
	}
	srv.registerRoutes()
	return srv
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)

	s.mux.HandleFunc("POST /register", s.handleCreateUser)
	s.mux.HandleFunc("POST /login", s.handleLoginUser)
	s.mux.HandleFunc("POST /refresh", s.handleRefreshAccessToken)
	s.mux.HandleFunc("POST /revoke", s.handleRevokeAccessToken)
	s.mux.HandleFunc("PATCH /users/{id}/password", s.middlewareAuth(s.handleChangePassword))
	s.mux.HandleFunc("PATCH /users/{id}/role", s.middlewareAuth(s.middlewareRequireRole("admin", s.handleUpdateUserRole)))
	s.mux.HandleFunc("GET /users", s.middlewareAuth(s.middlewareRequireRole("admin", s.handleGetAllUsers)))

	s.mux.HandleFunc("GET /items", s.middlewareAuth(s.handleGetItems))
	s.mux.HandleFunc("GET /items/search", s.middlewareAuth(s.handleSearchItems))
	s.mux.HandleFunc("POST /items", s.middlewareAuth(s.middlewareRequireRole("manager", s.handleCreateItem)))
	s.mux.HandleFunc("GET /items/{id}", s.middlewareAuth(s.handleGetItemFromID))
	s.mux.HandleFunc("DELETE /items/{id}", s.middlewareAuth(s.middlewareRequireRole("manager", s.handleDeleteItem)))
	s.mux.HandleFunc("PATCH /items/{id}", s.middlewareAuth(s.middlewareRequireRole("manager", s.handleUpdateItem)))

	s.mux.HandleFunc("POST /locations", s.middlewareAuth(s.middlewareRequireRole("manager", s.handleCreateLocation)))
	s.mux.HandleFunc("GET /locations", s.middlewareAuth(s.handleGetLocations))
	s.mux.HandleFunc("GET /locations/{id}", s.middlewareAuth(s.handleGetLocationFromID))
	s.mux.HandleFunc("PATCH /locations/{id}", s.middlewareAuth(s.middlewareRequireRole("manager", s.handleUpdateLocation)))
	s.mux.HandleFunc("DELETE /locations/{id}", s.middlewareAuth(s.middlewareRequireRole("manager", s.handleDeleteLocation)))

	s.mux.HandleFunc("GET /stock", s.middlewareAuth(s.handleGetAllStock))
	s.mux.HandleFunc("GET /stock/alerts", s.middlewareAuth(s.handleGetLowStockItems))
	s.mux.HandleFunc("GET /stock/item/{id}", s.middlewareAuth(s.handleGetItemStock))
	s.mux.HandleFunc("GET /stock/location/{id}", s.middlewareAuth(s.handleGetLocationStock))
	s.mux.HandleFunc("POST /stock/receive", s.middlewareAuth(s.middlewareRequireRole("manager", s.handleReceiveStock)))
	s.mux.HandleFunc("POST /stock/move", s.middlewareAuth(s.handleMoveStock))

	s.mux.HandleFunc("GET /movements", s.middlewareAuth(s.middlewareRequireRole("manager", s.handleGetMovementHistory)))
	s.mux.HandleFunc("GET /movements/item/{id}", s.middlewareAuth(s.middlewareRequireRole("manager", s.handleGetItemMovementHistory)))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, `{"status": "ok"}`)
}

func respondWithJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, status int, message string, err error) {
	if err != nil {
		log.Println(err)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, status, errorResponse{Error: message})
}
