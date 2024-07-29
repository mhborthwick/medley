package spotify

type CreateConfig struct {
	UserID    string   `pkl:"userID"`
	Token     string   `pkl:"token"`
	Playlists []string `pkl:"playlists"`
}

type SyncConfig struct {
	UserID      string   `pkl:"userID"`
	Token       string   `pkl:"token"`
	Playlists   []string `pkl:"playlists"`
	Destination string   `pkl:"destination"`
}
