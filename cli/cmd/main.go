package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/alecthomas/kong"
	"github.com/apple/pkl-go/pkl"
	"github.com/mhborthwick/medley/cli/pkg/spotify"
)

var CLI struct {
	Create struct {
		Path string `arg:"" name:"path" help:"Path to pkl file." type:"path"`
	} `cmd:"" help:"Create playlist."`
	Sync struct {
		Path string `arg:"" name:"path" help:"Path to pkl file." type:"path"`
	} `cmd:"" help:"Sync playlist."`
}

func handleError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

func GetToken() (string, error) {
	client := &http.Client{}
	url := "http://localhost:1337/api/token"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var parsed TokenResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	return parsed.AccessToken, nil
}

func main() {
	ctx := kong.Parse(&CLI)
	evaluator, err := pkl.NewEvaluator(context.Background(), pkl.PreconfiguredOptions)
	if err != nil {
		panic(err)
	}
	defer evaluator.Close()
	switch ctx.Command() {
	case "create <path>":
		startNow := time.Now()
		fmt.Println("Evaluating from: " + CLI.Create.Path)

		var cfg spotify.CreateConfig
		if err = evaluator.EvaluateModule(context.Background(), pkl.FileSource(CLI.Create.Path), &cfg); err != nil {
			panic(err)
		}

		// get token from authserver
		token, err := GetToken()
		handleError(err)

		spotifyClient := spotify.Spotify{
			URL:    "https://api.spotify.com",
			Token:  token,
			UserID: cfg.UserID,
			Client: &http.Client{},
		}

		var all []string

		for _, p := range cfg.Playlists {
			id, err := spotify.GetID(p)
			handleError(err)
			baseURL := fmt.Sprintf("%s/v1/playlists/%s/tracks", spotifyClient.URL, id)
			nextURL := baseURL
			// you have to paginate these requests
			// because spotify caps you at 20 songs per request
			for nextURL != "" {
				body, err := spotifyClient.GetPlaylistItems(nextURL)
				handleError(err)
				uris, err := spotify.GetURIs(body)
				handleError(err)
				all = append(all, uris...)
				nextURL, err = spotify.GetNextURL(body)
				handleError(err)
			}
		}

		playlistID, err := spotifyClient.CreatePlaylist()
		handleError(err)

		// cleans duplicate songs
		uniqueURIsMap := make(map[string]bool)
		unique := make([]string, 0, len(uniqueURIsMap))
		for _, uri := range all {
			if _, found := uniqueURIsMap[uri]; !found {
				uniqueURIsMap[uri] = true
				unique = append(unique, uri)
			}
		}

		var payloads [][]string

		// creates multiple payloads with <=100 songs to send in batches
		// because spotify caps you at 100 songs per request
		for len(unique) > 0 {
			var payload []string
			if len(unique) >= 100 {
				payload, unique = unique[:100], unique[100:]
			} else {
				payload, unique = unique, nil
			}
			payloads = append(payloads, payload)
		}

		for _, p := range payloads {
			_, err = spotifyClient.AddItemsToPlaylist(p, playlistID, false)
			handleError(err)
		}

		fmt.Println("Playlist:", "https://open.spotify.com/playlist/"+playlistID)
		fmt.Println("Created in:", time.Since(startNow))
	case "sync <path>":
		startNow := time.Now()
		fmt.Println("Evaluating from: " + CLI.Sync.Path)

		var cfg spotify.SyncConfig
		if err = evaluator.EvaluateModule(context.Background(), pkl.FileSource(CLI.Sync.Path), &cfg); err != nil {
			panic(err)
		}

		// get token from authserver
		token, err := GetToken()
		handleError(err)

		spotifyClient := spotify.Spotify{
			URL:    "https://api.spotify.com",
			Token:  token,
			UserID: cfg.UserID,
			Client: &http.Client{},
		}

		// get all uris from target playlist
		var target []string

		id, err := spotify.GetID(cfg.Destination)
		handleError(err)
		baseURL := fmt.Sprintf("%s/v1/playlists/%s/tracks", spotifyClient.URL, id)
		nextURL := baseURL
		for nextURL != "" {
			body, err := spotifyClient.GetPlaylistItems(nextURL)
			handleError(err)
			uris, err := spotify.GetURIs(body)
			handleError(err)
			target = append(target, uris...)
			nextURL, err = spotify.GetNextURL(body)
			handleError(err)
		}

		// create target map
		targetMap := make(map[string]bool)
		for _, t := range target {
			targetMap[t] = false
		}

		// get all uris from provided playlists
		var all []string

		for _, p := range cfg.Playlists {
			id, err := spotify.GetID(p)
			handleError(err)
			baseURL := fmt.Sprintf("%s/v1/playlists/%s/tracks", spotifyClient.URL, id)
			nextURL := baseURL
			// you have to paginate these requests
			// because spotify caps you at 20 songs per request
			for nextURL != "" {
				body, err := spotifyClient.GetPlaylistItems(nextURL)
				handleError(err)
				uris, err := spotify.GetURIs(body)
				handleError(err)
				all = append(all, uris...)
				nextURL, err = spotify.GetNextURL(body)
				handleError(err)
			}
		}

		// if uri in target playlist
		// set value to true
		// if not, add to toAdd slice
		toAdd := []string{}

		for _, a := range all {
			_, ok := targetMap[a]
			if ok {
				targetMap[a] = true
			} else {
				toAdd = append(toAdd, a)
			}
		}

		// cleans duplicate "toAdd songs"
		// we won't need to clean "toRemove songs"
		// because the way we evaluate them
		uniqueToAddMap := make(map[string]bool)
		uniqueToAdd := make([]string, 0, len(uniqueToAddMap))
		for _, uri := range toAdd {
			if _, found := uniqueToAddMap[uri]; !found {
				uniqueToAddMap[uri] = true
				uniqueToAdd = append(uniqueToAdd, uri)
			}
		}

		// creates multiple payloads with <=100 songs to send in batches
		// because spotify caps you at 100 songs per request
		var toAddPayloads [][]string

		for len(uniqueToAdd) > 0 {
			var payload []string
			if len(uniqueToAdd) >= 100 {
				payload, uniqueToAdd = uniqueToAdd[:100], uniqueToAdd[100:]
			} else {
				payload, uniqueToAdd = uniqueToAdd, nil
			}
			toAddPayloads = append(toAddPayloads, payload)
		}

		fmt.Println("adding", toAddPayloads)

		// get values still set to false
		// these should be deleted
		toRemove := []string{}

		for k, v := range targetMap {
			if !v {
				toRemove = append(toRemove, k)
			}
		}

		// creates multiple payloads with <=100 songs to send in batches
		// because spotify caps you at 100 songs per request
		var toRemovePayloads [][]string
		for len(toRemove) > 0 {
			var payload []string
			if len(toRemove) >= 100 {
				payload, toRemove = toRemove[:100], toRemove[100:]
			} else {
				payload, toRemove = toRemove, nil
			}
			toRemovePayloads = append(toRemovePayloads, payload)
		}

		fmt.Println("removing", toRemovePayloads)

		// handle deletion
		for _, p := range toRemovePayloads {
			_, err = spotifyClient.DeleteItemsFromPlaylist(p, cfg.Destination)
			handleError(err)
		}

		// reverse items in toAddPayloads
		slices.Reverse(toAddPayloads)

		// handle addition
		for _, p := range toAddPayloads {
			_, err = spotifyClient.AddItemsToPlaylist(p, cfg.Destination, true)
			handleError(err)
		}
		fmt.Println("Playlist:", "https://open.spotify.com/playlist/"+cfg.Destination)
		fmt.Println("Created in:", time.Since(startNow))
	default:
		panic(ctx.Command())
	}
}
