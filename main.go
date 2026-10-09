package main

import (
	"log"

	"pacman-ai/assets"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 224
	screenHeight = 288
)

type Game struct{}

func (game *Game) Update() error {
	return nil
}

func (game *Game) Draw(screen *ebiten.Image) {}

func (game *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Pacman")

	icons, err := assets.LoadWindowIcons()
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowIcon(icons)

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
