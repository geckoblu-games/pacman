package game

import (
	"fmt"
	"image"
	"image/color"
	"strconv"

	"pacman/assets"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	// highScoreFontSize is the logical pixel size of the HUD font.
	highScoreFontSize = 14
	// highScoreColumn is the zero-based starting column of the record label.
	highScoreColumn = 9
	// oneUpColumn is the zero-based starting column of the player label.
	oneUpColumn = 3
	// twoUpColumn is the zero-based starting column of the arcade's second-player label.
	twoUpColumn = 22
	// scoreRightColumn is the zero-based column whose right edge aligns the score.
	scoreRightColumn = 6
	// highScoreRightColumn aligns the record to the right edge of this column.
	highScoreRightColumn = 15
	// highScoreTop is the top HUD label offset in logical pixels.
	highScoreTop = 2
	// readyRow is the zero-based grid row of the corridor below the ghost house where READY! is centered.
	readyRow = 20
)

// readyColor is the yellow used for the READY! intro message.
var readyColor = color.RGBA{R: 255, G: 255, B: 0, A: 255}

// fruitScoreColor is the arcade pink used for the bonus score popup.
var fruitScoreColor = color.RGBA{R: 255, G: 184, B: 255, A: 255}

// ghostScoreColor is the light blue used for captured-ghost scores.
var ghostScoreColor = color.RGBA{G: 255, B: 255, A: 255}

// ScenePlay handles input and renders the attract screen or the maze, actors,
// collectibles, and HUD of a match.
type ScenePlay struct {
	// State references the session and match owned by Game.
	State *GameState
	// images contains loaded sprites and lazily generated debug overlays.
	images assets.GameImages
	// highScoreFace is the reusable font face for HUD labels, scores and messages.
	highScoreFace *text.GoTextFace
	// smallFace is the half-height font face for the attract legend's PTS suffix.
	smallFace *text.GoTextFace
	// pacMan draws the player's normal and death sprites.
	pacMan *pacManSprite
	// sounds plays the simulation's cues and background loops.
	sounds *soundPlayer
}

// newScenePlay loads rendering and audio resources and binds the scene to the shared state.
func newScenePlay(state *GameState) (*ScenePlay, error) {
	images, err := assets.LoadGameImages(image.Pt(ScreenWidth, ScreenHeight))
	if err != nil {
		return nil, err
	}
	highScoreFace, err := assets.LoadTextFace("PressStart2P-Regular.ttf", highScoreFontSize)
	if err != nil {
		return nil, err
	}
	smallFace, err := assets.LoadTextFace("PressStart2P-Regular.ttf", attractSmallFontSize)
	if err != nil {
		return nil, err
	}
	sounds, err := newSoundPlayer()
	if err != nil {
		return nil, err
	}
	return &ScenePlay{
		State:         state,
		images:        images,
		highScoreFace: highScoreFace,
		smallFace:     smallFace,
		pacMan:        newPacManSprite(images.PacManFrames, images.PacManDeathFrames),
		sounds:        sounds,
	}, nil
}

// Update handles input, advances timers, updates actors during active play, and then
// presents the resulting sound cues and background loop. Enter replaces the arcade's
// coin insert: from the attract screen or GAME OVER it starts a new match immediately.
// M toggles mute.
func (scene *ScenePlay) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && (scene.State.Match == nil || scene.State.Phase == PhaseGameOver) {
		scene.State.StartNewMatch()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		scene.State.TogglePause()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		scene.sounds.ToggleMute()
	}
	if !scene.State.Paused {
		previousPhase := scene.State.Phase
		scene.State.Tick()
		if scene.State.Phase == PhasePlaying && previousPhase != PhaseGhostCaptured && scene.State.Match != nil {
			scene.State.UpdateActors(scene.pacManDirection(), GhostMovementEnabled)
		}
	}
	scene.sounds.Update(scene.State)
	return nil
}

// pacManDirection maps held arrow or WASD keys to one buffered cardinal direction.
func (scene *ScenePlay) pacManDirection() Direction {
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyUp), ebiten.IsKeyPressed(ebiten.KeyW):
		return DirectionUp
	case ebiten.IsKeyPressed(ebiten.KeyLeft), ebiten.IsKeyPressed(ebiten.KeyA):
		return DirectionLeft
	case ebiten.IsKeyPressed(ebiten.KeyDown), ebiten.IsKeyPressed(ebiten.KeyS):
		return DirectionDown
	case ebiten.IsKeyPressed(ebiten.KeyRight), ebiten.IsKeyPressed(ebiten.KeyD):
		return DirectionRight
	default:
		return DirectionNone
	}
}

