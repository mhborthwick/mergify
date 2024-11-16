package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss"
)

var CLI struct {
	Create struct {
	} `cmd:"" help:"Create a new playlist."`
}

var style = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#7D56F4")).
	PaddingTop(2).
	PaddingLeft(4).
	Width(22)

func main() {
	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "create":
		fmt.Println(style.Render("Hello, World!"))
	default:
		panic(ctx.Command())
	}
}
