package handlers

import (
	"net/http"
)

func SetupRoutes(handler *LinkHandler) *http.ServeMux {
	// custom Mux
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/shorten", handler.CreateShortLink)

	mux.HandleFunc("GET /{slug}", handler.RedirectURL)

	// mux.HandleFunc("GET /api/v1/analytics/{slug}", handler.Analytics)

	return mux
}
