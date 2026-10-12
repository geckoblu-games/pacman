// Package game implements the Ebitengine game loop and tile-based rendering.
package game

import (
	"image"

	"pacman-ai/assets"

	"github.com/hajimehoshi/ebiten/v2"
)

// Game owns the loaded images and lazily initializes debug layers.
// Use New to initialize the wall image before passing a Game to Ebitengine.
type Game struct {
	images assets.GameImages
}

// New loads the transparent wall layer once and validates its size for 1:1 rendering.
// walls.png preserves every blue pixel from doc/ReferenceMap.png and restores
// grid-covered pixels where neighboring blue segments continue across the gap.
// All other pixels remain transparent, including the ghost-house doorway.
// Extracting only blue pixels from the reference would discard these repairs.
func New() (*Game, error) {
	images, err := assets.LoadGameImages(image.Pt(ScreenWidth, ScreenHeight))
	if err != nil {
		return nil, err
	}
	return &Game{images: images}, nil
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
	screen.DrawImage(game.images.Walls, nil)
	game.drawCoins(screen)
	game.drawEnergizers(screen)
	if DebugGridEnabled {
		if game.images.DebugGrid == nil {
			game.images.DebugGrid = newDebugGrid()
		}
		screen.DrawImage(game.images.DebugGrid, nil)
	}
}

func (game *Game) drawCoins(screen *ebiten.Image) {
	coinSize := game.images.Coin.Bounds().Size()
	for _, coord := range CoinCoords {
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		x := float64(coord[0]*CellSize) + (float64(CellSize)-float64(coinSize.X))/2
		y := float64(coord[1]*CellSize) + (float64(CellSize)-float64(coinSize.Y))/2
		op.GeoM.Translate(x, y)
		screen.DrawImage(game.images.Coin, op)
	}
}

func (game *Game) drawEnergizers(screen *ebiten.Image) {
	spriteSize := game.images.Energizer.Bounds().Size()
	for _, coord := range EnergizerCoords {
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		x := float64(coord[0]*CellSize) + (float64(CellSize)-float64(spriteSize.X))/2
		y := float64(coord[1]*CellSize) + (float64(CellSize)-float64(spriteSize.Y))/2
		op.GeoM.Translate(x, y)
		screen.DrawImage(game.images.Energizer, op)
	}
}

// Layout keeps the internal resolution fixed, regardless of the window's dimensions.
// Ebitengine scales this logical screen to the available window area.
func (game *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
