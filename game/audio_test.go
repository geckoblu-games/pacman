package game

import (
	"testing"

	"pacman/assets"
)

// TestSelectBackgroundLoop checks the loop chosen for each situation: silence outside
// unpaused play, the siren level rising per eaten quarter, the frightened track, and the
// retreating track taking precedence while eyes return.
func TestSelectBackgroundLoop(t *testing.T) {
	state := GameState{}
	if got := selectBackgroundLoop(&state); got != loopNone {
		t.Fatalf("no match should be silent, got %v", got)
	}
	state.StartNewMatch()
	if got := selectBackgroundLoop(&state); got != loopNone {
		t.Fatalf("the READY! intro should be silent, got %v", got)
	}
	state.BeginPlaying()
	if got := selectBackgroundLoop(&state); got != loopSiren {
		t.Fatalf("fresh level should play the slowest siren, got %v", got)
	}
	total := len(CoinCoords) + len(EnergizerCoords)
	for level := 1; level < assets.SirenLevels; level++ {
		state.Match.Level.Remaining = total - level*total/assets.SirenLevels
		if got := selectBackgroundLoop(&state); got != loopSiren+backgroundLoop(level) {
			t.Fatalf("eating %d quarters should select siren level %d, got %v", level, level, got)
		}
	}
	state.Match.Level.Remaining = 1
	if got := selectBackgroundLoop(&state); got != loopSiren+assets.SirenLevels-1 {
		t.Fatalf("the last collectible should keep the fastest siren, got %v", got)
	}
	state.Match.Blinky.Frighten(state.Match.Level)
	if got := selectBackgroundLoop(&state); got != loopFrightened {
		t.Fatalf("frightened ghosts should select the frightened loop, got %v", got)
	}
	state.Match.Clyde.Mode = GhostEaten
	if got := selectBackgroundLoop(&state); got != loopRetreating {
		t.Fatalf("returning eyes should take precedence, got %v", got)
	}
	state.TogglePause()
	if got := selectBackgroundLoop(&state); got != loopNone {
		t.Fatalf("a paused match should be silent, got %v", got)
	}
	state.TogglePause()
	state.LoseLife()
	if got := selectBackgroundLoop(&state); got != loopNone {
		t.Fatalf("the death sequence should have no loop, got %v", got)
	}
}
