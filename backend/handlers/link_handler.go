package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/adityasrc/snip/backend/repository"
	"github.com/adityasrc/snip/backend/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
)

type ShortentRequest struct {
	LongURL string `json:"long_url"` // json tag
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

type LinkHandler struct {
	DB *pgxpool.Pool
}

// function as method
func (h *LinkHandler) CreateShortLink(w http.ResponseWriter, r *http.Request) {

	var req ShortentRequest

	// Decoding the json request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	id, err := repository.GetNextID(h.DB)
	if err != nil {
		fmt.Println("GetNextID Error:", err)
		http.Error(w, "failed to generate id", http.StatusInternalServerError)
		return
	}

	slug := utils.ConvertToBase62(id)
	fmt.Println("Inserting:", id, req.LongURL, slug)
	linkErr := repository.SaveLink(h.DB, id, req.LongURL, slug)

	if linkErr != nil {
		fmt.Println("SaveLink Error:", linkErr)
		http.Error(w, "failed to create short_url", http.StatusInternalServerError)
		return
	}

	resp := ShortenResponse{ShortURL: "localhost:4000/" + slug}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	// fmt.Println("Received URL:", req.LongURL)

	// w.WriteHeader(http.StatusOK)
	// w.Write([]byte("URL received successfully"))

}
