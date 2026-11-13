package game

import "math/rand"

const (
	// InitialLives counts the active life and reserve lives at the start of a match.
	InitialLives = 3
	// LevelClearTicks is the exact 60 Hz duration of the level-clear maze flash (2 seconds).
	LevelClearTicks = 120
	// LevelClearFlashes counts the white periods shown during the level-clear flash.
	LevelClearFlashes = 4
	// ReadyTicks is the approximate 60 Hz duration of the READY! intro (2 seconds).
	ReadyTicks = 120
)

// Phase identifies the current stage of the match lifecycle.
type Phase uint8

const (
	// PhaseWaiting is the zero-value state with no match started.
	PhaseWaiting Phase = iota
	// PhaseReady shows the READY! intro with controls locked until Tick enters play.
	PhaseReady
	// PhasePlaying permits scoring, collectible consumption, and life loss.
	PhasePlaying
	// PhaseLevelComplete flashes the cleared maze until Tick advances the level.
	PhaseLevelComplete
	// PhaseGameOver marks a match with no remaining lives.
	PhaseGameOver
)

// GameState stores the session record, current phase, and optional active match.
type GameState struct {
	// HighScore is the record, optionally loaded and persisted by the application.
	HighScore int
	// Phase controls which match operations are currently valid.
	Phase Phase
	// PhaseTicks counts 60 Hz ticks elapsed in the current timed phase (ready or level complete).
	PhaseTicks int
	// Paused freezes Tick and gameplay input until TogglePause clears it.
	Paused bool
	// Match is nil until a match is started.
	Match *MatchState
	// Random draws bonus durations; nil creates an unpredictably seeded generator on first use.
	Random *rand.Rand
}

// MatchState stores the score, remaining lives, and current level for one match.
type MatchState struct {
	// Score is the total points earned in this match.
	Score int
	// Lives includes the life currently in play.
	Lives int
	// Level contains the current parameters and collectible presence flags.
	Level LevelState
	// PacMan stores the player's position and movement intent.
	PacMan PacManState
	// Fruit stores the current level's bonus visibility and score popup.
	Fruit FruitState
}

// StartNewMatch resets score, lives, and level while retaining the session record.
// The new match starts unpaused in the ready intro.
func (state *GameState) StartNewMatch() {
	state.Match = &MatchState{Lives: InitialLives, Level: NewLevelState(1), PacMan: NewPacManState()}
	state.Paused = false
	state.enterPhase(PhaseReady)
}

// TogglePause flips the pause flag of a match that is not over and reports the new
// value. Without a match or after game over it clears the flag and returns false.
func (state *GameState) TogglePause() bool {
	if state.Match == nil || state.Phase == PhaseGameOver {
		state.Paused = false
		return false
	}
	state.Paused = !state.Paused
	return state.Paused
}

// enterPhase switches phase and restarts the timed-phase tick counter.
func (state *GameState) enterPhase(phase Phase) {
	state.Phase = phase
	state.PhaseTicks = 0
}

// Tick advances the timed phases by one 60 Hz tick: the level-clear flash
// advances the level after LevelClearTicks, the ready intro begins play after ReadyTicks,
// and active play runs the bonus timers. Other phases and a paused match are unaffected.
func (state *GameState) Tick() {
	if state.Match == nil || state.Paused {
		return
	}
	switch state.Phase {
	case PhaseLevelComplete:
		state.PhaseTicks++
		if state.PhaseTicks >= LevelClearTicks {
			state.AdvanceLevel()
		}
	case PhaseReady:
		state.PhaseTicks++
		if state.PhaseTicks >= ReadyTicks {
			state.BeginPlaying()
		}
	case PhasePlaying:
		state.Match.Fruit.Tick()
	}
}

// MazeFlashWhite reports whether the level-clear flash currently shows the white maze.
// The flash splits LevelClearTicks into 2*LevelClearFlashes equal segments, starting white.
func (state *GameState) MazeFlashWhite() bool {
	if state.Match == nil || state.Phase != PhaseLevelComplete {
		return false
	}
	segmentTicks := LevelClearTicks / (2 * LevelClearFlashes)
	return (state.PhaseTicks/segmentTicks)%2 == 0
}

