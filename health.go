package main

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gorilla/mux"
)

func uptimeString() string {
	return time.Since(startTime).String()
}

func memoryString() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return fmt.Sprintf("Alloc=%v MiB, TotalAlloc=%v MiB, Sys=%v MiB, NumGC=%v",
		m.Alloc/1024/1024,
		m.TotalAlloc/1024/1024,
		m.Sys/1024/1024,
		m.NumGC)
}

func healthUptime(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", uptimeString())
}

func healthMemory(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", memoryString())
}

func healthCombined(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Uptime: %s - Memory: %s", uptimeString(), memoryString())
}

func handleHealth(router *mux.Router) {
	router.HandleFunc("/health/uptime", healthUptime).Methods("GET")
	router.HandleFunc("/health/memory", healthMemory).Methods("GET")
	router.HandleFunc("/health/combined", healthCombined).Methods("GET")
}
