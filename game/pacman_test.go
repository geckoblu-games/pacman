package game

import (
	"math"
	"testing"
)

// TestPacManMovesAtLevelSpeed checks deterministic movement at the configured level speed.
func TestPacManMovesAtLevelSpeed(t *testing.T) {
	player := NewPacManState()
	level := NewLevelState(1)
	for range simulationTicksPerSecond / 2 {
		player.Update(DirectionLeft, level)
	}
	wantX := float64(ScreenWidth)/2 - BaseSpeedPixelsPerSecond*float64(CellSize)/8*
		float64(level.Parameters.PacManSpeedPercent)/100/2
	if player.X < wantX-0.001 || player.X > wantX+0.001 || player.Y != float64(pacManStartRow*CellSize+CellSize/2) {
		t.Fatalf("unexpected half-second position: got (%v, %v), want x %v", player.X, player.Y, wantX)
	}
	if player.Direction != DirectionLeft {
		t.Fatalf("expected leftward movement, got direction %d", player.Direction)
	}
}

// TestPacManReversesInCorridor checks that opposite input reverses movement immediately.
func TestPacManReversesInCorridor(t *testing.T) {
	player := NewPacManState()
	level := NewLevelState(1)
	player.X = float64(10*CellSize + CellSize/2)
	player.Y = float64(32*CellSize + CellSize/2)
	player.Direction = DirectionRight
	player.lastDirection = DirectionRight
	previousX := player.X

	player.Update(DirectionLeft, level)

	if player.Direction != DirectionLeft || player.lastDirection != DirectionLeft || player.X >= previousX {
		t.Fatalf("opposite input should reverse Pac-Man in the corridor: x=%v direction=%d", player.X, player.Direction)
	}
}

// TestPacManTunnelCrossingStaysContinuous checks both directions across the side tunnel.
func TestPacManTunnelCrossingStaysContinuous(t *testing.T) {
	tests := []struct {
		// direction is the direction Pac-Man travels through the tunnel.
		direction Direction
		// startX is the center of the edge tile before crossing.
		startX float64
		// firstX is the expected continuous coordinate before reaching the wrapped tile center.
		firstX float64
		// crossedX is the expected position after crossing and applying the remaining movement.
		crossedX float64
		// wrappedColumn is the logical column while the center crosses the screen edge.
		wrappedColumn int
	}{
		{DirectionLeft, float64(CellSize) / 2, -2, 436, GridColumns - 1},
		{DirectionRight, float64(ScreenWidth) - float64(CellSize)/2, float64(ScreenWidth) + 2, 12, 0},
	}
	for _, test := range tests {
		t.Run(directionName(test.direction), func(t *testing.T) {
			player := NewPacManState()
			player.X = test.startX
			player.Y = float64(17*CellSize + CellSize/2)
			player.Direction = test.direction
			player.lastDirection = test.direction

			player.advance(10)
			if player.X != test.firstX {
				t.Fatalf("tunnel approach should remain continuous: got x=%v, want %v", player.X, test.firstX)
			}
			if column, row := player.OccupiedTile(); column != test.wrappedColumn || row != 17 {
				t.Fatalf("tunnel position should occupy wrapped cell %d,17: got %d,%d", test.wrappedColumn, column, row)
			}
			player.advance(10)
			if player.X != test.crossedX {
				t.Fatalf("tunnel crossing should preserve remaining movement: got x=%v, want %v", player.X, test.crossedX)
			}
		})
	}
}

// TestPacManBuffersTurnsUntilOpenTile checks buffered turns and wall blocking.
func TestPacManBuffersTurnsUntilOpenTile(t *testing.T) {
	player := NewPacManState()
	level := NewLevelState(1)
	for range 20 {
		player.Update(DirectionUp, level)
	}
	if player.Direction != DirectionUp || player.X != float64(12*CellSize+CellSize/2) || player.Y >= float64(pacManStartRow*CellSize+CellSize/2) {
		t.Fatalf("buffered up turn should happen at the open junction: position=(%v,%v), direction=%d", player.X, player.Y, player.Direction)
	}

	for range simulationTicksPerSecond {
		player.Update(DirectionNone, level)
	}
	if player.Direction != DirectionNone || player.Y != float64(23*CellSize+CellSize/2) {
		t.Fatalf("Pac-Man should stop at the wall: position=(%v,%v), direction=%d", player.X, player.Y, player.Direction)
	}
	if player.animationFrame != pacManHalfMouthFrame(DirectionUp) {
		t.Fatalf("stopped Pac-Man should show the half-open up frame, got frame %d", player.animationFrame)
	}
}

