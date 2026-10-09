package main

import (
	"log"

	"pacman-ai/assets"
	"pacman-ai/game"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("Pacman")

	icons, err := assets.LoadWindowIcons()
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowIcon(icons)

	if err := ebiten.RunGame(&game.Game{}); err != nil {
		log.Fatal(err)
	}
}
