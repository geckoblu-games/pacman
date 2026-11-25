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

// TestAttractRevealTimeline checks the cast, legend, energizer and copyright reveal ticks.
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
	if before.LegendVisible() || before.EnergizerVisible() || before.DecoyVisible() ||
		before.CopyrightVisible() || before.ChaseActive() {
		t.Fatal("legend, energizers and actors should be hidden before the legend tick")
	}
	if !legend.LegendVisible() || !legend.EnergizerVisible() || legend.DecoyVisible() ||
		legend.CopyrightVisible() || legend.ChaseActive() {
		t.Fatal("legend tick should show the point legend and its energizer, but not the decoy or actors")
	}
	if runAttract(attractChaseTick - 1).CopyrightVisible() {
		t.Fatal("copyright should remain hidden until the chase starts")
	}
	chase := runAttract(attractChaseTick)
	if !chase.CopyrightVisible() || !chase.DecoyVisible() {
		t.Fatal("copyright and decoy energizer should appear when the chase starts")
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
	if attract.Finished() {
		t.Fatalf("presentation should not be finished before tick %d", AttractCycleTicks)
	}
	attract.Tick()
	finished := attract
	attract.Tick()
	if !finished.Finished() || finished.Ticks != AttractCycleTicks || attract != finished {
		t.Fatalf("presentation should finish and stop at tick %d", AttractCycleTicks)
	}
}

// TestAttractLoopScenes checks that the attract loop opens with the presentation,
// continues with the controls scene when it finishes, and returns to the presentation.
func TestAttractLoopScenes(t *testing.T) {
	state := GameState{}
	state.EnterAttract()
	if state.AttractScene != AttractScenePresentation || state.Attract != newAttractState() {
		t.Fatal("the attract loop should open with a fresh presentation")
	}
	for tick := 0; tick < AttractCycleTicks-1; tick++ {
		state.Tick()
		if state.AttractScene != AttractScenePresentation || state.Attract.Ticks != tick+1 {
			t.Fatalf("tick %d: presentation should run for %d ticks", tick, AttractCycleTicks)
		}
	}
	state.Tick()
	if state.Phase != PhaseAttract || state.AttractScene != AttractSceneControls || state.Controls != newControlsState(false) {
		t.Fatal("a finished presentation should start a fresh controls scene")
	}
	for tick := 0; tick < state.Controls.CycleTicks()-1; tick++ {
		state.Tick()
		if state.AttractScene != AttractSceneControls || state.Controls.Ticks != tick+1 {
			t.Fatalf("tick %d: controls scene should run for %d ticks", tick, state.Controls.CycleTicks())
		}
	}
	state.Tick()
	if state.Phase != PhaseAttract || state.AttractScene != AttractScenePresentation ||
		state.Attract != newAttractState() || state.Match != nil {
		t.Fatal("a finished controls scene should loop back to a fresh presentation")
	}
}

// TestTouchControlsLoop checks the selected mode survives scene resets and uses its shorter duration.
func TestTouchControlsLoop(t *testing.T) {
	state := GameState{}
	state.Controls.Touch = true
	state.EnterAttract()
	for tick := 0; tick < AttractCycleTicks; tick++ {
		state.Tick()
	}
	if state.AttractScene != AttractSceneControls || !state.Controls.Touch {
		t.Fatal("presentation should lead to touch controls")
	}
	for tick := 0; tick < state.Controls.CycleTicks(); tick++ {
		state.Tick()
	}
	if state.AttractScene != AttractScenePresentation || !state.Controls.Touch {
		t.Fatal("touch controls should return to presentation and preserve the mode")
	}
}

// TestEnterAttractAndGameOverReturn checks the match discard, last-score memory, and the
// timed return from GAME OVER to the attract screen.
func TestEnterAttractAndGameOverReturn(t *testing.T) {
	state := GameState{}
	state.EnterAttract()
	if state.Phase != PhaseAttract || state.Match != nil || state.LastScore != 0 ||
		state.AttractScene != AttractScenePresentation || state.Attract != newAttractState() {
		t.Fatal("entering attract without a match should show 00 and a fresh loop")
	}
	state.Tick()
	if state.Attract.Ticks != 1 || state.PhaseTicks != 1 {
		t.Fatal("ticking in attract should advance the current scene")
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
		state.AttractScene != AttractScenePresentation || state.Attract != newAttractState() {
		t.Fatal("game over should return to a fresh attract loop remembering the last score and record")
	}
	if state.TogglePause() || state.Paused {
		t.Fatal("attract screen should reject pause")
	}
	state.Tick()
	if state.Attract.Ticks != 1 {
		t.Fatal("attract screen should keep running after a rejected pause")
	}
}
