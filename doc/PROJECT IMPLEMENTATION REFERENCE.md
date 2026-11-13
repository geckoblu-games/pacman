# Project Implementation Reference

## Project Overview

Pacman is a Go desktop application built with Ebitengine v2.8.8. It renders the maze at a fixed logical resolution and owns a renderer-independent game state with match lifecycle, scoring, lives, level progression, collectible tracking and consumption, and Pac-Man movement. Ghosts, player-ghost collisions, and audio are not implemented yet.

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
- `go.mod` declares module `pacman-ai`, Go 1.22.0, and Ebitengine v2.8.8. `go.sum` records dependency checksums.
- `README.md` contains user-facing requirements, run/build instructions, and a short project structure overview.

### `game/`

- `game.go` owns the Ebitengine `Game` implementation and session state. Its `Update` and `Draw` callbacks delegate to its `ScenePlay`, while `Layout` returns the logical dimensions. `Update` observes phase entry before and after the scene update, calling the optional `SaveHighScore` callback at level completion and game over. Checking before the update also handles a checkpoint reached between updates, before input can start a new match or advance a timed phase. Save failures are logged once per checkpoint, not once per tick. `game_test.go` verifies checkpoint persistence and retry timing as well as the AllowedPath matrix, bottom-HUD geometry, bonus sprite loading, shared state, the automatic READY! intro at startup, and delegation of timed level advancement and rendering.
- `sceneplay.go` defines `ScenePlay`, which holds a pointer to `Game.State`, the `assets.GameImages` bundle, the HUD font face, and the Pac-Man sprite. It maps arrow/WASD input to a requested direction, calls `GameState.Tick` every frame to drive the level-clear flash and READY! intro, advances the player during active play, composes the screen in `Draw`, and owns all drawing helpers, including the `READY!` message, debug overlays, and their lazy caches.
- `pacman.go` defines renderer-independent player position, direction, speed conversion, buffered tile turns, immediate reversal on opposite input, early/late diagonal cornering, wall blocking, continuous side-tunnel traversal, and the deterministic mouth-animation cycle. Tunnel coordinates pass beyond one screen edge before wrapping while preserving the remaining movement distance.
- `pacman_test.go` checks level-speed movement, corridor reversal, buffered turns, early and late diagonal cornering, wall blocking, respawn positioning, and the mouth-animation cycle without initializing Ebitengine.
- `spritepacman.go` draws the selected directional atlas frame at the simulation position and repeats it across the opposite screen edge when the sprite is clipped by a tunnel boundary; a stopped Pac-Man keeps the half-open frame for the last direction.
- `state.go` defines the pure Go session and match state and their transitions, using `LevelState`, plus the tick-driven level-clear and READY! timers; `state_test.go` tests them without initializing Ebitengine.
- `level.go` contains level-specific types, data, and behavior: `LevelState`, `LevelParameters`, `NewLevelState`, bonus-symbol history, collectible consumption, the private level parameter table, `ParametersForLevel`, and the base speed. Its code has no rendering dependencies. `LevelState.ConsumeCollectible` updates presence flags and the remaining count and returns 10 or 50 points, or zero when nothing is consumed. Match phase checks, score updates, and the transition to `PhaseLevelComplete` remain in `GameState`.
- `level_test.go` verifies collectible initialization and idempotence and compares all level parameter rows with the dossier.
- `constants.go` defines logical display/grid dimensions, scale and debug switches.
- `constsantsmaze.go` defines `AllowedPath` as a `[GridRows][GridColumns]bool` matrix indexed `[row][column]`, plus the `EnergizerCoords` and `CoinCoords` coordinate arrays.

### `assets/`

