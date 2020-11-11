package v1

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/mamau/restream/helpers"
	"github.com/mamau/restream/routes/response"
	"github.com/mamau/restream/routes/validator"
	"github.com/mamau/restream/routes/validator/contraints"
	"github.com/mamau/restream/stream"
	"go.uber.org/zap"
	"net/http"
)

func streamStart(w http.ResponseWriter, r *http.Request) {
	var strm = stream.NewStream()
	if !validator.Validate(w, r, &contraints.StreamStart{Stream: strm}) {
		return
	}
	if err := stream.GetLive().SetStream(strm); err != nil {
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

func streamSchedulingDownload(w http.ResponseWriter, r *http.Request) {
	var strm = stream.NewScheduledStream()
	if !validator.Validate(w, r, &contraints.ScheduleStart{Stream: strm}) {
		return
	}
	strm.ScheduleDownload()
	response.JsonStruct(w, fmt.Sprintf("Stream %v scheduled...", strm.Name), http.StatusOK)
}

func streamStop(w http.ResponseWriter, r *http.Request) {
	type dataStream struct {
		Name string
	}
	var ds dataStream

	if err := helpers.JsonRequestToMap(r, &ds); err != nil {
		zap.L().Error("error while parse request",
			zap.String("stream", ds.Name),
			zap.String("error", err.Error()),
		)
		response.Json(w, "error while parse request", http.StatusBadRequest)
		return
	}

	strm, err := stream.GetLive().DeleteStream(ds.Name)
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
	response.JsonStruct(w, stream.GetLive().AllStreams(), http.StatusOK)
}

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Post("/stream-start", streamStart)
	r.Post("/stream-stop", streamStop)
	r.Post("/stream-schedule-download", streamSchedulingDownload)
	r.Get("/streams", streams)

	return r
}
