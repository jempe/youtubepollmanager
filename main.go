package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

var localConfig Configuration

type Configuration struct {
	ChannelID   string `json:"channel_id"`
	MainVideoID string `json:"main_video_id"`
}

func main() {
	file, err := ioutil.ReadFile("./config/config.json")
	fatalError(err)

	json.Unmarshal(file, &localConfig)

	// Replace "PATH_TO_CLIENT_SECRET_JSON_FILE" with the path to your client secret JSON file
	// You can obtain this file by creating a new OAuth 2 client ID in the Google Cloud Console
	// and downloading the JSON file.
	b, err := ioutil.ReadFile("config/client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// Set up the OAuth 2 config and client
	config, err := google.ConfigFromJSON(b, youtube.YoutubeForceSslScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	// Set up the YouTube API service
	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating YouTube client: %v", err)
	}

	// Call the API to get the video's statistics
	videoId := localConfig.MainVideoID
	call := service.Videos.List([]string{"statistics", "snippet"}).Id(videoId)
	response, err := call.Do()
	if err != nil {
		log.Fatalf("Error making API call: %v", err)
	}

	// Print the video's statistics
	video := response.Items[0]
	fmt.Printf("Video: %s\n", video.Snippet.Title)
	fmt.Printf("Views: %d\n", video.Statistics.ViewCount)
	fmt.Printf("Likes: %d\n", video.Statistics.LikeCount)
	fmt.Printf("Dislikes: %d\n", video.Statistics.DislikeCount)
}

// Helper function to get an OAuth 2 client from the config
func getClient(config *oauth2.Config) *http.Client {
	tokenFile := "TOKEN_FILE_NAME.json"
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Helper function to retrieve a token from a local file
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Helper function to retrieve a token from the web
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%s\n and then copy and paste the code from the return URL here", authURL)

	var authCode string
	fmt.Scan(&authCode)

	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Helper function to save a token to a file
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func logAndExit(message string) {
	log.Println(message)
	os.Exit(1)
}

func fatalError(err error) {
	if err != nil {
		logAndExit(err.Error())
	}
}
