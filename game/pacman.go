package game

import "math"

const (
	// simulationTicksPerSecond is the fixed rate used to convert actor speeds to per-tick movement.
	simulationTicksPerSecond = 60
	// pacManAnimationTicksPerFrame is the number of simulation ticks each mouth frame lasts.
	pacManAnimationTicksPerFrame = 6
	// pacManClosedMouthFrame is the final atlas cell shared by all directions.
	pacManClosedMouthFrame = 8
	// pacManEarlyCornerPixels is the early-turn allowance in source 8-pixel tiles.
	pacManEarlyCornerPixels = 3
	// pacManLateCornerPixels is the late-turn allowance in source 8-pixel tiles.
	pacManLateCornerPixels = 4
	// pacManCoinPauseTicks is the movement pause after eating a regular coin.
	pacManCoinPauseTicks = 1
	// pacManEnergizerPauseTicks is the movement pause after eating an energizer.
	pacManEnergizerPauseTicks = 3
	// pacManStartRow is the zero-based maze row of Pac-Man's starting position.
	pacManStartRow = 26
)

// Direction identifies one cardinal direction or no movement.
type Direction uint8

const (
	// DirectionNone means the player is stationary.
	DirectionNone Direction = iota
	// DirectionUp moves toward decreasing row coordinates.
	DirectionUp
	// DirectionLeft moves toward decreasing column coordinates.
	DirectionLeft
	// DirectionDown moves toward increasing row coordinates.
	DirectionDown
	// DirectionRight moves toward increasing column coordinates.
	DirectionRight
)

// PacManState stores Pac-Man's logical-pixel center and current turn intent.
type PacManState struct {
	// X is Pac-Man's horizontal center coordinate in logical pixels.
	X float64
	// Y is Pac-Man's vertical center coordinate in logical pixels.
	Y float64
	// Direction is the direction Pac-Man is currently moving.
	Direction Direction
	// lastDirection preserves the facing direction after movement stops.
	lastDirection Direction
	// RequestedDirection is buffered until Pac-Man can turn into an open tile.
	RequestedDirection Direction
	// EatingPauseTicks counts remaining stationary ticks after consuming a collectible.
	EatingPauseTicks int
	// animationFrame is the current first-row atlas cell selected for rendering.
	animationFrame int
	// animationPhase is the current position in the five-step mouth cycle.
	animationPhase int
	// animationTicks counts ticks elapsed within the current mouth frame.
	animationTicks int
	// corneringCorrection moves the player laterally toward the new lane center.
	corneringCorrection Direction
	// corneringTarget is the logical-pixel coordinate of the lane center being reached.
	corneringTarget float64
}

// OccupiedTile returns the tile containing Pac-Man's center point.
func (player PacManState) OccupiedTile() (int, int) {
	column := tileIndex(player.X, 0)
	row := tileIndex(player.Y, 0)
	if row == 17 {
		if column < 0 {
			column = GridColumns - 1
		} else if column >= GridColumns {
			column = 0
		}
	}
	return column, row
}

// NewPacManState creates Pac-Man at the existing lower-maze starting position.
func NewPacManState() PacManState {
	return PacManState{
		X:              float64(ScreenWidth) / 2,
		Y:              float64(pacManStartRow*CellSize + CellSize/2),
		Direction:      DirectionLeft,
		lastDirection:  DirectionLeft,
		animationFrame: pacManHalfMouthFrame(DirectionLeft),
	}
}

// Update buffers a requested turn and advances Pac-Man by one 60 Hz tick.
func (player *PacManState) Update(requested Direction, level LevelState) {
	if requested > DirectionNone && requested <= DirectionRight {
		player.RequestedDirection = requested
	}
	speed := actorSpeed(level.Parameters.PacManSpeedPercent)
	previousX, previousY := player.X, player.Y
	player.advance(speed)
	player.animate(player.X != previousX || player.Y != previousY)
}

// animate advances the mouth cycle by one tick after movement, or resets it to the
// half-open frame of the facing direction when Pac-Man stood still.
func (player *PacManState) animate(moved bool) {
	if !moved {
		player.animationTicks = 0
		player.animationPhase = 0
		player.animationFrame = pacManHalfMouthFrame(player.lastDirection)
		return
	}
	player.animationTicks++
	if player.animationTicks == pacManAnimationTicksPerFrame {
		player.animationTicks = 0
		player.animationPhase = (player.animationPhase + 1) % 5
	}
	player.animationFrame = pacManFrameForPhase(player.lastDirection, player.animationPhase)
}

// actorSpeed converts a table percentage to logical pixels per simulation tick.
func actorSpeed(percent int) float64 {
	return BaseSpeedPixelsPerSecond * float64(CellSize) / 8 *
		float64(percent) / 100 / simulationTicksPerSecond
}

