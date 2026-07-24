package main

import (
	// "fmt"
	"log"
	"os"

	"github.com/adityasrc/snip/backend/database"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	log.Println(dbURL)

	res, err := database.Connect(dbURL)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("Db connected")
	defer res.Close()

}