// BeginPlaying enters active play only when a match is in the ready phase.
// Tick calls it when the intro expires; callers may use it to skip the intro.
func (state *GameState) BeginPlaying() bool {
	if state.Match == nil || state.Phase != PhaseReady {
		return false
	}
	state.enterPhase(PhasePlaying)
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

// ConsumeCollectible awards points for a remaining dot, spawns the bonus when the
// eaten count reaches 70 or 170, and starts the level-clear flash on the final one;
// Pac-Man keeps his current position for that phase and any bonus is removed.
// It rejects inactive play and returns whether a coin or energizer was consumed.
func (state *GameState) ConsumeCollectible(column, row int) bool {
	if state.Match == nil || state.Phase != PhasePlaying {
		return false
	}
	points := state.Match.Level.ConsumeCollectible(column, row)
	if points == 0 {
		return false
	}
	if points == 10 {
		state.Match.PacMan.EatingPauseTicks = pacManCoinPauseTicks
	} else if points == 50 {
		state.Match.PacMan.EatingPauseTicks = pacManEnergizerPauseTicks
	}
	state.AddScore(points)
	eaten := len(CoinCoords) + len(EnergizerCoords) - state.Match.Level.Remaining
	if eaten == FruitFirstDotsEaten || eaten == FruitSecondDotsEaten {
		state.spawnFruit()
	}
	if state.Match.Level.Remaining == 0 {
		state.Match.Fruit = FruitState{}
		state.enterPhase(PhaseLevelComplete)
	}
	return true
}

// spawnFruit shows the level's bonus for a random duration between FruitMinTicks
// and FruitMaxTicks inclusive, discarding any previous score popup.
func (state *GameState) spawnFruit() {
	if state.Random == nil {
		state.Random = rand.New(rand.NewSource(rand.Int63()))
	}
	state.Match.Fruit = FruitState{
		Visible: true,
		Ticks:   FruitMinTicks + state.Random.Intn(FruitMaxTicks-FruitMinTicks+1),
	}
}

// ConsumeFruit awards the level's bonus points when Pac-Man's center overlaps a visible
// bonus during active play, replacing it with a score popup for FruitScoreTicks and
// pausing Pac-Man for one tick. It returns whether the bonus was eaten.
func (state *GameState) ConsumeFruit() bool {
	if state.Match == nil || state.Phase != PhasePlaying {
		return false
	}
	player := &state.Match.PacMan
	if !state.Match.Fruit.Overlaps(player.X, player.Y) {
		return false
	}
	points := state.Match.Level.Parameters.BonusPoints
	state.Match.Fruit = FruitState{ScoreTicks: FruitScoreTicks, Points: points}
	player.EatingPauseTicks = max(player.EatingPauseTicks, pacManFruitPauseTicks)
	state.AddScore(points)
	return true
}

// AdvanceLevel replaces a cleared level, resets Pac-Man to his starting position and
// the bonus, and enters the ready intro, preserving score and lives.
func (state *GameState) AdvanceLevel() bool {
	if state.Match == nil || state.Phase != PhaseLevelComplete {
		return false
	}
	state.Match.Level = NewLevelState(state.Match.Level.Number + 1)
	state.Match.PacMan = NewPacManState()
	state.Match.Fruit = FruitState{}
	state.enterPhase(PhaseReady)
	return true
}

// LoseLife preserves the level, removes any bonus without rescheduling it, and enters
// the ready intro or game over during active play.
func (state *GameState) LoseLife() bool {
	if state.Match == nil || state.Phase != PhasePlaying {
		return false
	}
	state.Match.Lives--
	state.Match.PacMan = NewPacManState()
	state.Match.Fruit = FruitState{}
	if state.Match.Lives == 0 {
		state.enterPhase(PhaseGameOver)
	} else {
		state.enterPhase(PhaseReady)
	}
	return true
}
