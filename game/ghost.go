package game

import (
	"fmt"
	"math"
	"math/rand"
)

const (
	// ghostHouseBounceSpeedPercent is the fixed waiting speed as a percentage of base speed.
	ghostHouseBounceSpeedPercent = 40
	// ghostHouseBounceTop and ghostHouseBounceBottom bound the waiting center in logical pixels.
	ghostHouseBounceTop    = 17 * CellSize
	ghostHouseBounceBottom = 18 * CellSize
)

// GhostMode identifies the ghost's current behavior.
type GhostMode uint8

const (
	// GhostScatter targets the ghost's home corner.
	GhostScatter GhostMode = iota
	// GhostChase uses the ghost's pursuit target.
	GhostChase
	// GhostFrightened wanders and can be eaten.
	GhostFrightened
	// GhostEaten returns to the house as eyes.
	GhostEaten
)

// GhostKind identifies a ghost's targeting and house-release behavior.
type GhostKind uint8

const (
	// GhostBlinky identifies the red ghost.
	GhostBlinky GhostKind = iota
	// GhostPinky identifies the pink ghost.
	GhostPinky
	// GhostInky identifies the cyan ghost.
	GhostInky
	// GhostClyde identifies the orange ghost.
	GhostClyde
)

// ghostClock holds the mode schedule and frightened randomness shared by ghosts.
type ghostClock struct {
	// scheduleIndex and scheduleTicks track the shared scatter/chase sequence.
	scheduleIndex, scheduleTicks int
	// frightTicks counts remaining frightened ticks, including while eyes return.
	frightTicks int
	// random is reset for each level and life.
	random *rand.Rand
}

// GhostState contains a ghost's renderer-independent movement and mode state.
type GhostState struct {
	// ghostClock is shared by every ghost in the match.
	*ghostClock
	// Kind selects the ghost's individual behavior.
	Kind GhostKind
	// blinky supplies the current position used by Inky's chase target.
	blinky *GhostState
	// X and Y are the ghost's center in logical pixels.
	X, Y float64
	// Direction is the current direction of travel.
	Direction Direction
	// Mode is the current visible and collision state.
	Mode GhostMode
	// ElroySuspended disables Elroy until Clyde leaves after a death.
	ElroySuspended bool
	// reversePending defers a mode-change reversal until the next tile entry.
	reversePending bool
	// turn is the direction chosen for the center of the currently entered tile.
	turn Direction
	// nextTurn is the direction chosen for the center of the next tile.
	nextTurn Direction
	// houseStage tracks eyes entering, returning to their slots, waiting, and leaving.
	houseStage int
	// exitRight records a schedule change while the ghost is inside the house.
	exitRight bool
	// animationTicks advances independently of movement.
	animationTicks int
}

// NewBlinky creates Blinky above the ghost-house door, facing left.
func NewBlinky() GhostState {
	return GhostState{
		ghostClock: &ghostClock{random: rand.New(rand.NewSource(1))},
		X:          float64(ScreenWidth) / 2, Y: float64(14*CellSize + CellSize/2),
		Direction: DirectionLeft, turn: DirectionLeft, nextTurn: DirectionLeft,
	}
}

// newPinky creates Pinky in the house with the match's shared clock.
func newPinky(clock *ghostClock, afterDeath bool) GhostState {
	stage := 3
	if afterDeath {
		stage = 4
	}
	return GhostState{
		ghostClock: clock, Kind: GhostPinky,
		X: float64(ScreenWidth) / 2, Y: float64(17*CellSize + CellSize/2),
		Direction: DirectionUp, houseStage: stage,
	}
}

// newInky creates Inky on the left side of the house with Blinky's shared clock.
func newInky(blinky *GhostState) GhostState {
	return GhostState{
		ghostClock: blinky.ghostClock, blinky: blinky, Kind: GhostInky,
		X: float64(ScreenWidth)/2 - 2*CellSize, Y: float64(17*CellSize + CellSize/2),
		Direction: DirectionUp, houseStage: 4,
	}
}