// Draw renders the attract screen in PhaseAttract. Otherwise it composes debug
// backgrounds, walls, collectibles, the bonus or its score popup, Pac-Man, ghosts, the
// grid, the HUD, and the READY! intro or PAUSED message. The maze uses the white walls
// during white flash segments.
func (scene *ScenePlay) Draw(screen *ebiten.Image) {
	if scene.State.Phase == PhaseAttract {
		scene.drawAttract(screen)
		return
	}
	if DebugAllowedPathEnabled {
		scene.drawDebugAllowedPath(screen)
	}
	if scene.State.MazeFlashWhite() {
		screen.DrawImage(scene.images.WallsWhite, nil)
	} else {
		screen.DrawImage(scene.images.Walls, nil)
	}
	scene.drawCoins(screen)
	scene.drawEnergizers(screen)
	scene.drawFruit(screen)
	if scene.State.Match != nil {
		if scene.State.PacManVisible() {
			scene.pacMan.Draw(screen, scene.State.Match.PacMan, scene.State.DeathFrame())
		}
		if scene.State.GhostVisible(GhostBlinky) {
			scene.drawGhost(screen, scene.State.Match.Blinky, scene.State.Match.Level)
		}
		if scene.State.GhostVisible(GhostPinky) {
			scene.drawGhost(screen, scene.State.Match.Pinky, scene.State.Match.Level)
		}
		if scene.State.GhostVisible(GhostInky) {
			scene.drawGhost(screen, scene.State.Match.Inky, scene.State.Match.Level)
		}
		if scene.State.GhostVisible(GhostClyde) {
			scene.drawGhost(screen, scene.State.Match.Clyde, scene.State.Match.Level)
		}
	}
	scene.drawGhostCaptureScore(screen)
	if DebugGridEnabled {
		if scene.images.DebugGrid == nil {
			scene.images.DebugGrid = newDebugGrid()
		}
		screen.DrawImage(scene.images.DebugGrid, nil)
	}
	scene.drawTopLabel(screen, "1UP", oneUpColumn)
	scene.drawTopLabel(screen, "HIGH SCORE", highScoreColumn)
	scene.drawTopLabel(screen, "2UP", twoUpColumn)
	scene.drawScore(screen, scene.State.HighScore, highScoreRightColumn)
	if scene.State.Match != nil {
		scene.drawScore(screen, scene.State.Match.Score, scoreRightColumn)
	}
	scene.drawLives(screen)
	scene.drawBonusSymbols(screen)
	scene.drawReady(screen)
}

// drawGhostCaptureScore draws the captured ghost's score at its collision position.
func (scene *ScenePlay) drawGhostCaptureScore(screen *ebiten.Image) {
	if scene.State.Match == nil || scene.State.Phase != PhaseGhostCaptured {
		return
	}
	capture := scene.State.Match.GhostCapture
	message := strconv.Itoa(capture.Points)
	positions, count := pacManDrawPositions(capture.X, len(message)*highScoreFontSize)
	for _, x := range positions[:count] {
		op := &text.DrawOptions{}
		op.PrimaryAlign = text.AlignCenter
		op.GeoM.Translate(x, capture.Y-float64(highScoreFontSize)/2)
		op.ColorScale.ScaleWithColor(ghostScoreColor)
		text.Draw(screen, message, scene.highScoreFace, op)
	}
}

// drawGhost draws the current ghost frame, including a wrapped tunnel copy. Frightened
// ghosts flash white during the level's final flash periods.
func (scene *ScenePlay) drawGhost(screen *ebiten.Image, ghost GhostState, level LevelState) {
	flashing := ghost.Mode == GhostFrightened &&
		ghost.frightTicks <= level.Parameters.FrightFlashes*30 && (ghost.frightTicks/15)%2 == 0
	frame := scene.ghostFrame(ghost.Kind, ghost.Direction, ghost.Mode, ghost.animationTicks, flashing)
	size := frame.Bounds().Size()
	positions, count := pacManDrawPositions(ghost.X, size.X)
	for _, x := range positions[:count] {
		drawSpriteCentered(screen, frame, x, ghost.Y)
	}
}

