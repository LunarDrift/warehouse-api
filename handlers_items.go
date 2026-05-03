package main

import (
	"encoding/json"
	"net/http"

	"github.com/LunarDrift/warehouse-api/internal/database"
	"github.com/google/uuid"
)

func (s *Server) handleCreateItem(w http.ResponseWriter, r *http.Request) {
	var params itemParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
	}
	item, err := s.dbQueries.CreateItem(r.Context(), database.CreateItemParams{
		Sku:               params.Sku,
		Name:              params.Name,
		Description:       params.Description,
		LowStockThreshold: params.LowStockThreshold,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create item", err)
		return
	}
	respondWithJSON(w, http.StatusCreated, item)
}

func (s *Server) handleGetItems(w http.ResponseWriter, r *http.Request) {
	items, err := s.dbQueries.GetAllItems(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch items", err)
		return
	}
	respondWithJSON(w, http.StatusOK, items)
}

func (s *Server) handleGetItemFromID(w http.ResponseWriter, r *http.Request) {
	itemIDStr := r.PathValue("id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid item ID", err)
		return
	}
	item, err := s.dbQueries.GetItemFromID(r.Context(), itemID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch item", err)
		return
	}
	respondWithJSON(w, http.StatusOK, item)
}

func (s *Server) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	itemID, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid item ID", err)
		return
	}
	err = s.dbQueries.DeleteItem(r.Context(), itemID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete item", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUpdateItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	itemID, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid item ID", err)
		return
	}

	var params itemParams
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	item, err := s.dbQueries.UpdateItem(r.Context(), database.UpdateItemParams{
		ID:          itemID,
		Sku:         params.Sku,
		Name:        params.Name,
		Description: params.Description,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update item", err)
		return
	}
	respondWithJSON(w, http.StatusOK, item)
}

func (s *Server) handleSearchItems(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	items, err := s.dbQueries.SearchItems(r.Context(), query)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not search items", err)
		return
	}
	respondWithJSON(w, http.StatusOK, items)
}
