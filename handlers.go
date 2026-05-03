package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/LunarDrift/warehouse-api/internal/auth"
	"github.com/LunarDrift/warehouse-api/internal/database"
	"github.com/google/uuid"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, `{"status": "ok"}`)
}

// #########################################################################################################################
// USER HANDLERS
// #########################################################################################################################
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

	accessToken, err := auth.MakeJWT(dbUser.ID, s.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not make JWT", err)
		return
	}

	// TODO: Token database table
	// refreshToken := auth.MakeRefreshToken()

	type loginResponse struct {
		ID        uuid.UUID `json:"id"`
		UserName  string    `json:"username"`
		Role      string    `json:"role"`
		CreatedAt time.Time `json:"created_at"`
		Token     string    `json:"token"`
	}
	loginResp := loginResponse{
		ID:        dbUser.ID,
		UserName:  dbUser.Username,
		Role:      dbUser.Role,
		CreatedAt: dbUser.CreatedAt,
		Token:     accessToken,
	}
	respondWithJSON(w, http.StatusOK, loginResp)
}

// #########################################################################################################################
// ITEM HANDLERS
// #########################################################################################################################
// TODO: Implement low-stock threshold

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

// #########################################################################################################################
// LOCATION HANDLERS
// #########################################################################################################################
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

// #########################################################################################################################
// STOCK HANDLERS
// #########################################################################################################################
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
