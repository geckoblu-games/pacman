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
	// controlsFirstRow is the grid row of the first key entry.
	controlsFirstRow = 9
	// controlsRowSpacing is the number of grid rows between consecutive entries.
	controlsRowSpacing = 3
	// controlsKeyColumn is the starting column of the key labels.
	controlsKeyColumn = 6
	// controlsDescriptionColumn is the starting column of the action descriptions.
	controlsDescriptionColumn = 13
	// controlsStartRow gives the prompt the same gap from the last entry as the title
	// has from the first entry.
	controlsStartRow = controlsFirstRow + (len(ControlsEntries)-1)*controlsRowSpacing +
		(controlsFirstRow - controlsTitleRow)
)

// controlsKeyColor is the Pac-Man yellow used for the key labels.
var controlsKeyColor = color.RGBA{R: 255, G: 255, A: 255}

// drawControls renders the controls scene: the top HUD with the previous match's score
// and the credit line as on the presentation, the heading from the first tick, then each
// key label and its description as ControlsState reveals them, followed by a flashing
// start prompt.
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
	for index, entry := range ControlsEntries {
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
		op.GeoM.Translate(float64(ScreenWidth)/2, float64(controlsStartRow*CellSize+highScoreTop))
		op.ColorScale.ScaleWithColor(controlsKeyColor)
		text.Draw(screen, "PRESS ENTER TO START", scene.highScoreFace, op)
	}
}
