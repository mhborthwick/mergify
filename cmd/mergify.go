package main

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhborthwick/mergify/pkg/spotify"
)

var style = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#7D56F4")).
	PaddingTop(2).
	PaddingLeft(4).
	Width(16)

var cli CLI

type CLI struct {
	Token     string   `json:"token" hidden:""`
	Playlists []string `json:"playlists" hidden:""`
	Create    struct {
	} `cmd:"" help:"Combines the tracks from the playlists in your CLI config into a new playlist"`
}

func ExitIfError(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func main() {
	homeDir, err := os.UserHomeDir()
	ExitIfError(err)
	pathToConfig := path.Join(homeDir, ".mergify", "config.json")
	_, err = os.Stat(pathToConfig)
	ExitIfError(err)
	ctx := kong.Parse(&cli, kong.Configuration(kong.JSON, pathToConfig))
	switch ctx.Command() {
	case "create":
		fmt.Println(style.Render("Mergify!"))
		s := spotify.Spotify{}
		s.Token = cli.Token
		s.Client = &http.Client{}
		userID, err := s.GetUserID()
		ExitIfError(err)
		playlistIDs, err := s.GetPlaylistIDsByName(userID, cli.Playlists)
		ExitIfError(err)
		trackIDs, err := s.GetPlaylistTrackIDs(playlistIDs)
		ExitIfError(err)
		playlistID, err := s.CreatePlaylist(userID, trackIDs)
		ExitIfError(err)
		_, err = s.AddTracksToPlaylist(playlistID, trackIDs, 100)
		ExitIfError(err)
		url := fmt.Sprintf("Created playlist: https://open.spotify.com/playlist/%s", playlistID)
		text := lipgloss.NewStyle().SetString(url).Bold(true)
		fmt.Println(text)
	default:
		panic(ctx.Command())
	}
}
