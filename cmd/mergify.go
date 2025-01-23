package main

import (
	"errors"
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
	Token     string   `json:"token"`
	Playlists []string `json:"playlists"`
	Create    struct {
	} `cmd:"" help:"Create a new playlist."`
}

func ExitIfError(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func (cli *CLI) HasToken() error {
	if cli.Token == "" {
		return errors.New("token is required")
	}
	return nil
}

func main() {
	homeDir, err := os.UserHomeDir()
	ExitIfError(err)
	pathToConfig := path.Join(homeDir, ".mergify", "config.json")
	_, err = os.Stat(pathToConfig)
	ExitIfError(err)
	ctx := kong.Parse(&cli, kong.Configuration(kong.JSON, pathToConfig))
	err = cli.HasToken()
	ExitIfError(err)
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
		playlistID, err := s.CreatePlaylist(userID)
		ExitIfError(err)
		temp, err := s.AddTracksToPlaylist(playlistID, trackIDs, 100)
		fmt.Println(temp)
		ExitIfError(err)
	default:
		panic(ctx.Command())
	}
}
