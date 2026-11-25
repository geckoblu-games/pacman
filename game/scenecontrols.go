package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	// controlsTitle is the heading of the controls scene, centered on controlsTitleRow.
	controlsTitle = "CONTROLS"
	// controlsTitleRow is the grid row of the heading.
	controlsTitleRow = 5
	// controlsFirstRow is the grid row of the first entry.
	controlsFirstRow = 9
	// controlsRowSpacing is the number of grid rows between consecutive entries.
	controlsRowSpacing = 3
	// controlsKeyColumn is the starting column of the key and gesture labels; it leaves
	// room for the nine-character `2 FINGERS` label before the descriptions.
	controlsKeyColumn = 5
	// controlsDescriptionColumn is the starting column of the action descriptions, one
	// column clear of the longest label.
	controlsDescriptionColumn = 14
)

// controlsKeyColor is the Pac-Man yellow used for the key labels.
var controlsKeyColor = color.RGBA{R: 255, G: 255, A: 255}

// drawControls renders the controls scene: the top HUD with the previous match's score
// and the credit line as on the presentation, the heading from the first tick, then each
// key or gesture label and its description as ControlsState reveals them, followed by a
// flashing start prompt.
func (scene *ScenePlay) drawControls(screen *ebiten.Image) {
	controls := scene.State.Controls
	scene.drawTopLabel(screen, "1UP", oneUpColumn)
	scene.drawTopLabel(screen, "HIGH SCORE", highScoreColumn)
	scene.drawTopLabel(screen, "2UP", twoUpColumn)
	scene.drawScore(screen, scene.State.LastScore, scoreRightColumn)
	scene.drawScore(screen, scene.State.HighScore, highScoreRightColumn)
	scene.drawAttractText(screen, attractCredit, attractCreditColumn, attractCreditRow, color.White)
	op := &text.DrawOptions{}
	op.PrimaryAlign = text.AlignCenter
	op.GeoM.Translate(float64(ScreenWidth)/2, float64(controlsTitleRow*CellSize+highScoreTop))
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, controlsTitle, scene.highScoreFace, op)
	for index, entry := range controls.Entries() {
		row := controlsFirstRow + index*controlsRowSpacing
		if controls.KeyRevealed(index) {
			scene.drawAttractText(screen, entry.Key, controlsKeyColumn, row, controlsKeyColor)
		}
		if controls.KeyDescribed(index) {
			scene.drawAttractText(screen, entry.Description, controlsDescriptionColumn, row, color.White)
		}
	}
	if controls.StartPromptVisible() {
		op := &text.DrawOptions{}
		op.PrimaryAlign = text.AlignCenter
		startRow := controlsFirstRow + (len(controls.Entries())-1)*controlsRowSpacing +
			(controlsFirstRow - controlsTitleRow)
		op.GeoM.Translate(float64(ScreenWidth)/2, float64(startRow*CellSize+highScoreTop))
		op.ColorScale.ScaleWithColor(controlsKeyColor)
		text.Draw(screen, controls.StartPrompt(), scene.highScoreFace, op)
	}
}
