package handlers

import (
	"net/http"
)

func SetupRoutes(handler *LinkHandler) *http.ServeMux {
	// custom Mux
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/shorten", handler.CreateShortLink)

	mux.HandleFunc("GET /{slug}", handler.RedirectURL)

	mux.HandleFunc("GET /api/v1/analytics/{id}", handler.Analytics)

	mux.HandleFunc("POST /api/v1/signup", handler.Signup)

	mux.HandleFunc("POST /api/v1/signin", handler.Signin)

	// mux.HandleFunc("GET /api/v1/user/{id}", handler.GetUser)

	return mux
}
