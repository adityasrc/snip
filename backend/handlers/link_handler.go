package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/adityasrc/snip/backend/repository"
	"github.com/adityasrc/snip/backend/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RequestPayload struct {
	LongURL string `json:"long_url"` // json tag

}

type ResponsePayload struct {
	ShortURL string `json:"short_url"`
}

type LinkHandler struct {
	DB *pgxpool.Pool
}

// function as method
func (h *LinkHandler) CreateShortLink(w http.ResponseWriter, r *http.Request) {

	var req RequestPayload

	// Decoding the json request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.LongURL == "" {
		utils.JSONError(w, "URL cannot be empty", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(req.LongURL, "http://") && !strings.HasPrefix(req.LongURL, "https://") {
		req.LongURL = "https://" + req.LongURL
	}

	id, err := repository.GetNextID(h.DB)
	if err != nil {
		utils.JSONError(w, "Failed to generate id", http.StatusInternalServerError)
		return
	}

	slug := utils.ConvertToBase62(id)
	linkErr := repository.SaveLink(h.DB, id, req.LongURL, slug)

	if linkErr != nil {
		fmt.Println("SaveLink Error:", linkErr)
		utils.JSONError(w, "Failed to create short_url", http.StatusInternalServerError)
		return
	}

	res := ResponsePayload{ShortURL: "localhost:4000/" + slug}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)

}

func (h *LinkHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("slug")
	longURL, err := repository.GetLink(h.DB, slug)

	if err != nil {
		http.Error(w, "Invalid link", http.StatusInternalServerError)
	}

	http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
}
