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
	// DeathFreezeTicks holds the collision scene for one second.
	DeathFreezeTicks = 60
	// DeathAnimationTicks is the two-second implosion and spark sequence.
	DeathAnimationTicks = 120
	// DeathEmptyTicks holds the empty maze for one second before life loss.
	DeathEmptyTicks = 60
	// DeathFrameTicks is the duration of each of the ten death animation frames.
	DeathFrameTicks = 12
	// GhostCaptureTicks is the one-second score display after eating a ghost.
	GhostCaptureTicks = 60
	// GameOverTicks is the five-second GAME OVER display before the attract screen returns.
	GameOverTicks = 300
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
	// PhaseDying freezes the collision, animates Pac-Man, and holds the empty maze.
	PhaseDying
	// PhaseGhostCaptured freezes play while the captured ghost's score is shown.
	PhaseGhostCaptured
	// PhaseGameOver shows the result of a match with no remaining lives until Tick
	// returns to the attract screen or a new match starts.
	PhaseGameOver
	// PhaseAttract runs the attract-mode presentation without a match until a new match starts.
	PhaseAttract
)

// GameState stores the session record, current phase, and optional active match.
type GameState struct {
	// HighScore is the record, optionally loaded and persisted by the application.
	HighScore int
	// LastScore is the final score of the previous match, shown as 1UP on the attract screen.
	LastScore int
	// Phase controls which match operations are currently valid.
	Phase Phase
	// PhaseTicks counts 60 Hz ticks elapsed in the current timed phase.
	PhaseTicks int
	// Paused freezes Tick and gameplay input until TogglePause clears it.
	Paused bool
	// Match is nil until a match is started and while the attract screen is shown.
	Match *MatchState
	// Attract stores the attract presentation; it is meaningful only in PhaseAttract.
	Attract AttractState
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
	// Blinky stores the red ghost and references the shared mode schedule.
	Blinky GhostState
	// Pinky stores the pink ghost and references the same mode schedule.
	Pinky GhostState
	// Inky stores the cyan ghost and references Blinky and the shared mode schedule.
	Inky GhostState
	// Clyde stores the orange ghost and references the shared mode schedule.
	Clyde GhostState
	// InkyDots is Inky's personal dot count, retained through life loss.
	InkyDots int
	// ClydeDots is Clyde's personal dot count, retained through life loss.
	ClydeDots int
	// GlobalGhostCounter is active after a death until Clyde is waiting at 32 dots.
	GlobalGhostCounter bool
	// GlobalGhostDots counts collectibles eaten since the last death.
	GlobalGhostDots int
	// NoDotTicks counts active ticks since a collectible or anti-stall release.
	NoDotTicks int
	// GhostStreak counts ghosts captured during the current energizer.
	GhostStreak int
	// GhostCapture stores the score popup while a ghost capture is frozen.
	GhostCapture GhostCaptureState
}

// GhostCaptureState identifies the displayed score and the ghost's collision position.
type GhostCaptureState struct {
	// X and Y are the captured ghost's center in logical pixels.
	X, Y float64
	// Points is the score awarded for the capture.
	Points int
	// Kind identifies the ghost hidden during the capture.
	Kind GhostKind
}

// StartNewMatch resets score, lives, and level while retaining the session record.
// The new match starts unpaused in the ready intro, replacing the attract screen or a
// finished match.
func (state *GameState) StartNewMatch() {
	state.Match = &MatchState{Lives: InitialLives, Level: NewLevelState(1), PacMan: NewPacManState(), Blinky: NewBlinky()}
	state.Match.Pinky = newPinky(state.Match.Blinky.ghostClock, false)
	state.Match.Inky = newInky(&state.Match.Blinky)
	state.Match.Clyde = newClyde(state.Match.Blinky.ghostClock)
	state.Match.releasePersonalGhosts()
	state.Paused = false
	state.enterPhase(PhaseReady)
}

