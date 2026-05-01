package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/LunarDrift/warehouse-api/internal/database"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Couldn't find .env file: ", err)
	}

	connectionString := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET")

	sqlDB, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("Could not connect to database: ", err)
	}
	defer sqlDB.Close()

	queries := database.New(sqlDB)

	srv := NewServer(sqlDB, queries, jwtSecret)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", srv.mux))
}
