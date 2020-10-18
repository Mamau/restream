package v1

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/mamau/restream/stream"
	"log"
	"net/http"
	"strings"
)

const ValidBearer = "123456"

type Response struct {
	Message string `json:"message"`
}

var stream4eg = stream.InitStream()

func jsonResponse(w http.ResponseWriter, data interface{}, c int) {
	dj, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	fmt.Fprintf(w, "%s", dj)
}

func RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Make sure an Authorization header was provided
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		token = strings.TrimPrefix(token, "Bearer ")
		// This is where token validation would be done. For this boilerplate,
		// we just check and make sure the token matches a hardcoded string
		if token != ValidBearer {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		// Assuming that passed, we can execute the authenticated handler
		next.ServeHTTP(w, r)
	})
}

func streamStart(w http.ResponseWriter, r *http.Request) {
	stream4eg.Start(r.URL.Query())

	response := Response{
		Message: "Stream starting...",
	}
	jsonResponse(w, response, http.StatusOK)
}

func streamStop(w http.ResponseWriter, r *http.Request) {
	stream4eg.Stop()
	response := Response{
		Message: "Stream stopping...",
	}
	jsonResponse(w, response, http.StatusOK)
}

func NewRouter() http.Handler {
	r := chi.NewRouter()

	//r.Use(RequireAuthentication)
	r.Post("/stream-start", streamStart)
	r.Post("/stream-stop", streamStop)

	return r
}
