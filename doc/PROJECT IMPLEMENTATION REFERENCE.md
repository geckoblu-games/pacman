# Project Implementation Reference

## Project Overview

Pacman is a Go desktop application built with Ebitengine v2.8.8. Its renderer-independent game state owns the match lifecycle, scoring, collectibles, the bonus fruit, Pac-Man and Blinky. The other ghosts and audio are not implemented yet.

## Project Structure

```text
.
|-- main.go
|-- go.mod
|-- go.sum
|-- README.md
|-- highscore/
|   |-- highscore.go
|   `-- highscore_test.go
|-- assets/
|   |-- assets.go
|   |-- walls.go
|   |-- walls_test.go
|   |-- classic-tileset.png
|   |-- walls.png
|   |-- icon_16.png
|   |-- icon_32.png
|   |-- icon_64.png
|   |-- icon_128.png
|   `-- PressStart2P-Regular.ttf
|-- game/
|   |-- game.go
|   |-- game_test.go
|   |-- fruit.go
|   |-- ghost.go
|   |-- ghost_test.go
|   |-- pacman.go
|   |-- pacman_test.go
|   |-- sceneplay.go
|   |-- spritepacman.go
|   |-- state.go
|   |-- state_test.go
|   |-- level.go
|   |-- level_test.go
|   |-- constants.go
|   `-- constsantsmaze.go
`-- doc/
    |-- PROJECT IMPLEMENTATION REFERENCE.md
    |-- ReferenceMap.png
    |-- TilesWalls_sketch.pdf
    `-- The Pac-Man Dossier.html
```

### Root files

- `main.go` is the executable entry point. It calculates the initial window dimensions from the game's logical dimensions and scale, sets the title and window icons, constructs the game, restores the high score, and starts Ebitengine's event loop. It supplies the game's save callback and saves again after the event loop returns, including normal window closure and run errors. Asset initialization and run errors are fatal; persistence errors are logged without preventing play or normal exit.
- `go.mod` declares module `pacman`, Go 1.22.0, and Ebitengine v2.8.8. `go.sum` records dependency checksums.
- `README.md` contains user-facing requirements, run/build instructions, and a short project structure overview.

### `game/`

- `game.go` owns the Ebitengine `Game` implementation and session state. Its `Update` and `Draw` callbacks delegate to its `ScenePlay`, while `Layout` returns the logical dimensions. `Update` observes phase entry before and after the scene update, calling the optional `SaveHighScore` callback at level completion and game over. Checking before the update also handles a checkpoint reached between updates, before input can start a new match or advance a timed phase. Save failures are logged once per checkpoint, not once per tick. `game_test.go` verifies checkpoint persistence and retry timing as well as the AllowedPath matrix, bottom-HUD geometry, bonus sprite loading, shared state, the automatic READY! intro at startup, and delegation of timed level advancement and rendering.
- `sceneplay.go` maps input, advances the state and actors while unpaused, and composes the screen. Blinky's normal, frightened and returning-eyes sprites use the simulation state and are duplicated across tunnel edges. It renders the capture score, death sequence and red GAME OVER according to the simulation phase.
- `fruit.go` defines the renderer-independent bonus: `FruitState` (visibility, remaining ticks, score-popup ticks and value), its per-tick timer advance, the half-cell overlap test, the fixed center `FruitX`/`FruitY` (screen center, zero-based row 20), the 70/170 eaten-collectible triggers, the 540–600 tick duration range, the 120-tick popup, and the one-tick eating pause. Spawning, scoring, and phase guards live in `GameState`.
- `ghost.go` owns Blinky's movement, shared direction selection, mode timers, frightened behavior, Elroy, house return, and actor collision updates. `ghost_test.go` exercises these rules headlessly.
- `pacman.go` defines renderer-independent player position, direction, speed conversion, buffered tile turns, immediate reversal on opposite input, early/late diagonal cornering, wall blocking, continuous side-tunnel traversal, the starting row, and the deterministic mouth-animation cycle. Tunnel coordinates pass beyond one screen edge before wrapping while preserving the remaining movement distance.
- `pacman_test.go` checks level-speed movement, corridor reversal, buffered turns, early and late diagonal cornering, wall blocking, respawn positioning, and the mouth-animation cycle without initializing Ebitengine.
- `spritepacman.go` draws the selected directional or death frame at the simulation position and repeats it across the opposite screen edge when clipped. The first death frames rotate the mouth toward up.
- `state.go` defines the pure Go session and match state and their transitions, using `LevelState` and `FruitState`, plus the tick-driven level-clear and READY! timers, bonus spawning and consumption, and the injectable random generator; `state_test.go` tests them without initializing Ebitengine.
- `level.go` contains level-specific types, data, and behavior: `LevelState`, `LevelParameters`, `NewLevelState`, bonus-symbol history, collectible consumption, the private level parameter table, `ParametersForLevel`, and the base speed. Its code has no rendering dependencies. `LevelState.ConsumeCollectible` updates presence flags and the remaining count and returns 10 or 50 points, or zero when nothing is consumed. Match phase checks, score updates, and the transition to `PhaseLevelComplete` remain in `GameState`.
- `level_test.go` verifies collectible initialization and idempotence and compares all level parameter rows with the dossier.
- `constants.go` defines logical display/grid dimensions, scale, debug switches and `GhostMovementEnabled` (default true).
- `constsantsmaze.go` defines `AllowedPath` as a `[GridRows][GridColumns]bool` matrix indexed `[row][column]`, plus the `EnergizerCoords` and `CoinCoords` coordinate arrays.

### `assets/`

- `assets.go` embeds PNG and TTF files at build time and exposes helpers to read bytes, decode and crop PNG images, load the window icon sizes, and construct the `GameImages` bundle.
- `classic-tileset.png` supplies Pac-Man, Blinky, frightened ghosts, returning eyes, energizers, lives, and bonus symbols. Blinky's row contains two frames each for up, down, left and right; the frightened row contains two blue frames, two white frames and four directional eye frames. Nine procedural implosion frames and a spark are generated at initialization; they are not exact arcade sprites.
- `walls.png` is the pixel reference for the transparent maze-wall layer and must match the logical screen dimensions exactly. It contains 19,168 opaque blue pixels (`#2121FF`), 128 opaque pink ghost-house-door pixels (`#FFB8FF`), and 238,752 transparent pixels.
- `walls.go` scans the reference into same-color horizontal runs and draws them once as integer-aligned Ebitengine vector rectangles without antialiasing. `GameImages.Walls` preserves the original pixels; `GameImages.WallsWhite` replaces only blue pixels with opaque white, preserving the pink door and transparency. The renderer draws the white version during the white segments of the level-clear flash and the blue version otherwise. `walls_test.go` reads back both generated images inside an Ebitengine loop and checks every RGBA pixel against the reference and the specified blue-to-white replacement.
- The four `icon_*.png` files provide window-manager icon choices at multiple sizes.
- `PressStart2P-Regular.ttf` is embedded and loaded once for the top-center `HIGH SCORE` label.

