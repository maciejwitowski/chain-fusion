package main

import (
	"log"
	"net/http"
	"strconv"
)

// LittleServer starts a server which doubles a number
// curl to this server and see responses
func LittleServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/double", doubleHandler)

	server := http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}

	log.Println("Server starting on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func doubleHandler(w http.ResponseWriter, r *http.Request) {
	valueStr := r.URL.Query().Get("value")

	log.Println("Received request with value:", valueStr)

	if valueStr == "" {
		http.Error(w, "Missing 'value' param", http.StatusBadRequest)
		return
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		http.Error(w, "'value' must be integer", http.StatusBadRequest)
		return
	}

	doubled := value * 2
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(strconv.Itoa(doubled)))
	if err != nil {
		http.Error(w, "Error writing value", http.StatusInternalServerError)
	}
}
