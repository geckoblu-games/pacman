package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// pacManStartRow is the zero-based maze row where Pac-Man is drawn initially.
	pacManStartRow = 26
)

// pacManSprite draws Pac-Man from the nine first-row atlas cells.
type pacManSprite struct {
	// frames are the directional mouth images and shared closed-mouth image.
	frames [9]*ebiten.Image
}

// newPacManSprite binds the loaded atlas frames to the player's renderer.
func newPacManSprite(frames [9]*ebiten.Image) *pacManSprite {
	return &pacManSprite{frames: frames}
}

// Draw renders the selected mouth frame at Pac-Man's simulation position.
func (sprite *pacManSprite) Draw(screen *ebiten.Image, player PacManState) {
	frame := player.animationFrame
	if frame < 0 || frame >= len(sprite.frames) {
		frame = pacManClosedMouthFrame
	}
	image := sprite.frames[frame]
	if image == nil {
		return
	}
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	size := image.Bounds().Size()
	drawAt := func(x float64) {
		op.GeoM.Reset()
		op.GeoM.Translate(-float64(size.X)/2, -float64(size.Y)/2)
		op.GeoM.Translate(x, player.Y)
		screen.DrawImage(image, op)
	}
	positions, count := pacManDrawPositions(player.X, size.X)
	for _, x := range positions[:count] {
		drawAt(x)
	}
}

// pacManDrawPositions returns the primary and wrapped centers for a clipped sprite.
func pacManDrawPositions(centerX float64, spriteWidth int) ([3]float64, int) {
	positions := [3]float64{centerX}
	count := 1
	halfWidth := float64(spriteWidth) / 2
	if centerX-halfWidth < 0 {
		positions[count] = centerX + ScreenWidth
		count++
	}
	if centerX+halfWidth > ScreenWidth {
		positions[count] = centerX - ScreenWidth
		count++
	}
	return positions, count
}
