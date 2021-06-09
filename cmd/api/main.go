package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"time"

	defaultresolver "github.com/Chatterino/api/internal/resolvers/default"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/go-chi/chi/v5"
)

var (
	httpClient = &http.Client{
		Timeout: 15 * time.Second,
	}
	startTime = time.Now()

	bind    = flag.String("l", ":1234", "bind address")
	baseURL = flag.String("b", "", "base url (useful if being proxied through something like nginx). Value needs to be full url up to the application (e.g. https://braize.pajlada.com/chatterino)")

	prefix string
)

func BaseURL() string {
	if *baseURL != "" {
		return *baseURL
	}

	if value, exists := utils.LookupEnv("BASE_URL"); exists {
		return value
	}

	return ""
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func BindAddress() string {
	if isFlagPassed("l") {
		return *bind
	}

	if value, exists := utils.LookupEnv("BIND_ADDRESS"); exists {
		return value
	}

	return ":1234"
}

func mountRouter(r *chi.Mux) *chi.Mux {
	if BaseURL() == "" {
		return r
	}

	// figure out prefix from address
	u, err := url.Parse(BaseURL())
	if err != nil {
		log.Fatal(err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		log.Fatal("Scheme must be included in base url")
	}

	prefix = u.Path
	ur := chi.NewRouter()
	ur.Mount(prefix, r)

	log.Printf("Listening on %s (Prefix=%s, BaseURL=%s)\n", BindAddress(), prefix, BaseURL())

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
	flag.Parse()

	log.Printf("Listening on %s (Prefix=%s, BaseURL=%s)\n", BindAddress(), prefix, BaseURL())

	router := chi.NewRouter()

	handleTwitchEmotes(router)
	handleHealth(router)

	defaultresolver.Initialize(router, BaseURL())

	listen(BindAddress(), mountRouter(router))
}
