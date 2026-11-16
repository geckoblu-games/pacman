package game

import (
	"fmt"
	"math"
	"math/rand"
)

// GhostMode identifies the ghost's current behavior.
type GhostMode uint8

const (
	// GhostScatter targets the ghost's home corner.
	GhostScatter GhostMode = iota
	// GhostChase targets Pac-Man.
	GhostChase
	// GhostFrightened wanders and can be eaten.
	GhostFrightened
	// GhostEaten returns to the house as eyes.
	GhostEaten
)

// GhostState contains Blinky's renderer-independent movement and mode state.
type GhostState struct {
	// X and Y are the ghost's center in logical pixels.
	X, Y float64
	// Direction is the current direction of travel.
	Direction Direction
	// Mode is the current visible and collision state.
	Mode GhostMode
	// ElroySuspended disables Elroy until Clyde leaves after a death.
	ElroySuspended bool
	// scheduleIndex and scheduleTicks track the shared scatter/chase sequence.
	scheduleIndex, scheduleTicks int
	// frightTicks counts remaining frightened ticks, including while eyes return.
	frightTicks int
	// reversePending defers a mode-change reversal until the next tile entry.
	reversePending bool
	// turn is the direction chosen for the center of the currently entered tile.
	turn Direction
	// nextTurn is the direction chosen for the center of the next tile.
	nextTurn Direction
	// houseStage tracks eyes entering and Blinky leaving the ghost house.
	houseStage int
	// random is reset for each level and life.
	random *rand.Rand
	// animationTicks advances independently of movement.
	animationTicks int
}

// NewBlinky creates Blinky above the ghost-house door, facing left.
func NewBlinky() GhostState {
	return GhostState{
		X: float64(ScreenWidth) / 2, Y: float64(14*CellSize + CellSize/2),
		Direction: DirectionLeft, turn: DirectionLeft, nextTurn: DirectionLeft,
		random: rand.New(rand.NewSource(1)),
	}
}

// ghostSchedule returns the scatter/chase durations in ticks for a level.
func ghostSchedule(level int) [8]int {
	first, thirdChase, fourth := 7, 20, 5*simulationTicksPerSecond
	if level >= 5 {
		first, thirdChase = 5, 1037
	} else if level >= 2 {
		thirdChase = 1033
	}
	if level >= 2 {
		fourth = 1
	}
	return [8]int{first * 60, 20 * 60, first * 60, 20 * 60, 5 * 60, thirdChase * 60, fourth, 0}
}

// scheduledMode reports the active scatter/chase mode independently of fright.
func (ghost GhostState) scheduledMode() GhostMode {
	if ghost.scheduleIndex%2 == 0 {
		return GhostScatter
	}
	return GhostChase
}

// Tick advances the mode schedule, frightened timer, and animation at 60 Hz.
func (ghost *GhostState) Tick(level LevelState) {
	ghost.animationTicks++
	if ghost.frightTicks > 0 {
		ghost.frightTicks--
		if ghost.frightTicks == 0 && ghost.Mode == GhostFrightened {
			ghost.Mode = ghost.scheduledMode()
		}
		return
	}
	durations := ghostSchedule(level.Number)
	if ghost.scheduleIndex >= len(durations)-1 {
		return
	}
	ghost.scheduleTicks++
	if ghost.scheduleTicks >= durations[ghost.scheduleIndex] {
		ghost.scheduleTicks = 0
		ghost.scheduleIndex++
		if ghost.Mode != GhostEaten && ghost.houseStage == 0 {
			ghost.Mode = ghost.scheduledMode()
			ghost.reversePending = true
		}
	}
}

// Frighten resets the vulnerability timer and signals a reversal on energizer consumption.
func (ghost *GhostState) Frighten(level LevelState) {
	ghost.frightTicks = level.Parameters.FrightTimeSeconds * simulationTicksPerSecond
	if ghost.Mode == GhostEaten || ghost.houseStage != 0 {
		return
	}
	ghost.reversePending = true
	if ghost.frightTicks > 0 {
		ghost.Mode = GhostFrightened
	}
}

// OccupiedTile returns the cell containing the ghost's center, wrapping the tunnel.
func (ghost GhostState) OccupiedTile() (int, int) {
	column, row := int(math.Floor(ghost.X/CellSize)), int(math.Floor(ghost.Y/CellSize))
	if row == 17 {
		if column < 0 {
			column = GridColumns - 1
		} else if column >= GridColumns {
			column = 0
		}
	}
	return column, row
}

