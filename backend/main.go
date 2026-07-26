package main

import (
	// "fmt"
	"log"
	"net/http"
	"os"

	"github.com/adityasrc/snip/backend/database"
	"github.com/adityasrc/snip/backend/handlers"
	"github.com/joho/godotenv"
)

func main() {

	//godotenv check
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	log.Println(dbURL)

	//database connection
	db, err := database.Connect(dbURL)
	LinkHandler := &handlers.LinkHandler{DB: db}
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("Db connected")
	defer db.Close()

	mux := handlers.SetupRoutes(LinkHandler)

	log.Println("Server is running on port 4000")
	if err := http.ListenAndServe(":4000", mux); err != nil {
		log.Fatal(err)
	}
}
