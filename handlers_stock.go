package main

import (
	"encoding/json"
	"net/http"

	"github.com/LunarDrift/warehouse-api/internal/database"
	"github.com/google/uuid"
)

func (s *Server) handleReceiveStock(w http.ResponseWriter, r *http.Request) {
	var params receiveStockParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	stockInfo, err := s.dbQueries.ReceiveStock(r.Context(), database.ReceiveStockParams{
		ItemID:     params.ItemID,
		LocationID: params.LocationID,
		Quantity:   params.Quantity,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not receive item", err)
		return
	}
	respondWithJSON(w, http.StatusCreated, stockInfo)
}

func (s *Server) handleGetAllStock(w http.ResponseWriter, r *http.Request) {
	items, err := s.dbQueries.GetAllStockWithThreshold(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch stock info", err)
		return
	}
	result := make([]stockResponse, len(items))
	for i, item := range items {
		result[i] = stockResponse{
			ID:                item.ID,
			ItemID:            item.ItemID,
			LocationID:        item.LocationID,
			Quantity:          item.Quantity,
			Name:              item.Name,
			Sku:               item.Sku,
			LowStockThreshold: item.LowStockThreshold,
			LowStockWarning:   item.Quantity < item.LowStockThreshold,
		}
	}
	respondWithJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetItemStock(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	itemID, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid location ID", err)
		return
	}
	stock, err := s.dbQueries.GetStockForItemAllLocations(r.Context(), itemID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch stock info for that item", err)
		return
	}
	respondWithJSON(w, http.StatusOK, stock)
}

func (s *Server) handleGetLocationStock(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	locID, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid location ID", err)
		return
	}
	stock, err := s.dbQueries.GetStockForItemSpecificLocation(r.Context(), locID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch stock info for that location", err)
		return
	}
	respondWithJSON(w, http.StatusOK, stock)
}

func (s *Server) handleGetLowStockItems(w http.ResponseWriter, r *http.Request) {
	lowStockItems, err := s.dbQueries.GetLowStockItems(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch items", err)
		return
	}
	type response struct {
		ID                uuid.UUID `json:"id"`
		ItemID            uuid.UUID `json:"item_id"`
		LocationID        uuid.UUID `json:"location_id"`
		Quantity          int32     `json:"quantity"`
		Name              string    `json:"item_name"`
		Sku               string    `json:"sku"`
		LowStockThreshold int32     `json:"low_stock_threshold"`
	}
	result := make([]response, len(lowStockItems))
	for i, itemRow := range lowStockItems {
		result[i] = response{
			ID:                itemRow.ID,
			ItemID:            itemRow.ItemID,
			LocationID:        itemRow.LocationID,
			Quantity:          itemRow.Quantity,
			Name:              itemRow.Name,
			Sku:               itemRow.Sku,
			LowStockThreshold: itemRow.LowStockThreshold,
		}
	}
	respondWithJSON(w, http.StatusOK, result)
}

// #########################################################################################################################
// STOCK MOVEMENT HANDLERS
// #########################################################################################################################
func (s *Server) handleMoveStock(w http.ResponseWriter, r *http.Request) {
	var params moveStockRequest
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	user, ok := r.Context().Value(userKey).(database.User)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not get user from context", nil)
		return
	}

	movement, err := s.processMoveStock(r.Context(), params, user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not move stock", err)
		return
	}
	respondWithJSON(w, http.StatusOK, movement)
}

func (s *Server) handleGetMovementHistory(w http.ResponseWriter, r *http.Request) {
	movements, err := s.dbQueries.GetAllMovements(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch movement history", err)
		return
	}
	respondWithJSON(w, http.StatusOK, movements)
}

func (s *Server) handleGetItemMovementHistory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	itemID, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid location ID", err)
		return
	}

	movementHistory, err := s.dbQueries.GetMovementsForItem(r.Context(), itemID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch movement history", err)
		return
	}
	respondWithJSON(w, http.StatusOK, movementHistory)
}
