// Command pacman configures the desktop window and starts the game.
package main

import (
	"log"
	"math"

	"pacman-ai/assets"
	"pacman-ai/game"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// Window dimensions must be integers, so fractional scaling is rounded to whole pixels.
	// A scale of 1 matches the internal resolution; OS HiDPI scaling may still apply.
	windowWidth := int(math.Round(float64(game.ScreenWidth) * game.ScreenScale))
	windowHeight := int(math.Round(float64(game.ScreenHeight) * game.ScreenScale))
	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("Pacman")

	icons, err := assets.LoadWindowIcons()
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowIcon(icons)

	// Load required game assets before entering the loop so initialization errors are reported here.
	pacman, err := game.New()
	if err != nil {
		log.Fatal(err)
	}
	if err := ebiten.RunGame(pacman); err != nil {
		log.Fatal(err)
	}
}
