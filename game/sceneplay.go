package game

import (
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

// ScenePlay handles match input and renders the maze, collectibles, and HUD.
type ScenePlay struct {
	// State references the session and match owned by Game.
	State *GameState
	// images contains loaded sprites and lazily generated debug overlays.
	images assets.GameImages
	// highScoreFace is the reusable font face for HUD labels and scores.
	highScoreFace *text.GoTextFace
	// pacMan is the player's idle sprite drawn over the maze collectibles.
	pacMan *pacManSprite
}

// newScenePlay loads rendering resources and binds the scene to the shared state.
func newScenePlay(state *GameState) (*ScenePlay, error) {
	images, err := assets.LoadGameImages(image.Pt(ScreenWidth, ScreenHeight))
	if err != nil {
		return nil, err
	}
	highScoreFace, err := assets.LoadTextFace("PressStart2P-Regular.ttf", highScoreFontSize)
	if err != nil {
		return nil, err
	}
	return &ScenePlay{
		State:         state,
		images:        images,
		highScoreFace: highScoreFace,
		pacMan:        newPacManSprite(images.PacManFrames),
	}, nil
}

// Update starts a new match on Enter after game over, toggles the pause on P, and
// while unpaused advances the timed level-clear and ready phases and moves Pac-Man
// only during active play.
func (scene *ScenePlay) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && (scene.State.Match == nil || scene.State.Phase == PhaseGameOver) {
		scene.State.StartNewMatch()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		scene.State.TogglePause()
	}
	if scene.State.Paused {
		return nil
	}
	scene.State.Tick()
	if scene.State.Phase == PhasePlaying && scene.State.Match != nil {
		scene.updatePacMan(scene.pacManDirection())
	}
	return nil
}

// updatePacMan advances the player, consumes the occupied tile and any overlapped
// bonus, and honors eating pauses.
func (scene *ScenePlay) updatePacMan(requested Direction) {
	player := &scene.State.Match.PacMan
	if player.EatingPauseTicks > 0 {
		player.EatingPauseTicks--
		return
	}
	player.Update(requested, scene.State.Match.Level)
	column, row := player.OccupiedTile()
	scene.State.ConsumeCollectible(column, row)
	scene.State.ConsumeFruit()
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

// Draw composes debug backgrounds, walls, collectibles, the bonus or its score popup,
// Pac-Man, the grid, the HUD, and the READY! intro or PAUSED message. The maze uses
// the white walls during white flash segments.
func (scene *ScenePlay) Draw(screen *ebiten.Image) {
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
		scene.pacMan.Draw(screen, scene.State.Match.PacMan)
	}
	if DebugGridEnabled {
		if scene.images.DebugGrid == nil {
			scene.images.DebugGrid = newDebugGrid()
		}
		screen.DrawImage(scene.images.DebugGrid, nil)
	}
	scene.drawTopLabel(screen, "1UP", oneUpColumn)
	scene.drawTopLabel(screen, "HIGH SCORE", highScoreColumn)
	scene.drawScore(screen, scene.State.HighScore, highScoreRightColumn)
	if scene.State.Match != nil {
		scene.drawScore(screen, scene.State.Match.Score, scoreRightColumn)
	}
	scene.drawLives(screen)
	scene.drawBonusSymbols(screen)
	scene.drawReady(screen)
}

// drawReady centers a yellow message on readyRow: PAUSED while the match is paused,
// otherwise READY! during the ready intro.
func (scene *ScenePlay) drawReady(screen *ebiten.Image) {
	if scene.State.Match == nil {
		return
	}
	message := ""
	switch {
	case scene.State.Paused:
		message = "PAUSED"
	case scene.State.Phase == PhaseReady:
		message = "READY!"
	default:
		return
	}
	op := &text.DrawOptions{}
	op.PrimaryAlign = text.AlignCenter
	op.GeoM.Translate(float64(ScreenWidth)/2, float64(readyRow*CellSize+(CellSize-highScoreFontSize)/2))
	op.ColorScale.ScaleWithColor(readyColor)
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

// drawScore renders a decimal score aligned to the right edge of a grid column.
func (scene *ScenePlay) drawScore(screen *ebiten.Image, score, rightColumn int) {
	op := &text.DrawOptions{}
	op.PrimaryAlign = text.AlignEnd
	op.GeoM.Translate(float64((rightColumn+1)*CellSize), float64(highScoreTop+CellSize))
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, strconv.Itoa(score), scene.highScoreFace, op)
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

// drawTopLabel renders a label at the common top HUD offset.
func (scene *ScenePlay) drawTopLabel(screen *ebiten.Image, label string, column int) {
	scene.drawHUDText(screen, label, column, highScoreTop)
}

// drawHUDText renders white HUD text at a grid column and logical pixel offset.
func (scene *ScenePlay) drawHUDText(screen *ebiten.Image, label string, column, top int) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(column*CellSize), float64(top))
	op.ColorScale.ScaleWithColor(color.White)
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
