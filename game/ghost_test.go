package game

import (
	"math"
	"math/rand"
	"testing"
)

// TestGhostModesAndFright checks scheduling, pause, reversal and zero-duration energizers.
func TestGhostModesAndFright(t *testing.T) {
	level := NewLevelState(1)
	ghost := NewBlinky()
	for range 419 {
		ghost.Tick(level)
	}
	if ghost.Mode != GhostScatter {
		t.Fatal("scatter ended too early")
	}
	ghost.Frighten(level)
	for range 360 {
		ghost.Tick(level)
	}
	if ghost.Mode != GhostScatter || ghost.scheduleTicks != 419 {
		t.Fatal("fright should pause and resume the scatter schedule")
	}
	ghost.Tick(level)
	if ghost.Mode != GhostChase || !ghost.reversePending {
		t.Fatal("first chase should begin after 420 active schedule ticks")
	}
	ghost.reversePending = false
	ghost.Frighten(NewLevelState(17))
	if ghost.Mode != GhostChase || !ghost.reversePending || ghost.frightTicks != 0 {
		t.Fatal("zero-duration fright should only signal a reversal")
	}
}

// TestGhostReversesAtTileEntry checks mode changes do not reverse in mid-tile.
func TestGhostReversesAtTileEntry(t *testing.T) {
	level := NewLevelState(1)
	ghost := NewBlinky()
	ghost.X, ghost.Y = float64(10*CellSize+CellSize/2), float64(8*CellSize+CellSize/2)
	ghost.Direction, ghost.turn, ghost.nextTurn = DirectionRight, DirectionRight, DirectionRight
	ghost.Frighten(level)
	for range 6 {
		ghost.Update(level, NewPacManState())
		if ghost.Direction != DirectionRight {
			t.Fatal("reversal occurred before entering the next tile")
		}
	}
	ghost.Update(level, NewPacManState())
	if ghost.Direction != DirectionLeft || ghost.reversePending {
		t.Fatal("reversal should occur upon entering the next tile")
	}
}

// TestFrozenGhostTimers checks disabling movement does not pause mode changes.
func TestFrozenGhostTimers(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	state.Match.PacMan.EatingPauseTicks = 500
	x, y := state.Match.Blinky.X, state.Match.Blinky.Y
	for range 420 {
		state.Tick()
		state.UpdateActors(DirectionNone, false)
	}
	ghost := &state.Match.Blinky
	if ghost.X != x || ghost.Y != y || ghost.Mode != GhostChase || !ghost.reversePending {
		t.Fatal("frozen Blinky should change mode without moving")
	}
	coord := EnergizerCoords[0]
	state.ConsumeCollectible(coord[0], coord[1])
	if ghost.Mode != GhostFrightened || ghost.frightTicks != 360 {
		t.Fatal("energizers should still affect a frozen ghost")
	}
}

// TestGhostDirection checks distance ties, no reversals and the red-zone exception.
func TestGhostDirection(t *testing.T) {
	random := rand.New(rand.NewSource(1))
	if got := ghostDirection(6, 8, DirectionUp, 5, 7, false, false, random); got != DirectionUp {
		t.Fatalf("tie should prefer up over left, got %v", got)
	}
	if got := ghostDirection(12, 14, DirectionLeft, 12, 0, false, false, random); got == DirectionUp || got == DirectionRight {
		t.Fatalf("red zone should forbid up and reversing, got %v", got)
	}
	if got := ghostDirection(12, 14, DirectionLeft, 12, 0, false, true, random); got != DirectionUp {
		t.Fatalf("returning eyes should be allowed to turn up, got %v", got)
	}
}

// TestGhostDirectionHasLegalExit checks connected maze entries have a legal forward choice.
func TestGhostDirectionHasLegalExit(t *testing.T) {
	random := rand.New(rand.NewSource(1))
	for row := range GridRows {
		for column := range GridColumns {
			if !isWalkable(column, row) {
				continue
			}
			for direction := DirectionUp; direction <= DirectionRight; direction++ {
				if !canMoveFrom(column, row, oppositeDirection(direction)) {
					continue
				}
				for _, frightened := range []bool{false, true} {
					if ghostDirection(column, row, direction, 13, 14, frightened, false, random) == DirectionNone {
						t.Fatalf("no legal exit at %d,%d moving %v, frightened=%v", column, row, direction, frightened)
					}
				}
			}
		}
	}
}

// TestFrightenedGhostStaysOnPaths checks movement across repeated mode reversals.
func TestFrightenedGhostStaysOnPaths(t *testing.T) {
	player := NewPacManState()
	for _, number := range []int{1, 5} {
		level := NewLevelState(number)
		for _, interval := range []int{137, 509} {
			for seed := int64(1); seed <= 8; seed++ {
				ghost := NewBlinky()
				ghost.random = rand.New(rand.NewSource(seed))
				for tick := 0; tick < 3000; tick++ {
					if tick%interval == 83 {
						ghost.Frighten(level)
					}
					before := ghost
					ghost.Tick(level)
					ghost.Update(level, player)
					column, row := ghost.OccupiedTile()
					if !isWalkable(column, row) {
						t.Fatalf("level %d interval %d seed %d tick %d: ghost entered wall %d,%d, before %+v, after %+v", number, interval, seed, tick, column, row, before, ghost)
					}
				}
			}
		}
	}
}

