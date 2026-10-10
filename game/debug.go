package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// drawDebugAllowedPath caches the reference map's gray backgrounds on first use.
// AllowedPath entries are zero-based {column, row} coordinates; cells outside
// the path stay transparent. The cached image assumes the path does not change.
func (game *Game) drawDebugAllowedPath(screen *ebiten.Image) {
	if game.debugAllowedPath == nil {
		game.debugAllowedPath = ebiten.NewImage(ScreenWidth, ScreenHeight)
		pathColor := color.RGBA{R: 79, G: 79, B: 79, A: 255}
		for _, cell := range AllowedPath {
			startX := cell[0] * CellSize
			startY := cell[1] * CellSize
			for pixelY := startY; pixelY < startY+CellSize; pixelY++ {
				for pixelX := startX; pixelX < startX+CellSize; pixelX++ {
					game.debugAllowedPath.Set(pixelX, pixelY, pathColor)
				}
			}
		}
	}
	screen.DrawImage(game.debugAllowedPath, nil)
}

// newDebugGrid creates a transparent dashed overlay, cached by Game.Draw.
// DebugGridDashSize sets equal dash and gap lengths in logical pixels and must be positive.
func newDebugGrid() *ebiten.Image {
	grid := ebiten.NewImage(ScreenWidth, ScreenHeight)
	gridColor := color.RGBA{R: 160, G: 160, B: 160, A: 160}

	// Clamp the outer right and bottom lines to the last drawable pixel.
	for column := 0; column <= GridColumns; column++ {
		pixelX := min(column*CellSize, ScreenWidth-1)
		for pixelY := 0; pixelY < ScreenHeight; pixelY++ {
			if (pixelY/DebugGridDashSize)%2 == 0 {
				grid.Set(pixelX, pixelY, gridColor)
			}
		}
	}

	for row := 0; row <= GridRows; row++ {
		pixelY := min(row*CellSize, ScreenHeight-1)
		for pixelX := 0; pixelX < ScreenWidth; pixelX++ {
			if (pixelX/DebugGridDashSize)%2 == 0 {
				grid.Set(pixelX, pixelY, gridColor)
			}
		}
	}

	return grid
}
