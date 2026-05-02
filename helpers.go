package main

import "github.com/LunarDrift/warehouse-api/internal/database"

func dbLocationToLocation(loc database.Location) Location {
	return Location{
		ID:          loc.ID,
		Name:        loc.Name,
		Description: loc.Description.String,
		CreatedAt:   loc.CreatedAt,
	}
}