// TestGhostReversalBeforeIntersection checks that reversal replans the return turn.
func TestGhostReversalBeforeIntersection(t *testing.T) {
	level := NewLevelState(1)
	for _, frightened := range []bool{false, true} {
		ghost := NewBlinky()
		ghost.X = float64(10*CellSize) - 0.5
		ghost.Y = float64(14*CellSize + CellSize/2)
		ghost.Direction, ghost.turn, ghost.nextTurn = DirectionRight, DirectionRight, DirectionRight
		if frightened {
			ghost.Frighten(level)
		} else {
			ghost.reversePending = true
		}
		turned := false
		for tick := 0; tick < 30; tick++ {
			ghost.Update(level, NewPacManState())
			column, row := ghost.OccupiedTile()
			if !isWalkable(column, row) {
				t.Fatalf("frightened=%v: reversed ghost entered wall %d,%d at tick %d", frightened, column, row, tick)
			}
			if ghost.Direction == DirectionDown {
				turned = true
				break
			}
		}
		if !turned {
			t.Fatalf("frightened=%v: reversed ghost should turn down at the intersection", frightened)
		}
	}
}

// TestBlinkyMovement checks sustained navigation, legal paths, and tunnel traversal.
func TestBlinkyMovement(t *testing.T) {
	level := NewLevelState(1)
	player := NewPacManState()
	ghost := NewBlinky()
	startX := ghost.X
	visited := map[[2]int]bool{}
	for tick := 0; tick < 10000; tick++ {
		ghost.Tick(level)
		ghost.Update(level, player)
		column, row := ghost.OccupiedTile()
		if !isWalkable(column, row) {
			t.Fatalf("tick %d: ghost left the maze at %d,%d, position %.2f,%.2f", tick, column, row, ghost.X, ghost.Y)
		}
		visited[[2]int{column, row}] = true
	}
	if ghost.X == startX || len(visited) < 30 {
		t.Fatalf("ghost failed to navigate the maze: visited %d tiles", len(visited))
	}
}

// TestBlinkyLookahead checks that an intersection decision survives a later target change.
func TestBlinkyLookahead(t *testing.T) {
	level := NewLevelState(1)
	ghost := NewBlinky()
	ghost.Mode = GhostChase
	ghost.X, ghost.Y = float64(10*CellSize+CellSize/2), float64(8*CellSize+CellSize/2)
	ghost.Direction, ghost.turn, ghost.nextTurn = DirectionRight, DirectionRight, DirectionRight
	player := NewPacManState()
	player.X, player.Y = float64(12*CellSize+CellSize/2), float64(4*CellSize+CellSize/2)
	for {
		ghost.Update(level, player)
		column, _ := ghost.OccupiedTile()
		if column == 11 {
			break
		}
	}
	if ghost.nextTurn != DirectionUp {
		t.Fatal("Blinky should choose up for the approaching intersection")
	}
	player.X, player.Y = float64(26*CellSize+CellSize/2), float64(8*CellSize+CellSize/2)
	for tick := 0; tick < 30; tick++ {
		ghost.Update(level, player)
		if ghost.Direction == DirectionUp {
			return
		}
	}
	t.Fatal("Blinky should retain his previous decision at the intersection")
}

// TestGhostPassThrough checks opposite tile swaps do not trigger swept collisions.
func TestGhostPassThrough(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	boundary := float64(7 * CellSize)
	state.Match.PacMan.X, state.Match.PacMan.Y = boundary-0.5, float64(8*CellSize+CellSize/2)
	state.Match.PacMan.Direction = DirectionRight
	ghost := &state.Match.Blinky
	ghost.X, ghost.Y = boundary+0.5, state.Match.PacMan.Y
	ghost.Direction, ghost.turn, ghost.nextTurn = DirectionLeft, DirectionLeft, DirectionLeft
	state.UpdateActors(DirectionNone, true)
	playerColumn, _ := state.Match.PacMan.OccupiedTile()
	ghostColumn, _ := ghost.OccupiedTile()
	if state.Match.Lives != InitialLives || playerColumn != 7 || ghostColumn != 6 {
		t.Fatal("actors swapping tiles in one tick should pass through")
	}
}

// TestGhostTunnel checks continuous wrapping and the tunnel speed penalty.
func TestGhostTunnel(t *testing.T) {
	level := NewLevelState(1)
	ghost := NewBlinky()
	ghost.X, ghost.Y = float64(CellSize/2), float64(17*CellSize+CellSize/2)
	ghost.Direction, ghost.turn, ghost.nextTurn = DirectionLeft, DirectionLeft, DirectionLeft
	if ghost.speedPercent(level) != 40 {
		t.Fatal("Blinky should use tunnel speed at the edge")
	}
	player := NewPacManState()
	wrapped := false
	for tick := 0; tick < 30; tick++ {
		previous := ghost.X
		ghost.Update(level, player)
		if ghost.X > previous+float64(ScreenWidth)/2 {
			wrapped = true
			if ghost.X >= float64(ScreenWidth) {
				t.Fatal("wrapped center should enter at the opposite edge")
			}
			break
		}
	}
	if !wrapped {
		t.Fatal("ghost should wrap through the side tunnel")
	}
}

// TestGhostCollisionAndFreeze checks that a frozen ghost remains dangerous and vulnerable.
func TestGhostCollisionAndFreeze(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	state.Match.PacMan.X, state.Match.PacMan.Y = state.Match.Blinky.X, state.Match.Blinky.Y
	state.Match.PacMan.EatingPauseTicks = 1
	state.UpdateActors(DirectionNone, false)
	if state.Match.Lives != 3 || state.Phase != PhaseDying {
		t.Fatal("a collision with stationary Blinky should start the death sequence")
	}
	finishDeath(t, &state)
	if state.Match.Lives != 2 || !state.Match.Blinky.ElroySuspended {
		t.Fatal("death should cost a life and suspend Elroy")
	}
	state.BeginPlaying()
	ghost := &state.Match.Blinky
	ghost.Frighten(state.Match.Level)
	x, y := ghost.X, ghost.Y
	state.Tick()
	state.Match.PacMan.X, state.Match.PacMan.Y = x, y
	state.Match.PacMan.EatingPauseTicks = 1
	state.UpdateActors(DirectionNone, false)
	if ghost.X != x || ghost.Y != y || ghost.Mode != GhostFrightened ||
		state.Phase != PhaseGhostCaptured || state.Match.Score != 200 {
		t.Fatal("a frozen frightened ghost should start the capture pause without moving")
	}
}

