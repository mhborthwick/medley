package spotify

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Spotify struct {
	URL    string
	Token  string
	UserID string
	Client *http.Client
}

type GetPlaylistItemsResponseBody struct {
	Items []struct {
		Track Track `json:"track"`
	} `json:"items"`
}

type CreatePlaylistRequestBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
}

type CreatePlaylistResponseBody struct {
	ID string `json:"id"`
}

type AddItemsToPlaylistRequestBody struct {
	URIs     []string `json:"uris"`
	Position *int     `json:"position"`
}

type DeleteItemsFromPlaylistRequestBody struct {
	Tracks []Track `json:"tracks"`
}

// GetPlaylistItems gets the items (tracks) within a Spotify playlist.
func (s Spotify) GetPlaylistItems(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	token := "Bearer " + s.Token
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	res, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// CreatePlaylist creates a new empty Spotify playlist.
func (s Spotify) CreatePlaylist() (string, error) {
	currentTime := time.Now().Unix()
	currentTimeString := strconv.FormatInt(currentTime, 10)
	name := "Playlist " + currentTimeString
	requestData := CreatePlaylistRequestBody{
		Name:        name,
		Description: "Created with Playlists Combiner - https://github.com/mhborthwick/spotify-playlists-combiner",
		Public:      false,
	}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", s.URL+"/v1/users/"+s.UserID+"/playlists", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	token := "Bearer " + s.Token
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	res, err := s.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var parsed CreatePlaylistResponseBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	return parsed.ID, nil
}

// AddItemsToPlaylist adds items (tracks) to a playlist.
func (s Spotify) AddItemsToPlaylist(uris []string, playlistID string, prepend bool) ([]byte, error) {
	var requestData AddItemsToPlaylistRequestBody
	if prepend {
		position := 0
		requestData = AddItemsToPlaylistRequestBody{
			URIs:     uris,
			Position: &position,
		}
	} else {
		requestData = AddItemsToPlaylistRequestBody{
			URIs: uris,
		}
	}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", s.URL+"/v1/playlists/"+playlistID+"/tracks", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	token := "Bearer " + s.Token
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	res, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// DeleteItemsFromPlaylist deletes items (tracks) from a playlist.
func (s Spotify) DeleteItemsFromPlaylist(uris []string, playlistID string) ([]byte, error) {
	tracks := make([]Track, len(uris))
	for i, uri := range uris {
		tracks[i] = Track{URI: uri}
	}
	requestData := DeleteItemsFromPlaylistRequestBody{
		Tracks: tracks,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", s.URL+"/v1/playlists/"+playlistID+"/tracks", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	token := "Bearer " + s.Token
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	res, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
