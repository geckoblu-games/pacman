package game

import (
	"math/rand"
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

// finishDeath advances an unpaused death sequence to READY or game over.
func finishDeath(t *testing.T, state *GameState) {
	t.Helper()
	if state.Phase != PhaseDying {
		t.Fatal("expected death sequence")
	}
	for range DeathFreezeTicks + DeathAnimationTicks + DeathEmptyTicks {
		if state.Phase != PhaseDying {
			return
		}
		state.Tick()
	}
	if state.Phase == PhaseDying {
		t.Fatal("death sequence did not finish")
	}
}

// TestDeathSequence checks frozen collision, animation, empty maze, pause, and respawn.
func TestDeathSequence(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	state.Match.Fruit = FruitState{Visible: true, Ticks: 300}
	state.Match.PacMan.X, state.Match.PacMan.Y = state.Match.Blinky.X, state.Match.Blinky.Y
	state.Match.PacMan.EatingPauseTicks = 1
	state.UpdateActors(DirectionNone, false)
	if state.Phase != PhaseDying || state.PhaseTicks != 0 || state.Match.Lives != InitialLives {
		t.Fatal("collision should start death without deducting a life")
	}
	player, ghost, fruit := state.Match.PacMan, state.Match.Blinky, state.Match.Fruit
	state.TogglePause()
	for range 10 {
		state.Tick()
		state.UpdateActors(DirectionRight, true)
	}
	if state.PhaseTicks != 0 || state.Match.PacMan != player || state.Match.Blinky != ghost || state.Match.Fruit != fruit {
		t.Fatal("pause should freeze the collision scene and death timer")
	}
	state.TogglePause()
	for tick := 0; tick < DeathFreezeTicks; tick++ {
		if !state.PacManVisible() || !state.GhostsVisible() || state.DeathFrame() != -1 ||
			state.Match.PacMan != player || state.Match.Blinky != ghost || state.Match.Fruit != fruit {
			t.Fatalf("collision scene changed during freeze at tick %d", tick)
		}
		state.UpdateActors(DirectionRight, true)
		state.Tick()
	}
	if !state.PacManVisible() || state.GhostsVisible() || state.Match.Fruit != (FruitState{}) || state.DeathFrame() != 0 {
		t.Fatal("animation should begin with ghost and fruit hidden")
	}
	for tick := 0; tick < DeathAnimationTicks; tick++ {
		if state.DeathFrame() != tick/DeathFrameTicks || !state.PacManVisible() || state.GhostsVisible() {
			t.Fatalf("unexpected death animation at tick %d", tick)
		}
		state.Tick()
	}
	if state.PacManVisible() || state.GhostsVisible() || state.Match.Lives != InitialLives {
		t.Fatal("maze should be empty before a life is deducted")
	}
	for range DeathEmptyTicks - 1 {
		state.Tick()
	}
	if state.Phase != PhaseDying || state.Match.Lives != InitialLives {
		t.Fatal("empty maze ended early")
	}
	state.Tick()
	if state.Phase != PhaseReady || state.Match.Lives != InitialLives-1 ||
		state.Match.PacMan != NewPacManState() || !state.Match.Blinky.ElroySuspended {
		t.Fatal("death should deduct one life and reset actors into READY")
	}
}

// TestDeathGameOver checks the final life is deducted only after the empty maze.
func TestDeathGameOver(t *testing.T) {
	state := GameState{}
	state.StartNewMatch()
	state.BeginPlaying()
	state.Match.Lives = 1
	state.LoseLife()
	for range DeathFreezeTicks + DeathAnimationTicks + DeathEmptyTicks - 1 {
		state.Tick()
	}
	if state.Phase != PhaseDying || state.Match.Lives != 1 {
		t.Fatal("final life should remain until the death sequence ends")
	}
	state.Tick()
	if state.Phase != PhaseGameOver || state.Match.Lives != 0 || state.PacManVisible() || state.GhostsVisible() {
		t.Fatal("final death should enter game over with no actors")
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
		if !state.LoseLife() {
			t.Fatal("expected death to start")
		}
		finishDeath(t, &state)
		if state.Match.Lives != lives {
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
	finishDeath(t, &state)
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

// eatCoins consumes the next count remaining coins of the current level in CoinCoords order.
func eatCoins(t *testing.T, state *GameState, count int) {
	t.Helper()
	for index := range CoinCoords {
		if count == 0 {
			return
		}
		if !state.Match.Level.Coins[index] {
			continue
		}
		if !state.ConsumeCollectible(CoinCoords[index][0], CoinCoords[index][1]) {
			t.Fatalf("failed to consume coin %d", index)
		}
		count--
	}
	if count > 0 {
		t.Fatalf("not enough coins left to eat %d more", count)
	}
}

// TestFruitSpawnsAtDotThresholds checks the 70/170 triggers, the 9–10 second range,
// seeded determinism, and expiry.
func TestFruitSpawnsAtDotThresholds(t *testing.T) {
	state := GameState{Random: rand.New(rand.NewSource(7))}
	state.StartNewMatch()
	state.BeginPlaying()
	eatCoins(t, &state, FruitFirstDotsEaten-1)
	if state.Match.Fruit.Visible {
		t.Fatal("bonus should not appear before 70 collectibles")
	}
	eatCoins(t, &state, 1)
	fruit := state.Match.Fruit
	if !fruit.Visible || fruit.Ticks < FruitMinTicks || fruit.Ticks > FruitMaxTicks || fruit.ScoreTicks != 0 {
		t.Fatalf("unexpected bonus after 70 collectibles: %+v", fruit)
	}
	other := GameState{Random: rand.New(rand.NewSource(7))}
	other.StartNewMatch()
	other.BeginPlaying()
	eatCoins(t, &other, FruitFirstDotsEaten)
	if other.Match.Fruit != fruit {
		t.Fatal("identical seeds should produce identical bonus durations")
	}
	for tick := 0; tick < fruit.Ticks-1; tick++ {
		state.Tick()
	}
	if !state.Match.Fruit.Visible || state.Match.Fruit.Ticks != 1 {
		t.Fatalf("bonus should remain until its last tick: %+v", state.Match.Fruit)
	}
	state.Tick()
	if state.Match.Fruit != (FruitState{}) {
		t.Fatalf("bonus should expire after its duration: %+v", state.Match.Fruit)
	}
	eatCoins(t, &state, FruitSecondDotsEaten-FruitFirstDotsEaten-1)
	if state.Match.Fruit.Visible {
		t.Fatal("bonus should not reappear before 170 collectibles")
	}
	eatCoins(t, &state, 1)
	if !state.Match.Fruit.Visible {
		t.Fatal("bonus should appear after 170 collectibles")
	}
	seen := map[int]bool{}
	for seed := int64(0); seed < 64; seed++ {
		sample := GameState{Random: rand.New(rand.NewSource(seed))}
		sample.StartNewMatch()
		sample.BeginPlaying()
		eatCoins(t, &sample, FruitFirstDotsEaten)
		if ticks := sample.Match.Fruit.Ticks; ticks < FruitMinTicks || ticks > FruitMaxTicks {
			t.Fatalf("seed %d: duration %d outside 540–600 ticks", seed, ticks)
		}
		seen[sample.Match.Fruit.Ticks] = true
	}
	if len(seen) < 2 {
		t.Fatal("bonus duration should be randomized")
	}
	lazy := GameState{}
	lazy.StartNewMatch()
	lazy.BeginPlaying()
	eatCoins(t, &lazy, FruitFirstDotsEaten)
	if lazy.Random == nil || !lazy.Match.Fruit.Visible {
		t.Fatal("a nil generator should be created on first spawn")
	}
}

// TestConsumeFruit checks the half-cell overlap rule, scoring, the popup, and the pause.
func TestConsumeFruit(t *testing.T) {
	state := GameState{Random: rand.New(rand.NewSource(1))}
	state.StartNewMatch()
	state.BeginPlaying()
	eatCoins(t, &state, FruitFirstDotsEaten)
	score := state.Match.Score
	player := &state.Match.PacMan
	player.EatingPauseTicks = 0
	player.X, player.Y = FruitX+float64(CellSize)/2+0.01, FruitY
	if state.ConsumeFruit() || !state.Match.Fruit.Visible {
		t.Fatal("bonus should not be eaten beyond half a cell")
	}
	player.X, player.Y = FruitX-float64(CellSize)/2, FruitY
	if !state.ConsumeFruit() {
		t.Fatal("bonus should be eaten within half a cell")
	}
	want := FruitState{ScoreTicks: FruitScoreTicks, Points: 100}
	if state.Match.Fruit != want || state.Match.Score != score+100 || state.HighScore != score+100 {
		t.Fatalf("unexpected state after eating the bonus: fruit=%+v score=%d", state.Match.Fruit, state.Match.Score)
	}
	if player.EatingPauseTicks != pacManFruitPauseTicks {
		t.Fatalf("expected a %d-tick pause, got %d", pacManFruitPauseTicks, player.EatingPauseTicks)
	}
	if state.ConsumeFruit() || state.Match.Score != score+100 {
		t.Fatal("an eaten bonus should not score again")
	}
	for tick := 0; tick < FruitScoreTicks; tick++ {
		if state.Match.Fruit.ScoreTicks == 0 {
			t.Fatalf("popup ended early at tick %d", tick)
		}
		state.Tick()
	}
	if state.Match.Fruit != (FruitState{}) {
		t.Fatalf("popup should end after %d ticks: %+v", FruitScoreTicks, state.Match.Fruit)
	}
	state.Match.Level = NewLevelState(13)
	state.Match.Fruit = FruitState{Visible: true, Ticks: FruitMinTicks}
	player.X, player.Y = FruitX, FruitY+float64(CellSize)/2
	if !state.ConsumeFruit() || state.Match.Fruit.Points != 5000 || state.Match.Score != score+5100 {
		t.Fatalf("bonus value should follow the level table: %+v", state.Match.Fruit)
	}
	state.Match.Fruit = FruitState{Visible: true, Ticks: FruitMinTicks}
	state.Phase = PhaseReady
	if state.ConsumeFruit() || !state.Match.Fruit.Visible {
		t.Fatal("bonus should only be eaten during play")
	}
}

// TestFruitClearedOnLifeLossAndLevelChange checks removal on death, level clear, and advancement.
func TestFruitClearedOnLifeLossAndLevelChange(t *testing.T) {
	state := GameState{Random: rand.New(rand.NewSource(3))}
	state.StartNewMatch()
	state.BeginPlaying()
	eatCoins(t, &state, FruitFirstDotsEaten)
	state.LoseLife()
	finishDeath(t, &state)
	if state.Match.Fruit != (FruitState{}) {
		t.Fatal("losing a life should remove the bonus")
	}
	state.BeginPlaying()
	for tick := 0; tick < FruitMaxTicks; tick++ {
		state.Tick()
	}
	if state.Match.Fruit.Visible {
		t.Fatal("a removed bonus should not reappear")
	}
	eatCoins(t, &state, FruitSecondDotsEaten-FruitFirstDotsEaten)
	if !state.Match.Fruit.Visible {
		t.Fatal("the second trigger should still spawn a bonus")
	}
	clearLevel(t, &state)
	if state.Match.Fruit != (FruitState{}) {
		t.Fatal("completing the level should remove the bonus")
	}
	state.Match.Fruit = FruitState{ScoreTicks: 5, Points: 100}
	state.AdvanceLevel()
	if state.Match.Fruit != (FruitState{}) {
		t.Fatal("advancing should reset the bonus state")
	}
}

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

// TestReadyIntroTiming checks that the opening ready intro waits MatchStartTicks and
// every later ready entry waits ReadyTicks before play begins.
func TestReadyIntroTiming(t *testing.T) {
	expectIntro := func(state *GameState, context string, duration int) {
		if state.ReadyDuration() != duration {
			t.Fatalf("%s: expected intro duration %d, got %d", context, duration, state.ReadyDuration())
		}
		for tick := 0; tick < duration; tick++ {
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
	expectIntro(&state, "new match", MatchStartTicks)
	state.LoseLife()
	finishDeath(t, &state)
	expectIntro(&state, "life lost", ReadyTicks)
	clearLevel(t, &state)
	for tick := 0; tick < LevelClearTicks; tick++ {
		state.Tick()
	}
	expectIntro(&state, "next level", ReadyTicks)
	for tick := 0; tick < 10; tick++ {
		state.Tick()
	}
	if state.Phase != PhasePlaying || state.PhaseTicks != 0 {
		t.Fatal("ticking during play should not change phase or counter")
	}
	state.LoseLife()
	finishDeath(t, &state)
	state.BeginPlaying()
	state.LoseLife()
	finishDeath(t, &state)
	if state.Phase != PhaseGameOver {
		t.Fatal("expected game over")
	}
	state.Tick()
	if state.Phase != PhaseGameOver || state.MazeFlashWhite() {
		t.Fatal("ticking after game over should only count toward the attract return")
	}
}

// TestTogglePause checks that pausing freezes Tick in timed phases and play, that
// unpausing resumes exactly where it stopped, and that it is rejected without a match,
// after game over, and cleared by a new match.
func TestTogglePause(t *testing.T) {
	state := GameState{}
	if state.TogglePause() || state.Paused {
		t.Fatal("pause should be rejected without a match")
	}
	state.StartNewMatch()
	state.Tick()
	if !state.TogglePause() || !state.Paused {
		t.Fatal("expected the match to pause during the intro")
	}
	for range ReadyTicks {
		state.Tick()
	}
	if state.Phase != PhaseReady || state.PhaseTicks != 1 {
		t.Fatalf("paused intro should not advance: phase %v tick %d", state.Phase, state.PhaseTicks)
	}
	if state.TogglePause() || state.Paused {
		t.Fatal("second toggle should resume")
	}
	state.Tick()
	if state.PhaseTicks != 2 {
		t.Fatalf("resumed intro should continue from tick 1, got %d", state.PhaseTicks)
	}
	state.BeginPlaying()
	state.Match.Fruit.Visible = true
	state.Match.Fruit.Ticks = 10
	state.TogglePause()
	state.Tick()
	if state.Match.Fruit.Ticks != 10 {
		t.Fatal("paused play should not advance the bonus timer")
	}
	state.TogglePause()
	state.Tick()
	if state.Match.Fruit.Ticks != 9 {
		t.Fatal("resumed play should advance the bonus timer")
	}
	state.TogglePause()
	state.StartNewMatch()
	if state.Paused {
		t.Fatal("a new match should start unpaused")
	}
	state.BeginPlaying()
	for state.Phase != PhaseGameOver {
		state.LoseLife()
		finishDeath(t, &state)
		state.BeginPlaying()
	}
	if state.TogglePause() || state.Paused {
		t.Fatal("pause should be rejected after game over")
	}
}

// TestSoundCues checks that the simulation records the start jingle, waka, ghost, fruit
// and death cues at the documented moments, that TakeSoundCues drains them, and that
// rejected operations and the attract screen add nothing.
func TestSoundCues(t *testing.T) {
	state := GameState{}
	if cues := state.TakeSoundCues(); len(cues) != 0 {
		t.Fatalf("fresh state should have no cues, got %v", cues)
	}
	state.StartNewMatch()
	if cues := state.TakeSoundCues(); !reflect.DeepEqual(cues, []SoundCue{SoundCueStart}) {
		t.Fatalf("new match should request the start jingle, got %v", cues)
	}
	if state.ConsumeCollectible(CoinCoords[0][0], CoinCoords[0][1]) || len(state.SoundCues) != 0 {
		t.Fatal("locked controls should not record a waka")
	}
	state.BeginPlaying()
	state.ConsumeCollectible(CoinCoords[0][0], CoinCoords[0][1])
	state.ConsumeCollectible(CoinCoords[0][0], CoinCoords[0][1])
	state.ConsumeCollectible(EnergizerCoords[0][0], EnergizerCoords[0][1])
	if cues := state.TakeSoundCues(); !reflect.DeepEqual(cues, []SoundCue{SoundCueWaka, SoundCueWaka}) {
		t.Fatalf("each consumed collectible should record one waka, got %v", cues)
	}
	ghost := &state.Match.Blinky
	state.Match.PacMan.X, state.Match.PacMan.Y = ghost.X, ghost.Y
	state.Match.PacMan.EatingPauseTicks = 1
	state.UpdateActors(DirectionNone, false)
	if state.Phase != PhaseGhostCaptured {
		t.Fatal("expected a ghost capture")
	}
	if cues := state.TakeSoundCues(); !reflect.DeepEqual(cues, []SoundCue{SoundCueGhostEaten}) {
		t.Fatalf("capture should record the ghost-eaten cue, got %v", cues)
	}
	for range GhostCaptureTicks {
		state.Tick()
	}
	state.Match.Fruit = FruitState{Visible: true, Ticks: 300}
	state.Match.PacMan.X, state.Match.PacMan.Y = FruitX, FruitY
	if !state.ConsumeFruit() {
		t.Fatal("expected the bonus to be eaten")
	}
	if cues := state.TakeSoundCues(); !reflect.DeepEqual(cues, []SoundCue{SoundCueFruitEaten}) {
		t.Fatalf("eating the bonus should record the fruit cue, got %v", cues)
	}
	state.LoseLife()
	for tick := 0; tick < DeathFreezeTicks-1; tick++ {
		state.Tick()
	}
	if len(state.SoundCues) != 0 {
		t.Fatal("the collision freeze should be silent")
	}
	state.Tick()
	if cues := state.TakeSoundCues(); !reflect.DeepEqual(cues, []SoundCue{SoundCueDying}) {
		t.Fatalf("the death animation start should record the dying cue, got %v", cues)
	}
	finishDeath(t, &state)
	if len(state.SoundCues) != 0 {
		t.Fatal("the later READY! intro should not repeat the start jingle")
	}
	state.addSoundCue(SoundCueWaka)
	state.EnterAttract()
	for range 10 {
		state.Tick()
	}
	if len(state.SoundCues) != 0 {
		t.Fatal("the attract screen should discard pending cues and stay silent")
	}
}
