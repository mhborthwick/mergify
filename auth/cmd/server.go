package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

type AuthServer struct {
	AccessToken string
	// RefreshToken string
	source oauth2.TokenSource
	config *oauth2.Config
	state  string
}

func (server *AuthServer) Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<a href='/login'>Login to Spotify</a>")
}

func (server *AuthServer) Authorize(w http.ResponseWriter, r *http.Request) {
	server.state = GetRandomString()
	http.Redirect(w, r, server.config.AuthCodeURL(server.state), http.StatusSeeOther)
}

func (server *AuthServer) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if code == "" {
		fmt.Println("Unauthorized")
		return
	}
	token, err := server.config.Exchange(r.Context(), code)
	if err != nil {
		fmt.Println("Unauthorized")
		return
	}
	server.source = server.config.TokenSource(context.Background(), token)
	tokenJSON, _ := json.Marshal(token)
	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<script>
				const accessToken = "%s";
				function copyToClipboard() {
					navigator.clipboard.writeText(accessToken)
						.catch(err => alert("Failed to copy token: " + err));
				}
			</script>
		</head>
		<body>
			<div>
				<pre>"%s"</pre>
				<button onclick="copyToClipboard()">Copy Access Token</button>
			</div>
		</body>
		</html>
	`, token.AccessToken, string(tokenJSON))
}

func (server *AuthServer) APIToken(w http.ResponseWriter, r *http.Request) {
	token, err := server.source.Token()
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "Could not retrieve token")
		return
	}
	server.source = server.config.TokenSource(context.Background(), token)
	w.Header().Set("Content-Type", "application/json")
	tokenResponse := map[string]string{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
	}
	json.NewEncoder(w).Encode(tokenResponse)
}

func GetRandomString() string {
	return uuid.NewString()
}

func ExitIfError(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func HasClientCredentials(clientID, clientSecret string) error {
	if clientID == "" {
		return errors.New("client_id is required")
	}
	if clientSecret == "" {
		return errors.New("client_secret is required")
	}
	return nil
}

func main() {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	err := HasClientCredentials(clientID, clientSecret)
	ExitIfError(err)
	server := AuthServer{}
	server.config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:3000/callback",
		Endpoint:     spotify.Endpoint,
		Scopes: []string{
			"user-read-email",
			"user-read-private",
			"playlist-modify-public",
			"playlist-modify-private",
		},
	}

	http.HandleFunc("/", server.Index)
	http.HandleFunc("/login", server.Authorize)
	http.HandleFunc("/callback", server.Callback)
	http.HandleFunc("/api/token", server.APIToken)

	log.Print("Listening on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
