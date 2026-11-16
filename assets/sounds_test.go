package assets

import "testing"

// TestLoadGameSounds checks that every sound decodes to whole stereo frames, that the
// siren variants shorten as their pitch rises, and that silence trimming keeps the
// loud content of a known sound.
func TestLoadGameSounds(t *testing.T) {
	sounds, err := LoadGameSounds()
	if err != nil {
		t.Fatal(err)
	}
	for name, pcm := range map[string][]byte{
		"start": sounds.Start, "waka": sounds.Waka, "fruit": sounds.FruitEaten,
		"ghost": sounds.GhostEaten, "dying": sounds.Dying, "frightened": sounds.Frightened,
		"retreating": sounds.Retreating,
	} {
		if len(pcm) == 0 || len(pcm)%pcmFrameBytes != 0 {
			t.Errorf("%s: expected non-empty whole-frame PCM, got %d bytes", name, len(pcm))
		}
	}
	for level := 1; level < SirenLevels; level++ {
		if len(sounds.Sirens[level]) >= len(sounds.Sirens[level-1]) {
			t.Errorf("siren level %d should be shorter than level %d", level, level-1)
		}
	}
	// The captured-ghost effect carries about three seconds of trailing silence in the file.
	if seconds := float64(len(sounds.GhostEaten)) / (SoundSampleRate * pcmFrameBytes); seconds > 1 {
		t.Errorf("ghost eaten sound should be trimmed below one second, got %.2fs", seconds)
	}
}

// TestTrimSilence checks frame-aligned trimming of quiet edges and all-silent input.
func TestTrimSilence(t *testing.T) {
	loud := []byte{0x00, 0x10, 0x00, 0x00}
	quiet := []byte{0x01, 0x00, 0xff, 0xff}
	var pcm []byte
	pcm = append(pcm, quiet...)
	pcm = append(pcm, quiet...)
	pcm = append(pcm, loud...)
	pcm = append(pcm, quiet...)
	pcm = append(pcm, loud...)
	pcm = append(pcm, quiet...)
	pcm = append(pcm, 0x7f)
	if got := trimSilence(pcm); len(got) != 3*pcmFrameBytes {
		t.Fatalf("expected the three inner frames, got %d bytes", len(got))
	}
	if got := trimSilence(append(quiet, quiet...)); len(got) != 0 {
		t.Fatalf("all-silent input should trim to nothing, got %d bytes", len(got))
	}
}