// elroyStage returns Blinky's active speed stage for the remaining collectibles.
func (ghost GhostState) elroyStage(level LevelState) int {
	if ghost.ElroySuspended {
		return 0
	}
	if level.Remaining <= level.Parameters.Elroy2DotsLeft {
		return 2
	}
	if level.Remaining <= level.Parameters.Elroy1DotsLeft {
		return 1
	}
	return 0
}

// speedPercent selects the mode, tunnel, or Elroy speed for Blinky.
func (ghost GhostState) speedPercent(level LevelState) int {
	if ghost.Mode == GhostEaten {
		return 200
	}
	column, row := ghost.OccupiedTile()
	if row == 17 && (column <= 5 || column >= 22) {
		return level.Parameters.GhostTunnelSpeedPercent
	}
	if ghost.Mode == GhostFrightened {
		return level.Parameters.FrightGhostSpeedPercent
	}
	switch ghost.elroyStage(level) {
	case 1:
		return level.Parameters.Elroy1SpeedPercent
	case 2:
		return level.Parameters.Elroy2SpeedPercent
	default:
		return level.Parameters.GhostSpeedPercent
	}
}

// target returns the current tile sought by Blinky.
func (ghost GhostState) target(level LevelState, player PacManState) (int, int) {
	if ghost.Mode == GhostEaten {
		return 13, 14
	}
	if ghost.Mode == GhostScatter && ghost.elroyStage(level) == 0 {
		return 25, 0
	}
	return player.OccupiedTile()
}

// ghostDirection chooses a legal exit using the common ghost lookahead rules.
func ghostDirection(column, row int, direction Direction, targetX, targetY int, frightened, eaten bool, random *rand.Rand) Direction {
	order := [4]Direction{DirectionUp, DirectionLeft, DirectionDown, DirectionRight}
	if frightened {
		choices := 0
		var only Direction
		for _, candidate := range order {
			if candidate != oppositeDirection(direction) && canMoveFrom(column, row, candidate) {
				choices++
				only = candidate
			}
		}
		if choices <= 1 {
			return only
		}
		clockwise := [4]Direction{DirectionUp, DirectionRight, DirectionDown, DirectionLeft}
		start := random.Intn(4)
		for i := range order {
			order[i] = clockwise[(start+i)%4]
		}
	}
	best, bestDistance := DirectionNone, math.MaxInt
	for _, candidate := range order {
		if candidate == oppositeDirection(direction) || !canMoveFrom(column, row, candidate) {
			continue
		}
		if !frightened && !eaten && candidate == DirectionUp && (column == 12 || column == 15) && (row == 14 || row == 26) {
			continue
		}
		x, y, _ := adjacentTile(column, row, candidate)
		distance := (x-targetX)*(x-targetX) + (y-targetY)*(y-targetY)
		if frightened || distance < bestDistance {
			best, bestDistance = candidate, distance
			if frightened {
				break
			}
		}
	}
	return best
}

// planTurn selects a legal turn for a tile the ghost will reach next.
func (ghost *GhostState) planTurn(column, row int, direction Direction, level LevelState, player PacManState) Direction {
	targetX, targetY := ghost.target(level, player)
	turn := ghostDirection(column, row, direction, targetX, targetY, ghost.Mode == GhostFrightened, ghost.Mode == GhostEaten, ghost.random)
	if turn == DirectionNone {
		panic(fmt.Sprintf("ghost has no legal exit at tile (%d,%d) moving %d", column, row, direction))
	}
	return turn
}

