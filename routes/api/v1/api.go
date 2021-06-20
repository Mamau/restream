package v1

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/mamau/restream/helpers"
	"github.com/mamau/restream/routes/response"
	"github.com/mamau/restream/routes/validator"
	"github.com/mamau/restream/routes/validator/contraints"
	"github.com/mamau/restream/stream"
	"github.com/mamau/restream/stream/scheduler"
	"log"
	"net/http"
)

func streamStart(w http.ResponseWriter, r *http.Request) {
	var strm = stream.NewStream()
	if !validator.Validate(w, r, &contraints.StreamStart{Stream: strm}) {
		return
	}

	if !strm.Start() {
		response.Json(w, "error while starting", http.StatusBadRequest)
		return
	}
	response.Json(w, "Stream starting...", http.StatusOK)
}

func streamSchedulingDownload(w http.ResponseWriter, r *http.Request) {
	var strm = scheduler.NewScheduledStream()
	if !validator.Validate(w, r, &contraints.ScheduleStart{Stream: strm}) {
		return
	}
	if err := strm.ScheduleDownload(); err != nil {
		response.Json(w, err.Error(), http.StatusBadRequest)
		return
	}
	response.JsonStruct(w, fmt.Sprintf("Stream %v scheduled...", strm.Name), http.StatusOK)
}

func streamStop(w http.ResponseWriter, r *http.Request) {
	type dataStream struct {
		Name string
	}
	var ds dataStream

	if err := helpers.JsonRequestToMap(r, &ds); err != nil {
		log.Printf("error while parse request stream: %s, error: %s\n", ds.Name, err.Error())
		response.Json(w, "error while parse request", http.StatusBadRequest)
		return
	}

	if strm, err := stream.GetLive().GetStream(ds.Name); err == nil {
		strm.Stop()
		response.Json(w, "Stream stopping...", http.StatusOK)
	}
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
