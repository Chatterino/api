package main

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5"
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
	w.Write([]byte(uptimeString()))
}

func healthMemory(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(memoryString()))
}

func healthCombined(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Uptime: %s - Memory: %s", uptimeString(), memoryString())))
}

func handleHealth(router *chi.Mux) {
	router.Get("/health/uptime", healthUptime)
	router.Get("/health/memory", healthMemory)
	router.Get("/health/combined", healthCombined)
}