### `highscore/`

- `highscore.go` owns platform data-path resolution and filesystem persistence, independently of simulation and rendering. `DefaultPath` uses an absolute `XDG_DATA_HOME` or falls back to `~/.local/share` on Linux, `LOCALAPPDATA` on Windows, and `$HOME/Library/Application Support` on macOS and iOS, followed by `geckoblu/pacman/highscore`. Relative `XDG_DATA_HOME` values are ignored; missing or relative resolved data roots are errors. On iOS, `HOME` must identify the app sandbox; the implementation deliberately avoids `os.UserHomeDir`'s root-directory fallback. Android has no inferred default: the native host must pass a file inside `Context.getFilesDir()` to `New`, which accepts an explicit path on all platforms. Unsupported default-path platforms return an error.
- `Store.Load` reads a nonnegative decimal integer; a missing file means zero, while malformed data and other read errors are reported. `Store.Save` only writes a record higher than the last successful load or save. It creates parent directories with mode `0755` and writes the integer and newline to a unique temporary file with mode `0600` (Unix permissions). The temporary file is synced and closed before `os.Rename` replaces the destination in the same directory. On Unix this replacement is atomic and the directory is then synced. On Windows, Go's replacement operation is used, but no Unix atomicity guarantee is asserted; directory sync is omitted because directory handles opened by `os.Open` do not support it. Temporary files are cleaned up on failure. Failed saves remain eligible for retry. The store assumes a single application instance; concurrent processes do not coordinate updates.
- `highscore_test.go` tests platform path selection, invalid roots, Android's explicit-private-path requirement, first-run behavior, restart loading, exact file contents, unchanged-record suppression, invalid data, failed-save retries, and temporary-file cleanup.

The persistence module can be used by Android and iOS hosts, but the application entry point remains desktop-only. There is no mobile packaging or native directory/lifecycle bridge in this repository. A mobile host must load the record, connect checkpoint saves, and save on appropriate lifecycle events; mobile termination does not guarantee that the desktop post-`RunGame` save will run.

### `doc/`

