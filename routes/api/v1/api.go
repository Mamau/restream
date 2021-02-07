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
	"io/ioutil"
	"log"
	"net/http"
)

func streamStart(w http.ResponseWriter, r *http.Request) {
	var strm = stream.NewStream()
	if !validator.Validate(w, r, &contraints.StreamStart{Stream: strm}) {
		return
	}

	if err := strm.Start(); err != nil {
		response.Json(w, err.Error(), http.StatusBadRequest)
		return
	}
	response.Json(w, "Stream starting...", http.StatusOK)
}

func startChannel(w http.ResponseWriter, r *http.Request) {
	var strm = stream.NewStream()
	if !validator.Validate(w, r, &contraints.ChannelStart{Stream: strm}) {
		return
	}

	if err := strm.StartViaSelenium(false); err != nil {
		response.Json(w, err.Error(), http.StatusBadRequest)
		return
	}

	response.Json(w, fmt.Sprintf("Channel %s starting...", strm.Name), http.StatusOK)
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

func index(w http.ResponseWriter, r *http.Request) {
	indexFile, _ := ioutil.ReadFile("./dist/index.html")
	_, err := w.Write(indexFile)
	if err != nil {
		log.Fatalf("Error wrtie response %v", err)
	}
}

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/player", index)
	r.Post("/stream-start", streamStart)
	r.Post("/start-channel", startChannel)
	r.Post("/stream-stop", streamStop)
	r.Post("/stream-schedule-download", streamSchedulingDownload)
	r.Get("/streams", streams)

	return r
}