// TestGhostCapturePause checks score, actor visibility, paused timers, and returning eyes.
func TestGhostCapturePause(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	state.Match.Fruit = FruitState{Visible: true, Ticks: 300}
	ghost := &state.Match.Blinky
	ghost.Frighten(state.Match.Level)
	state.Match.PacMan.X, state.Match.PacMan.Y = ghost.X, ghost.Y
	state.Match.PacMan.EatingPauseTicks = 1
	state.UpdateActors(DirectionNone, false)
	if state.Phase != PhaseGhostCaptured || state.PacManVisible() || state.Match.Score != 200 ||
		state.Match.GhostCapture != (GhostCaptureState{X: ghost.X, Y: ghost.Y, Points: 200}) {
		t.Fatal("capture should hide Pac-Man and show 200 points at the ghost")
	}
	player, beforeGhost, fruit := state.Match.PacMan, *ghost, state.Match.Fruit
	state.TogglePause()
	for range 10 {
		state.Tick()
	}
	if state.PhaseTicks != 0 {
		t.Fatal("pause should stop the capture countdown")
	}
	state.TogglePause()
	for tick := 0; tick < GhostCaptureTicks; tick++ {
		if state.Phase != PhaseGhostCaptured || state.PhaseTicks != tick ||
			state.Match.PacMan != player || *ghost != beforeGhost || state.Match.Fruit != fruit {
			t.Fatalf("capture should freeze actors and timers at tick %d", tick)
		}
		state.UpdateActors(DirectionRight, true)
		state.Tick()
	}
	if state.Phase != PhasePlaying || !state.PacManVisible() || ghost.Mode != GhostEaten ||
		ghost.X != beforeGhost.X || ghost.Y != beforeGhost.Y || ghost.frightTicks != beforeGhost.frightTicks ||
		state.Match.GhostCapture != (GhostCaptureState{}) {
		t.Fatal("capture should resume play with eyes and the frightened timer preserved")
	}
	state.UpdateActors(DirectionNone, true)
	if ghost.X == beforeGhost.X && ghost.Y == beforeGhost.Y {
		t.Fatal("eyes should move when play resumes")
	}
}

// TestGhostCaptureScores checks the score sequence for prior captures in one energizer.
func TestGhostCaptureScores(t *testing.T) {
	for streak, points := range []int{200, 400, 800, 1600} {
		state := GameState{}
		state.StartNewMatch()
		state.BeginPlaying()
		state.Match.GhostStreak = streak
		ghost := &state.Match.Blinky
		ghost.Frighten(state.Match.Level)
		state.Match.PacMan.X, state.Match.PacMan.Y = ghost.X, ghost.Y
		state.Match.PacMan.EatingPauseTicks = 1
		state.UpdateActors(DirectionNone, false)
		if state.Match.Score != points || state.HighScore != points ||
			state.Match.GhostCapture.Points != points || state.Match.GhostStreak != streak+1 {
			t.Fatalf("capture %d should award %d points", streak+1, points)
		}
	}
}

// TestGhostStreakResets checks new energizers and fright expiration restart scoring.
func TestGhostStreakResets(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	state.Match.GhostStreak = 3
	coord := EnergizerCoords[0]
	if !state.ConsumeCollectible(coord[0], coord[1]) || state.Match.GhostStreak != 0 {
		t.Fatal("a new energizer should reset the ghost streak")
	}
	state.Match.GhostStreak = 2
	state.Match.Blinky.frightTicks = 1
	state.Tick()
	if state.Match.GhostStreak != 0 {
		t.Fatal("fright expiration should reset the ghost streak")
	}
}

// TestBlinkyElroy checks both thresholds, scatter targeting, and suspension after death.
func TestBlinkyElroy(t *testing.T) {
	level := NewLevelState(1)
	ghost := NewBlinky()
	player := NewPacManState()
	level.Remaining = 20
	if ghost.speedPercent(level) != 80 {
		t.Fatal("first Elroy speed should begin at 20 dots")
	}
	x, y := ghost.target(level, player)
	px, py := player.OccupiedTile()
	if x != px || y != py {
		t.Fatal("Elroy should target Pac-Man during scatter")
	}
	level.Remaining = 10
	if ghost.speedPercent(level) != 85 {
		t.Fatal("second Elroy speed should begin at 10 dots")
	}
	ghost.ElroySuspended = true
	if ghost.speedPercent(level) != 75 {
		t.Fatal("Elroy should remain suspended after death")
	}
}

// TestEatenBlinkyReturns checks eyes reach the house and Blinky exits again.
func TestEatenBlinkyReturns(t *testing.T) {
	level := NewLevelState(1)
	ghost := NewBlinky()
	ghost.Mode = GhostEaten
	player := NewPacManState()
	entered := false
	for tick := 0; tick < 5000; tick++ {
		ghost.Tick(level)
		ghost.Update(level, player)
		if ghost.houseStage != 0 {
			entered = true
		}

		if entered && ghost.houseStage == 0 && ghost.Mode != GhostEaten {
			if math.Abs(ghost.X-float64(ScreenWidth)/2) > 0.001 {
				t.Fatal("Blinky should exit at the center of the house")
			}
			return
		}
	}
	t.Fatalf("eyes did not return and exit: %+v", ghost)
}

