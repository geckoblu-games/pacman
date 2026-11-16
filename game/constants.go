package game

const (
	// GridColumns counts logical columns across the maze and reserved HUD areas.
	GridColumns = 28
	// GridRows counts logical rows including the top and bottom HUD areas.
	GridRows = 36
	// CellSize is measured in logical pixels and matches the reference map's grid.
	CellSize = 16

	// ScreenWidth is the internal render width in logical pixels, before window scaling.
	ScreenWidth = GridColumns * CellSize
	// ScreenHeight is the internal render height in logical pixels, before window scaling.
	ScreenHeight = GridRows * CellSize

	// ScreenScale controls the initial window size and supports positive fractional values.
	ScreenScale float64 = 1.2

	// DebugGridEnabled toggles the grid independently of the gray path backgrounds.
	DebugGridEnabled = false
	// DebugGridDashSize defines equal dash and gap lengths; it must be greater than zero.
	DebugGridDashSize = 1

	// DebugAllowedPathEnabled toggles the reference map's gray path backgrounds.
	DebugAllowedPathEnabled = false
)

// GhostMovementEnabled controls ghost movement without pausing their other behavior.
var GhostMovementEnabled = true
