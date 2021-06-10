package main

import (
	"log"
	"net/http"
	"net/url"
	"time"

	defaultresolver "github.com/Chatterino/api/internal/resolvers/default"
	"github.com/Chatterino/api/pkg/config"
	"github.com/go-chi/chi/v5"
)

var (
	httpClient = &http.Client{
		Timeout: 15 * time.Second,
	}
	startTime = time.Now()

	prefix string
)

func mountRouter(r *chi.Mux) *chi.Mux {
	if config.Cfg.BaseURL == "" {
		log.Printf("Listening on %s (Prefix=%s, BaseURL=%s)\n", config.Cfg.BindAddress, prefix, config.Cfg.BaseURL)
		return r
	}

	// figure out prefix from address
	u, err := url.Parse(config.Cfg.BaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		log.Fatal("Scheme must be included in base url")
	}

	prefix = u.Path
	ur := chi.NewRouter()
	ur.Mount(prefix, r)

	log.Printf("Listening on %s (Prefix=%s, BaseURL=%s)\n", config.Cfg.BindAddress, prefix, config.Cfg.BaseURL)

	return ur
}

func listen(bind string, router *chi.Mux) {
	srv := &http.Server{
		Handler:      router,
		Addr:         bind,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func main() {
	router := chi.NewRouter()

	handleTwitchEmotes(router)
	handleHealth(router)
	defaultresolver.Initialize(router, config.Cfg.BaseURL)

	listen(config.Cfg.BindAddress, mountRouter(router))
}