// TestEyesSpeed checks actual movement at every level, including both tunnels.
func TestEyesSpeed(t *testing.T) {
	player := NewPacManState()
	for number := 1; number <= 256; number++ {
		level := NewLevelState(number)
		for _, position := range [][2]int{{4, 4}, {3, 17}, {24, 17}} {
			ghost := NewBlinky()
			ghost.Mode = GhostEaten
			ghost.X = float64(position[0]*CellSize + CellSize/2)
			ghost.Y = float64(position[1]*CellSize + CellSize/2)
			ghost.Direction, ghost.turn, ghost.nextTurn = DirectionRight, DirectionRight, DirectionRight
			start := ghost.X
			ghost.Update(level, player)
			// 151.5151525 original pixels/s scaled from 8 to 16 pixels per tile at 60 Hz.
			const expectedStep = 151.5151525 * 2 / 60
			if delta := ghost.X - start; math.Abs(delta-expectedStep) > 0.0000001 {
				t.Fatalf("level %d, tile %v: eyes moved %v, want %v", number, position, delta, expectedStep)
			}
		}
	}
}

// TestEyesIgnorePacMan checks return paths do not depend on Pac-Man or random choices.
func TestEyesIgnorePacMan(t *testing.T) {
	level := NewLevelState(1)
	first, second := NewBlinky(), NewBlinky()
	first.Mode, second.Mode = GhostEaten, GhostEaten
	first.random, second.random = rand.New(rand.NewSource(1)), rand.New(rand.NewSource(99))
	first.X, first.Y = float64(CellSize+CellSize/2), float64(4*CellSize+CellSize/2)
	second.X, second.Y = first.X, first.Y
	first.Direction, first.turn, first.nextTurn = DirectionRight, DirectionRight, DirectionRight
	second.Direction, second.turn, second.nextTurn = DirectionRight, DirectionRight, DirectionRight
	firstPlayer, secondPlayer := NewPacManState(), NewPacManState()
	secondPlayer.X, secondPlayer.Y = float64(26*CellSize+CellSize/2), float64(4*CellSize+CellSize/2)
	for tick := 0; tick < 200; tick++ {
		first.Update(level, firstPlayer)
		second.Update(level, secondPlayer)
		if first.X != second.X || first.Y != second.Y || first.Direction != second.Direction {
			t.Fatalf("eyes followed different paths at tick %d", tick)
		}
		if first.houseStage != 0 {
			return
		}
	}
}

// TestEyesReturnFromMaze checks returns from several moving positions and directions.
func TestEyesReturnFromMaze(t *testing.T) {
	level := NewLevelState(1)
	player := NewPacManState()
	ghost := NewBlinky()
	for sample := 0; sample < 16; sample++ {
		for range 200 {
			ghost.Tick(level)
			ghost.Update(level, player)
		}
		eyes := ghost
		eyes.Mode = GhostEaten
		eyes.reversePending = false
		for tick := 0; tick < 5000 && eyes.houseStage == 0; tick++ {
			eyes.Tick(level)
			eyes.Update(level, player)
		}
		if eyes.houseStage == 0 {
			t.Fatalf("eyes from sample %d did not reach house: start %+v, end %+v", sample, ghost, eyes)
		}
	}
}

// TestPinkyTargets checks all chase directions, the up bug, and the scatter corner.
func TestPinkyTargets(t *testing.T) {
	level := NewLevelState(1)
	level.Remaining = 1
	ghost := newPinky(NewBlinky().ghostClock, false)
	player := NewPacManState()
	player.X, player.Y = float64(10*CellSize+CellSize/2), float64(10*CellSize+CellSize/2)
	ghost.Mode = GhostChase
	for _, test := range []struct {
		direction Direction
		x, y      int
	}{
		{DirectionUp, 6, 6},
		{DirectionLeft, 6, 10},
		{DirectionDown, 10, 14},
		{DirectionRight, 14, 10},
	} {
		player.Direction = test.direction
		if x, y := ghost.target(level, player); x != test.x || y != test.y {
			t.Errorf("direction %v: got (%d,%d), want (%d,%d)", test.direction, x, y, test.x, test.y)
		}
	}
	player.Direction, player.lastDirection = DirectionNone, DirectionUp
	if x, y := ghost.target(level, player); x != 6 || y != 6 {
		t.Fatal("a stopped Pac-Man should retain his facing direction")
	}
	ghost.Mode = GhostScatter
	if x, y := ghost.target(level, player); x != 2 || y != 0 || ghost.speedPercent(level) != level.Parameters.GhostSpeedPercent {
		t.Fatal("Pinky should use her scatter corner and normal speed even at Elroy thresholds")
	}
}

// TestPinkySharedModes checks both ghosts share and pause one schedule and fright timer.
func TestPinkySharedModes(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	match := state.Match
	if match.Blinky.ghostClock != match.Pinky.ghostClock {
		t.Fatal("ghosts must share the same clock")
	}
	match.Pinky.houseStage = 0
	for range 419 {
		state.Tick()
	}
	coord := EnergizerCoords[0]
	state.ConsumeCollectible(coord[0], coord[1])
	if match.Blinky.Mode != GhostFrightened || match.Pinky.Mode != GhostFrightened {
		t.Fatal("energizers should frighten both ghosts")
	}
	for range 360 {
		state.Tick()
	}
	if match.Blinky.scheduleTicks != 419 || match.Blinky.Mode != GhostScatter || match.Pinky.Mode != GhostScatter {
		t.Fatal("fright should pause the shared schedule and restore both ghosts")
	}
	state.Tick()
	if match.Blinky.Mode != GhostChase || match.Pinky.Mode != GhostChase ||
		!match.Blinky.reversePending || !match.Pinky.reversePending {
		t.Fatal("the shared transition should reverse both ghosts")
	}
}

