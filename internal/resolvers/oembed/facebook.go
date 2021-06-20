package oembed

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var (
	facebookAppAccessToken string
)

func initFacebookAppAccessToken(appID string, appSecret string) error {
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
