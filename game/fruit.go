package game

import "math"

const (
	// FruitFirstDotsEaten is the eaten collectible count that spawns the first bonus of a level.
	FruitFirstDotsEaten = 70
	// FruitSecondDotsEaten is the eaten collectible count that spawns the second bonus of a level.
	FruitSecondDotsEaten = 170
	// FruitMinTicks is the shortest randomized visibility of a bonus (9 seconds at 60 Hz).
	FruitMinTicks = 9 * simulationTicksPerSecond
	// FruitMaxTicks is the longest randomized visibility of a bonus (10 seconds at 60 Hz).
	FruitMaxTicks = 10 * simulationTicksPerSecond
	// FruitScoreTicks is how long the awarded value replaces an eaten bonus (2 seconds at 60 Hz).
	FruitScoreTicks = 2 * simulationTicksPerSecond
	// fruitRow is the zero-based maze row of the corridor below the ghost house where the bonus sits.
	fruitRow = 20
	// pacManFruitPauseTicks is the movement pause after eating a bonus.
	pacManFruitPauseTicks = 1
)

// FruitX is the bonus's horizontal center in logical pixels, between columns 13 and 14.
const FruitX = float64(ScreenWidth) / 2

// FruitY is the bonus's vertical center in logical pixels on fruitRow.
const FruitY = float64(fruitRow*CellSize + CellSize/2)

// FruitState tracks the level's visible bonus and the value shown after it is eaten.
// The zero value has neither a bonus nor a score popup.
type FruitState struct {
	// Visible reports whether a bonus is currently on the maze.
	Visible bool
	// Ticks counts the remaining visibility ticks while Visible is true.
	Ticks int
	// ScoreTicks counts the remaining ticks of the score popup; zero hides it.
	ScoreTicks int
	// Points is the value displayed by the score popup.
	Points int
}

// Tick advances visibility and popup timers by one 60 Hz tick during active play.
// An expired popup also clears Points, returning the state to its zero value.
func (fruit *FruitState) Tick() {
	if fruit.Visible {
		fruit.Ticks--
		if fruit.Ticks <= 0 {
			fruit.Visible = false
			fruit.Ticks = 0
		}
	}
	if fruit.ScoreTicks > 0 {
		fruit.ScoreTicks--
		if fruit.ScoreTicks == 0 {
			fruit.Points = 0
		}
	}
}

// Overlaps reports whether a logical-pixel center lies within half a cell of the bonus center.
func (fruit FruitState) Overlaps(x, y float64) bool {
	return fruit.Visible &&
		math.Abs(x-FruitX) <= float64(CellSize)/2 &&
		math.Abs(y-FruitY) <= float64(CellSize)/2
}