// pacManFrameForPhase maps the five mouth-cycle phases to the oriented atlas cells.
func pacManFrameForPhase(direction Direction, phase int) int {
	switch phase {
	case 1, 3:
		return pacManHalfMouthFrame(direction)
	case 2:
		return pacManWideMouthFrame(direction)
	default:
		return pacManClosedMouthFrame
	}
}

// pacManHalfMouthFrame returns the atlas cell for a direction's half-open mouth.
func pacManHalfMouthFrame(direction Direction) int {
	switch direction {
	case DirectionUp:
		return 0
	case DirectionDown:
		return 2
	case DirectionRight:
		return 6
	default:
		return 4
	}
}

// pacManWideMouthFrame returns the atlas cell for a direction's fully open mouth.
func pacManWideMouthFrame(direction Direction) int {
	switch direction {
	case DirectionUp:
		return 1
	case DirectionDown:
		return 3
	case DirectionRight:
		return 7
	default:
		return 5
	}
}

// advance moves through connected tiles, stopping at walls and wrapping in the tunnel.
func (player *PacManState) advance(distance float64) {
	for distance > 0 {
		if player.Direction != DirectionNone && player.RequestedDirection == oppositeDirection(player.Direction) {
			player.Direction = player.RequestedDirection
			player.lastDirection = player.RequestedDirection
		}
		if player.corneringCorrection != DirectionNone {
			distance = player.advanceCornering(distance)
			if distance <= 0 {
				return
			}
			continue
		}

		column := tileIndex(player.X, 0)
		row := tileIndex(player.Y, 0)
		atCenter := coordinateAtTileCenter(player.X) && coordinateAtTileCenter(player.Y)
		if player.Direction == DirectionNone {
			if !atCenter || !canMoveFrom(column, row, player.RequestedDirection) {
				return
			}
			player.Direction = player.RequestedDirection
			player.lastDirection = player.RequestedDirection
		}
		if player.tryStartCornering() && player.corneringCorrection != DirectionNone {
			continue
		}

		deltaColumn, deltaRow := directionDelta(player.Direction)
		column = tileIndex(player.X, deltaColumn)
		row = tileIndex(player.Y, deltaRow)
		nextColumn, nextRow, wraps := adjacentTile(column, row, player.Direction)
		if !isWalkable(nextColumn, nextRow) {
			if atCenter {
				player.Direction = DirectionNone
			}
			return
		}
		if wraps {
			targetX := -float64(CellSize) / 2
			if player.Direction == DirectionRight {
				targetX = float64(ScreenWidth) + float64(CellSize)/2
			}
			remaining := math.Abs(targetX - player.X)
			step := math.Min(distance, remaining)
			deltaColumn, _ := directionDelta(player.Direction)
			player.X += float64(deltaColumn) * step
			distance -= step
			if step == remaining {
				if player.Direction == DirectionLeft {
					player.X += float64(ScreenWidth)
				} else {
					player.X -= float64(ScreenWidth)
				}
			}
			continue
		}

		targetX := float64(nextColumn*CellSize + CellSize/2)
		targetY := float64(nextRow*CellSize + CellSize/2)
		step := math.Min(distance, math.Abs(targetX-player.X)+math.Abs(targetY-player.Y))
		deltaX, deltaY := directionDelta(player.Direction)
		player.X += float64(deltaX) * step
		player.Y += float64(deltaY) * step
		distance -= step
	}
}

// oppositeDirection returns the direction directly opposite a cardinal direction.
func oppositeDirection(direction Direction) Direction {
	switch direction {
	case DirectionUp:
		return DirectionDown
	case DirectionDown:
		return DirectionUp
	case DirectionLeft:
		return DirectionRight
	case DirectionRight:
		return DirectionLeft
	default:
		return DirectionNone
	}
}

// tryStartCornering accepts a buffered perpendicular turn around the next tile center.
func (player *PacManState) tryStartCornering() bool {
	if !perpendicularDirections(player.Direction, player.RequestedDirection) {
		return false
	}

	coordinate := player.X
	perpendicularCoordinate := player.Y
	if player.Direction == DirectionUp || player.Direction == DirectionDown {
		coordinate = player.Y
		perpendicularCoordinate = player.X
	}
	if !coordinateAtTileCenter(perpendicularCoordinate) {
		return false
	}
	position := (coordinate - float64(CellSize)/2) / float64(CellSize)
	firstCenter := int(math.Floor(position))
	earlyAllowance, lateAllowance := corneringWindow(player.Direction)
	deltaColumn, deltaRow := directionDelta(player.Direction)
	axisDirection := deltaColumn
	if axisDirection == 0 {
		axisDirection = deltaRow
	}
	for centerIndex := firstCenter; centerIndex <= firstCenter+1; centerIndex++ {
		centerCoordinate := float64(centerIndex*CellSize + CellSize/2)
		offset := (coordinate - centerCoordinate) * float64(axisDirection)
		if offset < -earlyAllowance || offset > lateAllowance {
			continue
		}

		column := centerIndex
		row := tileIndex(perpendicularCoordinate, 0)
		if deltaColumn == 0 {
			column = tileIndex(perpendicularCoordinate, 0)
			row = centerIndex
		}
		if !isWalkable(column, row) || !canMoveFrom(column, row, player.RequestedDirection) {
			continue
		}

		player.Direction = player.RequestedDirection
		player.lastDirection = player.RequestedDirection
		player.corneringTarget = centerCoordinate
		if centerCoordinate > coordinate {
			if deltaColumn != 0 {
				player.corneringCorrection = DirectionRight
			} else {
				player.corneringCorrection = DirectionDown
			}
		} else if centerCoordinate < coordinate {
			if deltaColumn != 0 {
				player.corneringCorrection = DirectionLeft
			} else {
				player.corneringCorrection = DirectionUp
			}
		}
		return true
	}
	return false
}

