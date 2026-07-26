package handlers

import (
	"encoding/json"
	"log"
	// "fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/adityasrc/snip/backend/repository"
	"github.com/adityasrc/snip/backend/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RequestPayload struct {
	LongURL   string `json:"long_url"`              // json tag
	ExpiresIn int    `json:"expiry_date,omitempty"` // TTL in seconds
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

	// expiry date
	var expiry *time.Time
	if req.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)
		expiry = &t
	}

	id, err := repository.GetNextID(h.DB)
	if err != nil {
		utils.JSONError(w, "Failed to generate id", http.StatusInternalServerError)
		return
	}

	slug := utils.ConvertToBase62(id)
	linkErr := repository.SaveLink(h.DB, id, req.LongURL, slug, expiry)

	if linkErr != nil {
		utils.JSONError(w, "Failed to create short_url", http.StatusInternalServerError)
		return
	}

	res := ResponsePayload{ShortURL: "localhost:4000/" + slug}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)

}

func (h *LinkHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("slug") // extract slug

	linkID, longURL, err := repository.GetLink(h.DB, slug) // fetch link
	if err != nil {
		utils.JSONError(w, "Link not found or expired", http.StatusNotFound)
		return
	}

	// Extract Metdadata
	user_agent := r.UserAgent()
	referer := r.Referer()

	var clientIP string
	proxy := r.Header.Get("X-Forwarded-For")

	if proxy != "" {
		ips := strings.Split(proxy, ",")
		clientIP = strings.TrimSpace(ips[0])
	} else {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			clientIP = r.RemoteAddr
		} else {
			clientIP = host
		}
	}

	go func() {
		err = repository.SaveClick(h.DB, linkID, user_agent, referer, clientIP)
		if err != nil {
			log.Println("Analytics Error:", err)
		}
	}()

	http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
}
