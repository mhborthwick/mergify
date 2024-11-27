package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

var (
	source oauth2.TokenSource
	config *oauth2.Config
	state  string
)

var clientConfig ClientConfig

type ClientConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func GetRandomString() string {
	return uuid.NewString()
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<a href='/login'>Click to Login</a>")
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

func APIToken(w http.ResponseWriter, r *http.Request) {
	token, err := source.Token()
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "Could not retrieve token")
		return
	}
	source = config.TokenSource(context.Background(), token)
	w.Header().Set("Content-Type", "application/json")
	tokenResponse := map[string]string{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
	}
	json.NewEncoder(w).Encode(tokenResponse)

}

func ExitIfError(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func (cfg *ClientConfig) CheckClientCredentials() error {
	if cfg.ClientID == "" {
		return errors.New("client_id is required")
	}
	if cfg.ClientSecret == "" {
		return errors.New("client_secret is required")
	}
	return nil
}

func main() {
	home, err := os.UserHomeDir()
	ExitIfError(err)
	pathToConfig := path.Join(home, ".mergify", "config.json")
	_, err = os.Stat(pathToConfig)
	ExitIfError(err)
	content, _ := os.ReadFile(pathToConfig)
	json.Unmarshal(content, &clientConfig)
	err = clientConfig.CheckClientCredentials()
	ExitIfError(err)
	config = &oauth2.Config{
		ClientID:     clientConfig.ClientID,
		ClientSecret: clientConfig.ClientSecret,
		RedirectURL:  "http://localhost:3000/callback",
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

	log.Print("Listening on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