// TestPinkyHouseRelease checks startup, post-death checkpoints, and the persistent counter.
func TestPinkyHouseRelease(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	if state.Match.Pinky.houseStage != 3 || state.Match.GlobalGhostCounter {
		t.Fatal("Pinky should immediately exit on a new level")
	}
	state.BeginPlaying()
	state.LoseLife()
	finishDeath(t, &state)
	state.BeginPlaying()
	match := state.Match
	if !match.GlobalGhostCounter || match.Pinky.houseStage != 4 {
		t.Fatal("Pinky should wait for the global counter after a death")
	}
	for index, coord := range CoinCoords[:32] {
		if !state.ConsumeCollectible(coord[0], coord[1]) {
			t.Fatalf("dot %d was not consumed", index+1)
		}
		if index == 5 && match.Pinky.houseStage != 4 {
			t.Fatal("Pinky left before the seventh dot")
		}
		if index == 6 && match.Pinky.houseStage != 3 {
			t.Fatal("the seventh dot should release Pinky")
		}
	}
	if !match.GlobalGhostCounter || match.GlobalGhostDots != 32 {
		t.Fatal("the global counter must remain active without Clyde")
	}
	match.Pinky.houseStage = 4
	for range 239 {
		state.Tick()
	}
	if match.Pinky.houseStage != 4 {
		t.Fatal("Pinky should wait until the anti-stall timer expires")
	}
	state.Tick()
	if match.Pinky.houseStage != 3 {
		t.Fatal("the anti-stall timer should release Pinky after four seconds")
	}
	state.enterPhase(PhaseLevelComplete)
	state.AdvanceLevel()
	if state.Match.GlobalGhostCounter || state.Match.GlobalGhostDots != 0 || state.Match.Pinky.houseStage != 3 {
		t.Fatal("a new level should reset the counter and release Pinky immediately")
	}
}

// TestPinkyAntiStall checks the shorter level-five timer and collectible reset.
func TestPinkyAntiStall(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.Match.Level = NewLevelState(5)
	state.BeginPlaying()
	state.LoseLife()
	finishDeath(t, &state)
	state.BeginPlaying()
	for range 179 {
		state.Tick()
	}
	coord := EnergizerCoords[0]
	state.ConsumeCollectible(coord[0], coord[1])
	for range 179 {
		state.Tick()
	}
	if state.Match.Pinky.houseStage != 4 {
		t.Fatal("an energizer should reset the anti-stall timer")
	}
	state.Tick()
	if state.Match.Pinky.houseStage != 3 {
		t.Fatal("level five should release Pinky after three seconds without dots")
	}
}

// TestPinkyExitDirection checks mode changes inside the house determine the exit turn.
func TestPinkyExitDirection(t *testing.T) {
	for _, changed := range []bool{false, true} {
		state := GameState{}
		state.StartNewMatch()
		state.BeginPlaying()
		match := state.Match
		match.Pinky.houseStage = 4
		if changed {
			for range 420 {
				state.Tick()
			}
		}
		match.releasePinky()
		for tick := 0; tick < 100 && match.Pinky.houseStage != 0; tick++ {
			match.Pinky.Update(match.Level, match.PacMan)
		}
		want := DirectionLeft
		if changed {
			want = DirectionRight
		}
		if match.Pinky.houseStage != 0 || match.Pinky.Direction != want {
			t.Fatalf("changed=%v: exit direction %v, want %v", changed, match.Pinky.Direction, want)
		}
		for range 1000 {
			match.Pinky.Tick(match.Level)
			match.Pinky.Update(match.Level, match.PacMan)
			x, y := match.Pinky.OccupiedTile()
			if !isWalkable(x, y) {
				t.Fatalf("changed=%v: Pinky entered wall (%d,%d)", changed, x, y)
			}
		}
	}
}

// TestPinkyCapture checks ghost-specific visibility, frozen timers and sequential scoring.
func TestPinkyCapture(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	match := state.Match
	match.Pinky.houseStage = 0
	match.Pinky.X, match.Pinky.Y = float64(3*CellSize+CellSize/2), float64(4*CellSize+CellSize/2)
	coord := EnergizerCoords[0]
	state.ConsumeCollectible(coord[0], coord[1])
	match.PacMan.X, match.PacMan.Y = match.Blinky.X, match.Blinky.Y
	match.PacMan.EatingPauseTicks = 1
	state.UpdateActors(DirectionNone, false)
	if state.Phase != PhaseGhostCaptured || state.GhostVisible(GhostBlinky) || !state.GhostVisible(GhostPinky) {
		t.Fatal("only captured Blinky should be hidden")
	}
	fright, other := match.Blinky.frightTicks, match.Pinky
	for range GhostCaptureTicks {
		state.Tick()
	}
	if match.Blinky.Mode != GhostEaten || match.Pinky != other || match.Blinky.frightTicks != fright {
		t.Fatal("capture should freeze Pinky and the shared timer and turn only Blinky into eyes")
	}
	match.PacMan.X, match.PacMan.Y = match.Pinky.X, match.Pinky.Y
	match.PacMan.EatingPauseTicks = 1
	state.UpdateActors(DirectionNone, false)
	if state.Phase != PhaseGhostCaptured || state.GhostVisible(GhostPinky) || !state.GhostVisible(GhostBlinky) ||
		match.GhostCapture.Kind != GhostPinky || match.GhostCapture.Points != 400 {
		t.Fatal("captured Pinky should be hidden and award the second score")
	}
	for range GhostCaptureTicks {
		state.Tick()
	}
	if match.Pinky.Mode != GhostEaten || match.Blinky.Mode != GhostEaten {
		t.Fatal("Pinky should become eyes after her own capture pause")
	}
}

// TestPinkyEyesReturn checks that revived Pinky exits immediately or waits after death.
func TestPinkyEyesReturn(t *testing.T) {
	for _, global := range []bool{false, true} {
		state := GameState{}
		state.StartNewMatch()
		state.BeginPlaying()
		match := state.Match
		match.GlobalGhostCounter = global
		match.Pinky = newPinky(match.Blinky.ghostClock, false)
		match.Pinky.houseStage = 0
		match.Pinky.X, match.Pinky.Y = match.Blinky.X, match.Blinky.Y
		match.Pinky.Direction, match.Pinky.turn, match.Pinky.nextTurn = DirectionLeft, DirectionLeft, DirectionLeft
		match.Pinky.Mode = GhostEaten
		match.PacMan.EatingPauseTicks = 5000
		entered := false
		for tick := 0; tick < 5000; tick++ {
			state.UpdateActors(DirectionNone, true)
			if match.Pinky.houseStage != 0 {
				entered = true
			}
			if match.Pinky.Mode != GhostEaten {
				break
			}
		}
		if !entered || (global && match.Pinky.houseStage != 4) || (!global && match.Pinky.houseStage != 3) {
			t.Fatalf("global=%v: unexpected Pinky state after returning: %+v", global, match.Pinky)
		}
	}
}