// ghostFrame selects the sprite for a ghost's kind, facing, mode and animation tick;
// flashing selects the white frightened frames instead of the blue ones.
func (scene *ScenePlay) ghostFrame(kind GhostKind, facing Direction, mode GhostMode, animationTicks int, flashing bool) *ebiten.Image {
	direction := 3
	switch facing {
	case DirectionUp:
		direction = 0
	case DirectionDown:
		direction = 1
	case DirectionLeft:
		direction = 2
	}
	index := direction*2 + (animationTicks/6)%2
	frame := scene.images.BlinkyFrames[index]
	switch kind {
	case GhostPinky:
		frame = scene.images.PinkyFrames[index]
	case GhostInky:
		frame = scene.images.InkyFrames[index]
	case GhostClyde:
		frame = scene.images.ClydeFrames[index]
	}
	if mode == GhostEaten {
		frame = scene.images.GhostEyes[direction]
	} else if mode == GhostFrightened {
		index := (animationTicks / 6) % 2
		if flashing {
			index += 2
		}
		frame = scene.images.FrightenedFrames[index]
	}
	return frame
}

// drawSpriteCentered draws a sprite with its center at the given logical-pixel position.
func drawSpriteCentered(screen, sprite *ebiten.Image, x, y float64) {
	size := sprite.Bounds().Size()
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	op.GeoM.Translate(x-float64(size.X)/2, y-float64(size.Y)/2)
	screen.DrawImage(sprite, op)
}

// drawReady centers PAUSED, READY!, or the red GAME OVER message on readyRow.
func (scene *ScenePlay) drawReady(screen *ebiten.Image) {
	if scene.State.Match == nil {
		return
	}
	message := ""
	messageColor := readyColor
	switch {
	case scene.State.Paused:
		message = "PAUSED"
	case scene.State.Phase == PhaseReady:
		message = "READY!"
	case scene.State.Phase == PhaseGameOver:
		message = "GAME OVER"
		messageColor = color.RGBA{R: 255, A: 255}
	default:
		return
	}
	op := &text.DrawOptions{}
	op.PrimaryAlign = text.AlignCenter
	op.GeoM.Translate(float64(ScreenWidth)/2, float64(readyRow*CellSize+(CellSize-highScoreFontSize)/2))
	op.ColorScale.ScaleWithColor(messageColor)
	text.Draw(screen, message, scene.highScoreFace, op)
}

// drawFruit draws the visible bonus sprite centered on its position, or the pink
// value of an eaten bonus centered there while its popup timer runs.
func (scene *ScenePlay) drawFruit(screen *ebiten.Image) {
	if scene.State.Match == nil {
		return
	}
	fruit := scene.State.Match.Fruit
	if fruit.Visible {
		sprite := scene.images.BonusSymbols[scene.State.Match.Level.Parameters.BonusSymbol]
		spriteSize := sprite.Bounds().Size()
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		op.GeoM.Translate(FruitX-float64(spriteSize.X)/2, FruitY-float64(spriteSize.Y)/2)
		screen.DrawImage(sprite, op)
		return
	}
	if fruit.ScoreTicks > 0 {
		op := &text.DrawOptions{}
		op.PrimaryAlign = text.AlignCenter
		op.GeoM.Translate(FruitX, FruitY-float64(highScoreFontSize)/2)
		op.ColorScale.ScaleWithColor(fruitScoreColor)
		text.Draw(screen, strconv.Itoa(fruit.Points), scene.highScoreFace, op)
	}
}

// bottomHUDPosition centers a life or bonus sprite in a mirrored bottom HUD slot.
func bottomHUDPosition(spriteSize image.Point, index int, fromRight bool) (float64, float64) {
	column := 3 + index*2
	x := float64(column*CellSize) + float64(CellSize-spriteSize.X)/2 - float64(CellSize)/2
	if fromRight {
		x = float64(ScreenWidth-spriteSize.X) - x
	}
	y := float64((GridRows-2)*CellSize) + float64(CellSize-spriteSize.Y)/2 + float64(CellSize)/2
	return x, y
}

// drawLives draws reserve lives, excluding the life currently in play.
func (scene *ScenePlay) drawLives(screen *ebiten.Image) {
	if scene.State.Match == nil {
		return
	}
	spriteSize := scene.images.Life.Bounds().Size()
	for index := 0; index < scene.State.Match.Lives-1; index++ {
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		x, y := bottomHUDPosition(spriteSize, index, false)
		op.GeoM.Translate(x, y)
		screen.DrawImage(scene.images.Life, op)
	}
}