This directory contains source/reference material: the reference map image, a wall-tile sketch PDF, and a local copy of The Pac-Man Dossier. They are not read by the runtime; the runtime maze is `assets/walls.png` and the path coordinates are compiled into `game/constsantsmaze.go`.

## Runtime and Rendering Flow

1. `main` derives the initial window size by multiplying `ScreenWidth` and `ScreenHeight` by `ScreenScale` and rounding each dimension to an integer.
2. It configures the title and loads all four window icons through the `assets` package.
3. `game.New` creates `ScenePlay` with a pointer to its session state. The scene calls `assets.LoadGameImages` with the expected logical screen size and loads the HUD font. The image loader validates `walls.png`, generates pixel-perfect blue and white wall images from rectangles, and crops sprites from `classic-tileset.png`. A missing, undecodable, or incorrectly sized image prevents startup with an error.
4. `game.New` starts in `PhaseReady`. `ScenePlay.Update` handles Enter and P, advances phase, fruit and ghost timers with `Tick`, then updates both actors during play. Pac-Man consumes collectibles and fruit before Blinky moves; a single tile-occupancy collision check follows both movements. A frightened collision starts a score pause; a fatal collision starts the timed death sequence; the final collectible starts the level flash immediately. On leaving the score pause, actors appear at their capture positions before moving on the next tick. There is no audio yet.
5. `Draw` composes the maze, collectibles, fruit, Pac-Man, Blinky, overlays and HUD. The death sequence shows the frozen scene, then Pac-Man's generated implosion and spark, then an empty maze; after the last life it displays red `GAME OVER`. Blinky is also hidden during level clear and is drawn as eyes when eaten. Drawing does not mutate simulation state.
6. `Layout` always returns the logical dimensions. Ebitengine scales that fixed-resolution image to the window; resizing does not change the game's grid or logical coordinates.
7. Persistence is configured by `main`, not by `game.New`, so constructing a game in tests never accesses user data. The loaded record is assigned before entering the event loop. Entering `PhaseLevelComplete` saves the final collectible's score immediately, before the flash advances; entering `PhaseGameOver` is another checkpoint. After `RunGame` returns, `main` attempts a final save even during an unfinished match. Abrupt process termination is not a normal window closure and may lose points since the last successful checkpoint.

## Implementation Decisions

### Logical coordinate system

The display is represented as a 28-column by 36-row grid with 16 logical pixels per cell. This gives a fixed internal resolution of 448 x 576 pixels. Grid coordinates are zero-based, with the origin at the top-left. To get a cell's top-left logical pixel position, multiply its column and row by `CellSize`.

`ScreenScale` controls only the initial native window dimensions. It does not alter the internal rendering resolution or the cell-coordinate system. The current scale is 1.0.

### Embedded asset loading

The `assets` package uses `embed.FS` with `//go:embed *.png *.ttf`, so runtime assets are shipped in the executable rather than loaded from external paths. `ReadFile` adds filename context to read errors. `LoadImage` decodes PNG bytes without changing pixel dimensions or colors and adds context to decode errors. `LoadWindowIcons` requests all expected icon sizes in order and fails if any one is unavailable or invalid.

`assets.LoadGameImages` loads and validates the wall reference, generates both reusable wall variants, crops the energizer sprite from the atlas, and creates the reusable coin image once during construction. `assets.LoadTextFace` parses the embedded Press Start 2P font when `game.New` initializes its scene. `ScenePlay` retains the image bundle and font face; no asset file access, wall-image generation, or font parsing occurs in the frame loop.

### Debug overlays

`DebugAllowedPathEnabled` and `DebugGridEnabled` control independent overlays and are both true in the current configuration.

- `AllowedPath` is a 36-row by 28-column boolean matrix extracted from the reference map, with `true` for each of 300 traversable cells. It includes tunnel and energizer cells while excluding walls, HUD regions, and the black ghost-house interior. Movement checks use direct `[row][column]` indexing with bounds validation; the matrix is treated as read-only.
- The allowed-path overlay paints each `true` matrix cell as a 16 x 16 block in opaque RGB (79, 79, 79). It is created lazily on first use and then cached in `GameImages`. Its cached pixels assume `AllowedPath` remains unchanged.
- The grid overlay is a translucent gray image generated lazily on first use and cached in `GameImages`. Both horizontal and vertical lines use equal dash and gap lengths controlled by `DebugGridDashSize`. The outer right and bottom lines are clamped to the last drawable pixel to stay within image bounds. The dash size must be positive; its current value is one logical pixel.

The overlays are composed around, rather than baked into, the permanent wall image. The path layer is drawn before the walls and the grid after them, allowing the grid to remain visible over the maze.

