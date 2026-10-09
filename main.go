package main

import (
	"log"
	"math"

	"pacman-ai/assets"
	"pacman-ai/game"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	windowWidth := int(math.Round(float64(game.ScreenWidth) * game.ScreenScale))
	windowHeight := int(math.Round(float64(game.ScreenHeight) * game.ScreenScale))
	ebiten.SetWindowSize(windowWidth, windowHeight)
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
