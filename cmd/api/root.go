package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func rootIndex(w http.ResponseWriter, r *http.Request) {
	const text = `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Chatterino API</title>
</head>
<body>
  <h1>Chatterino API</h1>
  <p>
    Source code for this project is hosted on <a href="https://github.com/Chatterino/api">GitHub</a>
  </p>
  <p>
    Links:
    <ul>
      <li><a href="legal/terms-of-service">Terms of Service</a></li>
      <li><a href="legal/privacy-policy">Privacy Policy</a></li>
    </ul>
  </p>
</body>
</html>`
	w.Write([]byte(text))
}

func handleRoot(router *chi.Mux) {
	router.Get("/", rootIndex)
}
