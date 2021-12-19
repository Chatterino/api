package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func legalTermsOfService(w http.ResponseWriter, r *http.Request) {
	const text = `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Chatterino API - Terms of Service</title>
</head>
<body>
  <h1>Chatterino API - Terms of Service</h1>
  <p>
    Using the Chatterino API from the Chatterino Client is perfectly fine.
	If you intend on using the Chatterino API from outside the Chatterino Client, or from a modified version of the Chatterino Client, please reach out to the API owner with details of what you want to do to ensure you don't end up DoSing the API.
  </p>
  <p>
    The Chatterino API uses the following external APIs to provide users with tooltips:<br>
    <ul>
    <li>YouTube Videos and Channels through the YouTube Data API. YouTube's Terms of Service can be found here: <a href="https://www.youtube.com/t/terms">https://www.youtube.com/t/terms</a>. Note that no data is forwarded about you the user to the YouTube Data API.</li>
    </ul>
	This list of external APIs is not exhaustive.
  </p>
  <p>
    Terms of Service changelog:
	<ul>
    <li>2021-12-19: Created</li>
	</ul>
  </p>
  <p>
    <a href="privacy-policy">Our Privacy Policy</a>
  </p>
</body>
</html>`
	w.Write([]byte(text))
}

func legalPrivacyPolicy(w http.ResponseWriter, r *http.Request) {
	const text = `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Chatterino API - Privacy Policy</title>
</head>
<body>
  <h1>Chatterino API - Privacy Policy</h1>
  <p>
    The Chatterino API does not collect any personal information from the users of this API.
    When you make a request through Chatterino, we make an external request based on your payload and return the relevant data to you.
  </p>
  <p>
    When a link is sent to our API for resolving, we only store that link to make the tooltip, then cache that link with its data for up to 24 hours to lessen the load on our API and on any external APIs that we use.
	No information about who made the request is stored anywhere.
  </p>
  <p>
    The Chatterino API uses the following external APIs to provide users with tooltips:<br>
    <ul>
    <li>YouTube Videos and Channels through the YouTube Data API. Google's Privacy policy can be found here: <a href="https://policies.google.com/privacy">https://policies.google.com/privacy</a>. Note that no data is forwarded about you the user to the YouTube Data API.</li>
    </ul>
	This list of external APIs is not exhaustive.
  </p>
  <p>
    Privacy Policy changelog:
	<ul>
    <li>2021-12-19: Created</li>
	</ul>
  </p>
  <p>
    <a href="terms-of-service">Our Terms of Service</a>
  </p>
</body>
</html>`
	w.Write([]byte(text))
}

func handleLegal(router *chi.Mux) {
	router.Get("/legal/terms-of-service", legalTermsOfService)
	router.Get("/legal/privacy-policy", legalPrivacyPolicy)
}
