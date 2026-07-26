package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/adityasrc/snip/backend/database"
	"github.com/joho/godotenv"
)

func ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Pong\n")
}

func main() {

	//godotenv check
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	log.Println(dbURL)

	//database connection
	res, err := database.Connect(dbURL)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("Db connected")
	defer res.Close()

	// DefaultServeMux
	// http.HandleFunc("/ping", ping)
	// log.Println("Server is running on port 4000")
	// if err := http.ListenAndServe(":4000", nil); err != nil {
	// 	log.Fatal(err)
	// }

	// custom Mux
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", ping)

	log.Println("Server is running on port 4000")
	if err := http.ListenAndServe(":4000", mux); err != nil {
		log.Fatal(err)
	}

}