// drawBonusSymbols draws up to seven level symbols, current level on the right.
func (scene *ScenePlay) drawBonusSymbols(screen *ebiten.Image) {
	if scene.State.Match == nil {
		return
	}
	for index, symbol := range scene.State.Match.Level.BonusSymbols() {
		sprite := scene.images.BonusSymbols[symbol]
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		x, y := bottomHUDPosition(sprite.Bounds().Size(), index, true)
		op.GeoM.Translate(x, y)
		screen.DrawImage(sprite, op)
	}
}

// drawScore renders a score aligned to the right edge of a grid column in the arcade's
// two-digit minimum format.
func (scene *ScenePlay) drawScore(screen *ebiten.Image, score, rightColumn int) {
	op := &text.DrawOptions{}
	op.PrimaryAlign = text.AlignEnd
	op.GeoM.Translate(float64((rightColumn+1)*CellSize), float64(highScoreTop+CellSize))
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, formatScore(score), scene.highScoreFace, op)
}

// formatScore renders a score with at least two digits, so zero appears as 00 like the arcade.
func formatScore(score int) string {
	return fmt.Sprintf("%02d", score)
}

// drawCoins draws only the coins still present in the current level.
func (scene *ScenePlay) drawCoins(screen *ebiten.Image) {
	if scene.State.Match == nil {
		return
	}
	coinSize := scene.images.Coin.Bounds().Size()
	for index, coord := range CoinCoords {
		if !scene.State.Match.Level.Coins[index] {
			continue
		}
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		x := float64(coord[0]*CellSize) + (float64(CellSize)-float64(coinSize.X))/2
		y := float64(coord[1]*CellSize) + (float64(CellSize)-float64(coinSize.Y))/2
		op.GeoM.Translate(x, y)
		screen.DrawImage(scene.images.Coin, op)
	}
}

// drawEnergizers draws only the energizers still present in the current level.
func (scene *ScenePlay) drawEnergizers(screen *ebiten.Image) {
	if scene.State.Match == nil {
		return
	}
	spriteSize := scene.images.Energizer.Bounds().Size()
	for index, coord := range EnergizerCoords {
		if !scene.State.Match.Level.Energizers[index] {
			continue
		}
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		x := float64(coord[0]*CellSize) + (float64(CellSize)-float64(spriteSize.X))/2
		y := float64(coord[1]*CellSize) + (float64(CellSize)-float64(spriteSize.Y))/2
		op.GeoM.Translate(x, y)
		screen.DrawImage(scene.images.Energizer, op)
	}
}

// drawTopLabel renders a white label at the common top HUD offset.
func (scene *ScenePlay) drawTopLabel(screen *ebiten.Image, label string, column int) {
	scene.drawHUDText(screen, label, column, highScoreTop, color.White)
}

// drawHUDText renders HUD text in the given color at a grid column and logical pixel offset.
func (scene *ScenePlay) drawHUDText(screen *ebiten.Image, label string, column, top int, textColor color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(column*CellSize), float64(top))
	op.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, label, scene.highScoreFace, op)
}

// drawDebugAllowedPath lazily caches and draws the maze's gray path backgrounds.
func (scene *ScenePlay) drawDebugAllowedPath(screen *ebiten.Image) {
	if scene.images.DebugAllowedPath == nil {
		scene.images.DebugAllowedPath = ebiten.NewImage(ScreenWidth, ScreenHeight)
		pathColor := color.RGBA{R: 79, G: 79, B: 79, A: 255}
		for row, cells := range AllowedPath {
			for column, allowed := range cells {
				if !allowed {
					continue
				}
				startX := column * CellSize
				startY := row * CellSize
				for pixelY := startY; pixelY < startY+CellSize; pixelY++ {
					for pixelX := startX; pixelX < startX+CellSize; pixelX++ {
						scene.images.DebugAllowedPath.Set(pixelX, pixelY, pathColor)
					}
				}
			}
		}
	}
	screen.DrawImage(scene.images.DebugAllowedPath, nil)
}

// newDebugGrid creates a transparent dashed overlay, cached by ScenePlay.Draw.
func newDebugGrid() *ebiten.Image {
	grid := ebiten.NewImage(ScreenWidth, ScreenHeight)
	gridColor := color.RGBA{R: 160, G: 160, B: 160, A: 160}

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
