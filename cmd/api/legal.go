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
    Usage of the Chatterino API from the Chatterino Client is perfectly fine.
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
	If you have any questions about these Terms of Service, you can <a href="https://github.com/Chatterino/api/issues/new">contact us</a>.
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
    The Chatterino API does not share any user data with any internal or external parties.
  <p>
    When a link is sent to our API for resolving, we only store that link to make the tooltip, then cache that link with its data for up to 24 hours to lessen the load on our API and on any external APIs that we use.
	No information about who made the request is stored anywhere.
  </p>
  <p>
	When you make a request to our API through the Chatterino Client, the following data is available to us:
	<ul>
	<li>Your IP address</li>
	<li>The version of your Chatterino Client</li>
	<li>The link you requested information about</li>
	</ul>
	Your IP address is <strong>NOT</strong> processed or stored.<br>
	The version of your Chatterino Client is <strong>NOT</strong> processed or stored, but we maintain the right to process it in the future in case we need to provide information in different formats based on the Chatterino Client version.<br>
	The link you requested information about is processed and stored together with the output data for up to 24 hours as cache to save processing power/bandwidth for us and any external APIs.<br>
	No data other than the link is shared with external APIs, which is strictly necessary for the Chatterino API to function.
  </p>
  <p>
    The Chatterino API uses the following external APIs to provide users with tooltips:
    <ul>
    <li>YouTube Videos and Channels through the YouTube Data API. Google's Privacy Policy can be found here: <a href="https://policies.google.com/privacy">https://policies.google.com/privacy</a>. Note that no data is forwarded about you the user to the YouTube Data API.</li>
    </ul>
	This list of external APIs is not exhaustive.<br>
  </p>
  <p>
	The Chatterino API uses these above-mentioned external APIs to serve content to you, sometimes in a transformative way. Examples:
	<ul>
	<li>When the Chatterino Client makes a request to the Chatterino API about a YouTube video link (e.g. https://www.youtube.com/watch?v=7fnTxm1u_34), we will make a request about that video (7fnTxm1u_34) to the YouTube Data API and read data such as the video title, the channel the video is uploaded by, the duration of the video, thumbnail link, and more to provide the user with that data in a tooltip format like this: <pre>William Onyeabor - When the Going is Smooth &amp; Good (Official Audio)
Channel: Luaka Bop
Duration: 00:12:55
Published: 11 Aug 2015
Views: 425,717
4,604 likes</pre></li>
	</ul>
  </p>
  <p>
	If you have any questions about this Privacy Policy, you can <a href="https://github.com/Chatterino/api/issues/new">contact us</a>.
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
