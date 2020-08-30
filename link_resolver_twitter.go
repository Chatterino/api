package main

import (
    "bytes"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "os"
    "time"

    "log"
    "net/url"
    "regexp"
    "strings"
    "text/template"
)

const tweeterTooltip = `<div style="text-align: left;">
<b>{{.Name}} (@{{.Username}})</b>
<br>
{{.Text}}
</div>
`

var (
    tweetRegexp = regexp.MustCompile(`(?i)\/.*\/status(?:es)?\/([^\/\?]+)`)
)

type TweetApiResponse struct {
    ID       string `json:"id_str"`
    Text     string `json:"full_text"`
    Entities struct {
        Media []struct {
            URL string `json:"media_url_https"`
        } `json:"media"`
    } `json:"entities"`
    User struct {
        Name     string `json:"name"`
        Username string `json:"screen_name"`
    } `json:"user"`
}

type tweetTooltipData struct {
    Text      string
    Name      string
    Username  string
    Thumbnail string
}

func init() {
    bearerKey, exists := os.LookupEnv("CHATTERINO_API_TWITTER_BEARER_TOKEN")
    if !exists {
       log.Println("No CHATTERINO_API_TWITTER_BEARER_TOKEN specified, won't do special responses for twitter")
       return
    }

    tooltipTemplate, err := template.New("tweetTooltip").Parse(tweeterTooltip)
    if err != nil {
        log.Println("Error initialization tweeter tooltip template:", err)
        return
    }

    load := func(tweetID string, r *http.Request) (interface{}, error, time.Duration) {
        log.Println("[Twitter] GET", tweetID)

        tweetResp, err := getTweetByID(tweetID, bearerKey)
        if err != nil {
            if err.Error() == "404" {
                var response LinkResolverResponse
                json.Unmarshal(rNoLinkInfoFound, &response)

                return &response, nil, 1 * time.Hour
            }
        }

        tweetData := tweet2Tooltip(tweetResp)
        var tooltip bytes.Buffer
        if err := tooltipTemplate.Execute(&tooltip, tweetData); err != nil {
            return &LinkResolverResponse{
                Status:  http.StatusInternalServerError,
                Message: "twitter template error " + clean(err.Error()),
            }, nil, noSpecialDur
        }

        return &LinkResolverResponse{
            Status:    http.StatusOK,
            Tooltip:   tooltip.String(),
            Thumbnail: tweetData.Thumbnail,
        }, nil, noSpecialDur
    }

    cache := newLoadingCache("twitter", load, 24*time.Hour)

    customURLManagers = append(customURLManagers, customURLManager{
        check: func(url *url.URL) bool {
            return strings.HasSuffix(url.Host, ".twitter.com") || url.Host == "twitter.com"
        },
        run: func(url *url.URL) ([]byte, error) {
            tweetID := getTweetIDFromURL(url)
            if tweetID == "" {
                return rNoLinkInfoFound, nil
            }

            apiResponse := cache.Get(tweetID, nil)
            return json.Marshal(apiResponse)
        },
    })
}

func getTweetIDFromURL(url *url.URL) string {
    match := tweetRegexp.FindAllStringSubmatch(url.Path, -1)
    if len(match) > 0 && len(match[0]) == 2 {
        return match[0][1]
    }
    return ""
}

func getTweetByID(id, bearer string) (*TweetApiResponse, error) {
    endpointUrl := fmt.Sprintf("https://api.twitter.com/1.1/statuses/show.json?id=%s&tweet_mode=extended", id)
    req, err := http.NewRequest("GET", endpointUrl, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Authorization", "Bearer "+bearer)
    resp, err := httpClient.Do(req)
    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

    if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
        return nil, fmt.Errorf("%d", resp.StatusCode)
    }

    var tweet *TweetApiResponse
    err = json.NewDecoder(resp.Body).Decode(&tweet)
    if err != nil {
        return nil, errors.New("unable to process response")
    }

    return tweet, nil
}

func tweet2Tooltip(tweet *TweetApiResponse) *tweetTooltipData {
    data := &tweetTooltipData{}
    data.Text = tweet.Text
    data.Name = tweet.User.Name
    data.Username = tweet.User.Username

    if len(tweet.Entities.Media) > 0 {
        data.Thumbnail = tweet.Entities.Media[0].URL
    }

    return data
}
