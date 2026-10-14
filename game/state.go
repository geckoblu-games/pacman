package game

// InitialLives counts the active life and reserve lives at the start of a match.
const InitialLives = 3

// Phase identifies the current stage of the match lifecycle.
type Phase uint8

const (
	// PhaseWaiting is the zero-value state with no match started.
	PhaseWaiting Phase = iota
	// PhaseReady waits for Enter before accepting gameplay actions.
	PhaseReady
	// PhasePlaying permits scoring, collectible consumption, and life loss.
	PhasePlaying
	// PhaseLevelComplete marks a cleared level awaiting advancement.
	PhaseLevelComplete
	// PhaseGameOver marks a match with no remaining lives.
	PhaseGameOver
)

// GameState stores the session record, current phase, and optional active match.
type GameState struct {
	// HighScore is the session record, preserved when starting another match.
	HighScore int
	// Phase controls which match operations are currently valid.
	Phase Phase
	// Match is nil until a match is started.
	Match *MatchState
}

// MatchState stores the score, remaining lives, and current level for one match.
type MatchState struct {
	// Score is the total points earned in this match.
	Score int
	// Lives includes the life currently in play.
	Lives int
	// Level contains the current parameters and collectible presence flags.
	Level LevelState
}

// StartNewMatch resets score, lives, and level while retaining the session record.
func (state *GameState) StartNewMatch() {
	state.Match = &MatchState{Lives: InitialLives, Level: NewLevelState(1)}
	state.Phase = PhaseReady
}

// BeginPlaying enters active play only when a match is in the ready phase.
func (state *GameState) BeginPlaying() bool {
	if state.Match == nil || state.Phase != PhaseReady {
		return false
	}
	state.Phase = PhasePlaying
	return true
}

// AddScore awards positive points during active play and updates the session record.
func (state *GameState) AddScore(points int) {
	if state.Match == nil || state.Phase != PhasePlaying || points <= 0 {
		return
	}
	state.Match.Score += points
	if state.Match.Score > state.HighScore {
		state.HighScore = state.Match.Score
	}
}

// ConsumeCollectible awards points for a remaining dot and marks level completion.
// It rejects inactive play and returns whether a coin or energizer was consumed.
func (state *GameState) ConsumeCollectible(column, row int) bool {
	if state.Match == nil || state.Phase != PhasePlaying {
		return false
	}
	points := state.Match.Level.ConsumeCollectible(column, row)
	if points == 0 {
		return false
	}
	state.AddScore(points)
	if state.Match.Level.Remaining == 0 {
		state.Phase = PhaseLevelComplete
	}
	return true
}

// AdvanceLevel replaces a cleared level and enters ready, preserving score and lives.
func (state *GameState) AdvanceLevel() bool {
	if state.Match == nil || state.Phase != PhaseLevelComplete {
		return false
	}
	state.Match.Level = NewLevelState(state.Match.Level.Number + 1)
	state.Phase = PhaseReady
	return true
}

// LoseLife preserves the level and enters ready or game over during active play.
func (state *GameState) LoseLife() bool {
	if state.Match == nil || state.Phase != PhasePlaying {
		return false
	}
	state.Match.Lives--
	if state.Match.Lives == 0 {
		state.Phase = PhaseGameOver
	} else {
		state.Phase = PhaseReady
	}
	return true
}
