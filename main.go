// Command pacman configures the desktop window and starts the game.
package main

import (
	"log"
	"math"

	"pacman-ai/assets"
	"pacman-ai/game"

	"github.com/hajimehoshi/ebiten/v2"
)

// main configures the window, loads the scene, and runs Ebitengine's frame loop.
func main() {
	windowWidth := int(math.Round(float64(game.ScreenWidth) * game.ScreenScale))
	windowHeight := int(math.Round(float64(game.ScreenHeight) * game.ScreenScale))
	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("Pac-Man")

	icons, err := assets.LoadWindowIcons()
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowIcon(icons)

	pacman, err := game.New()
	if err != nil {
		log.Fatal(err)
	}
	if err := ebiten.RunGame(pacman); err != nil {
		log.Fatal(err)
	}
}
