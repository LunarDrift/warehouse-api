package main

import (
	"context"

	"github.com/LunarDrift/warehouse-api/internal/database"
	"github.com/google/uuid"
)

func dbLocationToLocation(loc database.Location) Location {
	return Location{
		ID:          loc.ID,
		Name:        loc.Name,
		Description: loc.Description.String,
		CreatedAt:   loc.CreatedAt,
	}
}

func (s *Server) processMoveStock(ctx context.Context, params moveStockRequest, movedBy uuid.UUID) (database.StockMovement, error) {
	tx, err := s.sqlDB.Begin()
	if err != nil {
		return database.StockMovement{}, err
	}
	defer tx.Rollback()

	qtx := s.dbQueries.WithTx(tx)

	// decrement source
	_, err = qtx.MoveStock(ctx, database.MoveStockParams{
		Quantity:   -params.Quantity,
		ItemID:     params.ItemID,
		LocationID: params.FromLocationID,
	})
	if err != nil {
		return database.StockMovement{}, err
	}

	// increment destination; create row if it doesn't already exist with ReceiveStock
	_, err = qtx.ReceiveStock(ctx, database.ReceiveStockParams{
		ItemID:     params.ItemID,
		LocationID: params.ToLocationID,
		Quantity:   params.Quantity,
	})
	if err != nil {
		return database.StockMovement{}, err
	}

	movement, err := qtx.CreateMovement(ctx, database.CreateMovementParams{
		ItemID:         params.ItemID,
		FromLocationID: uuid.NullUUID{UUID: params.FromLocationID, Valid: true},
		ToLocationID:   uuid.NullUUID{UUID: params.ToLocationID, Valid: true},
		Quantity:       params.Quantity,
		MovedBy:        movedBy,
	})
	if err != nil {
		return database.StockMovement{}, err
	}

	return movement, tx.Commit()
}
