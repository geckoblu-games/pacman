package game

// BaseSpeedPixelsPerSecond is the dossier's 100% speed at the original 8-pixel tile scale.
const BaseSpeedPixelsPerSecond = 75.75757625

// LevelParameters contains one Table A.1 row; speeds are integer percentages.
type LevelParameters struct {
	// BonusSymbol names the level's fruit or other bonus sprite.
	BonusSymbol string
	// BonusPoints is the score awarded for consuming that bonus.
	BonusPoints int
	// PacManSpeedPercent is Pac-Man's normal movement speed.
	PacManSpeedPercent int
	// PacManDotsSpeedPercent includes the source's dot-eating slowdown.
	PacManDotsSpeedPercent int
	// GhostSpeedPercent is a ghost's normal movement speed.
	GhostSpeedPercent int
	// GhostTunnelSpeedPercent is a ghost's speed inside a side tunnel.
	GhostTunnelSpeedPercent int
	// Elroy1DotsLeft is the remaining-dot threshold for the first Elroy stage.
	Elroy1DotsLeft int
	// Elroy1SpeedPercent is Blinky's speed in the first Elroy stage.
	Elroy1SpeedPercent int
	// Elroy2DotsLeft is the remaining-dot threshold for the second Elroy stage.
	Elroy2DotsLeft int
	// Elroy2SpeedPercent is Blinky's speed in the second Elroy stage.
	Elroy2SpeedPercent int
	// FrightPacManSpeedPercent is Pac-Man's speed while ghosts are frightened.
	FrightPacManSpeedPercent int
	// FrightPacManDotsSpeedPercent includes dot-eating slowdown while ghosts are frightened.
	FrightPacManDotsSpeedPercent int
	// FrightGhostSpeedPercent is the movement speed of frightened ghosts.
	FrightGhostSpeedPercent int
	// FrightTimeSeconds is the vulnerability duration; zero disables it.
	FrightTimeSeconds int
	// FrightFlashes counts the warning flashes at the end of vulnerability.
	FrightFlashes int
}

// LevelState tracks one level's number, copied parameters, and remaining collectibles.
type LevelState struct {
	// Number is the one-based level number, including levels beyond 21.
	Number int
	// Parameters is a value copy of the corresponding level specification.
	Parameters LevelParameters
	// Coins records presence in the same index order as CoinCoords.
	Coins [len(CoinCoords)]bool
	// Energizers records presence in the same index order as EnergizerCoords.
	Energizers [len(EnergizerCoords)]bool
	// Remaining counts all unconsumed coins and energizers.
	Remaining int
}

// BonusSymbols returns up to seven current-to-oldest level symbols without deduplication.
func (state LevelState) BonusSymbols() []string {
	if state.Number < 1 {
		return nil
	}
	symbols := make([]string, min(state.Number, 7))
	symbols[0] = state.Parameters.BonusSymbol
	for index := 1; index < len(symbols); index++ {
		symbols[index] = ParametersForLevel(state.Number - index).BonusSymbol
	}
	return symbols
}

// NewLevelState loads a positive level number and restores all 244 collectibles.
// Invalid numbers panic; levels 21 and above share the final parameter row.
func NewLevelState(number int) LevelState {
	state := LevelState{
		Number:     number,
		Parameters: ParametersForLevel(number),
		Remaining:  len(CoinCoords) + len(EnergizerCoords),
	}
	for index := range state.Coins {
		state.Coins[index] = true
	}
	for index := range state.Energizers {
		state.Energizers[index] = true
	}
	return state
}

// ConsumeCollectible removes an item once and returns 10, 50, or zero points.
// Phase checks and score updates belong to GameState, not the level itself.
func (state *LevelState) ConsumeCollectible(column, row int) int {
	coord := [2]int{column, row}
	for index, position := range CoinCoords {
		if position == coord && state.Coins[index] {
			state.Coins[index] = false
			state.Remaining--
			return 10
		}
	}
	for index, position := range EnergizerCoords {
		if position == coord && state.Energizers[index] {
			state.Energizers[index] = false
			state.Remaining--
			return 50
		}
	}
	return 0
}

// ParametersForLevel copies a Table A.1 row, using the last row for levels 21+.
// A level number below one is invalid and causes a panic.
func ParametersForLevel(number int) LevelParameters {
	if number < 1 {
		panic("level number must be positive")
	}
	if number >= len(levelParameters) {
		return levelParameters[len(levelParameters)-1]
	}
	return levelParameters[number-1]
}

// levelParameters stores all 21 Table A.1 rows in source-column order; treat as read-only.
var levelParameters = [...]LevelParameters{
	{"Cherries", 100, 80, 71, 75, 40, 20, 80, 10, 85, 90, 79, 50, 6, 5},
	{"Strawberry", 300, 90, 79, 85, 45, 30, 90, 15, 95, 95, 83, 55, 5, 5},
	{"Peach", 500, 90, 79, 85, 45, 40, 90, 20, 95, 95, 83, 55, 4, 5},
	{"Peach", 500, 90, 79, 85, 45, 40, 90, 20, 95, 95, 83, 55, 3, 5},
	{"Apple", 700, 100, 87, 95, 50, 40, 100, 20, 105, 100, 87, 60, 2, 5},
	{"Apple", 700, 100, 87, 95, 50, 50, 100, 25, 105, 100, 87, 60, 5, 5},
	{"Grapes", 1000, 100, 87, 95, 50, 50, 100, 25, 105, 100, 87, 60, 2, 5},
	{"Grapes", 1000, 100, 87, 95, 50, 50, 100, 25, 105, 100, 87, 60, 2, 5},
	{"Galaxian", 2000, 100, 87, 95, 50, 60, 100, 30, 105, 100, 87, 60, 1, 3},
	{"Galaxian", 2000, 100, 87, 95, 50, 60, 100, 30, 105, 100, 87, 60, 5, 5},
	{"Bell", 3000, 100, 87, 95, 50, 60, 100, 30, 105, 100, 87, 60, 2, 5},
	{"Bell", 3000, 100, 87, 95, 50, 80, 100, 40, 105, 100, 87, 60, 1, 3},
	{"Key", 5000, 100, 87, 95, 50, 80, 100, 40, 105, 100, 87, 60, 1, 3},
	{"Key", 5000, 100, 87, 95, 50, 80, 100, 40, 105, 100, 87, 60, 3, 5},
	{"Key", 5000, 100, 87, 95, 50, 100, 100, 50, 105, 100, 87, 60, 1, 3},
	{"Key", 5000, 100, 87, 95, 50, 100, 100, 50, 105, 100, 87, 60, 1, 3},
	{"Key", 5000, 100, 87, 95, 50, 100, 100, 50, 105, 0, 0, 0, 0, 0},
	{"Key", 5000, 100, 87, 95, 50, 100, 100, 50, 105, 100, 87, 60, 1, 3},
	{"Key", 5000, 100, 87, 95, 50, 120, 100, 60, 105, 0, 0, 0, 0, 0},
	{"Key", 5000, 100, 87, 95, 50, 120, 100, 60, 105, 0, 0, 0, 0, 0},
	{"Key", 5000, 90, 79, 95, 50, 120, 100, 60, 105, 0, 0, 0, 0, 0},
}