## Current Scope and Extension Points

### Game state and lifecycle

`Game.State` owns a `GameState`. Its zero value has `PhaseWaiting` and no match. `HighScore` is retained across matches and is loaded and persisted across application runs by the desktop application. Simulation methods do not perform I/O. `PhaseTicks` counts the 60 Hz ticks elapsed in the current timed phase and is reset by every phase change. `Paused` freezes `Tick` and actor updates. `Random` controls fruit durations; Blinky has a separate deterministic generator reset on life loss or level start. `MatchState` contains score, lives, level, Pac-Man, Blinky and fruit.

- `StartNewMatch` creates level 1 with score 0 and 3 lives in `PhaseReady`, unpaused, preserving the session record.
- `TogglePause` flips `Paused` during a match that is not over and returns the new value. Without a match or after game over it clears the flag and returns false.
- `Tick` advances the timed phases by one tick. During death and ghost capture, fruit and ghost timers remain frozen. In `PhasePlaying` it advances them even when ghost movement is disabled. It does nothing without a match, while paused, or in untimed phases.
- `MazeFlashWhite` reports whether the maze is currently white. The 120-tick flash is divided into `2 * LevelClearFlashes` (eight) equal 15-tick segments; even segments are white and odd segments blue, so the flash begins white and ends blue, producing exactly four white flashes.
- `BeginPlaying` transitions from ready to playing. `Tick` invokes it when the intro expires; tests call it directly to skip the intro.
- `AddScore` accepts positive points only during play and updates the session record immediately. Bonus lives and score overflow rules are not implemented.
- `ConsumeCollectible` removes a dot once, scores it, sets Pac-Man's eating pause, activates frightened behavior on energizers, triggers fruit at 70 and 170 dots and starts level clear on the final dot.
- `ConsumeFruit` accepts only active play and awards `BonusPoints` when Pac-Man's center is within `CellSize/2` of `FruitX`/`FruitY` on both axes while the bonus is visible. It replaces the bonus with a popup (`ScoreTicks` = `FruitScoreTicks`, 120; `Points` = the awarded value), and sets Pac-Man's eating pause to at least one tick. A bonus that is not eaten disappears when its random `Ticks` (540–600 inclusive) run out.
- `AdvanceLevel` resets Pac-Man, Blinky and their timers, restores collectibles and enters READY while preserving score and lives.
- `LoseLife` enters `PhaseDying` without moving actors or deducting a life. The first 60 ticks preserve the collision pose and fruit; the next 120 hide ghosts and fruit and display ten 12-tick animation frames (including the spark); the last 60 show no actors. At tick 240 one life is deducted, actors and ghost timers reset, and READY or game over begins. Score and collectibles are preserved. Elroy is suspended until the next level because Clyde is not implemented.
- `PhaseGhostCaptured` lasts 60 ticks. Pac-Man and Blinky are hidden and the score is drawn at Blinky's collision position. All movement and fruit/ghost timers are frozen, including the frightened timer. The score is awarded immediately and doubles from 200 through 1,600 based on `GhostStreak`, which resets on another energizer, frightened expiry, life loss or a new level. When the pause ends Blinky becomes returning eyes. Only Blinky is currently implemented, so only the first award is normally reachable; there is no extra 12,000-point bonus.

The game-over result remains available until a new match starts. State methods are independent of assets, rendering, wall-clock time, and Ebitengine. Fields are exposed for rendering; callers should use transition methods to mutate gameplay state.

### Level parameters

`levelParameters` in `game/level.go` encodes all 21 rows of Table A.1 in source-column order. Go arrays cannot be constants, so this package-private table is treated as read-only. `NewLevelState` selects the appropriate row through `ParametersForLevel` and copies it into `LevelState.Parameters`. Levels 21 and above share the final row while retaining their actual level number; level numbers below 1 are invalid and rejected with a panic. New matches and level advancement load parameters automatically; losing a life preserves them.

`LevelParameters` retains every column: bonus symbol and points; Pac-Man normal and dot-eating speeds; ghost normal and tunnel speeds; both Elroy dot thresholds and speeds; frightened Pac-Man, dot-eating and ghost speeds; frightened duration and flash count. Speed fields are integer percentages, not pixel velocities. Dot-eating percentages retain the source's rounded approximations (71, 79, 83, and 87). `BaseSpeedPixelsPerSecond` stores the source's canonical 100% speed, 75.75757625 pixels/second, measured at the original 8-pixel tile scale; future movement must account for the renderer's 16-pixel cells.

