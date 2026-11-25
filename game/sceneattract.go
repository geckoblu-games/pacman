package game

import (
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	// attractSmallFontSize is the logical pixel size of the legend's small PTS suffix.
	attractSmallFontSize = 8
	// attractHeaderColumn and attractHeaderRow place the CHARACTER / NICKNAME header.
	attractHeaderColumn, attractHeaderRow = 7, 5
	// attractCastSpriteX is the source top-left x of the table ghost sprites.
	attractCastSpriteX = 36
	// attractCastSpriteCenterOffsetX is the source-pixel offset from the sprite's
	// documented x coordinate to its visual center in the reference image.
	attractCastSpriteCenterOffsetX = 4
	// attractCastSpriteCenterOffsetY is the source-pixel offset from the text row's
	// top to the visual center of its corresponding sprite.
	attractCastSpriteCenterOffsetY = 3
	// attractCharacterColumn is the starting column of each ghost's character description.
	attractCharacterColumn = 7
	// attractNicknameColumn is the starting column of each ghost's quoted nickname.
	attractNicknameColumn = 18
	// attractLegendIconColumn is the column of the legend's dot and energizer icons.
	attractLegendIconColumn = 10
	// attractLegendTextColumn is the starting column of the legend's point values.
	attractLegendTextColumn = 12
	// attractLegendSuffixColumn is the starting column of the legend's small PTS suffix.
	attractLegendSuffixColumn = 15
	// attractDotRow and attractEnergizerRow are the legend's two rows.
	attractDotRow, attractEnergizerRow = 24, 26
	// attractCopyrightColumn and attractCopyrightRow place the copyright notice.
	attractCopyrightColumn, attractCopyrightRow = 4, 31
	// attractCreditColumn and attractCreditRow place the credit counter on the last row.
	attractCreditColumn, attractCreditRow = 2, 35
	// attractCopyright reproduces the arcade's notice; there is no coin mechanism, so the
	// credit counter always reads zero.
	attractCopyright = "© 1980 MIDWAY MFG.CO."
	// attractCredit is the static credit line.
	attractCredit = "CREDIT  0"
)

// attractCastRows holds, in cast order, the text row of each ghost in the table.
var attractCastRows = [4]int{7, 10, 13, 16}

// attractGhostColors holds, in cast order, the text color of each ghost's description.
var attractGhostColors = [4]color.RGBA{
	{R: 255, A: 255},
	{R: 255, G: 184, B: 255, A: 255},
	{G: 255, B: 255, A: 255},
	{R: 255, G: 184, B: 82, A: 255},
}

// attractCopyrightColor is the arcade pink of the copyright notice.
var attractCopyrightColor = color.RGBA{R: 255, G: 184, B: 255, A: 255}

