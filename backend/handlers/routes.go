package handlers

import (
	"net/http"
)

func SetupRoutes(handler *LinkHandler) *http.ServeMux {
	// custom Mux
	mux := http.NewServeMux()

	// mux.HandleFunc("POST /api/v1/shorten", handler.CreateShortLink)
	mux.Handle("POST /api/v1/shorten", AuthMiddleware(http.HandlerFunc(handler.CreateShortLink)))

	mux.HandleFunc("GET /{slug}", handler.RedirectURL)

	mux.HandleFunc("GET /api/v1/analytics/{id}", handler.Analytics)

	mux.HandleFunc("POST /api/v1/signup", handler.Signup)

	mux.HandleFunc("POST /api/v1/signin", handler.Signin)

	mux.Handle("GET /api/v1/links", AuthMiddleware(http.HandlerFunc(handler.GetMyLinks)))

	// mux.HandleFunc("DELETE /api/v1/link/{id}", )

	// mux.HandleFunc("PUT /api/v1/link/{id}", )

	mux.HandleFunc("POST /api/v1/logout", handler.Logout)

	// mux.HandleFunc("GET /api/v1/analytics/dashboard")

	mux.HandleFunc("GET /ping", handler.Ping)

	return mux
}