// TestPinkyFrightenedPaths checks Pinky's navigation through repeated reversals.
func TestPinkyFrightenedPaths(t *testing.T) {
	for _, number := range []int{1, 5} {
		for seed := int64(1); seed <= 4; seed++ {
			level := NewLevelState(number)
			ghost := newPinky(NewBlinky().ghostClock, false)
			ghost.random = rand.New(rand.NewSource(seed))
			player := NewPacManState()
			for tick := 0; tick < 3000; tick++ {
				if tick%137 == 83 {
					ghost.Frighten(level)
				}
				ghost.Tick(level)
				ghost.Update(level, player)
				x, y := ghost.OccupiedTile()
				if ghost.houseStage == 0 && !isWalkable(x, y) {
					t.Fatalf("level %d seed %d tick %d: Pinky entered wall (%d,%d)", number, seed, tick, x, y)
				}
			}
		}
	}
}

// TestPinkyDoorCollision checks collision and capture while Pinky exits the house.
func TestPinkyDoorCollision(t *testing.T) {
	for _, frightened := range []bool{false, true} {
		state := GameState{}
		state.StartNewMatch()
		state.BeginPlaying()
		match := state.Match
		match.Blinky.X, match.Blinky.Y = float64(3*CellSize+CellSize/2), float64(4*CellSize+CellSize/2)
		match.Pinky.X, match.Pinky.Y = float64(ScreenWidth)/2, float64(15*CellSize)-1
		match.PacMan.X, match.PacMan.Y = match.Pinky.X, match.Pinky.Y
		match.PacMan.EatingPauseTicks = 1
		if frightened {
			coord := EnergizerCoords[0]
			state.ConsumeCollectible(coord[0], coord[1])
		}
		state.UpdateActors(DirectionNone, false)
		if !frightened {
			if state.Phase != PhaseDying {
				t.Fatal("Pinky should be dangerous as soon as she reaches the maze tile")
			}
			continue
		}
		if state.Phase != PhaseGhostCaptured || match.GhostCapture.Kind != GhostPinky {
			t.Fatal("frightened Pinky should be capturable at the door")
		}
		for range GhostCaptureTicks {
			state.Tick()
		}
		if match.Pinky.Mode != GhostEaten || match.Pinky.houseStage != 2 {
			t.Fatal("eyes captured while exiting should return into the house")
		}
	}
}

// TestPinkyFrozenCollision checks disabling movement does not disable danger or timers.
func TestPinkyFrozenCollision(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	match := state.Match
	match.Pinky.houseStage = 0
	match.Pinky.X, match.Pinky.Y = float64(3*CellSize+CellSize/2), float64(4*CellSize+CellSize/2)
	x, y := match.Pinky.X, match.Pinky.Y
	match.PacMan.EatingPauseTicks = 500
	for range 420 {
		state.Tick()
		state.UpdateActors(DirectionNone, false)
	}
	if match.Pinky.X != x || match.Pinky.Y != y || match.Pinky.Mode != GhostChase {
		t.Fatal("frozen Pinky should remain in place while her mode changes")
	}
	match.PacMan.X, match.PacMan.Y = x, y
	state.UpdateActors(DirectionNone, false)
	if state.Phase != PhaseDying {
		t.Fatal("a frozen Pinky should still be dangerous")
	}
}

// TestInkyTarget checks all offsets, the up bug, and Blinky's live position.
func TestInkyTarget(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	match := state.Match
	match.Level.Remaining = 1
	match.PacMan.X, match.PacMan.Y = float64(10*CellSize+CellSize/2), float64(10*CellSize+CellSize/2)
	match.Blinky.X, match.Blinky.Y = float64(6*CellSize+CellSize/2), float64(7*CellSize+CellSize/2)
	match.Inky.Mode = GhostChase
	for _, test := range []struct {
		// direction is Pac-Man's movement.
		direction Direction
		// x and y are the expected target coordinates.
		x, y int
	}{
		{DirectionUp, 10, 9},
		{DirectionLeft, 10, 13},
		{DirectionDown, 14, 17},
		{DirectionRight, 18, 13},
	} {
		match.PacMan.Direction = test.direction
		x, y := match.Inky.target(match.Level, match.PacMan)
		if x != test.x || y != test.y {
			t.Errorf("direction %v: got (%d,%d), want (%d,%d)", test.direction, x, y, test.x, test.y)
		}
	}
	match.PacMan.Direction, match.PacMan.lastDirection = DirectionNone, DirectionUp
	match.Blinky.X += 3 * CellSize
	if x, y := match.Inky.target(match.Level, match.PacMan); x != 7 || y != 9 {
		t.Fatal("Inky should use Pac-Man's last facing and Blinky's current tile")
	}
	match.Inky.Mode = GhostScatter
	if x, y := match.Inky.target(match.Level, match.PacMan); x != 27 || y != 35 ||
		match.Inky.speedPercent(match.Level) != match.Level.Parameters.GhostSpeedPercent {
		t.Fatal("Inky should use the bottom-right scatter target without Elroy speed")
	}
	match.Inky.Mode = GhostEaten
	if x, y := match.Inky.target(match.Level, match.PacMan); x != 13 || y != 14 {
		t.Fatal("Inky's eyes should use the shared house target")
	}
}

