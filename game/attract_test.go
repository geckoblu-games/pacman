package game

import "testing"

// runAttract advances a fresh presentation to the requested tick.
func runAttract(ticks int) AttractState {
	attract := newAttractState()
	for attract.Ticks < ticks {
		attract.Tick()
	}
	return attract
}

// TestAttractCastSpriteCenter checks cast icons use the reference image's text-row centers.
func TestAttractCastSpriteCenter(t *testing.T) {
	for _, row := range attractCastRows {
		x, y := attractCastSpriteCenter(row)
		if x != 80 || y != float64(row*CellSize+6) {
			t.Fatalf("cast sprite at row %d should be centered at (80, %d), got (%v, %v)",
				row, row*CellSize+6, x, y)
		}
	}
}

// TestAttractRevealTimeline checks the cast, legend and energizer reveal ticks.
func TestAttractRevealTimeline(t *testing.T) {
	for index := range attractGhostRevealTicks {
		before := runAttract(attractGhostRevealTicks[index] - 1)
		at := runAttract(attractGhostRevealTicks[index])
		if before.GhostRevealed(index) || !at.GhostRevealed(index) {
			t.Fatalf("ghost %d sprite should appear exactly at tick %d", index, attractGhostRevealTicks[index])
		}
		named := runAttract(attractGhostNameTicks[index])
		if at.GhostNamed(index) || !named.GhostNamed(index) {
			t.Fatalf("ghost %d name should appear exactly at tick %d", index, attractGhostNameTicks[index])
		}
	}
	before := runAttract(attractLegendTick - 1)
	legend := runAttract(attractLegendTick)
	if before.LegendVisible() || before.EnergizerVisible() || before.DecoyVisible() || before.ChaseActive() {
		t.Fatal("legend, energizers and actors should be hidden before the legend tick")
	}
	if !legend.LegendVisible() || !legend.EnergizerVisible() || !legend.DecoyVisible() || legend.ChaseActive() {
		t.Fatal("legend tick should show the legend and both steady energizers without actors")
	}
	for tick := attractLegendTick; tick < attractChaseTick; tick++ {
		if !runAttract(tick).EnergizerVisible() {
			t.Fatalf("tick %d: energizers should be steady before the chase", tick)
		}
	}
	for offset := 0; offset < 4*attractFlashHalfPeriod; offset++ {
		attract := runAttract(attractChaseTick + offset)
		want := (offset/attractFlashHalfPeriod)%2 == 0
		if attract.EnergizerVisible() != want || attract.DecoyVisible() != want || !attract.ChaseActive() {
			t.Fatalf("tick %d: energizer flash should be %v with actors on stage", attract.Ticks, want)
		}
	}
}

// TestAttractChaseAndReversal checks the entrance positions, the leftward chase, and the
// reversal when Pac-Man reaches the decoy energizer.
func TestAttractChaseAndReversal(t *testing.T) {
	start := runAttract(attractChaseTick - 1)
	if start.PacMan.X <= ScreenWidth || start.Ghosts[0].X <= start.PacMan.X || start.PacMan.Direction != DirectionLeft {
		t.Fatal("actors should wait off-screen right, Pac-Man leading, before the chase")
	}
	if start.PacMan.Y != attractChaseY() || start.PacMan.Y != float64(attractChaseRow*CellSize+CellSize/2) {
		t.Fatalf("chase should run along row %d, got y %v", attractChaseRow, start.PacMan.Y)
	}
	for index := 1; index < len(start.Ghosts); index++ {
		if start.Ghosts[index].Kind != GhostKind(index) ||
			start.Ghosts[index].X-start.Ghosts[index-1].X != attractGhostSpacing*attractSourceScale {
			t.Fatalf("ghost %d should follow at the source spacing in cast order", index)
		}
	}
	attract := start
	attract.Tick()
	if attract.PacMan.X >= start.PacMan.X || attract.Ghosts[0].X >= start.Ghosts[0].X ||
		attract.GhostDirection() != DirectionLeft || attract.GhostMode() != GhostScatter {
		t.Fatal("chase should move all actors left in normal mode")
	}
	for !attract.DecoyEaten {
		attract.Tick()
		if attract.Ticks > attractChaseTick+200 {
			t.Fatal("Pac-Man should reach the decoy energizer")
		}
	}
	if attract.Ticks < 655 || attract.Ticks > 680 {
		t.Fatalf("decoy should be eaten around tick 666, got %d", attract.Ticks)
	}
	if attract.PacMan.X > attractDecoyX() || attract.PacMan.Direction != DirectionRight ||
		attract.GhostDirection() != DirectionRight || attract.GhostMode() != GhostFrightened || attract.DecoyVisible() {
		t.Fatal("eating the decoy should reverse everyone, frighten the ghosts, and remove the decoy")
	}
	reversed := attract
	attract.Tick()
	if attract.PacMan.X <= reversed.PacMan.X || attract.Ghosts[0].X <= reversed.Ghosts[0].X ||
		attract.Ghosts[0].X-reversed.Ghosts[0].X >= attract.PacMan.X-reversed.PacMan.X {
		t.Fatal("after the reversal both move right with Pac-Man faster than the ghosts")
	}
}

