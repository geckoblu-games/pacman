package game

import (
	"reflect"
	"testing"
)

// TestBonusSymbols checks bonus history ordering, duplicates, limits, and resets.
func TestBonusSymbols(t *testing.T) {
	tests := []struct {
		// number is the one-based level used to build the history.
		number int
		// want lists the expected bonus names from current to oldest.
		want []string
	}{
		{1, []string{"Cherries"}},
		{2, []string{"Strawberry", "Cherries"}},
		{4, []string{"Peach", "Peach", "Strawberry", "Cherries"}},
		{7, []string{"Grapes", "Apple", "Apple", "Peach", "Peach", "Strawberry", "Cherries"}},
		{8, []string{"Grapes", "Grapes", "Apple", "Apple", "Peach", "Peach", "Strawberry"}},
		{13, []string{"Key", "Bell", "Bell", "Galaxian", "Galaxian", "Grapes", "Grapes"}},
		{19, []string{"Key", "Key", "Key", "Key", "Key", "Key", "Key"}},
		{22, []string{"Key", "Key", "Key", "Key", "Key", "Key", "Key"}},
		{256, []string{"Key", "Key", "Key", "Key", "Key", "Key", "Key"}},
	}
	for _, test := range tests {
		if got := NewLevelState(test.number).BonusSymbols(); !reflect.DeepEqual(got, test.want) {
			t.Errorf("level %d bonus symbols: got %v, want %v", test.number, got, test.want)
		}
	}
	if got := (LevelState{}).BonusSymbols(); len(got) != 0 {
		t.Fatal("uninitialized level should have no bonus symbols")
	}
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	before := state.Match.Level.BonusSymbols()
	state.LoseLife()
	if !reflect.DeepEqual(state.Match.Level.BonusSymbols(), before) {
		t.Fatal("losing a life should preserve bonus symbols")
	}
	state.Match.Level = NewLevelState(8)
	state.StartNewMatch()
	if got := state.Match.Level.BonusSymbols(); !reflect.DeepEqual(got, []string{"Cherries"}) {
		t.Fatal("new match should reset the bonus strip")
	}
}

// TestLevelParametersFollowMatchProgression checks parameter loading and the 21+ fallback.
func TestLevelParametersFollowMatchProgression(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	for number := 1; number <= 23; number++ {
		if state.Match.Level.Number != number || state.Match.Level.Parameters != ParametersForLevel(number) {
			t.Fatalf("wrong parameters for level %d", number)
		}
		before := state.Match.Level.BonusSymbols()
		state.BeginPlaying()
		for _, coord := range CoinCoords {
			state.ConsumeCollectible(coord[0], coord[1])
		}
		for _, coord := range EnergizerCoords {
			state.ConsumeCollectible(coord[0], coord[1])
		}
		if !state.AdvanceLevel() {
			t.Fatalf("failed to advance level %d", number)
		}
		after := state.Match.Level.BonusSymbols()
		if after[0] != ParametersForLevel(number+1).BonusSymbol || len(after) != min(number+1, 7) ||
			!reflect.DeepEqual(after[1:], before[:min(len(before), 6)]) {
			t.Fatalf("bonus strip did not shift after level %d: before=%v after=%v", number, before, after)
		}
	}
	for _, number := range []int{21, 22, 256, 1000} {
		currentLevel := NewLevelState(number)
		if currentLevel.Number != number || currentLevel.Parameters != ParametersForLevel(21) {
			t.Fatalf("level %d should keep its number and use 21+ parameters", number)
		}
	}
	state.Match.Level.Parameters.BonusPoints = 0
	state.StartNewMatch()
	if state.Match.Level.Parameters != ParametersForLevel(1) || ParametersForLevel(21).BonusPoints != 5000 {
		t.Fatal("match parameters should not mutate the shared table")
	}
}

// TestStartNewMatch checks initial state, match resets, and session-record retention.
func TestStartNewMatch(t *testing.T) {
	state := GameState{}
	if state.Phase != PhaseWaiting || state.Match != nil {
		t.Fatal("zero state should have no match")
	}
	state.StartNewMatch()
	if state.Phase != PhaseReady || state.Match.Score != 0 || state.Match.Lives != 3 || state.Match.Level.Number != 1 {
		t.Fatalf("unexpected new match: %+v", state)
	}
	if state.Match.Level.Remaining != 244 {
		t.Fatalf("expected 244 collectibles, got %d", state.Match.Level.Remaining)
	}
	state.BeginPlaying()
	state.AddScore(100)
	state.StartNewMatch()
	if state.HighScore != 100 || state.Match.Score != 0 || state.Match.Level.Number != 1 || state.Match.Lives != 3 {
		t.Fatal("new match should reset match data but retain high score")
	}
	state.BeginPlaying()
	state.AddScore(50)
	if state.HighScore != 100 {
		t.Fatal("lower score should not reduce high score")
	}
}