// TestInkyPersonalRelease checks level-one dots, revival, and zero limits on later levels.
func TestInkyPersonalRelease(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	match := state.Match
	if match.Inky.houseStage != 4 || match.Inky.X != float64(ScreenWidth)/2-2*CellSize {
		t.Fatal("Inky should start waiting on the left side of the house")
	}
	for _, coord := range CoinCoords[:29] {
		state.ConsumeCollectible(coord[0], coord[1])
	}
	if match.InkyDots != 29 || match.Inky.houseStage != 4 {
		t.Fatal("Inky should still wait after 29 collectibles")
	}
	coord := EnergizerCoords[0]
	state.ConsumeCollectible(coord[0], coord[1])
	if match.InkyDots != 30 || match.Inky.houseStage != 5 {
		t.Fatal("the thirtieth collectible should release Inky")
	}
	for tick := 0; tick < 200 && match.Inky.houseStage != 0; tick++ {
		match.Inky.Update(match.Level, match.PacMan)
	}
	if match.Inky.houseStage != 0 || match.Inky.Direction != DirectionLeft {
		t.Fatal("Inky should align with the door, exit, and turn left")
	}
	match.Inky.houseStage = 4
	match.Inky.X = float64(ScreenWidth) / 2
	state.UpdateActors(DirectionNone, false)
	if match.Inky.houseStage != 3 {
		t.Fatal("a revived Inky should retain his reached personal limit")
	}
	for number := 2; number <= 3; number++ {
		state.enterPhase(PhaseLevelComplete)
		state.AdvanceLevel()
		if match.Level.Number != number || match.InkyDots != 0 || match.Inky.houseStage != 5 ||
			match.Inky.blinky != &match.Blinky || match.Inky.ghostClock != match.Blinky.ghostClock {
			t.Fatalf("level %d should reset Inky and immediately release him", number)
		}
	}
}

// TestInkyPostDeathRelease checks the global checkpoint and retained personal count.
func TestInkyPostDeathRelease(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	for _, coord := range CoinCoords[:3] {
		state.ConsumeCollectible(coord[0], coord[1])
	}
	state.LoseLife()
	finishDeath(t, &state)
	state.BeginPlaying()
	match := state.Match
	if match.InkyDots != 3 || match.Inky.houseStage != 4 || !match.GlobalGhostCounter {
		t.Fatal("death should preserve but disable Inky's personal counter")
	}
	for index, coord := range CoinCoords[3:35] {
		state.ConsumeCollectible(coord[0], coord[1])
		if index == 15 && match.Inky.houseStage != 4 {
			t.Fatal("Inky should wait until the seventeenth post-death collectible")
		}
		if index == 16 && match.Inky.houseStage != 5 {
			t.Fatal("the seventeenth post-death collectible should release Inky")
		}
	}
	if !match.GlobalGhostCounter || match.InkyDots != 3 {
		t.Fatal("Clyde's absence should keep the global counter active and personal count unchanged")
	}
	match.Inky.houseStage = 4
	match.Inky.X = float64(ScreenWidth) / 2
	state.UpdateActors(DirectionNone, false)
	if match.Inky.houseStage != 4 {
		t.Fatal("Inky should not be released again after missing his checkpoint")
	}
	for range 240 {
		state.Tick()
	}
	if match.Inky.houseStage != 3 {
		t.Fatal("anti-stall should eventually release revived Inky")
	}
}

// TestInkyReleasePriority checks only the highest-priority waiting ghost is forced out.
func TestInkyReleasePriority(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	state.LoseLife()
	finishDeath(t, &state)
	state.BeginPlaying()
	for range 240 {
		state.Tick()
	}
	if state.Match.Pinky.houseStage != 3 || state.Match.Inky.houseStage != 4 {
		t.Fatal("the first anti-stall timeout should release Pinky only")
	}
	for range 240 {
		state.Tick()
	}
	if state.Match.Inky.houseStage != 5 {
		t.Fatal("the next timeout should release Inky")
	}
}

// TestInkyCounterAfterForcedExit checks a partial personal count survives revival.
func TestInkyCounterAfterForcedExit(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	match := state.Match
	for _, coord := range CoinCoords[:7] {
		state.ConsumeCollectible(coord[0], coord[1])
	}
	for range 240 {
		state.Tick()
	}
	if match.InkyDots != 7 || match.Inky.houseStage != 5 {
		t.Fatal("anti-stall should leave the personal count intact")
	}
	for tick := 0; tick < 200 && match.Inky.houseStage != 0; tick++ {
		match.Inky.Update(match.Level, match.PacMan)
	}
	if match.Inky.houseStage != 0 {
		t.Fatal("forced Inky should leave the house")
	}
	match.Inky.Mode = GhostEaten
	for tick := 0; tick < 5000 && match.Inky.Mode == GhostEaten; tick++ {
		match.Inky.Update(match.Level, match.PacMan)
	}
	if match.Inky.houseStage != 4 || match.InkyDots != 7 {
		t.Fatal("revived Inky should wait with his personal count intact")
	}
	for _, coord := range CoinCoords[7:29] {
		state.ConsumeCollectible(coord[0], coord[1])
	}
	if match.InkyDots != 29 || match.Inky.houseStage != 4 {
		t.Fatal("revived Inky should resume counting from the prior value")
	}
	coord := CoinCoords[29]
	state.ConsumeCollectible(coord[0], coord[1])
	if match.InkyDots != 30 || match.Inky.houseStage != 3 {
		t.Fatal("the resumed counter should release Inky at 30")
	}
}

