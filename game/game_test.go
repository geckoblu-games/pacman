package game

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

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

// TestNewStartsMatchAutomatically checks that a new game is playable without pressing Enter.
func TestNewStartsMatchAutomatically(t *testing.T) {
	game, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if game.State.Phase != PhasePlaying {
		t.Fatalf("new game should start in playing phase, got %v", game.State.Phase)
	}
}

// TestPacManSpriteLoaded checks that Pac-Man uses the atlas sprite and starts centered in the maze.
func TestPacManSpriteLoaded(t *testing.T) {
	game, err := New()
	if err != nil {
		t.Fatal(err)
	}
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

// TestScenePlayDelegation checks shared state, level advancement, and read-only drawing.
func TestScenePlayDelegation(t *testing.T) {
	game, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if game.scenePlay.State != &game.State {
		t.Fatal("scene should share the game's state")
	}
	game.State.HighScore = 100
	game.State.Phase = PhaseLevelComplete
	if err := game.Update(); err != nil {
		t.Fatal(err)
	}
	if game.State.Phase != PhaseReady || game.State.Match.Level.Number != 2 || game.State.HighScore != 100 {
		t.Fatal("game update should delegate level advancement to the scene")
	}
	screen := ebiten.NewImage(ScreenWidth, ScreenHeight)
	defer screen.Deallocate()
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
