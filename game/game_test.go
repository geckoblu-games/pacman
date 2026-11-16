package game

import (
	"errors"
	"image"
	"path/filepath"
	"testing"

	"pacman/highscore"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestHighScoreCheckpoints checks immediate level-clear persistence and game-over persistence.
func TestHighScoreCheckpoints(t *testing.T) {
	game, err := New()
	if err != nil {
		t.Fatal(err)
	}
	game.State.StartNewMatch()
	path := filepath.Join(t.TempDir(), "highscore")
	store := highscore.New(path)
	if err := store.Save(100); err != nil {
		t.Fatal(err)
	}
	game.State.HighScore, err = store.Load()
	if err != nil {
		t.Fatal(err)
	}
	game.SaveHighScore = store.Save
	game.State.BeginPlaying()
	game.State.AddScore(110)
	if err := game.Update(); err != nil {
		t.Fatal(err)
	}
	if score, err := highscore.New(path).Load(); err != nil || score != 100 {
		t.Fatalf("active play should not persist a new record: %d, %v", score, err)
	}
	level := &game.State.Match.Level
	level.Coins = [240]bool{true}
	level.Energizers = [4]bool{}
	level.Remaining = 1
	coord := CoinCoords[0]
	game.State.Match.PacMan = PacManState{
		X: float64(coord[0]*CellSize + CellSize/2),
		Y: float64(coord[1]*CellSize + CellSize/2),
	}
	if err := game.Update(); err != nil {
		t.Fatal(err)
	}
	if game.State.Phase != PhaseLevelComplete || game.State.PhaseTicks != 0 {
		t.Fatal("expected level completion on this update")
	}
	if score, err := highscore.New(path).Load(); err != nil || score != 120 {
		t.Fatalf("level completion should immediately persist the final dot: %d, %v", score, err)
	}
	game.State.AdvanceLevel()
	game.State.BeginPlaying()
	game.State.AddScore(20)
	game.State.Match.Lives = 1
	game.State.LoseLife()
	for range DeathFreezeTicks + DeathAnimationTicks + DeathEmptyTicks - 1 {
		if err := game.Update(); err != nil {
			t.Fatal(err)
		}
	}
	if score, err := highscore.New(path).Load(); err != nil || score != 120 {
		t.Fatalf("death sequence should not persist a game-over checkpoint early: %d, %v", score, err)
	}
	if err := game.Update(); err != nil {
		t.Fatal(err)
	}
	if score, err := highscore.New(path).Load(); err != nil || score != 140 {
		t.Fatalf("game over should persist the record: %d, %v", score, err)
	}
}

// TestHighScoreCheckpointFailure checks errors do not stop play or retry every tick.
func TestHighScoreCheckpointFailure(t *testing.T) {
	game, err := New()
	if err != nil {
		t.Fatal(err)
	}
	game.State.StartNewMatch()
	calls := 0
	game.SaveHighScore = func(score int) error {
		calls++
		if calls == 1 {
			return errors.New("test save failure")
		}
		return nil
	}
	game.State.Phase = PhaseLevelComplete
	for tick := 0; tick < LevelClearTicks; tick++ {
		if err := game.Update(); err != nil {
			t.Fatalf("save failure should not stop play: %v", err)
		}
	}
	if calls != 1 {
		t.Fatalf("expected one attempt at level completion, got %d", calls)
	}
	game.State.BeginPlaying()
	game.State.Match.Lives = 1
	game.State.LoseLife()
	for range DeathFreezeTicks + DeathAnimationTicks + DeathEmptyTicks {
		if err := game.Update(); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 2 {
		t.Fatalf("expected another attempt at the next checkpoint, got %d", calls)
	}
}

// TestCaptureResumeRenderingPosition checks the first resumed frame keeps both actors in place.
func TestCaptureResumeRenderingPosition(t *testing.T) {
	game, err := New()
	if err != nil {
		t.Fatal(err)
	}
	game.State.StartNewMatch()
	game.State.BeginPlaying()
	match := game.State.Match
	match.Blinky.Frighten(match.Level)
	match.PacMan.X, match.PacMan.Y = match.Blinky.X, match.Blinky.Y
	match.PacMan.EatingPauseTicks = 1
	game.State.UpdateActors(DirectionNone, false)
	playerX, playerY := match.PacMan.X, match.PacMan.Y
	ghostX, ghostY := match.Blinky.X, match.Blinky.Y
	for range GhostCaptureTicks {
		if err := game.Update(); err != nil {
			t.Fatal(err)
		}
	}
	if game.State.Phase != PhasePlaying || match.PacMan.X != playerX || match.PacMan.Y != playerY ||
		match.Blinky.X != ghostX || match.Blinky.Y != ghostY || match.Blinky.Mode != GhostEaten {
		t.Fatal("the first resumed frame should show Pac-Man and eyes at the capture positions")
	}
}

// TestBottomHUDPosition verifies icon mirroring, spacing, and screen bounds.
func TestBottomHUDPosition(t *testing.T) {
	size := image.Pt(26, 26)
	lifeX, lifeY := bottomHUDPosition(size, 0, false)
	bonusX, bonusY := bottomHUDPosition(size, 0, true)
	if lifeX != 35 || lifeY != 547 {
		t.Fatalf("life position changed: got (%v, %v)", lifeX, lifeY)
	}
	if bonusX != 387 || bonusY != lifeY || lifeX+bonusX+float64(size.X) != ScreenWidth {
		t.Fatalf("bonus position should mirror the first life: got (%v, %v)", bonusX, bonusY)
	}
	for index := 0; index < 7; index++ {
		x, y := bottomHUDPosition(size, index, true)
		if x != bonusX-float64(index*2*CellSize) || y != lifeY {
			t.Fatalf("incorrect bonus slot %d: (%v, %v)", index, x, y)
		}
		if x < 0 || x+float64(size.X) > ScreenWidth || y < 0 || y+float64(size.Y) > ScreenHeight {
			t.Fatalf("bonus slot %d is outside the screen", index)
		}
	}
}

// TestAllowedPathGrid checks the matrix dimensions, traversable-cell count, and collectible coverage.
func TestAllowedPathGrid(t *testing.T) {
	allowedCount := 0
	for _, row := range AllowedPath {
		for _, allowed := range row {
			if allowed {
				allowedCount++
			}
		}
	}
	if allowedCount != 300 {
		t.Fatalf("expected 300 allowed cells, got %d", allowedCount)
	}
	for _, coord := range CoinCoords {
		if !AllowedPath[coord[1]][coord[0]] {
			t.Fatalf("coin coordinate %v should be on an allowed cell", coord)
		}
	}
	for _, coord := range EnergizerCoords {
		if !AllowedPath[coord[1]][coord[0]] {
			t.Fatalf("energizer coordinate %v should be on an allowed cell", coord)
		}
	}
	if isWalkable(-1, 0) || isWalkable(0, GridRows) {
		t.Fatal("out-of-grid coordinates should not be walkable")
	}
}

// TestBonusSpritesLoaded checks that every level symbol has a 26-pixel atlas sprite.
func TestBonusSpritesLoaded(t *testing.T) {
	game, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if len(game.scenePlay.images.BonusSymbols) != 8 {
		t.Fatal("expected eight bonus sprites")
	}
	for number := 1; number <= 21; number++ {
		parameters := ParametersForLevel(number)
		sprite := game.scenePlay.images.BonusSymbols[parameters.BonusSymbol]
		if sprite == nil || sprite.Bounds().Size() != image.Pt(26, 26) {
			t.Fatalf("missing or incorrectly sized sprite for %q", parameters.BonusSymbol)
		}
	}
}

// TestNewStartsInAttract checks that a new game shows the attract screen, that the
// presentation loops without a match, and that starting a match enters the READY! intro.
func TestNewStartsInAttract(t *testing.T) {
	game, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if game.State.Phase != PhaseAttract || game.State.Match != nil {
		t.Fatalf("new game should start on the attract screen, got %v", game.State.Phase)
	}
	for tick := 0; tick < AttractCycleTicks+10; tick++ {
		if err := game.Update(); err != nil {
			t.Fatal(err)
		}
	}
	if game.State.Phase != PhaseAttract || game.State.Match != nil || game.State.Attract.Ticks != 10 {
		t.Fatalf("attract screen should loop without starting a match: phase %v tick %d", game.State.Phase, game.State.Attract.Ticks)
	}
	game.State.StartNewMatch()
	if game.State.Phase != PhaseReady || game.State.PhaseTicks != 0 {
		t.Fatalf("starting a match should enter the ready intro, got %v at tick %d", game.State.Phase, game.State.PhaseTicks)
	}
	for tick := 0; tick < ReadyTicks; tick++ {
		if err := game.Update(); err != nil {
			t.Fatal(err)
		}
	}
	if game.State.Phase != PhasePlaying {
		t.Fatalf("intro should end in playing phase, got %v", game.State.Phase)
	}
}

// TestAttractDrawReadOnly checks that drawing every attract tick does not mutate state.
func TestAttractDrawReadOnly(t *testing.T) {
	game, err := New()
	if err != nil {
		t.Fatal(err)
	}
	screen := ebiten.NewImage(ScreenWidth, ScreenHeight)
	defer screen.Deallocate()
	for tick := 0; tick < AttractCycleTicks; tick++ {
		before := game.State
		game.Draw(screen)
		if game.State != before {
			t.Fatalf("tick %d: drawing the attract screen changed the state", tick)
		}
		if err := game.Update(); err != nil {
			t.Fatal(err)
		}
	}
}

// TestPacManSpriteLoaded checks that Pac-Man uses the atlas sprite and starts centered in the maze.
func TestPacManSpriteLoaded(t *testing.T) {
	game, err := New()
	if err != nil {
		t.Fatal(err)
	}
	game.State.StartNewMatch()
	for index, frame := range game.scenePlay.pacMan.frames {
		if frame == nil || frame.Bounds().Size() != image.Pt(26, 26) {
			t.Fatalf("Pac-Man frame %d should be a loaded 26-pixel atlas image", index)
		}
	}
	if got, want := game.State.Match.PacMan.X, float64(ScreenWidth)/2; got != want {
		t.Fatalf("Pac-Man should start horizontally centered: got %v, want %v", got, want)
	}
	if got, want := game.State.Match.PacMan.Y, float64(pacManStartRow*CellSize+CellSize/2); got != want {
		t.Fatalf("Pac-Man should start in the lower corridor: got %v, want %v", got, want)
	}
}

// TestPacManSpriteWrapDraw checks that opposite sprite portions appear at both screen edges.
func TestPacManSpriteWrapDraw(t *testing.T) {
	tests := []struct {
		// centerX is Pac-Man's horizontal center coordinate.
		centerX float64
		// want contains the primary and wrapped sprite centers.
		want []float64
	}{
		{6, []float64{6, float64(ScreenWidth + 6)}},
		{float64(ScreenWidth - 6), []float64{float64(ScreenWidth - 6), -6}},
		{float64(ScreenWidth / 2), []float64{float64(ScreenWidth / 2)}},
	}
	for _, test := range tests {
		positions, count := pacManDrawPositions(test.centerX, 26)
		if count != len(test.want) {
			t.Fatalf("draw count at x=%v: got %d, want %d", test.centerX, count, len(test.want))
		}
		for index, want := range test.want {
			if positions[index] != want {
				t.Fatalf("draw position %d at x=%v: got %v, want %v", index, test.centerX, positions[index], want)
			}
		}
	}
}

// TestScenePlayDelegation checks shared state, timed level advancement, and read-only drawing.
func TestScenePlayDelegation(t *testing.T) {
	game, err := New()
	if err != nil {
		t.Fatal(err)
	}
	game.State.StartNewMatch()
	if game.scenePlay.State != &game.State {
		t.Fatal("scene should share the game's state")
	}
	game.State.HighScore = 100
	game.State.Phase = PhaseLevelComplete
	game.State.PhaseTicks = 0
	screen := ebiten.NewImage(ScreenWidth, ScreenHeight)
	defer screen.Deallocate()
	for tick := 0; tick < LevelClearTicks; tick++ {
		if game.State.Phase != PhaseLevelComplete {
			t.Fatalf("tick %d: level should still be flashing", tick)
		}
		game.Draw(screen)
		if err := game.Update(); err != nil {
			t.Fatal(err)
		}
	}
	if game.State.Phase != PhaseReady || game.State.Match.Level.Number != 2 || game.State.HighScore != 100 {
		t.Fatal("game update should delegate timed level advancement to the scene")
	}
	before := game.State
	matchBefore := *game.State.Match
	game.Draw(screen)
	if DebugGridEnabled && game.scenePlay.images.DebugGrid == nil {
		t.Fatal("game draw should initialize the scene's debug grid")
	}
	if game.State != before || *game.State.Match != matchBefore {
		t.Fatal("drawing should not change the game state")
	}
	game.scenePlay.drawDebugAllowedPath(screen)
	if game.scenePlay.images.DebugAllowedPath == nil {
		t.Fatal("scene should initialize the allowed-path overlay")
	}
}
