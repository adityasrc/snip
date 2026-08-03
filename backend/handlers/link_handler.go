package handlers

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	// "fmt"
	"github.com/adityasrc/snip/backend/repository"
	"github.com/adityasrc/snip/backend/utils"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	// "github.com/joho/godotenv"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mssola/user_agent"
	"github.com/oschwald/geoip2-golang"
	"golang.org/x/crypto/bcrypt"
)

type RequestPayload struct {
	LongURL   string `json:"long_url"`              // json tag
	ExpiresIn int    `json:"expiry_date,omitempty"` // TTL in seconds
	Email     string `json:"email" validate:"required,email"`
	Name      string `json:"name"`
	Password  string `json:"password" validate:"required,min=8,max=32"`
}

type ResponsePayload struct {
	ShortURL string `json:"short_url"`
}

type LinkHandler struct {
	DB       *pgxpool.Pool
	Geo      *geoip2.Reader
	Validate *validator.Validate
}

// We add jwt.RegisteredClaims as an embedded type, to provide fields like expiry time
type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// type Health struct {
// 	Status  string `json:"status"`
// 	Message string `json:"message"`
// }

// middleware signature
type Middleware func(http.Handler) http.Handler

type contextKey string

const UserEmailKey contextKey = "email"
const UserRoleKey contextKey = "role"

// middleware function
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))
		c, err := r.Cookie("token") // fetch token
		if err != nil {
			if err == http.ErrNoCookie {
				// If the cookie is not set, return an unauthorized status
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tokenStr := c.Value

		claims := &Claims{} // new instance of Claims

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
			return jwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

	})
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
	userEmail, _ := r.Context().Value(UserEmailKey).(string)
	if userEmail == "" {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	linkErr := repository.SaveLink(h.DB, id, req.LongURL, slug, expiry, userEmail)

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

func (h *LinkHandler) Signup(w http.ResponseWriter, r *http.Request) {

	var req RequestPayload

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	err := h.Validate.Struct(req)

	if err != nil {
		utils.JSONError(w, "Invalid Credentials", http.StatusBadRequest)
		return
	}

	if isUser := repository.CheckMail(h.DB, req.Email); isUser != false {
		utils.JSONError(w, "User already Exists", http.StatusConflict)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost) // dc is 10

	if err != nil {
		utils.JSONError(w, "Invalid Password", http.StatusBadRequest)
		return
	}
	email := strings.ToLower(req.Email)

	if err := repository.Signup(h.DB, req.Name, email, hash); err != nil {
		log.Println("DB Insert Error:", err)
		utils.JSONError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *LinkHandler) Signin(w http.ResponseWriter, r *http.Request) {

	var req RequestPayload
	var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	err := h.Validate.Struct(req)

	if err != nil {
		utils.JSONError(w, "Invalid Credentials", http.StatusBadRequest)
		return
	}

	pass, role, err := repository.Signin(h.DB, req.Email)
	if err != nil {
		utils.JSONError(w, "Invalid Email or Password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword(pass, []byte(req.Password))
	if err != nil {
		utils.JSONError(w, "Invalid Email or Password", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(5760 * time.Minute) // 4 days

	claims := &Claims{
		Email: req.Email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	w.WriteHeader(http.StatusOK)

}

func (h *LinkHandler) Ping(w http.ResponseWriter, r *http.Request) {

	res := map[string]string{
		"status":  "ok",
		"message": "Server is up and running",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		utils.JSONError(w, "Failed to encode JSON", http.StatusInternalServerError)
	}

}

func (h *LinkHandler) Logout(w http.ResponseWriter, r *http.Request) {

	cookie := http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // -1 ka matlab hai browser isko turant delete kar de
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)

	res := map[string]string{
		"status":  "ok",
		"message": "Logged out successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		utils.JSONError(w, "Failed to encode JSON", http.StatusInternalServerError)
	}

}

func (h *LinkHandler) GetMyLinks(w http.ResponseWriter, r *http.Request) {

	rawEmail := r.Context().Value(UserEmailKey)

	email, ok := rawEmail.(string)

	if !ok {
		utils.JSONError(w, "Unauthorized: Email not found in context", http.StatusUnauthorized)
		return
	}

	links, err := repository.Dashboard(h.DB, email)

	if err != nil {
		utils.JSONError(w, "Couldn't fetch links", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(links); err != nil {
		utils.JSONError(w, "Couldn't fetch links", http.StatusInternalServerError)
		return
	}
}