// TestPacManCorneringMovesDiagonally verifies a buffered turn starts before the junction center.
func TestPacManCorneringMovesDiagonally(t *testing.T) {
	player := NewPacManState()
	level := NewLevelState(1)
	player.Direction = DirectionRight
	player.lastDirection = DirectionRight
	player.X = float64(12*CellSize+CellSize/2) - float64(pacManEarlyCornerPixels*CellSize)/8
	player.Y = float64(pacManStartRow*CellSize + CellSize/2)
	previousX, previousY := player.X, player.Y

	player.Update(DirectionUp, level)

	deltaX := player.X - previousX
	deltaY := previousY - player.Y
	if player.Direction != DirectionUp || deltaX <= 0 || deltaY <= 0 || math.Abs(deltaX-deltaY) > 0.000001 {
		t.Fatalf("pre-turn should move diagonally toward the junction: position=(%v,%v), direction=%d", player.X, player.Y, player.Direction)
	}
}

// TestPacManCorneringCorrectsPostTurn verifies late turns return to the lane center diagonally.
func TestPacManCorneringCorrectsPostTurn(t *testing.T) {
	player := NewPacManState()
	level := NewLevelState(1)
	player.Direction = DirectionRight
	player.lastDirection = DirectionRight
	player.X = float64(12*CellSize+CellSize/2) + float64(pacManLateCornerPixels*CellSize)/8
	player.Y = float64(pacManStartRow*CellSize + CellSize/2)
	previousX, previousY := player.X, player.Y

	player.Update(DirectionUp, level)

	deltaX := previousX - player.X
	deltaY := previousY - player.Y
	if player.Direction != DirectionUp || deltaX <= 0 || deltaY <= 0 || math.Abs(deltaX-deltaY) > 0.000001 {
		t.Fatalf("post-turn should correct diagonally to the junction: position=(%v,%v), direction=%d", player.X, player.Y, player.Direction)
	}
}

// TestPacManCorneringApproachDirections checks the pre-turn window in all four travel directions.
func TestPacManCorneringApproachDirections(t *testing.T) {
	tests := []struct {
		// direction is Pac-Man's direction before cornering.
		direction Direction
		// requested is the perpendicular direction through the intersection.
		requested Direction
		// wantEarly is the expected early allowance in logical pixels.
		wantEarly float64
		// wantLate is the expected late allowance in logical pixels.
		wantLate float64
	}{
		{DirectionRight, DirectionUp, 6, 8},
		{DirectionLeft, DirectionUp, 8, 6},
		{DirectionUp, DirectionRight, 6, 8},
		{DirectionDown, DirectionRight, 8, 6},
	}
	level := NewLevelState(1)
	intersectionX := float64(6*CellSize + CellSize/2)
	intersectionY := float64(8*CellSize + CellSize/2)
	for _, test := range tests {
		t.Run(directionName(test.direction), func(t *testing.T) {
			early, late := corneringWindow(test.direction)
			if early != test.wantEarly || late != test.wantLate {
				t.Fatalf("cornering window: got (%v,%v), want (%v,%v)", early, late, test.wantEarly, test.wantLate)
			}
			player := NewPacManState()
			player.X, player.Y = intersectionX, intersectionY
			oldDeltaX, oldDeltaY := directionDelta(test.direction)
			if oldDeltaX != 0 {
				player.X -= float64(oldDeltaX) * early
			} else {
				player.Y -= float64(oldDeltaY) * early
			}
			player.Direction = test.direction
			player.lastDirection = test.direction
			previousX, previousY := player.X, player.Y

			player.Update(test.requested, level)

			deltaX, deltaY := player.X-previousX, player.Y-previousY
			requestedDeltaX, requestedDeltaY := directionDelta(test.requested)
			if deltaX*float64(oldDeltaX+requestedDeltaX) < 0 ||
				deltaY*float64(oldDeltaY+requestedDeltaY) < 0 ||
				math.Abs(math.Abs(deltaX)-math.Abs(deltaY)) > 0.000001 {
				t.Fatalf("cornering should move diagonally toward the lane: delta=(%v,%v)", deltaX, deltaY)
			}
		})
	}
}

// directionName returns a readable subtest name for a cardinal direction.
func directionName(direction Direction) string {
	switch direction {
	case DirectionRight:
		return "right"
	case DirectionUp:
		return "up"
	case DirectionDown:
		return "down"
	default:
		return "left"
	}
}