// newClyde creates Clyde on the right side of the house with the shared clock.
func newClyde(clock *ghostClock) GhostState {
	return GhostState{
		ghostClock: clock, Kind: GhostClyde,
		X: float64(ScreenWidth)/2 + 2*CellSize, Y: float64(17*CellSize + CellSize/2),
		Direction: DirectionUp, houseStage: 4,
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
func (clock ghostClock) scheduledMode() GhostMode {
	if clock.scheduleIndex%2 == 0 {
		return GhostScatter
	}
	return GhostChase
}

// tick advances the shared clock and reports schedule changes and fright expiration.
func (clock *ghostClock) tick(level LevelState) (bool, bool) {
	if clock.frightTicks > 0 {
		clock.frightTicks--
		return false, clock.frightTicks == 0
	}
	durations := ghostSchedule(level.Number)
	if clock.scheduleIndex >= len(durations)-1 {
		return false, false
	}
	clock.scheduleTicks++
	if clock.scheduleTicks >= durations[clock.scheduleIndex] {
		clock.scheduleTicks = 0
		clock.scheduleIndex++
		return true, false
	}
	return false, false
}

// applyClock advances animation and applies shared mode transitions to this ghost.
func (ghost *GhostState) applyClock(changed, frightEnded bool) {
	ghost.animationTicks++
	if frightEnded && ghost.Mode == GhostFrightened {
		ghost.Mode = ghost.scheduledMode()
	}
	if changed && ghost.Mode != GhostEaten {
		ghost.Mode = ghost.scheduledMode()
		if ghost.houseStage == 0 {
			ghost.reversePending = true
		} else {
			ghost.exitRight = true
		}
	}
}

// Tick advances the shared clock and this ghost; matches tick all ghosts together.
func (ghost *GhostState) Tick(level LevelState) {
	changed, ended := ghost.ghostClock.tick(level)
	ghost.applyClock(changed, ended)
}

// Frighten resets the vulnerability timer and signals a reversal on energizer consumption.
func (ghost *GhostState) Frighten(level LevelState) {
	ghost.frightTicks = level.Parameters.FrightTimeSeconds * simulationTicksPerSecond
	if ghost.Mode == GhostEaten {
		return
	}
	if ghost.houseStage == 0 {
		ghost.reversePending = true
	}
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
	if ghost.Kind != GhostBlinky || ghost.ElroySuspended {
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

// speedPercent selects the mode, tunnel, or Blinky's Elroy speed.
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

// target returns the current tile sought by the ghost.
func (ghost GhostState) target(level LevelState, player PacManState) (int, int) {
	if ghost.Mode == GhostEaten {
		return 13, 14
	}
	if ghost.Mode == GhostScatter && ghost.elroyStage(level) == 0 {
		switch ghost.Kind {
		case GhostPinky:
			return 2, 0
		case GhostInky:
			return 27, 35
		case GhostClyde:
			return 0, 35
		}
		return 25, 0
	}
	x, y := player.OccupiedTile()
	if ghost.Kind == GhostClyde {
		gx, gy := ghost.OccupiedTile()
		dx, dy := x-gx, y-gy
		if dx*dx+dy*dy < 64 {
			return 0, 35
		}
	}
	if ghost.Kind == GhostPinky || ghost.Kind == GhostInky {
		direction := player.Direction
		if direction == DirectionNone {
			direction = player.lastDirection
		}
		offset := 4
		if ghost.Kind == GhostInky {
			offset = 2
		}
		dx, dy := directionDelta(direction)
		x, y = x+offset*dx, y+offset*dy
		if direction == DirectionUp {
			x -= offset
		}
		if ghost.Kind == GhostInky {
			bx, by := ghost.blinky.OccupiedTile()
			x, y = 2*x-bx, 2*y-by
		}
	}
	return x, y
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
		if ghost.houseStage == 4 {
			distance = actorSpeed(ghostHouseBounceSpeedPercent)
		}
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

// moveInHouse returns eyes, bounces waiting ghosts, and moves released ghosts outside.
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
			switch ghost.Kind {
			case GhostInky, GhostClyde:
				ghost.houseStage = 6
				ghost.Direction = DirectionLeft
				if ghost.Kind == GhostClyde {
					ghost.Direction = DirectionRight
				}
			default:
				ghost.houseStage = 4
				ghost.Mode = ghost.scheduledMode()
				ghost.exitRight = false
				ghost.Direction = DirectionUp
			}
			if ghost.Kind == GhostBlinky {
				ghost.houseStage = 3
			}
		}
	case 3:
		ghost.Direction = DirectionUp
		ghost.Y = math.Max(topY, ghost.Y-distance)
		if ghost.Y == topY {
			ghost.houseStage = 0
			direction := DirectionLeft
			if ghost.exitRight {
				direction = DirectionRight
			}
			ghost.Direction, ghost.turn, ghost.nextTurn = direction, direction, direction
			ghost.reversePending, ghost.exitRight = false, false
		}
	case 4:
		if ghost.Y < ghostHouseBounceTop || ghost.Y > ghostHouseBounceBottom {
			panic(fmt.Sprintf("waiting ghost outside bounce bounds at y=%v", ghost.Y))
		}
		for distance > 0 {
			switch ghost.Direction {
			case DirectionUp:
				remaining := ghost.Y - float64(ghostHouseBounceTop)
				if distance >= remaining {
					ghost.Y = ghostHouseBounceTop
					distance -= remaining
					ghost.Direction = DirectionDown
				} else {
					ghost.Y -= distance
					distance = 0
				}
			case DirectionDown:
				remaining := float64(ghostHouseBounceBottom) - ghost.Y
				if distance >= remaining {
					ghost.Y = ghostHouseBounceBottom
					distance -= remaining
					ghost.Direction = DirectionUp
				} else {
					ghost.Y += distance
					distance = 0
				}
			default:
				panic(fmt.Sprintf("waiting ghost has invalid direction %d", ghost.Direction))
			}
		}
	case 5:
		if ghost.X < centerX {
			ghost.Direction = DirectionRight
			ghost.X = math.Min(centerX, ghost.X+distance)
		} else {
			ghost.Direction = DirectionLeft
			ghost.X = math.Max(centerX, ghost.X-distance)
		}
		if ghost.X == centerX {
			ghost.houseStage = 3
		}
	case 6:
		homeX := centerX - 2*CellSize
		if ghost.Kind == GhostClyde {
			homeX = centerX + 2*CellSize
		}
		if ghost.X < homeX {
			ghost.Direction = DirectionRight
			ghost.X = math.Min(homeX, ghost.X+distance)
		} else {
			ghost.Direction = DirectionLeft
			ghost.X = math.Max(homeX, ghost.X-distance)
		}
		if ghost.X == homeX {
			ghost.houseStage = 4
			ghost.Mode = ghost.scheduledMode()
			ghost.exitRight = false
			ghost.Direction = DirectionUp
		}
	}
}

// releasePinky starts Pinky's exit only when she is waiting in the house.
func (match *MatchState) releasePinky() {
	if match.Pinky.houseStage == 4 {
		match.Pinky.startExit()
	}
}

// startExit moves a waiting ghost to the center before leaving the house.
func (ghost *GhostState) startExit() {
	ghost.houseStage = 3
	if ghost.X != float64(ScreenWidth)/2 {
		ghost.houseStage = 5
	}
}

// releaseClyde starts his exit and restores Blinky's Elroy behavior.
func (match *MatchState) releaseClyde() {
	if match.Clyde.houseStage == 4 {
		match.Clyde.startExit()
		match.Blinky.ElroySuspended = false
	}
}

// releasePersonalGhosts applies the currently enabled personal dot limits.
func (match *MatchState) releasePersonalGhosts() {
	if match.GlobalGhostCounter {
		return
	}
	match.releasePinky()
	limit := 0
	if match.Level.Number == 1 {
		limit = 30
	}
	if match.Inky.houseStage == 4 {
		if match.InkyDots < limit {
			return
		}
		match.Inky.startExit()
	}
	limit = 0
	switch match.Level.Number {
	case 1:
		limit = 60
	case 2:
		limit = 50
	}
	if match.Clyde.houseStage == 4 && match.ClydeDots >= limit {
		match.releaseClyde()
	}
}

// releaseWaitingGhost forces the most-preferred waiting ghost to leave.
func (match *MatchState) releaseWaitingGhost() {
	if match.Pinky.houseStage == 4 {
		match.releasePinky()
	} else if match.Inky.houseStage == 4 {
		match.Inky.startExit()
	} else if match.Clyde.houseStage == 4 {
		match.releaseClyde()
	}
}

// UpdateActors advances the actors and checks discrete tile collisions per tick.
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
		match.Pinky.Update(match.Level, match.PacMan)
		match.Inky.Update(match.Level, match.PacMan)
		match.Clyde.Update(match.Level, match.PacMan)
	}
	match.releasePersonalGhosts()
	playerX, playerY := match.PacMan.OccupiedTile()
	for _, ghost := range []*GhostState{&match.Blinky, &match.Pinky, &match.Inky, &match.Clyde} {
		ghostX, ghostY := ghost.OccupiedTile()
		outside := ghost.houseStage == 0 || (ghost.houseStage == 3 && isWalkable(ghostX, ghostY))
		if playerX != ghostX || playerY != ghostY || ghost.Mode == GhostEaten || !outside {
			continue
		}
		if ghost.Mode == GhostFrightened {
			points := 200 << min(match.GhostStreak, 3)
			match.GhostStreak++
			match.GhostCapture = GhostCaptureState{X: ghost.X, Y: ghost.Y, Points: points, Kind: ghost.Kind}
			state.AddScore(points)
			state.enterPhase(PhaseGhostCaptured)
		} else {
			state.LoseLife()
		}
		return
	}
}
