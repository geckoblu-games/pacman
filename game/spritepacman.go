package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// pacManSprite draws Pac-Man's normal and death animation frames.
type pacManSprite struct {
	// frames are the directional mouth images and shared closed-mouth image.
	frames [9]*ebiten.Image
	// deathFrames are the implosion frames and final spark.
	deathFrames [10]*ebiten.Image
}

// newPacManSprite binds the normal and death frames to the player's renderer.
func newPacManSprite(frames [9]*ebiten.Image, deathFrames [10]*ebiten.Image) *pacManSprite {
	return &pacManSprite{frames: frames, deathFrames: deathFrames}
}

// Draw renders the normal or death frame at Pac-Man's simulation position.
func (sprite *pacManSprite) Draw(screen *ebiten.Image, player PacManState, deathFrame int) {
	frame := player.animationFrame
	if frame < 0 || frame >= len(sprite.frames) {
		frame = pacManClosedMouthFrame
	}
	image := sprite.frames[frame]
	if deathFrame >= 0 && deathFrame < len(sprite.deathFrames) {
		image = sprite.deathFrames[deathFrame]
	}
	if image == nil {
		return
	}
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	size := image.Bounds().Size()
	rotation := 0.0
	if deathFrame >= 0 && deathFrame < 2 {
		switch player.lastDirection {
		case DirectionRight:
			rotation = math.Pi / 2
		case DirectionDown:
			rotation = math.Pi
		case DirectionLeft:
			rotation = -math.Pi / 2
		}
		rotation *= float64(2-deathFrame) / 2
	}
	drawAt := func(x float64) {
		op.GeoM.Reset()
		op.GeoM.Translate(-float64(size.X)/2, -float64(size.Y)/2)
		op.GeoM.Rotate(rotation)
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
