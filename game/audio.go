package game

import (
	"bytes"
	"sync"

	"pacman/assets"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// backgroundLoop identifies the looping track selected from the simulation state.
type backgroundLoop uint8

const (
	// loopNone keeps every background track paused.
	loopNone backgroundLoop = iota
	// loopFrightened plays while ghosts are frightened and no eyes are returning.
	loopFrightened
	// loopRetreating plays while at least one captured ghost's eyes return to the house.
	loopRetreating
	// loopSiren is the slowest siren; the next assets.SirenLevels-1 values are faster ones.
	loopSiren
	// loopCount is the number of loop values including loopNone.
	loopCount = loopSiren + assets.SirenLevels
)

// soundPlayer presents the simulation's sound cues and background loops through
// Ebitengine. The simulation never depends on it.
type soundPlayer struct {
	// oneShots maps each cue to its restartable player.
	oneShots [soundCueCount]*audio.Player
	// wakaPending remembers a waka requested while the previous one was still playing,
	// so consecutive bites play back to back instead of overlapping or being dropped.
	wakaPending bool
	// loops holds one looping player per background track; the loopNone slot is nil.
	loops [loopCount]*audio.Player
	// current is the background track currently playing.
	current backgroundLoop
	// suspended lists the players paused by a game pause, to be resumed on unpause.
	suspended []*audio.Player
	// muted silences every player while remaining true.
	muted bool
}

// decodedSounds caches the decoded game sounds, which are immutable, so repeated
// scene construction in one process decodes them only once.
var decodedSounds struct {
	// once guards the single decoding.
	once sync.Once
	// sounds and err hold the decoding result.
	sounds assets.GameSounds
	err    error
}

// newSoundPlayer decodes the game sounds and prepares a player for each cue and loop,
// reusing the process-wide audio context when one already exists.
func newSoundPlayer() (*soundPlayer, error) {
	decodedSounds.once.Do(func() {
		decodedSounds.sounds, decodedSounds.err = assets.LoadGameSounds()
	})
	sounds, err := decodedSounds.sounds, decodedSounds.err
	if err != nil {
		return nil, err
	}
	context := audio.CurrentContext()
	if context == nil {
		context = audio.NewContext(assets.SoundSampleRate)
	}
	player := &soundPlayer{}
	for cue, pcm := range map[SoundCue][]byte{
		SoundCueStart:      sounds.Start,
		SoundCueWaka:       sounds.Waka,
		SoundCueFruitEaten: sounds.FruitEaten,
		SoundCueGhostEaten: sounds.GhostEaten,
		SoundCueDying:      sounds.Dying,
	} {
		player.oneShots[cue] = context.NewPlayerFromBytes(pcm)
	}
	loops := map[backgroundLoop][]byte{loopFrightened: sounds.Frightened, loopRetreating: sounds.Retreating}
	for level, pcm := range sounds.Sirens {
		loops[loopSiren+backgroundLoop(level)] = pcm
	}
	for loop, pcm := range loops {
		player.loops[loop], err = context.NewPlayer(audio.NewInfiniteLoop(bytes.NewReader(pcm), int64(len(pcm))))
		if err != nil {
			return nil, err
		}
	}
	return player, nil
}

// Update plays the cues recorded since the previous frame, keeps the waka queue
// flowing, suspends playback while the match is paused, and switches the background
// loop to the one selected from the state.
func (player *soundPlayer) Update(state *GameState) {
	for _, cue := range state.TakeSoundCues() {
		if cue == SoundCueWaka && player.oneShots[cue].IsPlaying() {
			player.wakaPending = true
			continue
		}
		player.restart(player.oneShots[cue])
	}
	if state.Paused {
		player.suspend()
		return
	}
	for _, suspended := range player.suspended {
		suspended.Play()
	}
	player.suspended = player.suspended[:0]
	if player.wakaPending && !player.oneShots[SoundCueWaka].IsPlaying() {
		player.wakaPending = false
		player.restart(player.oneShots[SoundCueWaka])
	}
	player.setLoop(selectBackgroundLoop(state))
}

// restart plays a one-shot from its beginning, cutting off any ongoing playback.
func (player *soundPlayer) restart(oneShot *audio.Player) {
	oneShot.Pause()
	if err := oneShot.SetPosition(0); err != nil {
		return
	}
	oneShot.Play()
}

// suspend pauses the players that are still playing, remembering them for resumption.
func (player *soundPlayer) suspend() {
	for _, oneShot := range player.oneShots {
		if oneShot.IsPlaying() {
			oneShot.Pause()
			player.suspended = append(player.suspended, oneShot)
		}
	}
	if player.current != loopNone && player.loops[player.current].IsPlaying() {
		player.loops[player.current].Pause()
		player.suspended = append(player.suspended, player.loops[player.current])
	}
}

// setLoop pauses the current background track and plays the selected one; loops keep
// their position across switches, which is inaudible for continuous tracks.
func (player *soundPlayer) setLoop(loop backgroundLoop) {
	if loop == player.current {
		return
	}
	if player.current != loopNone {
		player.loops[player.current].Pause()
	}
	player.current = loop
	if loop != loopNone {
		player.loops[loop].Play()
	}
}

// ToggleMute flips the mute flag and applies the resulting volume to every player.
func (player *soundPlayer) ToggleMute() bool {
	player.muted = !player.muted
	volume := 1.0
	if player.muted {
		volume = 0
	}
	for _, oneShot := range player.oneShots {
		oneShot.SetVolume(volume)
	}
	for _, loop := range player.loops {
		if loop != nil {
			loop.SetVolume(volume)
		}
	}
	return player.muted
}

// selectBackgroundLoop chooses the loop for the state: nothing outside unpaused play,
// the retreating track while any ghost's eyes return, the frightened track while
// ghosts are frightened, otherwise the siren level that rises with the eaten share of
// the level's collectibles (one level per quarter).
func selectBackgroundLoop(state *GameState) backgroundLoop {
	if state.Match == nil || state.Phase != PhasePlaying || state.Paused {
		return loopNone
	}
	match := state.Match
	for _, ghost := range []*GhostState{&match.Blinky, &match.Pinky, &match.Inky, &match.Clyde} {
		if ghost.Mode == GhostEaten {
			return loopRetreating
		}
	}
	if match.Blinky.frightTicks > 0 {
		return loopFrightened
	}
	total := len(CoinCoords) + len(EnergizerCoords)
	level := (total - match.Level.Remaining) * assets.SirenLevels / total
	return loopSiren + backgroundLoop(min(level, assets.SirenLevels-1))
}
