// Package game implements the Ebitengine game loop and tile-based rendering.
package game

import (
	"fmt"

	"pacman-ai/assets"

	"github.com/hajimehoshi/ebiten/v2"
)

// Game owns the permanent wall image and lazily initialized debug layers.
// Use New to initialize the wall image before passing a Game to Ebitengine.
type Game struct {
	walls            *ebiten.Image
	debugGrid        *ebiten.Image
	debugAllowedPath *ebiten.Image
}

// New loads the transparent wall layer once and validates its size for 1:1 rendering.
// walls.png preserves every blue pixel from doc/ReferenceMap.png and restores
// grid-covered pixels where neighboring blue segments continue across the gap.
// All other pixels remain transparent, including the ghost-house doorway.
// Extracting only blue pixels from the reference would discard these repairs.
func New() (*Game, error) {
	walls, err := assets.LoadImage("walls.png")
	if err != nil {
		return nil, err
	}
	bounds := walls.Bounds()
	if bounds.Dx() != ScreenWidth || bounds.Dy() != ScreenHeight {
		return nil, fmt.Errorf("wall image size is %dx%d, expected %dx%d", bounds.Dx(), bounds.Dy(), ScreenWidth, ScreenHeight)
	}
	return &Game{walls: ebiten.NewImageFromImage(walls)}, nil
}

// Update runs once per simulation tick; gameplay state updates are not implemented yet.
func (game *Game) Update() error {
	return nil
}

// Draw composes the maze and enabled debug layers in logical screen coordinates.
func (game *Game) Draw(screen *ebiten.Image) {
	// Debug path backgrounds go below the permanent walls; the debug grid goes above.
	// Each debug layer is independently controlled by its flag in constants.go.
	if DebugAllowedPathEnabled {
		game.drawDebugAllowedPath(screen)
	}
	screen.DrawImage(game.walls, nil)
	if DebugGridEnabled {
		if game.debugGrid == nil {
			game.debugGrid = newDebugGrid()
		}
		screen.DrawImage(game.debugGrid, nil)
	}
}

// Layout keeps the internal resolution fixed, regardless of the window's dimensions.
// Ebitengine scales this logical screen to the available window area.
func (game *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
