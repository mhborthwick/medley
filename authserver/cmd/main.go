package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

var (
	source oauth2.TokenSource
	config *oauth2.Config
	state  string
	rdb    *redis.Client
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
	err = rdb.HSet(r.Context(), "spotify_token", map[string]interface{}{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
		"expiry":        token.Expiry.Format(time.RFC3339),
	}).Err()
	if err != nil {
		fmt.Println("Failed to store token in Redis:", err)
	}
	fmt.Fprint(w, token.AccessToken)
}

func APIToken(w http.ResponseWriter, r *http.Request) {
	vals, err := rdb.HGetAll(r.Context(), "spotify_token").Result()
	if err != nil || len(vals) == 0 {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "could not retrieve token from redis")
		return
	}
	expiry, err := time.Parse(time.RFC3339, vals["expiry"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "invalid expiry time format in redis")
		return
	}
	token := &oauth2.Token{
		AccessToken:  vals["access_token"],
		RefreshToken: vals["refresh_token"],
		Expiry:       expiry,
	}
	source = config.TokenSource(context.Background(), token)
	newToken, err := source.Token()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "could not retrieve token")
		return
	}
	if newToken.AccessToken != vals["access_token"] {
		err = rdb.HSet(r.Context(), "spotify_token", map[string]interface{}{
			"access_token":  newToken.AccessToken,
			"refresh_token": newToken.RefreshToken,
			"expiry":        newToken.Expiry.Format(time.RFC3339),
		}).Err()
		if err != nil {
			fmt.Println("failed to store token in redis:", err)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	tokenResponse := map[string]string{
		"access_token": newToken.AccessToken,
	}
	json.NewEncoder(w).Encode(tokenResponse)
}

func main() {
	id := os.Getenv("CLIENT_ID")
	secret := os.Getenv("CLIENT_SECRET")
	rdbHost := os.Getenv("REDIS_HOST")

	if id == "" || secret == "" || rdbHost == "" {
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

	rdb = redis.NewClient(&redis.Options{
		Addr:     rdbHost + ":6379",
		Password: "", //TODO: add password
		DB:       0,
	})

	http.HandleFunc("/", Index)
	http.HandleFunc("/login", Authorize)
	http.HandleFunc("/callback", Callback)
	http.HandleFunc("/api/token", APIToken)

	log.Fatal(http.ListenAndServe(":1337", nil))
}
