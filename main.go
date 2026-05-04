package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/LunarDrift/warehouse-api/internal/database"
	"github.com/joho/godotenv"
	"github.com/pressly/goose"

	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	connectionString := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET")

	var sqlDB *sql.DB
	for i := range 5 {
		sqlDB, err = sql.Open("postgres", connectionString)
		if err == nil {
			if err = sqlDB.Ping(); err == nil {
				break
			}
		}
		log.Printf("DB not ready, retrying in 2 seconds... (attempt %d/5)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("Could not connect to database after retries: ", err)
	}
	defer sqlDB.Close()

	if err := goose.Up(sqlDB, "sql/schema"); err != nil {
		log.Fatal(err)
	}

	queries := database.New(sqlDB)

	srv := NewServer(sqlDB, queries, jwtSecret)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", srv.mux))
}
