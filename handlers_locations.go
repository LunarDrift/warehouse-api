package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/LunarDrift/warehouse-api/internal/database"
	"github.com/google/uuid"
)

func (s *Server) handleCreateLocation(w http.ResponseWriter, r *http.Request) {
	var params locationParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	description := sql.NullString{}
	if params.Description != nil {
		description = sql.NullString{String: *params.Description, Valid: true}
	}

	loc, err := s.dbQueries.CreateLocation(r.Context(), database.CreateLocationParams{
		Name:        params.Name,
		Description: description,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create location", err)
		return
	}
	location := dbLocationToLocation(loc)
	respondWithJSON(w, http.StatusCreated, location)
}

func (s *Server) handleGetLocations(w http.ResponseWriter, r *http.Request) {
	locs, err := s.dbQueries.GetAllLocations(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch locations", err)
		return
	}

	// map results to main.Location to remove Description.Valid field from payload
	result := make([]Location, len(locs))
	for i, loc := range locs {
		result[i] = dbLocationToLocation(loc)
	}
	respondWithJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetLocationFromID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	locID, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid location ID", err)
		return
	}

	loc, err := s.dbQueries.GetLocationFromID(r.Context(), locID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch location", err)
		return
	}
	location := dbLocationToLocation(loc)
	respondWithJSON(w, http.StatusOK, location)
}

func (s *Server) handleUpdateLocation(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	locID, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid location ID", err)
		return
	}

	var params locationParams
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	description := sql.NullString{}
	if params.Description != nil {
		description = sql.NullString{String: *params.Description, Valid: true}
	}

	loc, err := s.dbQueries.UpdateLocation(r.Context(), database.UpdateLocationParams{
		ID:          locID,
		Name:        params.Name,
		Description: description,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update location", err)
		return
	}
	location := dbLocationToLocation(loc)
	respondWithJSON(w, http.StatusOK, location)
}

func (s *Server) handleDeleteLocation(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	locID, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid location ID", err)
		return
	}
	err = s.dbQueries.DeleteLocation(r.Context(), locID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete location", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
