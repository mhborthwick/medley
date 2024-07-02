package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/apple/pkl-go/pkl"
	"github.com/mhborthwick/medley/cli/pkg/spotify"
)

var CLI struct {
	Create struct {
		Path string `arg:"" name:"path" help:"Path to pkl file." type:"path"`
	} `cmd:"" help:"Create playlist."`
}

func handleError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
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

		var cfg spotify.Config
		if err = evaluator.EvaluateModule(context.Background(), pkl.FileSource(CLI.Create.Path), &cfg); err != nil {
			panic(err)
		}

		spotifyClient := spotify.Spotify{
			URL:    "https://api.spotify.com",
			Token:  cfg.Token,
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
			_, err = spotifyClient.AddItemsToPlaylist(p, playlistID)
			handleError(err)
		}

		fmt.Println("Playlist:", "https://open.spotify.com/playlist/"+playlistID)
		fmt.Println("Created in:", time.Since(startNow))
	default:
		panic(ctx.Command())
	}
}