Source dashes are encoded as zero for all five frightened fields. This disables vulnerability on levels 17, 19, 20, and 21+, while level 18 retains its one-second window and three flashes. The simulation converts speeds and frightened durations to 60 Hz values.

The lives HUD displays reserve lives (`Lives - 1`, excluding the active life) centered on the penultimate grid row (zero-based row 34) with an additional half-row downward offset (8 logical pixels). Icons are centered on zero-based columns 3 and 5 from the left with a half-column leftward offset (8 logical pixels) for a new three-life match; further reserve lives would continue every two columns. The image is loaded once from `classic-tileset.png`, using the 26 x 26 sprite at row 0, column 4 (source x = 4, y = 0). No icons are drawn without a match or when there are no reserve lives.

`LevelState.BonusSymbols()` derives up to seven symbols in current-to-oldest order from the current level and the level parameter table. No mutable history is needed for sequential progression. The renderer places the current symbol on the far right, mirrored from the leftmost reserve-life icon, with previous levels spaced two columns to the left. For 26 x 26 sprites the current symbol's top-left is (387, 547), while the first reserve life is at (35, 547). Advancing inserts the new symbol on the right and shifts the previous six left; repeated symbols are not deduplicated. Losing lives preserves the strip and a new match resets it to one cherry. The counter is independent of fruit consumption, never exceeds seven icons, and does not emulate the level-256 overflow bug.

The application displays Blinky alongside Pac-Man. Chase targets Pac-Man's tile; scatter targets the upper-right corner, except during Elroy. The shared path selector looks one tile ahead, excludes reversal and restricted upturns, and breaks ties up, left, down, right. A forced reversal replans the return turn for the preceding tile rather than retaining the obsolete forward decision. Frightened turns use a seeded generator; the schedule pauses during fright. Energizers signal reversal even when the level has zero fright duration. Collision uses one discrete tile check after both actors move, preserving tile-swap pass-through. Eating Blinky shows his score for one second before his eyes return home; he revives and exits. Returning eyes navigate toward tile (13, 14) at 200% base speed on every level, without tunnel slowdown. At the original 8-pixel scale this is 151.5151525 pixels/second; the game's 16-pixel tiles double the logical pixel distance while preserving tiles per second. The house traversal remains simplified.

`GhostMovementEnabled = false` freezes Blinky's position, including returning eyes, but leaves mode timers, animation, energizers and collisions active. Pause freezes everything. Tunnel and frightened speeds and Elroy thresholds use the level table. After a death Elroy is suspended for the rest of the level until Clyde is implemented, per the design decision.

Natural ownership boundaries for future work are:

- Extend the existing shared direction selection and ghost state for Pinky, Inky and Clyde, including their targeting and house-release rules; preserve deterministic simulation and read-only drawing.
- Expand player controls and render additional actors/pellets in `Draw` with an explicit layer order.
- Use `AllowedPath` as the authoritative, validated movement grid; the boolean matrix is shared by movement checks and the debug overlay.
- Extend the existing HUD labels and use the loaded font when additional score, lives, or level text is implemented.

## Build and Verification

The README documents Go 1.22 or later and Ebitengine's platform requirements. On Linux, a C compiler and native graphics/audio development libraries may be required. Typical commands from the repository root are:

```sh
go run .
go build -o pacman .
```

The state tests cover new-match initialization, session record retention, life loss and game over, pause toggling and freezing, collectible idempotence, level completion and progression, and rejected operations outside active play. They also compare all 21 parameter rows directly with the local dossier and verify parameter loading through progression, new-match reset, and the 21+ fallback without mutating the shared table. Bonus-strip tests cover current-first ordering, duplicated symbols, the seven-icon limit, insertion and shifting on advancement, preservation after death, and reset on a new match. Bonus-fruit tests use seeded generators to cover the 70/170 triggers, the 540–600 tick range and its randomness, expiry, the half-cell overlap rule, level-dependent scoring, the popup duration, the one-tick pause, and removal on life loss, level clear, and advancement. The dossier is read only by tests, not by the runtime. State tests can run without a display or Ebitengine initialization:

```sh
go test game/state.go game/fruit.go game/pacman.go game/constants.go game/constsantsmaze.go game/level.go game/state_test.go game/level_test.go
```

Persistence unit tests also run without a display: `go test ./highscore`. All filesystem tests use temporary directories, not the user's high-score file.

`go test ./...` also tests bottom-HUD symmetry, spacing, screen bounds, and loading of every bonus sprite, and compiles the graphics integration. It requires Ebitengine's native dependencies and a display on Linux (or a virtual display such as `xvfb-run -a go test ./...`). Actual window rendering and actor gameplay are not covered by these tests.
