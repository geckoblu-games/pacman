package game

import "testing"

// runControls advances a fresh controls scene to the requested tick.
func runControls(ticks int) ControlsState {
	controls := newControlsState()
	for controls.Ticks < ticks {
		controls.Tick()
	}
	return controls
}

// TestControlsRevealTimeline checks that each key and its description appear in display
// order with the cast table's cadence.
func TestControlsRevealTimeline(t *testing.T) {
	for index := range attractGhostRevealTicks {
		if controlsKeyRevealTick(index) != attractGhostRevealTicks[index] ||
			controlsDescriptionRevealTick(index) != attractGhostNameTicks[index] {
			t.Fatalf("entry %d should follow the cast reveal timing", index)
		}
	}
	if runControls(0).KeyRevealed(0) {
		t.Fatal("the scene should start empty")
	}
	for index := range ControlsEntries {
		keyTick, descriptionTick := controlsKeyRevealTick(index), controlsDescriptionRevealTick(index)
		before, at := runControls(keyTick-1), runControls(keyTick)
		if before.KeyRevealed(index) || !at.KeyRevealed(index) || at.KeyDescribed(index) {
			t.Fatalf("entry %d key should appear exactly at tick %d without its description", index, keyTick)
		}
		if index > 0 && !at.KeyDescribed(index-1) {
			t.Fatalf("entry %d key should appear after the previous description", index)
		}
		described := runControls(descriptionTick)
		if runControls(descriptionTick-1).KeyDescribed(index) || !described.KeyDescribed(index) {
			t.Fatalf("entry %d description should appear exactly at tick %d", index, descriptionTick)
		}
	}
}

// TestControlsStartPrompt checks its reveal, flashing and exact three-second duration.
func TestControlsStartPrompt(t *testing.T) {
	if controlsStartTick != 600 || ControlsCycleTicks != 780 {
		t.Fatalf("start prompt should run from tick 600 through tick 779, got %d to %d",
			controlsStartTick, ControlsCycleTicks-1)
	}
	if (ControlsState{Ticks: 599}).StartPromptVisible() {
		t.Fatal("start prompt should be hidden before its reveal")
	}
	for tick := 600; tick < 780; tick++ {
		visible := (ControlsState{Ticks: tick}).StartPromptVisible()
		if want := ((tick-600)/30)%2 == 0; visible != want {
			t.Fatalf("tick %d: prompt visibility = %v, want %v", tick, visible, want)
		}
	}
	if runControls(ControlsCycleTicks - 1).Finished() {
		t.Fatal("scene should not finish before its length")
	}
	finished := runControls(ControlsCycleTicks)
	if finished.StartPromptVisible() {
		t.Fatal("start prompt should be hidden when the scene ends")
	}
	stopped := finished
	stopped.Tick()
	if !finished.Finished() || stopped != finished {
		t.Fatalf("scene should finish and stop at tick %d", ControlsCycleTicks)
	}
}
