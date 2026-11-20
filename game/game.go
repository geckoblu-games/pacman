// Package game implements the Ebitengine game loop and tile-based rendering.
package game

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

// Game owns the session state and delegates updates and rendering to ScenePlay.
// Use New to initialize the scene before passing a Game to Ebitengine.
type Game struct {
	// State contains the session record, the attract-loop scenes and the current match
	// shared with the scene.
	State GameState
	// SaveHighScore optionally persists the record at level completion and game over.
	SaveHighScore func(int) error
	// scenePlay handles input, frame updates, and rendering of the attract and play screens.
	scenePlay *ScenePlay
	// lastPhase tracks phase changes observed at update boundaries.
	lastPhase Phase
}

// New loads the scene's assets and starts the attract loop.
func New() (*Game, error) {
	game := &Game{}
	scene, err := newScenePlay(&game.State)
	if err != nil {
		return nil, err
	}
	game.scenePlay = scene
	game.State.EnterAttract()
	return game, nil
}

// Update delegates simulation and input, saving the record on checkpoint entry.
func (game *Game) Update() error {
	game.saveCheckpoint()
	if err := game.scenePlay.Update(); err != nil {
		return err
	}
	game.saveCheckpoint()
	return nil
}

// saveCheckpoint observes phase entry and reports save errors without stopping play.
func (game *Game) saveCheckpoint() {
	if game.State.Phase == game.lastPhase {
		return
	}
	game.lastPhase = game.State.Phase
	if game.SaveHighScore == nil || (game.State.Phase != PhaseLevelComplete && game.State.Phase != PhaseGameOver) {
		return
	}
	if err := game.SaveHighScore(game.State.HighScore); err != nil {
		log.Printf("save high score: %v", err)
	}
}

// Draw delegates rendering of the current frame to the scene.
func (game *Game) Draw(screen *ebiten.Image) {
	game.scenePlay.Draw(screen)
}

// Layout keeps the internal resolution fixed, regardless of the window's dimensions.
// Ebitengine scales this logical screen to the available window area.
func (game *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
