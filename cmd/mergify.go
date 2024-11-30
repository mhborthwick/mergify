package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss"
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
	Token  string `json:"token"`
	Create struct {
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
		fmt.Println(cli.Token)
	default:
		panic(ctx.Command())
	}
}
