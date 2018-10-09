package main

import (
	"errors"
	"net/http"

	"google.golang.org/api/googleapi/transport"
	youtube "google.golang.org/api/youtube/v3"
)

var youtubeClient *youtube.Service

func initializeYoutubeAPI() (err error) {
	youtubeHTTPClient := &http.Client{
		Transport: &transport.APIKey{Key: os.Getenv("YOUTUBE_API_KEY")},
	}
	youtubeClient, err = youtube.New(youtubeHTTPClient)

	return
}

func getYoutubeVideo(videoID string) (*youtube.Video, error) {
	youtubeResponse, err := youtubeClient.Videos.List("statistics,snippet,contentDetails").Id(videoID).Do()
	if err != nil {
		return nil, err
	}

	if len(youtubeResponse.Items) != 1 {
		return nil, errors.New("Videos response is not size 1")
	}

	return youtubeResponse.Items[0], nil
}
