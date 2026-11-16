package game

const (
	// AttractCycleTicks is the length of the attract presentation; the sequence restarts
	// from the empty screen when the counter reaches it (no demo play is shown).
	AttractCycleTicks = 1170
	// attractLegendTick is the tick at which the point legend and the decoy energizer appear.
	attractLegendTick = 450
	// attractChaseTick is the tick at which the actors enter and the energizers start flashing.
	attractChaseTick = 510
	// attractFlashHalfPeriod is the number of ticks the energizers stay visible, then hidden.
	attractFlashHalfPeriod = 10
	// attractCapturePauseTicks freezes the chase while a capture score is displayed.
	attractCapturePauseTicks = 45
	// attractChaseRow is the zero-based grid row of the chase corridor and the decoy energizer.
	attractChaseRow = 20
	// attractDecoyColumn is the zero-based grid column of the decoy energizer.
	attractDecoyColumn = 3

	// attractSourceScale converts the source's 8-pixel-tile pixel values to logical pixels.
	attractSourceScale = float64(CellSize) / 8
	// attractSpriteHalf is half the width of a 16-pixel source sprite, used to convert the
	// source's top-left sprite positions to centers.
	attractSpriteHalf = 8
	// attractPacManStartX is the source top-left x of Pac-Man when the actors enter.
	attractPacManStartX = 236
	// attractGhostStartX is the source top-left x of Blinky when the actors enter.
	attractGhostStartX = 258
	// attractGhostSpacing is the source distance between consecutive ghosts at entrance.
	attractGhostSpacing = 18
	// attractPacManSpeed is Pac-Man's chase speed in source pixels per tick.
	attractPacManSpeed = 1.35
	// attractGhostSpeed is the ghosts' chase speed in source pixels per tick.
	attractGhostSpeed = 1.42
	// attractFrightenedSpeed is the ghosts' speed after reversing, in source pixels per tick.
	attractFrightenedSpeed = 0.65
)

// attractGhostRevealTicks holds, in cast order, the tick each ghost's sprite appears in the table.
var attractGhostRevealTicks = [4]int{60, 150, 240, 330}

// attractGhostNameTicks holds, in cast order, the tick each ghost's character and nickname appear.
var attractGhostNameTicks = [4]int{120, 210, 300, 390}

// attractGhostCharacters holds, in cast order, the character description shown in the table.
var attractGhostCharacters = [4]string{"- SHADOW", "- SPEEDY", "- BASHFUL", "- POKEY"}

// attractGhostNicknames holds, in cast order, the quoted nickname shown in the table.
var attractGhostNicknames = [4]string{`"BLINKY"`, `"PINKY"`, `"INKY"`, `"CLYDE"`}

// AttractState stores the renderer-independent progress of the attract-mode presentation:
// the cast reveal timeline, the corridor chase and the capture sequence. It is driven by
// Tick at 60 Hz and restarts automatically after AttractCycleTicks.
type AttractState struct {
	// Ticks counts ticks since the presentation started and wraps at AttractCycleTicks.
	Ticks int
	// PacMan stores the chase actor's logical-pixel center, facing and mouth animation.
	PacMan PacManState
	// Ghosts stores the chase ghosts in cast order: Blinky, Pinky, Inky, Clyde.
	Ghosts [4]AttractGhost
	// DecoyEaten reports whether Pac-Man has eaten the decoy energizer and the chase reversed.
	DecoyEaten bool
	// Captured counts ghosts eaten so far during the reversal.
	Captured int
	// PauseTicks counts remaining frozen ticks while a capture score is displayed.
	PauseTicks int
	// CaptureX is the logical-pixel center of the displayed capture score.
	CaptureX float64
}

// AttractGhost stores a chase ghost's identity, horizontal center and capture flag.
type AttractGhost struct {
	// Kind identifies the ghost's sprite set.
	Kind GhostKind
	// X is the ghost's horizontal center in logical pixels.
	X float64
	// Captured reports whether Pac-Man has already eaten this ghost.
	Captured bool
}

// attractChaseY is the logical-pixel vertical center of the chase actors and the decoy.
func attractChaseY() float64 {
	return float64(attractChaseRow*CellSize + CellSize/2)
}

// attractDecoyX is the logical-pixel horizontal center of the decoy energizer.
func attractDecoyX() float64 {
	return float64(attractDecoyColumn*CellSize + CellSize/2)
}

// attractSourceCenter converts a source top-left sprite x to a logical-pixel center.
func attractSourceCenter(sourceX float64) float64 {
	return (sourceX + attractSpriteHalf) * attractSourceScale
}

