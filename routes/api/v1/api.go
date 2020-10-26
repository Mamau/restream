package v1

import (
	"github.com/go-chi/chi"
	"github.com/mamau/restream/helpers"
	"github.com/mamau/restream/routes/response"
	"github.com/mamau/restream/routes/validator"
	"github.com/mamau/restream/routes/validator/contraints"
	"github.com/mamau/restream/stream"
	"go.uber.org/zap"
	"net/http"
)

var Live = stream.NewLive()

func streamStart(w http.ResponseWriter, r *http.Request) {
	var strm = stream.NewStream()
	if !validator.Validate(w, r, contraints.StreamStart{Stream: strm}) {
		return
	}
	err := Live.SetStream(strm)
	if err != nil {
		zap.L().Error("cant start stream",
			zap.String("stream", strm.Name),
			zap.String("error", err.Error()),
		)
		response.Json(w, err.Error(), http.StatusBadRequest)
		return
	}

	strm.Start()
	response.Json(w, "Stream starting...", http.StatusOK)
}

func streamStop(w http.ResponseWriter, r *http.Request) {
	type dataStream struct {
		Name string
	}
	var ds dataStream

	err := helpers.JsonRequestToMap(r, &ds)
	if err != nil {
		zap.L().Error("error while parse request",
			zap.String("stream", ds.Name),
			zap.String("error", err.Error()),
		)
		response.Json(w, "error while parse request", http.StatusBadRequest)
		return
	}

	strm, err := Live.DeleteStream(ds.Name)
	if err != nil {
		zap.L().Error("error stopping stream",
			zap.String("error", err.Error()),
		)
		response.Json(w, err.Error(), http.StatusBadRequest)
		return
	}
	strm.Stop()
	response.Json(w, "Stream stopping...", http.StatusOK)
}

func streams(w http.ResponseWriter, r *http.Request) {
	response.JsonStruct(w, Live.AllStreams(), http.StatusOK)
}

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Post("/stream-start", streamStart)
	r.Post("/stream-stop", streamStop)
	r.Get("/streams", streams)

	return r
}
