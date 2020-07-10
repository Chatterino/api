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
<b>{{.Text}}</b>
<br><b>Author:</b> {{.Username}}
</div>
`

var (
    tweetRegexp = regexp.MustCompile(`(?i)\/.*\/status(?:es)?\/([^\/\?]+)`)
)

type TweetApiResponse struct {
    Data []struct {
        ID          string `json:"id"`
        AuthorID    string `json:"author_id"`
        Text        string `json:"text"`
        Attachments struct {
            MediaKeys []string `json:"media_keys"`
        } `json:"attachments"`
    } `json:"data"`
    Includes struct {
        Users []struct {
            Name     string `json:"name"`
            ID       string `json:"id"`
            Username string `json:"username"`
        } `json:"users"`
    } `json:"includes"`
    Errors []struct {
        Detail       string `json:"detail"`
        Title        string `json:"title"`
        ResourceType string `json:"resource_type"`
        Parameter    string `json:"parameter"`
        Value        string `json:"value"`
        Type         string `json:"type"`
    } `json:"errors"`
}

type tweetTooltipData struct {
    Text      string
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
            return &LinkResolverResponse{
                Status:  http.StatusInternalServerError,
                Message: "twitter error: " + clean(err.Error()),
            }, nil, 1 * time.Hour
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
    endpointUrl := fmt.Sprintf("https://api.twitter.com/labs/2/tweets?ids=%s&expansions=author_id", id)
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
        return nil, fmt.Errorf("responded with status %d", resp.StatusCode)
    }

    var tweet *TweetApiResponse
    err = json.NewDecoder(resp.Body).Decode(&tweet)
    if err != nil {
        return nil, errors.New("unable to process response")
    }

    if len(tweet.Errors) > 0 {
        log.Println("twitter err:", tweet.Errors)
        return nil, errors.New(strings.ToLower(tweet.Errors[0].Title))
    }

    return tweet, nil
}

func tweet2Tooltip(tweet *TweetApiResponse) *tweetTooltipData {
    data := &tweetTooltipData{}
    if len(tweet.Data) > 0 {
        data.Text = tweet.Data[0].Text
        for _, user := range tweet.Includes.Users {
            if user.ID == tweet.Data[0].AuthorID {
                data.Username = user.Username
                break
            }
        }
    }

    return data
}
