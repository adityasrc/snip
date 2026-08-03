package main

import (
	// "fmt"
	"log"
	"net/http"
	"os"

	"github.com/adityasrc/snip/backend/database"
	"github.com/adityasrc/snip/backend/handlers"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

func main() {

	//godotenv check
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}

	dbURL := os.Getenv("DATABASE_URL")

	//database connection
	db, geoReader, err := database.Connect(dbURL)
	LinkHandler := &handlers.LinkHandler{DB: db, Geo: geoReader, Validate: validator.New()}
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}

	defer db.Close()
	defer geoReader.Close()

	mux := handlers.SetupRoutes(LinkHandler)

	log.Println("Server is running on port 4000")
	if err := http.ListenAndServe(":4000", mux); err != nil {
		log.Fatal(err)
	}
}
