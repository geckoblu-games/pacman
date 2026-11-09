// Package game implements the Ebitengine game loop and tile-based rendering.
package game

import "github.com/hajimehoshi/ebiten/v2"

// Game owns the session state and delegates updates and rendering to ScenePlay.
// Use New to initialize the scene before passing a Game to Ebitengine.
type Game struct {
	// State contains the session record and current match shared with the scene.
	State GameState
	// scenePlay handles input, frame updates, and rendering.
	scenePlay *ScenePlay
}

// New loads the play scene's assets and starts a fresh match in the ready intro.
func New() (*Game, error) {
	game := &Game{}
	scene, err := newScenePlay(&game.State)
	if err != nil {
		return nil, err
	}
	game.scenePlay = scene
	game.State.StartNewMatch()
	return game, nil
}

// Update delegates each simulation tick and its input handling to the play scene.
func (game *Game) Update() error {
	return game.scenePlay.Update()
}

// Draw delegates rendering of the current frame to the play scene.
func (game *Game) Draw(screen *ebiten.Image) {
	game.scenePlay.Draw(screen)
}

// Layout keeps the internal resolution fixed, regardless of the window's dimensions.
// Ebitengine scales this logical screen to the available window area.
func (game *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
