package v1

import (
	"github.com/go-chi/chi"
	"github.com/mamau/restream/routes/response"
	"github.com/mamau/restream/routes/validator"
	"github.com/mamau/restream/routes/validator/contraints"
	"github.com/mamau/restream/stream"
	"net/http"
)

var stream4eg = stream.InitStream()

func streamStart(w http.ResponseWriter, r *http.Request) {
	if !validator.Validate(w, r, contraints.StreamStart{Stream: &stream4eg}) {
		return
	}
	stream4eg.Start()
	response.Json(w, "Stream starting...", http.StatusOK)
}

func streamStop(w http.ResponseWriter, r *http.Request) {
	stream4eg.Stop()
	response.Json(w, "Stream stopping...", http.StatusOK)
}

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Post("/stream-start", streamStart)
	r.Post("/stream-stop", streamStop)

	return r
}
