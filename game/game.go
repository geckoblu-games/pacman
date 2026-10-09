package game

import "github.com/hajimehoshi/ebiten/v2"

type Game struct {
	debugGrid        *ebiten.Image
	debugAllowedPath *ebiten.Image
}

func (game *Game) Update() error {
	return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
	if DebugAllowedPathEnabled {
		game.drawDebugAllowedPath(screen)
	}
	if DebugGridEnabled {
		if game.debugGrid == nil {
			game.debugGrid = newDebugGrid()
		}
		screen.DrawImage(game.debugGrid, nil)
	}
}

func (game *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
