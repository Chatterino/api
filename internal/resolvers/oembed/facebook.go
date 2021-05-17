package oembed

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/Chatterino/api/pkg/utils"
)

var (
	facebookAppAccessToken string
)

func loadFacebookCredentials() (appID string, appSecret string, exists bool) {
	if appID, exists = utils.LookupEnv("OEMBED_FACEBOOK_APP_ID"); !exists {
		log.Println("No CHATTERINO_API_OEMBED_FACEBOOK_APP_ID specified, won't do special responses for Facebook or Instagram oEmbed")
		return
	}

	if appSecret, exists = utils.LookupEnv("OEMBED_FACEBOOK_APP_SECRET"); !exists {
		log.Println("No CHATTERINO_API_OEMBED_FACEBOOK_APP_SECRET specified, won't do special responses for Facebook or Instagram oEmbed")
		return
	}

	return
}

func initFacebookAppAccessToken(appID, appSecret string) error {
	u, err := url.Parse("https://graph.facebook.com/oauth/access_token")
	if err != nil {
		return err
	}
	queryVariables := url.Values{}

	queryVariables.Set("client_id", appID)
	queryVariables.Set("client_secret", appSecret)
	queryVariables.Set("grant_type", "client_credentials")

	u.RawQuery = queryVariables.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	d := &facebookTokenResponse{}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("[oEmbed] error loading app access token", err)
		return err
	}

	err = json.Unmarshal(bytes, &d)
	if err != nil {
		return err
	}

	facebookAppAccessToken = d.AccessToken

	return nil
}