// TestInkyCaptureAndModes checks shared timers, third-score capture, and eye return.
func TestInkyCaptureAndModes(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	match := state.Match
	match.Inky.houseStage = 0
	match.Inky.X, match.Inky.Y = match.Blinky.X, match.Blinky.Y
	match.Inky.Direction, match.Inky.turn, match.Inky.nextTurn = DirectionLeft, DirectionLeft, DirectionLeft
	match.Blinky.X, match.Blinky.Y = float64(3*CellSize+CellSize/2), float64(4*CellSize+CellSize/2)
	coord := EnergizerCoords[0]
	state.ConsumeCollectible(coord[0], coord[1])
	if match.Inky.Mode != GhostFrightened || match.Inky.ghostClock != match.Blinky.ghostClock {
		t.Fatal("energizers should frighten Inky using the shared timer")
	}
	match.GhostStreak = 2
	match.PacMan.X, match.PacMan.Y = match.Inky.X, match.Inky.Y
	match.PacMan.EatingPauseTicks = 1
	state.UpdateActors(DirectionNone, false)
	if state.Phase != PhaseGhostCaptured || match.GhostCapture.Kind != GhostInky ||
		match.GhostCapture.Points != 800 || state.GhostVisible(GhostInky) ||
		!state.GhostVisible(GhostBlinky) || !state.GhostVisible(GhostPinky) {
		t.Fatal("Inky's capture should award 800 and hide only Inky")
	}
	fright, blinky, pinky := match.Inky.frightTicks, match.Blinky, match.Pinky
	for range GhostCaptureTicks {
		state.Tick()
	}
	if match.Inky.Mode != GhostEaten || match.Inky.frightTicks != fright || match.Blinky != blinky || match.Pinky != pinky {
		t.Fatal("capture should freeze other ghosts and timers before Inky becomes eyes")
	}
	for tick := 0; tick < 5000 && match.Inky.Mode == GhostEaten; tick++ {
		match.Inky.Update(match.Level, match.PacMan)
	}
	if match.Inky.Mode != GhostScatter || match.Inky.houseStage != 4 {
		t.Fatal("Inky's eyes should return and revive while his personal limit is unmet")
	}
}

// TestInkyDoorCollision checks fatal and frightened collisions while Inky exits.
func TestInkyDoorCollision(t *testing.T) {
	for _, frightened := range []bool{false, true} {
		state := GameState{}
		state.StartNewMatch()
		state.BeginPlaying()
		match := state.Match
		match.Blinky.X, match.Blinky.Y = float64(3*CellSize+CellSize/2), float64(4*CellSize+CellSize/2)
		match.Inky.houseStage = 3
		match.Inky.X, match.Inky.Y = float64(ScreenWidth)/2, float64(15*CellSize)-1
		match.PacMan.X, match.PacMan.Y = match.Inky.X, match.Inky.Y
		match.PacMan.EatingPauseTicks = 1
		if frightened {
			coord := EnergizerCoords[0]
			state.ConsumeCollectible(coord[0], coord[1])
		}
		state.UpdateActors(DirectionNone, false)
		if !frightened {
			if state.Phase != PhaseDying {
				t.Fatal("stationary Inky should be dangerous at the maze entrance")
			}
			continue
		}
		if state.Phase != PhaseGhostCaptured || match.GhostCapture.Kind != GhostInky {
			t.Fatal("frightened Inky should be capturable at the maze entrance")
		}
		for range GhostCaptureTicks {
			state.Tick()
		}
		if match.Inky.Mode != GhostEaten || match.Inky.houseStage != 2 {
			t.Fatal("Inky's eyes should return into the house after capture at the door")
		}
	}
}

// TestThreeGhostCaptureSequence checks a single energizer awards 200, 400, and 800.
func TestThreeGhostCaptureSequence(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	match := state.Match
	match.Pinky.houseStage, match.Inky.houseStage = 0, 0
	for index, ghost := range []*GhostState{&match.Blinky, &match.Pinky, &match.Inky} {
		ghost.X = float64((3+index)*CellSize + CellSize/2)
		ghost.Y = float64(4*CellSize + CellSize/2)
	}
	coord := EnergizerCoords[0]
	state.ConsumeCollectible(coord[0], coord[1])
	for index, ghost := range []*GhostState{&match.Blinky, &match.Pinky, &match.Inky} {
		match.PacMan.X, match.PacMan.Y = ghost.X, ghost.Y
		match.PacMan.EatingPauseTicks = 1
		state.UpdateActors(DirectionNone, false)
		want := 200 << index
		if state.Phase != PhaseGhostCaptured || match.GhostCapture.Kind != ghost.Kind || match.GhostCapture.Points != want {
			t.Fatalf("capture %d should award %d points for ghost %v", index+1, want, ghost.Kind)
		}
		for range GhostCaptureTicks {
			state.Tick()
		}
	}
	if match.Score != 1450 || match.GhostStreak != 3 {
		t.Fatal("three captures and an energizer should award 1450 points")
	}
}

// TestInkyNavigationAndFreeze checks long-running maze movement and disabled movement.
func TestInkyNavigationAndFreeze(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	match := state.Match
	match.Inky.startExit()
	match.PacMan.EatingPauseTicks = 5000
	x, y := match.Inky.X, match.Inky.Y
	for range 420 {
		state.Tick()
		state.UpdateActors(DirectionNone, false)
	}
	if match.Inky.X != x || match.Inky.Y != y || match.Inky.Mode != GhostChase || !match.Inky.exitRight {
		t.Fatal("disabled movement should preserve Inky's position but advance his mode")
	}
	exited := false
	for tick := 0; tick < 3000; tick++ {
		if tick%137 == 83 {
			match.Inky.Frighten(match.Level)
		}
		match.Inky.Tick(match.Level)
		match.Inky.Update(match.Level, match.PacMan)
		if !exited && match.Inky.houseStage == 0 {
			exited = true
			if match.Inky.Direction != DirectionRight {
				t.Fatal("a mode change inside the house should make Inky exit right")
			}
		}
		column, row := match.Inky.OccupiedTile()
		if match.Inky.houseStage == 0 && !isWalkable(column, row) {
			t.Fatalf("tick %d: Inky entered wall (%d,%d)", tick, column, row)
		}
	}
	if !exited {
		t.Fatal("Inky did not leave the house")
	}
}