// Update advances the ghost, retaining decisions made one tile ahead.
func (ghost *GhostState) Update(level LevelState, player PacManState) {
	distance := actorSpeed(ghost.speedPercent(level))
	if ghost.houseStage != 0 {
		ghost.moveInHouse(distance)
		return
	}
	for distance > 0 {
		column, row := ghost.OccupiedTile()
		dx, dy := directionDelta(ghost.Direction)
		axis := ghost.X
		if dy != 0 {
			axis = ghost.Y
		}
		index := tileIndex(axis, dx+dy)
		nextCenter := float64((index+dx+dy)*CellSize + CellSize/2)
		if dx != 0 && row == 17 {
			if ghost.Direction == DirectionLeft && ghost.X <= CellSize/2 {
				nextCenter = -float64(CellSize) / 2
			} else if ghost.Direction == DirectionRight && ghost.X >= float64(ScreenWidth-CellSize/2) {
				nextCenter = float64(ScreenWidth) + float64(CellSize)/2
			}
		}
		step := math.Min(distance, math.Abs(nextCenter-axis))
		if step < 0.0000001 {
			break
		}
		ghost.X += float64(dx) * step
		ghost.Y += float64(dy) * step
		distance -= step
		newColumn, newRow := ghost.OccupiedTile()
		if newColumn != column || newRow != row {
			if ghost.reversePending {
				ghost.Direction = oppositeDirection(ghost.Direction)
				ghost.turn = ghost.Direction
				ghost.nextTurn = ghost.planTurn(column, row, ghost.Direction, level, player)
				ghost.reversePending = false
				continue
			}
			ghost.turn = ghost.nextTurn
			aheadX, aheadY, _ := adjacentTile(newColumn, newRow, ghost.turn)
			ghost.nextTurn = ghost.planTurn(aheadX, aheadY, ghost.turn, level, player)
		}
		if step == math.Abs(nextCenter-axis) {
			if ghost.X <= -float64(CellSize)/2 {
				ghost.X += ScreenWidth
			} else if ghost.X >= float64(ScreenWidth)+float64(CellSize)/2 {
				ghost.X -= ScreenWidth
			} else {
				ghost.Direction = ghost.turn
				if ghost.Mode == GhostEaten && newColumn == 13 && newRow == 14 {
					ghost.houseStage = 1
					return
				}
			}
		}
	}
}

// moveInHouse returns eyes through the door and sends revived Blinky back out.
func (ghost *GhostState) moveInHouse(distance float64) {
	centerX := float64(ScreenWidth) / 2
	topY := float64(14*CellSize + CellSize/2)
	insideY := float64(17*CellSize + CellSize/2)
	switch ghost.houseStage {
	case 1:
		ghost.Direction = DirectionRight
		ghost.X = math.Min(centerX, ghost.X+distance)
		if ghost.X == centerX {
			ghost.houseStage = 2
		}
	case 2:
		ghost.Direction = DirectionDown
		ghost.Y = math.Min(insideY, ghost.Y+distance)
		if ghost.Y == insideY {
			ghost.houseStage = 3
			ghost.Mode = ghost.scheduledMode()
		}
	case 3:
		ghost.Direction = DirectionUp
		ghost.Y = math.Max(topY, ghost.Y-distance)
		if ghost.Y == topY {
			ghost.houseStage = 0
			ghost.Direction, ghost.turn, ghost.nextTurn = DirectionLeft, DirectionLeft, DirectionLeft
		}
	}
}

// UpdateActors advances both actors and checks one discrete tile collision per tick.
func (state *GameState) UpdateActors(requested Direction, ghostMovement bool) {
	if state.Match == nil || state.Phase != PhasePlaying || state.Paused {
		return
	}
	match := state.Match
	if match.PacMan.EatingPauseTicks > 0 {
		match.PacMan.EatingPauseTicks--
	} else {
		level := match.Level
		if match.Blinky.frightTicks > 0 {
			level.Parameters.PacManSpeedPercent = level.Parameters.FrightPacManSpeedPercent
		}
		match.PacMan.Update(requested, level)
		column, row := match.PacMan.OccupiedTile()
		state.ConsumeCollectible(column, row)
		state.ConsumeFruit()
	}
	if state.Phase != PhasePlaying {
		return
	}
	if ghostMovement {
		match.Blinky.Update(match.Level, match.PacMan)
	}
	playerX, playerY := match.PacMan.OccupiedTile()
	ghostX, ghostY := match.Blinky.OccupiedTile()
	if playerX != ghostX || playerY != ghostY || match.Blinky.Mode == GhostEaten || match.Blinky.houseStage != 0 {
		return
	}
	if match.Blinky.Mode == GhostFrightened {
		points := 200 << min(match.GhostStreak, 3)
		match.GhostStreak++
		match.GhostCapture = GhostCaptureState{X: match.Blinky.X, Y: match.Blinky.Y, Points: points}
		state.AddScore(points)
		state.enterPhase(PhaseGhostCaptured)
	} else {
		state.LoseLife()
	}
}