// EnterAttract discards the current match, if any, remembering its score as LastScore,
// and starts the attract presentation from its first tick, unpaused.
func (state *GameState) EnterAttract() {
	if state.Match != nil {
		state.LastScore = state.Match.Score
	}
	state.Match = nil
	state.Paused = false
	state.Attract = newAttractState()
	state.enterPhase(PhaseAttract)
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

// Tick advances the attract presentation, the timed phases, or active fruit and ghost
// timers by one 60 Hz tick. Game over returns to the attract screen after GameOverTicks.
// Other phases and a paused match are unaffected.
func (state *GameState) Tick() {
	if state.Phase == PhaseAttract {
		state.PhaseTicks++
		state.Attract.Tick()
		return
	}
	if state.Match == nil || state.Paused {
		return
	}
	switch state.Phase {
	case PhaseGameOver:
		state.PhaseTicks++
		if state.PhaseTicks >= GameOverTicks {
			state.EnterAttract()
		}
	case PhaseGhostCaptured:
		state.PhaseTicks++
		if state.PhaseTicks >= GhostCaptureTicks {
			var ghost *GhostState
			switch state.Match.GhostCapture.Kind {
			case GhostBlinky:
				ghost = &state.Match.Blinky
			case GhostPinky:
				ghost = &state.Match.Pinky
			case GhostInky:
				ghost = &state.Match.Inky
			case GhostClyde:
				ghost = &state.Match.Clyde
			default:
				panic("invalid captured ghost kind")
			}
			ghost.Mode = GhostEaten
			ghost.reversePending = false
			if ghost.houseStage == 3 {
				ghost.houseStage = 2
			}
			state.Match.GhostCapture = GhostCaptureState{}
			state.enterPhase(PhasePlaying)
		}
	case PhaseDying:
		state.PhaseTicks++
		if state.PhaseTicks == DeathFreezeTicks {
			state.Match.Fruit = FruitState{}
		}
		if state.PhaseTicks >= DeathFreezeTicks+DeathAnimationTicks+DeathEmptyTicks {
			state.finishDeath()
		}
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
		frightened := state.Match.Blinky.frightTicks > 0
		changed, ended := state.Match.Blinky.ghostClock.tick(state.Match.Level)
		state.Match.Blinky.applyClock(changed, ended)
		state.Match.Pinky.applyClock(changed, ended)
		state.Match.Inky.applyClock(changed, ended)
		state.Match.Clyde.applyClock(changed, ended)
		if frightened && state.Match.Blinky.frightTicks == 0 {
			state.Match.GhostStreak = 0
		}
		state.Match.NoDotTicks++
		limit := 4 * simulationTicksPerSecond
		if state.Match.Level.Number >= 5 {
			limit = 3 * simulationTicksPerSecond
		}
		if state.Match.NoDotTicks >= limit {
			state.Match.NoDotTicks = 0
			state.Match.releaseWaitingGhost()
		}
	}
}

// DeathFrame returns the zero-based animation frame, or -1 outside the animation.
func (state *GameState) DeathFrame() int {
	if state.Phase != PhaseDying || state.PhaseTicks < DeathFreezeTicks ||
		state.PhaseTicks >= DeathFreezeTicks+DeathAnimationTicks {
		return -1
	}
	return (state.PhaseTicks - DeathFreezeTicks) / DeathFrameTicks
}

// PacManVisible reports whether Pac-Man is visible in the current phase.
func (state *GameState) PacManVisible() bool {
	if state.Match == nil || state.Phase == PhaseGameOver || state.Phase == PhaseGhostCaptured {
		return false
	}
	return state.Phase != PhaseDying || state.PhaseTicks < DeathFreezeTicks+DeathAnimationTicks
}

// GhostsVisible reports whether ghosts are visible in the current phase.
func (state *GameState) GhostsVisible() bool {
	if state.Match == nil || state.Phase == PhaseGameOver || state.Phase == PhaseLevelComplete {
		return false
	}
	return state.Phase != PhaseDying || state.PhaseTicks < DeathFreezeTicks
}

// GhostVisible reports whether a particular ghost is visible in the current phase.
func (state *GameState) GhostVisible(kind GhostKind) bool {
	return state.GhostsVisible() && (state.Phase != PhaseGhostCaptured || state.Match.GhostCapture.Kind != kind)
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
	state.Match.NoDotTicks = 0
	if state.Match.GlobalGhostCounter {
		state.Match.GlobalGhostDots++
		if state.Match.GlobalGhostDots == 7 {
			state.Match.releasePinky()
		} else if state.Match.GlobalGhostDots == 17 && state.Match.Inky.houseStage == 4 {
			state.Match.Inky.startExit()
		} else if state.Match.GlobalGhostDots == 32 && state.Match.Clyde.houseStage == 4 {
			state.Match.GlobalGhostCounter = false
			state.Match.GlobalGhostDots = 0
			state.Match.releasePersonalGhosts()
		}
	} else if state.Match.Pinky.houseStage != 4 {
		if state.Match.Inky.houseStage == 4 {
			state.Match.InkyDots++
		} else if state.Match.Clyde.houseStage == 4 {
			state.Match.ClydeDots++
		}
		state.Match.releasePersonalGhosts()
	}
	if points == 10 {
		state.Match.PacMan.EatingPauseTicks = pacManCoinPauseTicks
	} else if points == 50 {
		state.Match.PacMan.EatingPauseTicks = pacManEnergizerPauseTicks
		state.Match.GhostStreak = 0
		state.Match.Blinky.Frighten(state.Match.Level)
		state.Match.Pinky.Frighten(state.Match.Level)
		state.Match.Inky.Frighten(state.Match.Level)
		state.Match.Clyde.Frighten(state.Match.Level)
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
	state.Match.Blinky = NewBlinky()
	state.Match.Pinky = newPinky(state.Match.Blinky.ghostClock, false)
	state.Match.Inky = newInky(&state.Match.Blinky)
	state.Match.Clyde = newClyde(state.Match.Blinky.ghostClock)
	state.Match.InkyDots = 0
	state.Match.ClydeDots = 0
	state.Match.GlobalGhostCounter = false
	state.Match.GlobalGhostDots = 0
	state.Match.NoDotTicks = 0
	state.Match.GhostStreak = 0
	state.Match.GhostCapture = GhostCaptureState{}
	state.Match.Fruit = FruitState{}
	state.Match.releasePersonalGhosts()
	state.enterPhase(PhaseReady)
	return true
}

// LoseLife begins the timed death sequence during active play without changing the level.
func (state *GameState) LoseLife() bool {
	if state.Match == nil || state.Phase != PhasePlaying {
		return false
	}
	state.enterPhase(PhaseDying)
	return true
}

// finishDeath deducts one life and resets the actors before READY or game over.
func (state *GameState) finishDeath() {
	state.Match.Lives--
	state.Match.PacMan = NewPacManState()
	state.Match.Blinky = NewBlinky()
	state.Match.Blinky.ElroySuspended = true
	state.Match.Pinky = newPinky(state.Match.Blinky.ghostClock, true)
	state.Match.Inky = newInky(&state.Match.Blinky)
	state.Match.Clyde = newClyde(state.Match.Blinky.ghostClock)
	state.Match.GlobalGhostCounter = true
	state.Match.GlobalGhostDots = 0
	state.Match.NoDotTicks = 0
	state.Match.GhostStreak = 0
	state.Match.GhostCapture = GhostCaptureState{}
	state.Match.Fruit = FruitState{}
	if state.Match.Lives == 0 {
		state.enterPhase(PhaseGameOver)
	} else {
		state.enterPhase(PhaseReady)
	}
}
