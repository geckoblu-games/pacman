// Command pacman configures the desktop window and starts the game.
package main

import (
	"log"
	"math"

	"pacman-ai/assets"
	"pacman-ai/game"
	"pacman-ai/highscore"

	"github.com/hajimehoshi/ebiten/v2"
)

// main configures the window, restores the record, runs the game, and saves on exit.
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
	path, err := highscore.DefaultPath()
	if err != nil {
		log.Printf("high-score persistence unavailable: %v", err)
	}
	var store *highscore.Store
	if err == nil {
		store = highscore.New(path)
		score, err := store.Load()
		if err != nil {
			log.Printf("load high score: %v", err)
		} else {
			pacman.State.HighScore = score
		}
		pacman.SaveHighScore = store.Save
	}
	runErr := ebiten.RunGame(pacman)
	if store != nil {
		if err := store.Save(pacman.State.HighScore); err != nil {
			log.Printf("save high score on exit: %v", err)
		}
	}
	if runErr != nil {
		log.Fatal(runErr)
	}
}
