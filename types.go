package main

import (
	"time"

	"github.com/google/uuid"
)

// main.User so I'm not sending hashed password in payload

type User struct {
	ID        uuid.UUID `json:"id"`
	UserName  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Location struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type locationParams struct {
	Name        string  `json:"name"`
	Description *string `json:"description"` // *string because description is optional/nullable; convert to sql.NullString in handlers
}

type userParams struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type itemParams struct {
	Sku         string `json:"sku"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type receiveStockParams struct {
	ItemID     uuid.UUID `json:"item_id"`
	LocationID uuid.UUID `json:"location_id"`
	Quantity   int32     `json:"quantity"`
}

type moveStockParams struct {
	Quantity   int32     `json:"quantity"`
	ItemID     uuid.UUID `json:"item_id"`
	LocationID uuid.UUID `json:"location_id"`
}
