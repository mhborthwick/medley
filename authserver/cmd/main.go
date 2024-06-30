package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

var (
	source oauth2.TokenSource
	config *oauth2.Config
	state  string
)

func GetRandomString() string {
	return uuid.NewString()
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Placeholder")
}

func Authorize(w http.ResponseWriter, r *http.Request) {
	state = GetRandomString()
	http.Redirect(w, r, config.AuthCodeURL(state), http.StatusSeeOther)
}

func Callback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if code == "" {
		fmt.Println("Unauthorized")
		return
	}
	token, err := config.Exchange(r.Context(), code)
	if err != nil {
		fmt.Println("Unauthorized")
		return
	}
	source = config.TokenSource(context.Background(), token)
	fmt.Println("source:", source) // TODO: store in Redis
	fmt.Fprint(w, token.AccessToken)
}

func APIToken(w http.ResponseWriter, r *http.Request) {
	// TODO: get token from Redis
	token, err := source.Token()
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "could not retrieve token")
		return
	}
	source = config.TokenSource(context.Background(), token)
	fmt.Println("source:", source) // TODO: store in Redis
	w.Header().Set("Content-Type", "application/json")
	tokenResponse := map[string]string{
		"access_token": token.AccessToken,
	}
	json.NewEncoder(w).Encode(tokenResponse)

}

func main() {
	id := os.Getenv("CLIENT_ID")
	secret := os.Getenv("CLIENT_SECRET")

	if id == "" || secret == "" {
		fmt.Println("Missing Env Vars")
		return
	}

	config = &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		RedirectURL:  "http://localhost:1337/callback",
		Endpoint:     spotify.Endpoint,
		Scopes: []string{
			"user-read-email",
			"user-read-private",
			"playlist-modify-public",
			"playlist-modify-private",
		},
	}

	http.HandleFunc("/", Index)
	http.HandleFunc("/login", Authorize)
	http.HandleFunc("/callback", Callback)
	http.HandleFunc("/api/token", APIToken)

	log.Fatal(http.ListenAndServe(":1337", nil))
}
