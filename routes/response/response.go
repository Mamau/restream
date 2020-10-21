package response

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type JsonResponse struct {
	Message string `json:"message"`
}

func JsonStruct(w http.ResponseWriter, data interface{}, c int) {
	send(w, data, c)
}

func Json(w http.ResponseWriter, message string, c int) {
	jr := JsonResponse{
		Message: message,
	}
	send(w, jr, c)
}

func send(w http.ResponseWriter, v interface{}, c int) {
	dj, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	_, err = fmt.Fprintf(w, "%s", dj)
}