- `assets.go` embeds PNG and TTF files at build time and exposes helpers to read bytes, decode and crop PNG images, load the window icon sizes, and construct the `GameImages` bundle.
- `classic-tileset.png` is the sprite atlas used for Pac-Man, energizers, reserve lives, and all eight bonus symbols. `LoadGameImages` crops the first nine 26 x 26 cells as Pac-Man animation frames. Their order is up half/wide, down half/wide, left half/wide, right half/wide, then shared closed-mouth. Bonus sprites are selected by the exact `BonusSymbol` names used in the level table, from row 6, columns 0 through 7: Cherries, Strawberry, Peach, Apple, Grapes, Galaxian, Bell, and Key. All sprites are loaded during initialization, not while drawing.
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
4. `game.New` starts a match in `PhaseReady`, so the READY! intro begins without a key press. `Game.Update` delegates to `ScenePlay.Update`, which first handles Enter (starting a fresh match only when none exists or after game over), then calls `GameState.Tick`. The tick advances the two timed phases: the level-clear flash lasts exactly `LevelClearTicks` (120) and then advances the level; the ready intro lasts `ReadyTicks` (120, approximately 2 seconds) and then unlocks play automatically. Gameplay controls are ignored outside `PhasePlaying`. During active play, arrow keys or WASD move Pac-Man at the configured level speed with buffered cornering. Entering a tile with a remaining coin or energizer removes it, awards its points, and pauses Pac-Man for one or three ticks respectively. There is no death animation or jingle yet; ghosts are not implemented, so nothing disappears at level clear.
5. `Draw` composes layers in this order: optional allowed-path background, permanent maze walls (white during white flash segments, otherwise blue), remaining coins and energizers, Pac-Man's sprite when a match exists, optional dashed grid, the white `1UP` and `HIGH SCORE` labels and their numeric scores, reserve lives, bonus symbols, and finally the yellow `READY!` message during `PhaseReady`. `READY!` is horizontally centered on the screen and vertically centered on zero-based row 20, the corridor below the ghost house. During the flash Pac-Man is drawn where he stopped; during the intro he is drawn at his starting position. Both labels use Press Start 2P at the same top offset; `1UP` begins at column 3 and `HIGH SCORE` at column 9. Scores appear one grid row below the labels. The current score is right-aligned to the right edge of zero-based column 6 (x = 112); the high score is right-aligned to the right edge of column 16 (x = 272). Collectible positions come from `CoinCoords` and `EnergizerCoords`; their visibility comes from the current level state. Debug images are generated lazily and cached. Pac-Man consumes the collectible in his occupied tile during `Update`; drawing remains read-only.
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

`Game.State` owns a `GameState`. Its zero value has `PhaseWaiting` and no match. `HighScore` is retained across matches and is loaded and persisted across application runs by the desktop application. Simulation methods do not perform I/O. `PhaseTicks` counts the 60 Hz ticks elapsed in the current timed phase and is reset by every phase change. `MatchState` contains score, total remaining lives (including the current life), a `LevelState`, and Pac-Man's state. The level contains its one-based number, a `LevelParameters` value, presence flags for all 240 coins and 4 energizers, and their remaining count.

- `StartNewMatch` creates level 1 with score 0 and 3 lives in `PhaseReady`, preserving the session record.
- `Tick` advances the timed phases by one tick. In `PhaseLevelComplete` it calls `AdvanceLevel` once `PhaseTicks` reaches `LevelClearTicks`; in `PhaseReady` it calls `BeginPlaying` once `PhaseTicks` reaches `ReadyTicks`. It does nothing without a match or in other phases.
- `MazeFlashWhite` reports whether the maze is currently white. The 120-tick flash is divided into `2 * LevelClearFlashes` (eight) equal 15-tick segments; even segments are white and odd segments blue, so the flash begins white and ends blue, producing exactly four white flashes.
- `BeginPlaying` transitions from ready to playing. `Tick` invokes it when the intro expires; tests call it directly to skip the intro.
- `AddScore` accepts positive points only during play and updates the session record immediately. Bonus lives and score overflow rules are not implemented.
- `ConsumeCollectible` accepts a tile coordinate during play, removes it exactly once, and awards 10 points for a coin or 50 for an energizer. The final collectible enters `PhaseLevelComplete` at tick 0; Pac-Man's position is left untouched so he is drawn where he stopped. Fruit is not part of level completion. Energizer frightened behavior is not implemented.
- `AdvanceLevel` only accepts a completed level, increments its number, loads the corresponding parameters, restores all collectibles, resets Pac-Man to his starting position, and enters the ready intro, preserving score and lives. Special late-level behavior is not implemented.
- `LoseLife` only accepts active play, decrements lives, resets Pac-Man, and enters the ready intro or `PhaseGameOver` at zero lives. Score and the entire current level, including consumed collectibles, are preserved. Ghost/timer respawn resets will be added with those systems, separately from level initialization.

The game-over result remains available until a new match starts. State methods are independent of assets, rendering, wall-clock time, and Ebitengine. Fields are exposed for rendering; callers should use transition methods to mutate gameplay state.

### Level parameters

`levelParameters` in `game/level.go` encodes all 21 rows of Table A.1 in source-column order. Go arrays cannot be constants, so this package-private table is treated as read-only. `NewLevelState` selects the appropriate row through `ParametersForLevel` and copies it into `LevelState.Parameters`. Levels 21 and above share the final row while retaining their actual level number; level numbers below 1 are invalid and rejected with a panic. New matches and level advancement load parameters automatically; losing a life preserves them.

