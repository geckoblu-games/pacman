package game

import "testing"

// TestControlsModes checks each instruction set, reveal timeline and scene duration.
func TestControlsModes(t *testing.T) {
	tests := []struct {
		name       string
		touch      bool
		keys       []string
		prompt     string
		start, end int
	}{
		{"keyboard", false, []string{"↑ / W", "↓ / S", "← / A", "→ / D", "P", "M"}, "PRESS ENTER TO START", 600, 780},
		{"touch", true, []string{"SWIPE", "TAP", "TAP HUD", "2 FINGERS"}, "TAP TO START", 420, 600},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controls := newControlsState(tt.touch)
			entries := controls.Entries()
			if len(entries) != len(tt.keys) || controls.StartPrompt() != tt.prompt ||
				controls.startTick() != tt.start || controls.CycleTicks() != tt.end {
				t.Fatalf("unexpected controls: entries=%v prompt=%q start=%d end=%d",
					entries, controls.StartPrompt(), controls.startTick(), controls.CycleTicks())
			}
			for index, key := range tt.keys {
				if entries[index].Key != key {
					t.Fatalf("entry %d: got %q, want %q", index, entries[index].Key, key)
				}
				keyTick := controlsKeyRevealTick(index)
				descriptionTick := controlsDescriptionRevealTick(index)
				if index < len(attractGhostRevealTicks) &&
					(keyTick != attractGhostRevealTicks[index] || descriptionTick != attractGhostNameTicks[index]) {
					t.Fatalf("entry %d should follow the cast reveal timing", index)
				}
				before := ControlsState{Touch: tt.touch, Ticks: keyTick - 1}
				at := ControlsState{Touch: tt.touch, Ticks: keyTick}
				described := ControlsState{Touch: tt.touch, Ticks: descriptionTick}
				if before.KeyRevealed(index) || !at.KeyRevealed(index) || at.KeyDescribed(index) ||
					(ControlsState{Ticks: descriptionTick - 1}).KeyDescribed(index) || !described.KeyDescribed(index) {
					t.Fatalf("entry %d revealed at incorrect ticks", index)
				}
			}
			for tick := 0; tick <= tt.end; tick++ {
				controls.Ticks = tick
				want := tick >= tt.start && tick < tt.end && ((tick-tt.start)/30)%2 == 0
				if controls.StartPromptVisible() != want || controls.Finished() != (tick >= tt.end) {
					t.Fatalf("tick %d: visibility=%v finished=%v", tick, controls.StartPromptVisible(), controls.Finished())
				}
			}
			controls.Tick()
			if controls.Ticks != tt.end {
				t.Fatalf("finished scene advanced to tick %d", controls.Ticks)
			}
		})
	}
}
