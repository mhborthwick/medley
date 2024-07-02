package spotify

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetID(t *testing.T) {
	t.Run("returns ID", func(t *testing.T) {
		valid := "https://open.spotify.com/playlist/5JusAvFUoQRDJ4K6UOjr4O"
		id, _ := GetID(valid)
		assert.Equal(t, id, "5JusAvFUoQRDJ4K6UOjr4O")
	})

	t.Run("returns nil", func(t *testing.T) {
		valid := "https://open.spotify.com/playlist/5JusAvFUoQRDJ4K6UOjr4O"
		_, err := GetID(valid)
		assert.Nil(t, err)
	})

	t.Run("returns empty string", func(t *testing.T) {
		invalid := "invalid"
		id, _ := GetID(invalid)
		assert.Equal(t, id, "")
	})

	t.Run("returns error", func(t *testing.T) {
		invalid := "invalid"
		_, err := GetID(invalid)
		assert.EqualError(t, err, "invalid playlist")
	})
}

func TestGetURIs(t *testing.T) {
	t.Run("returns uris", func(t *testing.T) {
		data := GetPlaylistItemsResponseBody{
			Items: []struct {
				Track Track `json:"track"`
			}{
				{
					Track: Track{URI: "123"},
				},
				{
					Track: Track{URI: "abc"},
				},
			},
		}
		jsonData, _ := json.Marshal(data)
		uris, _ := GetURIs(jsonData)
		assert.Equal(t, uris, []string{
			"123",
			"abc",
		})
	})

	t.Run("returns nil", func(t *testing.T) {
		data := GetPlaylistItemsResponseBody{
			Items: []struct {
				Track Track `json:"track"`
			}{
				{
					Track: Track{URI: "123"},
				},
				{
					Track: Track{URI: "abc"},
				},
			},
		}
		jsonData, _ := json.Marshal(data)
		_, err := GetURIs(jsonData)
		assert.Nil(t, err)
	})

	t.Run("returns nil", func(t *testing.T) {
		invalidJSON := []byte(`invalid`)
		err, _ := GetURIs(invalidJSON)
		assert.Nil(t, err)
	})

	t.Run("returns err", func(t *testing.T) {
		invalidJSON := []byte(`invalid`)
		_, err := GetURIs(invalidJSON)
		assert.Error(t, err)
	})
}
