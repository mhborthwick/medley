package spotify

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPlaylistItems(t *testing.T) {
	t.Run("returns body and nil", func(t *testing.T) {
		mockResponse := []byte(``)
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(mockResponse)
		}))
		defer mockServer.Close()
		spotifyClient := Spotify{
			URL:    mockServer.URL,
			Token:  "token",
			UserID: "me",
			Client: &http.Client{},
		}
		url := fmt.Sprintf("%s/v1/playlists/%s/tracks", spotifyClient.URL, "123")
		data, err := spotifyClient.GetPlaylistItems(url)
		assert.Equal(t, mockResponse, data)
		assert.Nil(t, err)
	})
}

func TestCreatePlaylist(t *testing.T) {
	t.Run("returns id and nil", func(t *testing.T) {
		mockResponse := []byte(`{"id": "123"}`)
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(mockResponse)
		}))
		defer mockServer.Close()
		spotifyClient := Spotify{
			URL:    mockServer.URL,
			Token:  "token",
			UserID: "me",
			Client: &http.Client{},
		}
		data, err := spotifyClient.CreatePlaylist()
		assert.Equal(t, "123", data)
		assert.Nil(t, err)
	})
}

func TestAddItemsToPlaylist(t *testing.T) {
	t.Run("returns body and nil", func(t *testing.T) {
		mockResponse := []byte(``)
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(mockResponse)
		}))
		defer mockServer.Close()
		spotifyClient := Spotify{
			URL:    mockServer.URL,
			Token:  "token",
			UserID: "me",
			Client: &http.Client{},
		}
		data, err := spotifyClient.AddItemsToPlaylist([]string{"abc", "def"}, "123")
		assert.Equal(t, mockResponse, data)
		assert.Nil(t, err)
	})
}
