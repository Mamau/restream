package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/mamau/restream/routes/response"
	"github.com/mamau/restream/routes/validator"
	"github.com/mamau/restream/routes/validator/contraints"
	"github.com/mamau/restream/stream"
	"net/http"
)

var streams = map[string]*stream.Stream{}

func streamStart(w http.ResponseWriter, r *http.Request) {
	var strm = stream.InitStream()
	if !validator.Validate(w, r, contraints.StreamStart{Stream: &strm}) {
		return
	}

	streams[strm.Name] = &strm
	strm.Start()

	response.Json(w, "Stream starting...", http.StatusOK)
}

func streamStop(w http.ResponseWriter, r *http.Request) {
	strm, err := getNameStream(r)
	if err != nil {
		response.Json(w, err.Error(), http.StatusBadRequest)
		return
	}
	strm.Stop()
	response.Json(w, "Stream stopping...", http.StatusOK)
}

func getNameStream(r *http.Request) (*stream.Stream, error) {
	type tmpStream struct {
		Name string
	}
	decoder := json.NewDecoder(r.Body)
	var ts tmpStream
	err := decoder.Decode(&ts)
	if err != nil {
		return &stream.Stream{}, err
	}
	strm, ok := streams[ts.Name]
	if !ok {
		return &stream.Stream{}, errors.New(fmt.Sprintf("Not found stream by name: %v", ts.Name))
	}
	delete(streams, ts.Name)
	return strm, nil
}

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Post("/stream-start", streamStart)
	r.Post("/stream-stop", streamStop)

	return r
}