// TestAttractCaptureSequence checks ordered captures, doubling scores, the frozen pause,
// hidden actors, and the automatic restart of the cycle.
func TestAttractCaptureSequence(t *testing.T) {
	attract := newAttractState()
	points := []int{}
	for attract.Ticks < AttractCycleTicks-1 {
		previous := attract
		attract.Tick()
		if attract.Captured > previous.Captured {
			if attract.PauseTicks != attractCapturePauseTicks || attract.PacManVisible() || attract.GhostVisible(attract.Captured-1) {
				t.Fatalf("capture %d should start a pause hiding Pac-Man and the ghost", attract.Captured)
			}
			if attract.Ghosts[attract.Captured-1].X != attract.CaptureX || attract.PacMan.X < attract.CaptureX {
				t.Fatalf("capture %d should show the score at the ghost's position", attract.Captured)
			}
			points = append(points, attract.CapturePoints())
			frozen := attract
			for tick := 0; tick < attractCapturePauseTicks; tick++ {
				attract.Tick()
				if attract.PacMan.X != frozen.PacMan.X || attract.Ghosts != frozen.Ghosts {
					t.Fatalf("capture %d: actors should not move during the pause", frozen.Captured)
				}
				if tick < attractCapturePauseTicks-1 && (attract.CapturePoints() != points[len(points)-1] || attract.PacManVisible()) {
					t.Fatalf("capture %d: score should stay visible during the pause", frozen.Captured)
				}
			}
			if attract.PauseTicks != 0 || attract.CapturePoints() != 0 || !attract.PacManVisible() {
				t.Fatalf("capture %d: pause should end after %d ticks", frozen.Captured, attractCapturePauseTicks)
			}
		}
	}
	if len(points) != 4 || points[0] != 200 || points[1] != 400 || points[2] != 800 || points[3] != 1600 {
		t.Fatalf("expected four doubling captures, got %v", points)
	}
	for index := range attract.Ghosts {
		if attract.GhostVisible(index) {
			t.Fatalf("ghost %d should remain hidden after its capture", index)
		}
	}
	if attract.PacMan.X <= ScreenWidth {
		t.Fatal("Pac-Man should have left the screen before the cycle ends")
	}
	attract.Tick()
	if attract != newAttractState() {
		t.Fatalf("cycle should restart from the initial presentation at tick %d", AttractCycleTicks)
	}
}

// TestEnterAttractAndGameOverReturn checks the match discard, last-score memory, and the
// timed return from GAME OVER to the attract screen.
func TestEnterAttractAndGameOverReturn(t *testing.T) {
	state := GameState{}
	state.EnterAttract()
	if state.Phase != PhaseAttract || state.Match != nil || state.LastScore != 0 || state.Attract != newAttractState() {
		t.Fatal("entering attract without a match should show 00 and a fresh presentation")
	}
	state.Tick()
	if state.Attract.Ticks != 1 || state.PhaseTicks != 1 {
		t.Fatal("ticking in attract should advance the presentation")
	}
	state.StartNewMatch()
	state.BeginPlaying()
	state.AddScore(370)
	state.Match.Lives = 1
	state.LoseLife()
	finishDeath(t, &state)
	for tick := 0; tick < GameOverTicks-1; tick++ {
		state.Tick()
		if state.Phase != PhaseGameOver || state.Match == nil {
			t.Fatalf("tick %d: game over should remain visible for %d ticks", tick, GameOverTicks)
		}
	}
	state.Tick()
	if state.Phase != PhaseAttract || state.Match != nil || state.LastScore != 370 || state.HighScore != 370 ||
		state.Attract != newAttractState() {
		t.Fatal("game over should return to a fresh attract screen remembering the last score and record")
	}
	if state.TogglePause() || state.Paused {
		t.Fatal("attract screen should reject pause")
	}
	state.Tick()
	if state.Attract.Ticks != 1 {
		t.Fatal("attract screen should keep running after a rejected pause")
	}
}
