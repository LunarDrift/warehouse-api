package main

import (
	"database/sql"
	"encoding/json"
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

	s.mux.HandleFunc("GET /items", s.handleGetItems)
	s.mux.HandleFunc("POST /items", s.middlewareAuth(s.middlewareRequireRole("manager", s.handleCreateItem)))
	s.mux.HandleFunc("GET /items/{id}", s.handleGetItemFromID)
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
