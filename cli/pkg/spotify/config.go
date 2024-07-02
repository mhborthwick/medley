package spotify

type Config struct {
	UserID    string   `pkl:"userID"`
	Token     string   `pkl:"token"`
	Playlists []string `pkl:"playlists"`
}
