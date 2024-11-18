package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss"
)

var style = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#7D56F4")).
	PaddingTop(2).
	PaddingLeft(4).
	Width(22)

var CLI struct {
	Create struct {
	} `cmd:"" help:"Create a new playlist."`
}

func main() {
	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "create":
		fmt.Println(style.Render("Hello, World!"))
	default:
		panic(ctx.Command())
	}
}
