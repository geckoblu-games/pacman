package game

// SoundCue identifies a one-shot sound requested by the simulation. Background loops
// are not cues; the presenter derives them from the current state each frame.
type SoundCue uint8

const (
	// SoundCueStart is the READY! jingle of a match's opening intro.
	SoundCueStart SoundCue = iota
	// SoundCueWaka is emitted for every eaten coin or energizer.
	SoundCueWaka
	// SoundCueFruitEaten is emitted when the bonus is eaten.
	SoundCueFruitEaten
	// SoundCueGhostEaten is emitted when a frightened ghost is captured.
	SoundCueGhostEaten
	// SoundCueDying is emitted when the death animation starts, after the collision freeze.
	SoundCueDying
	// soundCueCount is the number of defined cues.
	soundCueCount
)

// addSoundCue records a cue for the presenter to play; it has no other effect on the simulation.
func (state *GameState) addSoundCue(cue SoundCue) {
	state.SoundCues = append(state.SoundCues, cue)
}

// TakeSoundCues returns the cues recorded since the previous call and clears the list.
// The returned slice is only valid until the next cue is recorded.
func (state *GameState) TakeSoundCues() []SoundCue {
	cues := state.SoundCues
	state.SoundCues = state.SoundCues[:0]
	return cues
}
