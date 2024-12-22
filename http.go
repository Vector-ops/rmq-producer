package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
)

type Error struct {
	Status  int    `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func HttpServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /publish", CORS(handlePublish))

	log.Println("RMQ Publisher server running on port 8000")

	log.Fatal(http.ListenAndServe(":8000", mux))
}

func handlePublish(w http.ResponseWriter, r *http.Request) {

	logger.Info(
		"Received request to publish message",
		slog.String("method", r.Method),
		slog.String("url", r.URL.String()),
		slog.String("remote_addr", r.RemoteAddr),
	)
	message := MessageTriggerEvent{}
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		writeResponse(Error{Status: http.StatusInternalServerError, Message: err.Error()}, w, http.StatusInternalServerError)
		return
	}

	err = PublishMessage(&message)
	if err != nil {
		writeResponse(Error{Status: http.StatusBadRequest, Message: err.Error()}, w, http.StatusBadRequest)
		return
	}

	writeResponse("Message sent successfully", w, http.StatusOK)
}

func writeResponse(payload interface{}, w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	w.Write(data)

}

func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		if r.Method == "OPTIONS" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		next(w, r)
	}
}
