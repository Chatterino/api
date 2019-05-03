package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func healthUptime(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", time.Since(startTime))
}

func handleHealth(router *mux.Router) {
	router.HandleFunc("/health/uptime", healthUptime).Methods("GET")
}
