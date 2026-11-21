package game

const (
	// controlsFirstRevealTick is the tick the first key appears, matching the first cast ghost.
	controlsFirstRevealTick = 60
	// controlsRevealInterval separates consecutive key reveals, matching the cast cadence.
	controlsRevealInterval = 90
	// controlsDescriptionDelay is the delay between a key and its description, matching
	// the delay between a cast ghost's sprite and its name.
	controlsDescriptionDelay = 60
	// controlsStartTick is when the start prompt begins flashing, after the last key.
	controlsStartTick = controlsFirstRevealTick + len(ControlsEntries)*controlsRevealInterval
	// controlsStartFlashHalfPeriod is the duration of each visible or hidden prompt interval.
	controlsStartFlashHalfPeriod = 30
	// controlsStartDuration is the three-second duration of the flashing prompt.
	controlsStartDuration = 180
	// ControlsCycleTicks is the length of the controls scene, including the start prompt.
	ControlsCycleTicks = controlsStartTick + controlsStartDuration
)

// ControlsEntry pairs a key label with the action it performs.
type ControlsEntry struct {
	// Key is the label of the key or key pair, using arrow glyphs for the cursor keys.
	Key string
	// Description is the action performed by the key.
	Description string
}

// ControlsEntries lists the game's keys in display order: the four directions, then
// pause and mute.
var ControlsEntries = [6]ControlsEntry{
	{Key: "↑ / W", Description: "MOVE UP"},
	{Key: "↓ / S", Description: "MOVE DOWN"},
	{Key: "← / A", Description: "MOVE LEFT"},
	{Key: "→ / D", Description: "MOVE RIGHT"},
	{Key: "P", Description: "PAUSE"},
	{Key: "M", Description: "MUTE"},
}

// ControlsState stores the renderer-independent progress of the controls scene of the
// attract loop, a project addition without an arcade counterpart shown after the
// presentation. It is driven by Tick at 60 Hz; GameState switches scene when Finished
// reports true.
type ControlsState struct {
	// Ticks counts ticks since the scene started; it stops at ControlsCycleTicks.
	Ticks int
}

// newControlsState returns the scene at tick zero before any control is revealed.
func newControlsState() ControlsState {
	return ControlsState{}
}

// Tick advances the scene by one 60 Hz tick until it is finished.
func (controls *ControlsState) Tick() {
	if controls.Finished() {
		return
	}
	controls.Ticks++
}

// Finished reports whether the scene has run for ControlsCycleTicks.
func (controls ControlsState) Finished() bool {
	return controls.Ticks >= ControlsCycleTicks
}

// controlsKeyRevealTick returns the tick the display-order entry's key appears.
func controlsKeyRevealTick(index int) int {
	return controlsFirstRevealTick + index*controlsRevealInterval
}

// controlsDescriptionRevealTick returns the tick the display-order entry's description appears.
func controlsDescriptionRevealTick(index int) int {
	return controlsKeyRevealTick(index) + controlsDescriptionDelay
}

// KeyRevealed reports whether the display-order entry's key label has appeared.
func (controls ControlsState) KeyRevealed(index int) bool {
	return controls.Ticks >= controlsKeyRevealTick(index)
}

// KeyDescribed reports whether the display-order entry's description has appeared.
func (controls ControlsState) KeyDescribed(index int) bool {
	return controls.Ticks >= controlsDescriptionRevealTick(index)
}

// StartPromptVisible reports whether the start prompt is in a visible flash interval.
func (controls ControlsState) StartPromptVisible() bool {
	return controls.Ticks >= controlsStartTick && controls.Ticks < ControlsCycleTicks &&
		((controls.Ticks-controlsStartTick)/controlsStartFlashHalfPeriod)%2 == 0
}
