package assets

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

const (
	// SoundSampleRate is the playback rate of every decoded sound and of the audio context.
	SoundSampleRate = 48000
	// SirenLevels counts the siren pitch variants, from slowest to fastest.
	SirenLevels = 4
	// sirenSpeedStep is the playback speed increase between consecutive siren levels.
	sirenSpeedStep = 0.05
	// silenceThreshold is the absolute 16-bit sample amplitude treated as silence when trimming.
	silenceThreshold = 300
	// pcmFrameBytes is the size of one stereo 16-bit sample frame.
	pcmFrameBytes = 4
)

// GameSounds contains the decoded sounds as 16-bit little-endian stereo PCM at
// SoundSampleRate, with leading and trailing silence removed.
type GameSounds struct {
	// Start is the READY! jingle played once at the start of a match.
	Start []byte
	// Waka is the single bite played for each eaten collectible.
	Waka []byte
	// FruitEaten is played when the bonus is eaten.
	FruitEaten []byte
	// GhostEaten is played when a frightened ghost is captured.
	GhostEaten []byte
	// Dying accompanies Pac-Man's death animation.
	Dying []byte
	// Frightened is the loop played while ghosts are frightened.
	Frightened []byte
	// Retreating is the loop played while captured ghosts' eyes return to the house.
	Retreating []byte
	// Sirens holds the background siren loop at SirenLevels increasing pitches.
	Sirens [SirenLevels][]byte
}

// LoadSoundPCM decodes an embedded Ogg Vorbis file from the sounds directory into
// 16-bit little-endian stereo PCM at SoundSampleRate and trims surrounding silence.
// A speed above 1 shortens the sound and raises its pitch proportionally; it must be positive.
func LoadSoundPCM(name string, speed float64) ([]byte, error) {
	if speed <= 0 {
		return nil, fmt.Errorf("sound speed must be positive, got %v", speed)
	}
	data, err := ReadFile("sounds/" + name)
	if err != nil {
		return nil, err
	}
	stream, err := vorbis.DecodeWithoutResampling(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode sound %q: %w", name, err)
	}
	// Declaring a faster source rate makes the resampler compress the sound in time.
	sourceRate := int(float64(stream.SampleRate())*speed + 0.5)
	pcm, err := io.ReadAll(audio.ResampleReader(stream, stream.Length(), sourceRate, SoundSampleRate))
	if err != nil {
		return nil, fmt.Errorf("resample sound %q: %w", name, err)
	}
	return trimSilence(pcm), nil
}

// trimSilence removes leading and trailing sample frames whose channels both stay
// within silenceThreshold, keeping whole frames so channel alignment is preserved.
func trimSilence(pcm []byte) []byte {
	pcm = pcm[:len(pcm)/pcmFrameBytes*pcmFrameBytes]
	loud := func(frame []byte) bool {
		left := int16(binary.LittleEndian.Uint16(frame))
		right := int16(binary.LittleEndian.Uint16(frame[2:]))
		return left > silenceThreshold || left < -silenceThreshold || right > silenceThreshold || right < -silenceThreshold
	}
	start := 0
	for start < len(pcm) && !loud(pcm[start:start+pcmFrameBytes]) {
		start += pcmFrameBytes
	}
	end := len(pcm)
	for end > start && !loud(pcm[end-pcmFrameBytes:end]) {
		end -= pcmFrameBytes
	}
	return pcm[start:end]
}

// LoadGameSounds decodes every game sound, generating the siren pitch variants from
// the single embedded siren by increasing its speed sirenSpeedStep per level.
func LoadGameSounds() (GameSounds, error) {
	var sounds GameSounds
	var err error
	for _, entry := range []struct {
		// target receives the decoded PCM of the named file.
		target *[]byte
		// name is the file name inside the sounds directory.
		name string
	}{
		{&sounds.Start, "start.ogg"},
		{&sounds.Waka, "waka.ogg"},
		{&sounds.FruitEaten, "fruit_eaten.ogg"},
		{&sounds.GhostEaten, "ghost_eaten.ogg"},
		{&sounds.Dying, "dying.ogg"},
		{&sounds.Frightened, "frightened.ogg"},
		{&sounds.Retreating, "eaten.ogg"},
	} {
		if *entry.target, err = LoadSoundPCM(entry.name, 1); err != nil {
			return GameSounds{}, err
		}
	}
	for level := range sounds.Sirens {
		speed := 1 + float64(level)*sirenSpeedStep
		if sounds.Sirens[level], err = LoadSoundPCM("siren.ogg", speed); err != nil {
			return GameSounds{}, fmt.Errorf("load siren level %d: %w", level, err)
		}
	}
	return sounds, nil
}