// drawAttract renders the attract screen: the top HUD with the previous match's score,
// the cast table revealed over time, the point legend, credit line, decoy energizer,
// corridor chase (ghosts first, then Pac-Man in front) and capture scores. The
// copyright appears when the chase starts.
func (scene *ScenePlay) drawAttract(screen *ebiten.Image) {
	attract := scene.State.Attract
	scene.drawTopLabel(screen, "1UP", oneUpColumn)
	scene.drawTopLabel(screen, "HIGH SCORE", highScoreColumn)
	scene.drawTopLabel(screen, "2UP", twoUpColumn)
	scene.drawScore(screen, scene.State.LastScore, scoreRightColumn)
	scene.drawScore(screen, scene.State.HighScore, highScoreRightColumn)
	scene.drawAttractText(screen, "CHARACTER / NICKNAME", attractHeaderColumn, attractHeaderRow, color.White)
	for index, row := range attractCastRows {
		if attract.GhostRevealed(index) {
			frame := scene.ghostFrame(GhostKind(index), DirectionRight, GhostScatter, attract.Ticks, false)
			x, y := attractCastSpriteCenter(row)
			drawSpriteCentered(screen, frame, x, y)
		}
		if attract.GhostNamed(index) {
			scene.drawAttractCastText(screen, attractGhostCharacters[index], attractCharacterColumn, row, attractGhostColors[index])
			scene.drawAttractCastText(screen, attractGhostNicknames[index], attractNicknameColumn, row, attractGhostColors[index])
		}
	}
	if attract.LegendVisible() {
		drawSpriteCentered(screen, scene.images.Coin, tileCenter(attractLegendIconColumn), tileCenter(attractDotRow))
		scene.drawAttractPoints(screen, 10, attractDotRow)
		scene.drawAttractPoints(screen, 50, attractEnergizerRow)
	}
	if attract.EnergizerVisible() {
		drawSpriteCentered(screen, scene.images.Energizer, tileCenter(attractLegendIconColumn), tileCenter(attractEnergizerRow))
	}
	if attract.DecoyVisible() {
		drawSpriteCentered(screen, scene.images.Energizer, attractDecoyX(), attractChaseY())
	}
	if attract.CopyrightVisible() {
		scene.drawAttractText(screen, attractCopyright, attractCopyrightColumn, attractCopyrightRow, attractCopyrightColor)
	}
	scene.drawAttractText(screen, attractCredit, attractCreditColumn, attractCreditRow, color.White)
	for index, ghost := range attract.Ghosts {
		if !attract.GhostVisible(index) {
			continue
		}
		frame := scene.ghostFrame(ghost.Kind, attract.GhostDirection(), attract.GhostMode(), attract.Ticks, false)
		drawSpriteCentered(screen, frame, ghost.X, attractChaseY())
	}
	// Pac-Man is drawn after the ghosts so it stays in front when the sprites overlap
	// while it catches them.
	if attract.PacManVisible() {
		scene.pacMan.DrawUnwrapped(screen, attract.PacMan)
	}
	if points := attract.CapturePoints(); points > 0 {
		op := &text.DrawOptions{}
		op.PrimaryAlign = text.AlignCenter
		op.GeoM.Translate(attract.CaptureX, attractChaseY()-float64(highScoreFontSize)/2)
		op.ColorScale.ScaleWithColor(ghostScoreColor)
		text.Draw(screen, strconv.Itoa(points), scene.highScoreFace, op)
	}
}

// attractCastSpriteCenter returns the visual sprite center in logical pixels, matching
// the cast icon and text alignment shown in the attract-mode reference image.
func attractCastSpriteCenter(row int) (float64, float64) {
	x := float64(attractCastSpriteX+attractCastSpriteCenterOffsetX) * attractSourceScale
	y := float64(row*CellSize) + float64(attractCastSpriteCenterOffsetY)*attractSourceScale
	return x, y
}

// drawAttractText renders text at a grid column and row with the HUD's vertical offset.
func (scene *ScenePlay) drawAttractText(screen *ebiten.Image, label string, column, row int, textColor color.Color) {
	scene.drawHUDText(screen, label, column, row*CellSize+highScoreTop, textColor)
}

// drawAttractCastText renders a cast-table label with its visual top at the row's
// top edge, vertically centering it on the corresponding ghost sprite.
func (scene *ScenePlay) drawAttractCastText(screen *ebiten.Image, label string, column, row int, textColor color.Color) {
	scene.drawHUDText(screen, label, column, row*CellSize, textColor)
}

// drawAttractPoints renders a legend value followed by a small PTS suffix aligned to the
// value's baseline.
func (scene *ScenePlay) drawAttractPoints(screen *ebiten.Image, points, row int) {
	scene.drawAttractText(screen, strconv.Itoa(points), attractLegendTextColumn, row, color.White)
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(attractLegendSuffixColumn*CellSize), float64(row*CellSize+highScoreTop+highScoreFontSize-attractSmallFontSize))
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, "PTS", scene.smallFace, op)
}

// tileCenter returns the logical-pixel center coordinate of a zero-based column or row.
func tileCenter(index int) float64 {
	return float64(index*CellSize + CellSize/2)
}
