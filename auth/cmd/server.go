package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

type AuthServer struct {
	// accessToken string
	// refreshToken string
	client *http.Client
	source oauth2.TokenSource
	config *oauth2.Config
	state  string
}

const API = "https://api.spotify.com/v1"

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

func (server *AuthServer) Me(w http.ResponseWriter, r *http.Request) {
	if server.client == nil {
		server.client = &http.Client{}
	}
	token, err := server.source.Token()
	if err != nil {
		http.Error(w, "Could not retrieve token", http.StatusForbidden)
		return
	}
	req, err := http.NewRequest("GET", API+"/me", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := server.client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (server *AuthServer) GetPlaylistIDs(w http.ResponseWriter, r *http.Request) {
	if server.client == nil {
		server.client = &http.Client{}
	}
	token, err := server.source.Token()
	if err != nil {
		http.Error(w, "Could not retrieve token", http.StatusForbidden)
		return
	}
	endpoint := r.URL.RequestURI()
	req, err := http.NewRequest("GET", API+endpoint, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := server.client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	r := mux.NewRouter()

	r.HandleFunc("/", server.Index)
	r.HandleFunc("/login", server.Authorize)
	r.HandleFunc("/callback", server.Callback)
	r.HandleFunc("/api/token", server.APIToken)

	r.HandleFunc("/me", server.Me)
	r.HandleFunc("/users/{user}/playlists", server.GetPlaylistIDs)

	log.Print("Listening on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}
