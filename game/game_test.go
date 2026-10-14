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