`LevelParameters` retains every column: bonus symbol and points; Pac-Man normal and dot-eating speeds; ghost normal and tunnel speeds; both Elroy dot thresholds and speeds; frightened Pac-Man, dot-eating and ghost speeds; frightened duration and flash count. Speed fields are integer percentages, not pixel velocities. Dot-eating percentages retain the source's rounded approximations (71, 79, 83, and 87). `BaseSpeedPixelsPerSecond` stores the source's canonical 100% speed, 75.75757625 pixels/second, measured at the original 8-pixel tile scale; future movement must account for the renderer's 16-pixel cells.

Source dashes are encoded as zero for all five frightened fields. This disables vulnerability on levels 17, 19, 20, and 21+, while level 18 retains its one-second window and three flashes, matching the authoritative table rather than the prose summary. The table is loaded now, including unused parameters; actor movement, fruit, frightened behavior, and Elroy logic are not implemented by this change. Frightened durations remain in seconds until the future simulation converts them to fixed ticks.

The lives HUD displays reserve lives (`Lives - 1`, excluding the active life) centered on the penultimate grid row (zero-based row 34) with an additional half-row downward offset (8 logical pixels). Icons are centered on zero-based columns 3 and 5 from the left with a half-column leftward offset (8 logical pixels) for a new three-life match; further reserve lives would continue every two columns. The image is loaded once from `classic-tileset.png`, using the 26 x 26 sprite at row 0, column 4 (source x = 4, y = 0). No icons are drawn without a match or when there are no reserve lives.

`LevelState.BonusSymbols()` derives up to seven symbols in current-to-oldest order from the current level and the level parameter table. No mutable history is needed for sequential progression. The renderer places the current symbol on the far right, mirrored from the leftmost reserve-life icon, with previous levels spaced two columns to the left. For 26 x 26 sprites the current symbol's top-left is (387, 547), while the first reserve life is at (35, 547). Advancing inserts the new symbol on the right and shifts the previous six left; repeated symbols are not deduplicated. Losing lives preserves the strip and a new match resets it to one cherry. The counter is independent of fruit consumption, never exceeds seven icons, and does not emulate the level-256 overflow bug.

The application displays the maze, state-driven collectible sprites, Pac-Man's moving sprite, map-debug overlays, score labels and values, reserve-life icons, and the fruit/level history strip. Pac-Man consumes coins and energizers on entering their occupied tiles, awards the configured points, and pauses for one or three simulation ticks respectively. There is no ghost model, player-ghost collision handling, collectible fruit actor, or audio in the current code.

Natural ownership boundaries for future work are:

- Extend the existing logical state with actors and timers; keep `Update` responsible for advancing it and `Draw` read-only with respect to simulation state. Pac-Man moves through `AllowedPath` cells at the level's configured speed, buffers cardinal turns with arcade cornering, wraps across the side tunnel, consumes dots once, and observes the dossier's eating pauses. Ghosts and player-ghost collision handling are not implemented.
- Expand player controls and render additional actors/pellets in `Draw` with an explicit layer order.
- Use `AllowedPath` as the authoritative, validated movement grid; the boolean matrix is shared by movement checks and the debug overlay.
- Extend the existing HUD labels and use the loaded font when additional score, lives, or level text is implemented.

## Build and Verification

The README documents Go 1.22 or later and Ebitengine's platform requirements. On Linux, a C compiler and native graphics/audio development libraries may be required. Typical commands from the repository root are:

```sh
go run .
go build -o pacman .
```

The state tests cover new-match initialization, session record retention, life loss and game over, collectible idempotence, level completion and progression, and rejected operations outside active play. They also compare all 21 parameter rows directly with the local dossier and verify parameter loading through progression, new-match reset, and the 21+ fallback without mutating the shared table. Bonus-strip tests cover current-first ordering, duplicated symbols, the seven-icon limit, insertion and shifting on advancement, preservation after death, and reset on a new match. The dossier is read only by tests, not by the runtime. State tests can run without a display or Ebitengine initialization:

```sh
go test game/state.go game/constants.go game/constsantsmaze.go game/level.go game/state_test.go game/level_test.go
```

Persistence unit tests also run without a display: `go test ./highscore`. All filesystem tests use temporary directories, not the user's high-score file.

`go test ./...` also tests bottom-HUD symmetry, spacing, screen bounds, and loading of every bonus sprite, and compiles the graphics integration. It requires Ebitengine's native dependencies and a display on Linux (or a virtual display such as `xvfb-run -a go test ./...`). Actual window rendering and actor gameplay are not covered by these tests.