// newAttractState returns the presentation at tick zero with the actors off-screen right.
func newAttractState() AttractState {
	attract := AttractState{
		PacMan: PacManState{
			X:              attractSourceCenter(attractPacManStartX),
			Y:              attractChaseY(),
			Direction:      DirectionLeft,
			lastDirection:  DirectionLeft,
			animationFrame: pacManHalfMouthFrame(DirectionLeft),
		},
	}
	for index := range attract.Ghosts {
		attract.Ghosts[index] = AttractGhost{
			Kind: GhostKind(index),
			X:    attractSourceCenter(attractGhostStartX + float64(index*attractGhostSpacing)),
		}
	}
	return attract
}

// Tick advances the presentation by one 60 Hz tick: it restarts the cycle at
// AttractCycleTicks, counts down a capture pause, or moves the chase actors once the
// chase has started, eating the decoy and capturing ghosts when Pac-Man reaches them.
func (attract *AttractState) Tick() {
	attract.Ticks++
	if attract.Ticks >= AttractCycleTicks {
		*attract = newAttractState()
		return
	}
	if attract.Ticks < attractChaseTick {
		return
	}
	if attract.PauseTicks > 0 {
		attract.PauseTicks--
		return
	}
	if !attract.DecoyEaten {
		attract.PacMan.X -= attractPacManSpeed * attractSourceScale
		for index := range attract.Ghosts {
			attract.Ghosts[index].X -= attractGhostSpeed * attractSourceScale
		}
		if attract.PacMan.X <= attractDecoyX() {
			attract.DecoyEaten = true
			attract.PacMan.Direction = DirectionRight
			attract.PacMan.lastDirection = DirectionRight
		}
	} else {
		attract.PacMan.X += attractPacManSpeed * attractSourceScale
		for index := range attract.Ghosts {
			attract.Ghosts[index].X += attractFrightenedSpeed * attractSourceScale
		}
		for index := range attract.Ghosts {
			ghost := &attract.Ghosts[index]
			if ghost.Captured || attract.PacMan.X < ghost.X {
				continue
			}
			ghost.Captured = true
			attract.Captured++
			attract.PauseTicks = attractCapturePauseTicks
			attract.CaptureX = ghost.X
			break
		}
	}
	attract.PacMan.animate(true)
}

// GhostRevealed reports whether the cast-order ghost's table sprite has appeared.
func (attract AttractState) GhostRevealed(index int) bool {
	return attract.Ticks >= attractGhostRevealTicks[index]
}

// GhostNamed reports whether the cast-order ghost's character and nickname have appeared.
func (attract AttractState) GhostNamed(index int) bool {
	return attract.Ticks >= attractGhostNameTicks[index]
}

// LegendVisible reports whether the point legend has appeared.
func (attract AttractState) LegendVisible() bool {
	return attract.Ticks >= attractLegendTick
}

// EnergizerVisible reports whether the legend energizer is currently drawn: steady from
// the legend reveal, then flashing 10 ticks on and 10 off once the chase starts.
func (attract AttractState) EnergizerVisible() bool {
	if attract.Ticks < attractLegendTick {
		return false
	}
	if attract.Ticks < attractChaseTick {
		return true
	}
	return ((attract.Ticks-attractChaseTick)/attractFlashHalfPeriod)%2 == 0
}

// CopyrightVisible reports whether the attract copyright is shown, starting with the chase.
func (attract AttractState) CopyrightVisible() bool {
	return attract.ChaseActive()
}

// DecoyVisible reports whether the decoy energizer is drawn: it appears with the chase,
// flashes with the legend energizer and disappears once eaten.
func (attract AttractState) DecoyVisible() bool {
	return attract.ChaseActive() && !attract.DecoyEaten && attract.EnergizerVisible()
}

// ChaseActive reports whether the chase actors are on stage.
func (attract AttractState) ChaseActive() bool {
	return attract.Ticks >= attractChaseTick
}

// PacManVisible reports whether the chase Pac-Man is drawn; he is hidden during a capture pause.
func (attract AttractState) PacManVisible() bool {
	return attract.ChaseActive() && attract.PauseTicks == 0
}

// GhostVisible reports whether the cast-order chase ghost is drawn: on stage and not yet
// eaten. A captured ghost disappears permanently, starting with its capture pause.
func (attract AttractState) GhostVisible(index int) bool {
	return attract.ChaseActive() && !attract.Ghosts[index].Captured
}

// GhostMode returns the chase ghosts' shared mode: scatter before the decoy is eaten,
// frightened afterwards. Frightened attract ghosts never flash white.
func (attract AttractState) GhostMode() GhostMode {
	if attract.DecoyEaten {
		return GhostFrightened
	}
	return GhostScatter
}

// GhostDirection returns the chase ghosts' shared direction of travel.
func (attract AttractState) GhostDirection() Direction {
	if attract.DecoyEaten {
		return DirectionRight
	}
	return DirectionLeft
}

// CapturePoints returns the score displayed during a capture pause (200, 400, 800 or
// 1600 in capture order), or zero when no pause is active.
func (attract AttractState) CapturePoints() int {
	if attract.PauseTicks == 0 || attract.Captured == 0 {
		return 0
	}
	return 200 << (attract.Captured - 1)
}
