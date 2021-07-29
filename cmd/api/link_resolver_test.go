package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	defaultresolver "github.com/Chatterino/api/internal/resolvers/default"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/go-chi/chi/v5"
)

func TestResolveTwitchClip(t *testing.T) {
	router := chi.NewRouter()
	cfg := config.New()
	defaultresolver.Initialize(router, cfg, nil)
	ts := httptest.NewServer(router)
	defer ts.Close()
	fmt.Println(ts.URL)
	const url = `https%3A%2F%2Fclips.twitch.tv%2FGorgeousAntsyPizzaSaltBae`
	res, err := http.Get(ts.URL + "/link_resolver/" + url)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
	var jsonResponse resolver.Response
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		panic(err)
	}
	if jsonResponse.Status != 200 {
		t.Fatal("wrong status from api")
	}
}

func TestResolveTwitchClip2(t *testing.T) {
	router := chi.NewRouter()
	cfg := config.New()
	defaultresolver.Initialize(router, cfg, nil)
	ts := httptest.NewServer(router)
	defer ts.Close()
	const url = `https%3A%2F%2Ftwitch.tv%2Fpajlada%2Fclip%2FGorgeousAntsyPizzaSaltBae`
	// const url = `https%3A%2F%2Fclips.twitch.tv%2FGorgeousAntsyPizzaSaltBaee`
	res, err := http.Get(ts.URL + "/link_resolver/" + url)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var jsonResponse resolver.Response
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		panic(err)
	}
	if jsonResponse.Status != 200 {
		t.Fatal("wrong status from api")
	}
}

func TestResolveYouTubeChannelUserStandard(t *testing.T) {
	router := chi.NewRouter()
	cfg := config.New()
	defaultresolver.Initialize(router, cfg, nil)
	ts := httptest.NewServer(router)
	defer ts.Close()
	fmt.Println(ts.URL)
	const url = `https%3A%2F%2Fwww.youtube.com%2Fuser%2Fpenguinz0`
	res, err := http.Get(ts.URL + "/link_resolver/" + url)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
	var jsonResponse resolver.Response
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		panic(err)
	}
	if jsonResponse.Status != 200 {
		t.Fatal("wrong status from api")
	}
}

func TestResolveYouTubeChannelUserShortened(t *testing.T) {
	router := chi.NewRouter()
	cfg := config.New()
	defaultresolver.Initialize(router, cfg, nil)
	ts := httptest.NewServer(router)
	defer ts.Close()
	fmt.Println(ts.URL)
	const url = `https%3A%2F%2Fwww.youtube.com%2Fc%2FMizkifDaily`
	res, err := http.Get(ts.URL + "/link_resolver/" + url)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
	var jsonResponse resolver.Response
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		panic(err)
	}
	if jsonResponse.Status != 200 {
		t.Fatal("wrong status from api")
	}
}

func TestResolveYouTubeChannelIdentifier(t *testing.T) {
	router := chi.NewRouter()
	cfg := config.New()
	defaultresolver.Initialize(router, cfg, nil)
	ts := httptest.NewServer(router)
	defer ts.Close()
	fmt.Println(ts.URL)
	const url = `https%3A%2F%2Fwww.youtube.com%2Fchannel%2FUCoqDr5RdFOlomTQI2tkaDOA`
	res, err := http.Get(ts.URL + "/link_resolver/" + url)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
	var jsonResponse resolver.Response
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		panic(err)
	}
	if jsonResponse.Status != 200 {
		t.Fatal("wrong status from api")
	}
}

func TestResolve1M(t *testing.T) {
	// var resp *http.Response
	// var err error

	// fmt.Println("test resolve 1M")
	// resp, err = makeRequest("http://speedtest.tele2.net/100MB.zip")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// resp, err = makeRequest("http://httpbin.org/redirect/5")
	// fmt.Println(resp)
	// fmt.Println(resp.Request.URL)
	// resp, err = makeRequest("http://httpbin.org/image")
	// fmt.Println(resp)
	// resp, err = makeRequest("http://speedtest.tele2.net/1MB.zip")
	// fmt.Println(resp)
}

func TestDoRequest(t *testing.T) {
	// var err error
	// var data interface{}

	// data, err = doRequest("http://localhost:3000")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(string(data.([]byte)))

	// data, err = doRequest("http://httpbin.org/redirect/5")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(string(data.([]byte)))

	// data, err = doRequest("http://httpbin.org/redirect/15")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(string(data.([]byte)))

	// data, err = doRequest("http://speedtest.tele2.net/100MB.zip")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// fmt.Println(string(data.([]byte)))
}
