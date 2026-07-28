package handlers

import (
	"encoding/json"
	"log"
	"strconv"
	// "fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/adityasrc/snip/backend/repository"
	"github.com/adityasrc/snip/backend/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mssola/user_agent"
	"github.com/oschwald/geoip2-golang"
)

type RequestPayload struct {
	LongURL   string `json:"long_url"`              // json tag
	ExpiresIn int    `json:"expiry_date,omitempty"` // TTL in seconds
}

type ResponsePayload struct {
	ShortURL string `json:"short_url"`
}

type LinkHandler struct {
	DB  *pgxpool.Pool
	Geo *geoip2.Reader
}

// type StatItem struct {
// 	Name  string `json:"name"`
// 	Count int    `json:"count"`
// }

// type DashboardResponse struct {
// 	TotalClicks int        `json:"totalClicks"`
// 	Countries   []StatItem `json:"countries"`
// 	OS          []StatItem `json:"os"`
// 	Browsers    []StatItem `json:"browsers"`
// 	Device      []StatItem `json:"devices"`
// 	Referers    []StatItem `json:"referers"`
// }

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
	userAgent := r.UserAgent()
	ua := user_agent.New(userAgent)
	os := ua.OS()
	browser, _ := ua.Browser()
	isBot := ua.Bot()
	device := "Desktop"
	if isBot == true {
		device = "Bot"
	} else if ua.Mobile() {
		device = "Mobile"
	}

	referer := r.Referer()

	var ipString string
	proxy := r.Header.Get("X-Forwarded-For")

	if proxy != "" {
		ips := strings.Split(proxy, ",")
		ipString = strings.TrimSpace(ips[0])
	} else {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ipString = r.RemoteAddr
		} else {
			ipString = host
		}
	}
	// fetching country name
	parsedIP := net.ParseIP(ipString) // converted string IP to go internal IP
	countryName := "unknown"          //default name
	if parsedIP != nil {
		record, err := h.Geo.Country(parsedIP)
		if err == nil && record.Country.Names["en"] != "" { // "en"(English) key
			countryName = record.Country.Names["en"]
		}
	}

	go func() {
		err = repository.SaveClick(h.DB, linkID, userAgent, referer, ipString, os, browser, device, countryName, isBot)
		if err != nil {
			log.Println("Analytics Error:", err)
		}
	}()

	http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
}

func (h *LinkHandler) Analytics(w http.ResponseWriter, r *http.Request) {

	slugStr := r.PathValue("id")

	linkID, err := strconv.Atoi(slugStr)
	if err != nil {
		utils.JSONError(w, "Invalid Link ID", http.StatusBadRequest)
		return
	}

	count, _ := repository.CountHandler(h.DB, linkID)

	countries, _ := repository.ClickHandler(h.DB, linkID, "country")
	os, _ := repository.ClickHandler(h.DB, linkID, "os")
	browser, _ := repository.ClickHandler(h.DB, linkID, "browser")
	device, _ := repository.ClickHandler(h.DB, linkID, "device")
	referer, _ := repository.ClickHandler(h.DB, linkID, "referer")

	res := repository.DashboardResponse{TotalClicks: count, Countries: countries, OS: os, Browsers: browser, Device: device, Referers: referer}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)

}