// TestLoseLifePreservesLevel checks collectible preservation, respawn, and game over.
func TestLoseLifePreservesLevel(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	coord := CoinCoords[0]
	if !state.ConsumeCollectible(coord[0], coord[1]) {
		t.Fatal("expected to consume a coin")
	}
	level := state.Match.Level
	for lives := 2; lives >= 0; lives-- {
		if !state.LoseLife() || state.Match.Lives != lives {
			t.Fatalf("expected %d lives", lives)
		}
		if state.Match.Level != level || state.Match.Score != 10 {
			t.Fatal("losing a life changed level or score")
		}
		if lives > 0 {
			if state.Phase != PhaseReady || !state.BeginPlaying() {
				t.Fatal("expected ready phase before respawn")
			}
		}
	}
	if state.Phase != PhaseGameOver || state.LoseLife() || state.Match.Lives != 0 {
		t.Fatal("game over should prevent further life loss")
	}
	state.StartNewMatch()
	if !state.Match.Level.Coins[0] || state.HighScore != 10 {
		t.Fatal("new match should restore coins and retain record")
	}
}

// TestCompleteAndAdvanceLevel checks one-time consumption and preservation on advancement.
func TestCompleteAndAdvanceLevel(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	state.LoseLife()
	state.BeginPlaying()
	if state.AdvanceLevel() {
		t.Fatal("cannot advance an incomplete level")
	}
	for _, coord := range CoinCoords {
		if !state.ConsumeCollectible(coord[0], coord[1]) || state.ConsumeCollectible(coord[0], coord[1]) {
			t.Fatal("each coin should be consumed exactly once")
		}
	}
	if state.Phase != PhasePlaying || state.Match.Level.Remaining != 4 {
		t.Fatal("energizers are required to finish the level")
	}
	for _, coord := range EnergizerCoords {
		if !state.ConsumeCollectible(coord[0], coord[1]) || state.ConsumeCollectible(coord[0], coord[1]) {
			t.Fatal("each energizer should be consumed exactly once")
		}
	}
	if state.Phase != PhaseLevelComplete || state.Match.Level.Remaining != 0 || state.Match.Score != 2600 || state.HighScore != 2600 {
		t.Fatal("unexpected completed level state")
	}
	if !state.AdvanceLevel() || state.AdvanceLevel() {
		t.Fatal("level should advance exactly once")
	}
	if state.Phase != PhaseReady || state.Match.Level.Number != 2 || state.Match.Level.Remaining != 244 || state.Match.Score != 2600 || state.Match.Lives != 2 {
		t.Fatal("advancing should restore collectibles and preserve score and lives")
	}
}

// TestInactiveStateRejectsGameplay checks missing matches, inactive phases, and invalid input.
func TestInactiveStateRejectsGameplay(t *testing.T) {
	state := GameState{}
	coord := CoinCoords[0]
	for _, phase := range []Phase{PhaseWaiting, PhaseReady, PhaseLevelComplete, PhaseGameOver} {
		state.Phase = phase
		state.AddScore(100)
		if state.BeginPlaying() || state.LoseLife() || state.AdvanceLevel() || state.ConsumeCollectible(coord[0], coord[1]) || state.HighScore != 0 {
			t.Fatal("missing match should reject gameplay")
		}
	}
	state.StartNewMatch()
	state.AddScore(100)
	if state.LoseLife() || state.ConsumeCollectible(coord[0], coord[1]) || state.Match.Score != 0 {
		t.Fatal("ready phase should reject gameplay")
	}
	state.BeginPlaying()
	state.AddScore(0)
	state.AddScore(-10)
	if state.Match.Score != 0 || state.ConsumeCollectible(-1, -1) || state.Match.Level.Remaining != 244 {
		t.Fatal("invalid points and positions should have no effect")
	}
}