// TestPacManMouthCycle checks the closed, half-open, wide-open, half-open sequence.
func TestPacManMouthCycle(t *testing.T) {
	player := NewPacManState()
	level := NewLevelState(1)
	if player.animationFrame != pacManHalfMouthFrame(DirectionLeft) {
		t.Fatal("stationary Pac-Man should start with the half-open frame")
	}
	player.Update(DirectionNone, level)
	if player.animationFrame != pacManClosedMouthFrame {
		t.Fatalf("first movement frame should be closed, got frame %d", player.animationFrame)
	}
	for range pacManAnimationTicksPerFrame - 1 {
		player.Update(DirectionNone, level)
	}
	if player.animationFrame != pacManHalfMouthFrame(DirectionLeft) {
		t.Fatalf("sixth movement tick should show the half-open frame, got frame %d", player.animationFrame)
	}
	for _, want := range []int{
		pacManWideMouthFrame(DirectionLeft),
		pacManHalfMouthFrame(DirectionLeft),
		pacManClosedMouthFrame,
	} {
		for range pacManAnimationTicksPerFrame {
			player.Update(DirectionNone, level)
		}
		if player.animationFrame != want {
			t.Fatalf("mouth cycle frame: got %d, want %d", player.animationFrame, want)
		}
	}
}

// TestPacManResetsAfterLifeLoss checks that respawn restores the starting position.
func TestPacManResetsAfterLifeLoss(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	state.Match.PacMan.Update(DirectionLeft, state.Match.Level)
	if state.Match.PacMan.X == float64(ScreenWidth)/2 {
		t.Fatal("Pac-Man should move before losing a life")
	}
	if !state.LoseLife() {
		t.Fatal("expected to lose a life")
	}
	finishDeath(t, &state)
	if got, want := state.Match.PacMan, NewPacManState(); got != want {
		t.Fatalf("life loss should reset Pac-Man: got %+v, want %+v", got, want)
	}
}

// TestUpdateActorsConsumesCoinAndPausesMovement checks coin scoring and its one-tick pause.
func TestUpdateActorsConsumesCoinAndPausesMovement(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	column, row := CoinCoords[0][0], CoinCoords[0][1]
	state.Match.PacMan.X = float64(column*CellSize + CellSize/2)
	state.Match.PacMan.Y = float64(row*CellSize + CellSize/2)
	state.Match.PacMan.Direction = DirectionRight
	state.Match.PacMan.lastDirection = DirectionRight
	state.UpdateActors(DirectionNone, false)
	player := &state.Match.PacMan
	if state.Match.Level.Coins[0] || state.Match.Score != 10 || state.HighScore != 10 {
		t.Fatal("entering a coin tile should consume it and award ten points")
	}
	if player.EatingPauseTicks != pacManCoinPauseTicks {
		t.Fatalf("coin should pause Pac-Man for %d tick, got %d", pacManCoinPauseTicks, player.EatingPauseTicks)
	}
	positionAfterEating := player.X

	state.UpdateActors(DirectionNone, false)
	if player.X != positionAfterEating || player.EatingPauseTicks != 0 {
		t.Fatal("Pac-Man should remain stationary during the eating-pause tick")
	}
	state.UpdateActors(DirectionNone, false)
	if player.X <= positionAfterEating || state.Match.Score != 10 {
		t.Fatal("Pac-Man should resume movement without consuming the same coin twice")
	}
}

// TestUpdateActorsConsumesEnergizerAndPausesMovement checks the three-tick energizer pause.
func TestUpdateActorsConsumesEnergizerAndPausesMovement(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	column, row := EnergizerCoords[0][0], EnergizerCoords[0][1]
	state.Match.PacMan.X = float64(column*CellSize + CellSize/2)
	state.Match.PacMan.Y = float64(row*CellSize + CellSize/2)
	state.Match.PacMan.Direction = DirectionNone
	state.UpdateActors(DirectionNone, false)
	if state.Match.Level.Energizers[0] || state.Match.Score != 50 ||
		state.Match.PacMan.EatingPauseTicks != pacManEnergizerPauseTicks {
		t.Fatal("entering an energizer tile should award fifty points and pause for three ticks")
	}
	for range pacManEnergizerPauseTicks {
		state.UpdateActors(DirectionNone, false)
	}
	if state.Match.PacMan.EatingPauseTicks != 0 || state.Match.Score != 50 {
		t.Fatal("energizer pause should expire without consuming the energizer twice")
	}
}