// advanceCornering moves diagonally one pixel on each axis until lane alignment.
func (player *PacManState) advanceCornering(distance float64) float64 {
	coordinate := player.X
	if player.Direction == DirectionLeft || player.Direction == DirectionRight {
		coordinate = player.Y
	}
	step := math.Min(distance, math.Abs(player.corneringTarget-coordinate))
	directionX, directionY := directionDelta(player.Direction)
	correctionX, correctionY := directionDelta(player.corneringCorrection)
	player.X += float64(directionX+correctionX) * step
	player.Y += float64(directionY+correctionY) * step
	if math.Abs(player.corneringTarget-coordinate)-step < 0.000001 {
		if player.Direction == DirectionUp || player.Direction == DirectionDown {
			player.X = player.corneringTarget
		} else {
			player.Y = player.corneringTarget
		}
		player.corneringCorrection = DirectionNone
	}
	return distance - step
}

// corneringWindow returns early and late turn tolerances scaled to the logical cell size.
func corneringWindow(direction Direction) (float64, float64) {
	scale := float64(CellSize) / 8
	smallAllowance := float64(pacManEarlyCornerPixels) * scale
	largeAllowance := float64(pacManLateCornerPixels) * scale
	if direction == DirectionRight || direction == DirectionUp {
		return smallAllowance, largeAllowance
	}
	return largeAllowance, smallAllowance
}

// perpendicularDirections reports whether two cardinal directions form a right angle.
func perpendicularDirections(first, second Direction) bool {
	firstHorizontal := first == DirectionLeft || first == DirectionRight
	secondHorizontal := second == DirectionLeft || second == DirectionRight
	firstVertical := first == DirectionUp || first == DirectionDown
	secondVertical := second == DirectionUp || second == DirectionDown
	return (firstHorizontal && secondVertical) || (firstVertical && secondHorizontal)
}

// canMoveFrom reports whether a direction leads to a traversable adjacent tile.
func canMoveFrom(column, row int, direction Direction) bool {
	if direction == DirectionNone {
		return false
	}
	nextColumn, nextRow, _ := adjacentTile(column, row, direction)
	return isWalkable(nextColumn, nextRow)
}

// adjacentTile returns the next cell and wraps at either end of the side tunnel.
func adjacentTile(column, row int, direction Direction) (int, int, bool) {
	deltaColumn, deltaRow := directionDelta(direction)
	nextColumn := column + deltaColumn
	nextRow := row + deltaRow
	if row == 17 && direction == DirectionLeft && column == 0 {
		return GridColumns - 1, row, true
	}
	if row == 17 && direction == DirectionRight && column == GridColumns-1 {
		return 0, row, true
	}
	return nextColumn, nextRow, false
}

// directionDelta returns the column and row increments for a cardinal direction.
func directionDelta(direction Direction) (int, int) {
	switch direction {
	case DirectionUp:
		return 0, -1
	case DirectionLeft:
		return -1, 0
	case DirectionDown:
		return 0, 1
	case DirectionRight:
		return 1, 0
	default:
		return 0, 0
	}
}

// tileIndex maps a logical-pixel coordinate to the cell whose center bounds its movement.
func tileIndex(coordinate float64, movement int) int {
	position := (coordinate - float64(CellSize)/2) / float64(CellSize)
	switch {
	case movement > 0:
		return int(math.Floor(position))
	case movement < 0:
		return int(math.Ceil(position))
	default:
		return int(math.Round(position))
	}
}

// coordinateAtTileCenter reports whether a logical-pixel coordinate is tile-centered.
func coordinateAtTileCenter(coordinate float64) bool {
	centerIndex := math.Round((coordinate - float64(CellSize)/2) / float64(CellSize))
	center := centerIndex*float64(CellSize) + float64(CellSize)/2
	return math.Abs(coordinate-center) < 0.000001
}

// isWalkable reports whether a maze cell appears in the traversable-path reference data.
func isWalkable(column, row int) bool {
	if column < 0 || column >= GridColumns || row < 0 || row >= GridRows {
		return false
	}
	return AllowedPath[row][column]
}