// clearLevel consumes every collectible of the current level during active play.
func clearLevel(t *testing.T, state *GameState) {
	t.Helper()
	for _, coord := range CoinCoords {
		state.ConsumeCollectible(coord[0], coord[1])
	}
	for _, coord := range EnergizerCoords {
		state.ConsumeCollectible(coord[0], coord[1])
	}
	if state.Phase != PhaseLevelComplete || state.PhaseTicks != 0 {
		t.Fatalf("clearing the level should start the flash at tick 0: phase=%v ticks=%d", state.Phase, state.PhaseTicks)
	}
}

// TestLevelClearFlashTiming checks the 120-tick flash, its four white periods, and Pac-Man freezing.
func TestLevelClearFlashTiming(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	state.Match.PacMan.X = 100
	state.Match.PacMan.Y = 200
	clearLevel(t, &state)
	frozen := state.Match.PacMan
	if frozen.X != 100 || frozen.Y != 200 {
		t.Fatal("clearing the level should not move Pac-Man")
	}
	segmentTicks := LevelClearTicks / (2 * LevelClearFlashes)
	if segmentTicks != 15 {
		t.Fatalf("expected 15-tick flash segments, got %d", segmentTicks)
	}
	whiteStarts := 0
	previousWhite := false
	for tick := 0; tick < LevelClearTicks; tick++ {
		if state.Phase != PhaseLevelComplete || state.PhaseTicks != tick {
			t.Fatalf("tick %d: expected level-complete phase at tick %d, got phase %v tick %d", tick, tick, state.Phase, state.PhaseTicks)
		}
		if state.Match.PacMan != frozen {
			t.Fatalf("tick %d: Pac-Man should stay where the flash began", tick)
		}
		white := state.MazeFlashWhite()
		if want := (tick/segmentTicks)%2 == 0; white != want {
			t.Fatalf("tick %d: white=%v, want %v", tick, white, want)
		}
		if white && !previousWhite {
			whiteStarts++
		}
		previousWhite = white
		state.Tick()
	}
	if whiteStarts != LevelClearFlashes || previousWhite {
		t.Fatalf("expected %d white flashes ending blue, got %d (last white=%v)", LevelClearFlashes, whiteStarts, previousWhite)
	}
	if state.Phase != PhaseReady || state.PhaseTicks != 0 || state.Match.Level.Number != 2 || state.MazeFlashWhite() {
		t.Fatalf("flash should end in the ready intro of level 2: %+v", state)
	}
	if state.Match.PacMan != NewPacManState() {
		t.Fatal("Pac-Man should be at his starting position for the intro")
	}
}

// TestReadyIntroTiming checks that every ready entry waits ReadyTicks before play begins.
func TestReadyIntroTiming(t *testing.T) {
	expectIntro := func(state *GameState, context string) {
		for tick := 0; tick < ReadyTicks; tick++ {
			if state.Phase != PhaseReady || state.PhaseTicks != tick {
				t.Fatalf("%s: tick %d should still be in the ready intro, got phase %v tick %d", context, tick, state.Phase, state.PhaseTicks)
			}
			if state.ConsumeCollectible(CoinCoords[0][0], CoinCoords[0][1]) {
				t.Fatalf("%s: controls should be locked during the intro", context)
			}
			state.Tick()
		}
		if state.Phase != PhasePlaying || state.PhaseTicks != 0 {
			t.Fatalf("%s: intro should end in play: phase %v tick %d", context, state.Phase, state.PhaseTicks)
		}
	}
	state := GameState{}
	state.Tick()
	if state.Phase != PhaseWaiting || state.PhaseTicks != 0 {
		t.Fatal("ticking without a match should do nothing")
	}
	state.StartNewMatch()
	expectIntro(&state, "new match")
	state.LoseLife()
	expectIntro(&state, "life lost")
	clearLevel(t, &state)
	for tick := 0; tick < LevelClearTicks; tick++ {
		state.Tick()
	}
	expectIntro(&state, "next level")
	for tick := 0; tick < 10; tick++ {
		state.Tick()
	}
	if state.Phase != PhasePlaying || state.PhaseTicks != 0 {
		t.Fatal("ticking during play should not change phase or counter")
	}
	state.LoseLife()
	state.BeginPlaying()
	state.LoseLife()
	if state.Phase != PhaseGameOver {
		t.Fatal("expected game over")
	}
	state.Tick()
	if state.Phase != PhaseGameOver || state.MazeFlashWhite() {
		t.Fatal("ticking after game over should do nothing")
	}
}
