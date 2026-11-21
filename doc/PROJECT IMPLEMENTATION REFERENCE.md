# Project Implementation Reference

## Project Overview

Pacman is a Go desktop application built with Ebitengine v2. Its renderer-independent game state owns the attract loop (a controls legend and the attract-mode presentation), the match lifecycle, scoring, collectibles, the bonus fruit, Pac-Man and all four ghosts, and records one-shot sound cues; an Ebitengine audio player in the `game` package presents those cues and the background loops derived from the state.

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
|   |-- sounds.go
|   |-- sounds_test.go
|   |-- walls.go
|   |-- walls_test.go
|   |-- classic-tileset.png
|   |-- walls.png
|   |-- icon_16.png
|   |-- icon_32.png
|   |-- icon_64.png
|   |-- icon_128.png
|   |-- PressStart2P-Regular.ttf
|   `-- sounds/
|       |-- start.ogg
|       |-- waka.ogg
|       |-- fruit_eaten.ogg
|       |-- ghost_eaten.ogg
|       |-- dying.ogg
|       |-- frightened.ogg
|       |-- eaten.ogg
|       |-- siren.ogg
|       `-- high_score.ogg
|-- game/
|   |-- game.go
|   |-- game_test.go
|   |-- attract.go
|   |-- attract_test.go
|   |-- audio.go
|   |-- audio_test.go
|   |-- controls.go
|   |-- controls_test.go
|   |-- sound.go
|   |-- fruit.go
|   |-- ghost.go
|   |-- ghost_test.go
|   |-- pacman.go
|   |-- pacman_test.go
|   |-- sceneattract.go
|   |-- scenecontrols.go
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
    |-- Pacman_Clone_Design_Dossier.md
    |-- Pac-Man Attract Mode (Introductory Screen).md
    |-- Pac-Man Attract Mode (Introductory Screen).png
    |-- go-ebiten-implementation-proposal.md
    |-- ReferenceMap.png
    |-- TilesWalls_sketch.pdf
    |-- The Pac-Man Dossier.html
    |-- The Pac-Man Dossier.md
    `-- tpmd_images/
```

### Root files

- `main.go` is the executable entry point. It calculates the initial window dimensions from the game's logical dimensions and scale, sets the title and window icons, constructs the game, restores the high score, and starts Ebitengine's event loop. It supplies the game's save callback and saves again after the event loop returns, including normal window closure and run errors. Asset initialization and run errors are fatal; persistence errors are logged without preventing play or normal exit.
- `go.mod` declares module `pacman`, the Go toolchain version and Ebitengine v2; the audio packages pull in Ebitengine's `oto` output library and the `oggvorbis` decoder as indirect dependencies. `go.sum` records dependency checksums.
- `README.md` contains user-facing requirements, run/build instructions, controls, and asset attributions (font and sound sources).

### `game/`

- `game.go` owns the Ebitengine `Game` implementation and session state. Its `Update` and `Draw` callbacks delegate to its `ScenePlay`, while `Layout` returns the logical dimensions. `New` starts on the attract loop. `Update` observes phase entry before and after the scene update, calling the optional `SaveHighScore` callback at level completion and game over. Checking before the update also handles a checkpoint reached between updates, before input can start a new match or advance a timed phase. Save failures are logged once per checkpoint, not once per tick. `game_test.go` verifies checkpoint persistence and retry timing as well as the AllowedPath matrix, bottom-HUD geometry, bonus sprite loading, shared state, the attract loop at launch (presentation, then controls scene, then presentation again), the attract cast icons' alignment, read-only drawing of both attract scenes, the READY! intro of a started match, and delegation of timed level advancement and rendering.
- `attract.go` defines the renderer-independent attract presentation: `AttractState` with its tick counter, the chase Pac-Man (a `PacManState` reusing the mouth animation), the four `AttractGhost` entries, the decoy and capture progress; the reveal timeline tables; the source-pixel geometry and speeds converted to logical pixels; `Tick`, which stops at `AttractCycleTicks`, where `Finished` reports true; and the visibility, mode and capture-score queries used by the renderer. `attract_test.go` exercises the timeline, chase, reversal, capture sequence and completion headlessly, plus the attract-loop scene sequencing, `EnterAttract` and the timed return from game over.
- `controls.go` defines the renderer-independent controls scene of the attract loop, a project addition with no arcade counterpart: `ControlsEntries` (six key/description pairs in display order), `ControlsState` with its tick counter, `Tick`, which stops at `ControlsCycleTicks` (780), `Finished`, the `KeyRevealed`/`KeyDescribed` queries whose cadence matches the cast table's, and the flashing start prompt's visibility. `controls_test.go` checks the reveal timeline, its equality with the cast timing, the prompt's flashing and duration, and the stop.
- `sceneattract.go` draws the attract screen from `AttractState` and the session scores: top HUD, cast table, legend with a half-size `PTS` suffix, copyright and credit lines, decoy energizer, chase actors and capture scores. It reuses the ghost frame selection and Pac-Man sprite of the play scene without tunnel wrapping.
- `scenecontrols.go` draws the controls scene from `ControlsState` and the session scores: the same top HUD and `CREDIT  0` line as the presentation, the centered `CONTROLS` heading, each yellow key label and white description as they are revealed, and a centered flashing yellow start prompt.
- `sceneplay.go` maps input, advances the state and actors while unpaused, hands the state to the sound player every frame, and composes the screen. In `PhaseAttract` its `Draw` delegates to `scenecontrols.go` or `sceneattract.go` according to `AttractScene`. `Enter` replaces the arcade coin insert and starts a new match from the attract screen or `GAME OVER`; by project decision it does not play a coin-insert effect. `M` toggles mute. The top HUD shows `1UP`, `HIGH SCORE` and `2UP`; scores use the arcade's two-digit minimum, so zero reads `00`. Ghosts' normal, frightened and returning-eyes sprites use the simulation state and are duplicated across tunnel edges. It renders the capture score, death sequence and red GAME OVER according to the simulation phase.
- `fruit.go` defines the renderer-independent bonus: `FruitState` (visibility, remaining ticks, score-popup ticks and value), its per-tick timer advance, the half-cell overlap test, the fixed center `FruitX`/`FruitY` (screen center, zero-based row 20), the 70/170 eaten-collectible triggers, the 540–600 tick duration range, the 120-tick popup, and the one-tick eating pause. Spawning, scoring, and phase guards live in `GameState`.
- `ghost.go` owns ghost movement, shared direction selection and mode clock, targeting, frightened behavior, Elroy, house traversal, and actor collision updates. `ghost_test.go` exercises these rules headlessly.
- `pacman.go` defines renderer-independent player position, direction, speed conversion, buffered tile turns, immediate reversal on opposite input, early/late diagonal cornering, wall blocking, continuous side-tunnel traversal, the starting row, and the deterministic mouth-animation cycle. Tunnel coordinates pass beyond one screen edge before wrapping while preserving the remaining movement distance.
- `pacman_test.go` checks level-speed movement, corridor reversal, buffered turns, early and late diagonal cornering, wall blocking, respawn positioning, and the mouth-animation cycle without initializing Ebitengine.
- `spritepacman.go` draws the selected directional or death frame at the simulation position and repeats it across the opposite screen edge when clipped; `DrawUnwrapped` omits the wrapped copy for the attract chase. The first death frames rotate the mouth toward up.
- `state.go` defines the pure Go session and match state and their transitions, using `LevelState`, `FruitState`, `ControlsState` and `AttractState`, plus the attract-loop scene sequencing (`AttractScene`), the tick-driven attract, level-clear, READY! (opening and later) and game-over timers, bonus spawning and consumption, the injectable random generator, and the sound cues recorded at scoring and death moments; `state_test.go` tests them without initializing Ebitengine.
- `sound.go` defines the renderer-independent `SoundCue` values (start jingle, waka, fruit eaten, ghost eaten, dying), the private recorder used by the simulation, and `TakeSoundCues`, which drains the pending cues for the presenter.
- `audio.go` owns the Ebitengine playback: `soundPlayer` decodes the sounds once per process, creates one restartable player per cue and one infinite-loop player per background track, plays drained cues (queuing one waka while another plays so bites never overlap), suspends and resumes playback across pause, mutes on request, and switches the background loop chosen by `selectBackgroundLoop` from the state: nothing outside unpaused play, the retreating track while any ghost's eyes return, the frightened track while ghosts are frightened, otherwise one of four siren pitches rising with each eaten quarter of the level's collectibles. `audio_test.go` checks the loop selection headlessly.
- `level.go` contains level-specific types, data, and behavior: `LevelState`, `LevelParameters`, `NewLevelState`, bonus-symbol history, collectible consumption, the private level parameter table, `ParametersForLevel`, and the base speed. Its code has no rendering dependencies. `LevelState.ConsumeCollectible` updates presence flags and the remaining count and returns 10 or 50 points, or zero when nothing is consumed. Match phase checks, score updates, and the transition to `PhaseLevelComplete` remain in `GameState`.
- `level_test.go` verifies collectible initialization and idempotence and compares all level parameter rows with the dossier.
- `constants.go` defines logical display/grid dimensions, scale, debug switches (both off by default) and `GhostMovementEnabled` (default true).
- `constsantsmaze.go` defines `AllowedPath` as a `[GridRows][GridColumns]bool` matrix indexed `[row][column]`, plus the `EnergizerCoords` and `CoinCoords` coordinate arrays.

### `assets/`

- `assets.go` embeds PNG, TTF and Ogg files at build time and exposes helpers to read bytes, decode and crop PNG images, load the window icon sizes, and construct the `GameImages` bundle.
- `sounds.go` decodes the embedded Ogg Vorbis files from `sounds/` into 16-bit stereo PCM at `SoundSampleRate` (48 kHz), optionally time-compressing them through Ebitengine's resampler to raise their pitch, trims leading and trailing silence frame by frame, and builds the `GameSounds` bundle, including the four siren variants generated from the single `siren.ogg` at speeds 1.0, 1.05, 1.1 and 1.15. `sounds_test.go` checks decoding, the shortening siren variants and the trimming. `high_score.ogg` is embedded but unused: the record-beaten jingle of the source project is a non-arcade addition the project chose not to play.
- `classic-tileset.png` supplies Pac-Man, all four colored ghosts, frightened ghosts, returning eyes, energizers, lives, and bonus symbols. Each ghost's row contains two frames each for up, down, left and right; the frightened row contains two blue frames, two white frames and four directional eye frames. Nine procedural implosion frames and a spark are generated at initialization; they are not exact arcade sprites.
- `walls.png` is the pixel reference for the transparent maze-wall layer and must match the logical screen dimensions exactly. It contains 19,168 opaque blue pixels (`#2121FF`), 128 opaque pink ghost-house-door pixels (`#FFB8FF`), and 238,752 transparent pixels.
- `walls.go` scans the reference into same-color horizontal runs and draws them once as integer-aligned Ebitengine vector rectangles without antialiasing. `GameImages.Walls` preserves the original pixels; `GameImages.WallsWhite` replaces only blue pixels with opaque white, preserving the pink door and transparency. The renderer draws the white version during the white segments of the level-clear flash and the blue version otherwise. `walls_test.go` reads back both generated images inside an Ebitengine loop and checks every RGBA pixel against the reference and the specified blue-to-white replacement.
- The four `icon_*.png` files provide window-manager icon choices at multiple sizes.
- `PressStart2P-Regular.ttf` is embedded and loaded twice: at the HUD size for labels, scores and messages, and at half size for the attract legend's `PTS` suffix.

### `highscore/`

- `highscore.go` owns platform data-path resolution and filesystem persistence, independently of simulation and rendering. `DefaultPath` uses an absolute `XDG_DATA_HOME` or falls back to `~/.local/share` on Linux, `LOCALAPPDATA` on Windows, and `$HOME/Library/Application Support` on macOS and iOS, followed by `geckoblu/pacman/highscore`. Relative `XDG_DATA_HOME` values are ignored; missing or relative resolved data roots are errors. On iOS, `HOME` must identify the app sandbox; the implementation deliberately avoids `os.UserHomeDir`'s root-directory fallback. Android has no inferred default: the native host must pass a file inside `Context.getFilesDir()` to `New`, which accepts an explicit path on all platforms. Unsupported default-path platforms return an error.
- `Store.Load` reads a nonnegative decimal integer; a missing file means zero, while malformed data and other read errors are reported. `Store.Save` only writes a record higher than the last successful load or save. It creates parent directories with mode `0755` and writes the integer and newline to a unique temporary file with mode `0600` (Unix permissions). The temporary file is synced and closed before `os.Rename` replaces the destination in the same directory. On Unix this replacement is atomic and the directory is then synced. On Windows, Go's replacement operation is used, but no Unix atomicity guarantee is asserted; directory sync is omitted because directory handles opened by `os.Open` do not support it. Temporary files are cleaned up on failure. Failed saves remain eligible for retry. The store assumes a single application instance; concurrent processes do not coordinate updates.
- `highscore_test.go` tests platform path selection, invalid roots, Android's explicit-private-path requirement, first-run behavior, restart loading, exact file contents, unchanged-record suppression, invalid data, failed-save retries, and temporary-file cleanup.

The persistence module can be used by Android and iOS hosts, but the application entry point remains desktop-only. There is no mobile packaging or native directory/lifecycle bridge in this repository. A mobile host must load the record, connect checkpoint saves, and save on appropriate lifecycle events; mobile termination does not guarantee that the desktop post-`RunGame` save will run.

### `doc/`

This directory contains documentation and reference material. `Pacman_Clone_Design_Dossier.md` is the gameplay requirements baseline with FACT/ASSUMPTION/DECISION labels; `The Pac-Man Dossier.md` (with its `tpmd_images/` and the original HTML) is the local copy of the source article, read by `level_test.go` to validate the parameter table; `Pac-Man Attract Mode (Introductory Screen).md` is the source specification of the attract screen's layout and timeline, in the arcade's 8-pixel tile coordinates; its companion PNG is the visual reference for cast-table alignment; `go-ebiten-implementation-proposal.md` is a non-binding architecture proposal. The reference map image and the wall-tile sketch PDF were used to derive the maze. None of these files is read by the runtime; the runtime maze is `assets/walls.png` and the path coordinates are compiled into `game/constsantsmaze.go`.

## Runtime and Rendering Flow

1. `main` derives the initial window size by multiplying `ScreenWidth` and `ScreenHeight` by `ScreenScale` and rounding each dimension to an integer.
2. It configures the title and loads all four window icons through the `assets` package.
3. `game.New` creates `ScenePlay` with a pointer to its session state. The scene calls `assets.LoadGameImages` with the expected logical screen size, loads the HUD and small fonts, and builds the sound player, which decodes all sounds and creates the process-wide 48 kHz audio context (or reuses an existing one). The image loader validates `walls.png`, generates pixel-perfect blue and white wall images from rectangles, and crops sprites from `classic-tileset.png`. A missing, undecodable, or incorrectly sized image prevents startup with an error.
4. `game.New` starts in `PhaseAttract` with no match. `ScenePlay.Update` handles Enter, P and M, then, unless paused, calls `Tick`: in the attract phase this advances the current attract-loop scene and switches to the next one when it finishes; otherwise it advances phase, fruit and ghost timers, then actors are updated during play. Pac-Man consumes collectibles and fruit before the ghosts move; discrete tile-occupancy collision checks follow their movements, in Blinky, Pinky, Inky, Clyde order. A frightened collision starts a score pause; a fatal collision starts the timed death sequence; the final collectible starts the level flash immediately. On leaving the score pause, actors appear at their capture positions before moving on the next tick. `GAME OVER` returns to the attract screen after `GameOverTicks`. Finally the sound player drains the cues recorded during the frame, plays them, suspends or resumes playback according to `Paused`, and selects the background loop from the resulting state.
5. `Draw` renders the current attract-loop scene (controls legend or presentation) in `PhaseAttract`; otherwise it composes the maze, collectibles, fruit, Pac-Man, all four ghosts, overlays and HUD. The death sequence shows the frozen scene, then Pac-Man's generated implosion and spark, then an empty maze; after the last life it displays red `GAME OVER`. Ghosts are also hidden during level clear and drawn as eyes when eaten. Drawing does not mutate simulation state.
6. `Layout` always returns the logical dimensions. Ebitengine scales that fixed-resolution image to the window; resizing does not change the game's grid or logical coordinates.
7. Persistence is configured by `main`, not by `game.New`, so constructing a game in tests never accesses user data. The loaded record is assigned before entering the event loop. Entering `PhaseLevelComplete` saves the final collectible's score immediately, before the flash advances; entering `PhaseGameOver` is another checkpoint. After `RunGame` returns, `main` attempts a final save even during an unfinished match. Abrupt process termination is not a normal window closure and may lose points since the last successful checkpoint.

## Implementation Decisions

### Logical coordinate system

The display is represented as a 28-column by 36-row grid with 16 logical pixels per cell. This gives a fixed internal resolution of 448 x 576 pixels. Grid coordinates are zero-based, with the origin at the top-left. To get a cell's top-left logical pixel position, multiply its column and row by `CellSize`.

`ScreenScale` controls only the initial native window dimensions. It does not alter the internal rendering resolution or the cell-coordinate system. The current scale is 1.2, giving an initial 538 x 691 window; there is no runtime scale or fullscreen option.

### Embedded asset loading

The `assets` package uses `embed.FS` with `//go:embed *.png *.ttf sounds/*.ogg`, so runtime assets are shipped in the executable rather than loaded from external paths. `ReadFile` adds filename context to read errors. `LoadImage` decodes PNG bytes without changing pixel dimensions or colors and adds context to decode errors. `LoadWindowIcons` requests all expected icon sizes in order and fails if any one is unavailable or invalid.

`assets.LoadGameImages` loads and validates the wall reference, generates both reusable wall variants, crops the energizer sprite from the atlas, and creates the reusable coin image once during construction. `assets.LoadTextFace` parses the embedded Press Start 2P font when `game.New` initializes its scene. `assets.LoadGameSounds` decodes every sound into memory when the sound player is built, and the `game` package caches that result for the process. `ScenePlay` retains the image bundle, font faces and sound player; no asset file access, wall-image generation, font parsing or audio decoding occurs in the frame loop.

### Audio

Sound is a presentation concern: the simulation records `SoundCue` values when a collectible, the fruit or a frightened ghost is eaten, when the death animation starts (after the 60-tick collision freeze) and when a match starts, and `soundPlayer.Update` drains them once per frame. Background loops are not cues; they are a pure function of the current state (`selectBackgroundLoop`), so the simulation never tracks what is playing. All sounds come from the Warhorst/pacman project (see the README): there are a single waka bite, a single siren, a frightened loop, a retreating loop, and the one-shot start, fruit, ghost and death effects. The siren pitch levels are generated at load time by resampling because only one siren recording is available. By project decision, entering a match does not play a coin-insert sound; the opening jingle begins at once. The attract screen, the level-clear flash, the capture freeze and `GAME OVER` have no background loop; the capture freeze is covered by the ghost-eaten effect and the death animation by the dying effect, whose 2-second length matches `DeathAnimationTicks`. Pausing with `P` pauses every playing sound and resumes it in place; `M` sets every player's volume to zero or back to one without stopping playback. `high_score.ogg` is deliberately not played.

### Debug overlays

`DebugAllowedPathEnabled` and `DebugGridEnabled` control independent overlays and are both false in the current configuration; they are compile-time constants in `game/constants.go`.

- `AllowedPath` is a 36-row by 28-column boolean matrix extracted from the reference map, with `true` for each of 300 traversable cells. It includes tunnel and energizer cells while excluding walls, HUD regions, and the black ghost-house interior. Movement checks use direct `[row][column]` indexing with bounds validation; the matrix is treated as read-only.
- The allowed-path overlay paints each `true` matrix cell as a 16 x 16 block in opaque RGB (79, 79, 79). It is created lazily on first use and then cached in `GameImages`. Its cached pixels assume `AllowedPath` remains unchanged.
- The grid overlay is a translucent gray image generated lazily on first use and cached in `GameImages`. Both horizontal and vertical lines use equal dash and gap lengths controlled by `DebugGridDashSize`. The outer right and bottom lines are clamped to the last drawable pixel to stay within image bounds. The dash size must be positive; its current value is one logical pixel.

The overlays are composed around, rather than baked into, the permanent wall image. The path layer is drawn before the walls and the grid after them, allowing the grid to remain visible over the maze.

## Current Scope and Extension Points

### Game state and lifecycle

`Game.State` owns a `GameState`. Its zero value has `PhaseWaiting` and no match; `game.New` immediately calls `EnterAttract`. `HighScore` is retained across matches and is loaded and persisted across application runs by the desktop application. `LastScore` remembers the previous match's final score for the attract screen's `1UP` field. Simulation methods do not perform I/O. `PhaseTicks` counts the 60 Hz ticks elapsed in the current timed phase and is reset by every phase change. `Paused` freezes `Tick` and actor updates during a match. `Random` controls fruit durations; ghosts share a separate deterministic generator reset on life loss or level start. `MatchState` contains score, lives, level, Pac-Man, all four ghosts and fruit. `AttractScene`, `Controls` and `Attract` hold the attract loop described below.

- `EnterAttract` discards any match, storing its score in `LastScore`, clears pause, and enters `PhaseAttract` on the loop's opening scene, `AttractScenePresentation`, with fresh `AttractState` and `ControlsState`. `Tick` advances the current scene every tick regardless of `Paused` and, when it reports `Finished`, resets and enters the next one: the presentation is followed by `AttractSceneControls`, which loops back to the presentation. `TogglePause` is rejected without a match.
- `StartNewMatch` creates level 1 with score 0 and 3 lives in `PhaseReady`, unpaused, preserving the session record, with `Opening` set so the intro lasts `MatchStartTicks` (264, 4.4 seconds) and the start jingle cue recorded. It is the only way out of the attract loop, from either scene.
- `TogglePause` flips `Paused` during a match that is not over and returns the new value. Without a match or after game over it clears the flag and returns false.
- `Tick` advances the timed phases by one tick. During death and ghost capture, fruit and ghost timers remain frozen. In `PhasePlaying` it advances them even when ghost movement is disabled. In `PhaseGameOver` it counts `GameOverTicks` (300, five seconds) and then calls `EnterAttract`. It does nothing without a match, while paused, or in untimed phases.
- `MazeFlashWhite` reports whether the maze is currently white. The 120-tick flash is divided into `2 * LevelClearFlashes` (eight) equal 15-tick segments; even segments are white and odd segments blue, so the flash begins white and ends blue, producing exactly four white flashes.
- `BeginPlaying` transitions from ready to playing and clears `Opening`. `Tick` invokes it when the intro expires after `ReadyDuration` ticks (`MatchStartTicks` for the opening intro, `ReadyTicks`, 120, for intros after a lost life or a cleared level); tests call it directly to skip the intro.
- `TakeSoundCues` returns and clears the cues recorded since the previous call; `StartNewMatch` and `EnterAttract` discard pending cues first so a new match or the silent attract screen never replays stale effects.
- `AddScore` accepts positive points only during play and updates the session record immediately. Bonus lives and score overflow rules are not implemented.
- `ConsumeCollectible` removes a dot once, scores it, sets Pac-Man's eating pause, activates frightened behavior on energizers, triggers fruit at 70 and 170 dots and starts level clear on the final dot.
- `ConsumeFruit` accepts only active play and awards `BonusPoints` when Pac-Man's center is within `CellSize/2` of `FruitX`/`FruitY` on both axes while the bonus is visible. It replaces the bonus with a popup (`ScoreTicks` = `FruitScoreTicks`, 120; `Points` = the awarded value), and sets Pac-Man's eating pause to at least one tick. A bonus that is not eaten disappears when its random `Ticks` (540–600 inclusive) run out.
- `AdvanceLevel` resets Pac-Man, all ghosts and their timers, restores collectibles and enters READY while preserving score and lives.
- `LoseLife` enters `PhaseDying` without moving actors or deducting a life. The first 60 ticks preserve the collision pose and fruit; the next 120 hide ghosts and fruit and display ten 12-tick animation frames (including the spark); the last 60 show no actors. At tick 240 one life is deducted, actors and ghost timers reset, and READY or game over begins. Score and collectibles are preserved. Elroy is suspended until Clyde starts exiting the house.
- `PhaseGhostCaptured` lasts 60 ticks. Pac-Man and the captured ghost are hidden, the other ghosts remain visible, and the score is drawn at the captured ghost's collision position. All movement and fruit/ghost timers are frozen, including the frightened timer. The score is awarded immediately and doubles from 200 through 1,600 based on `GhostStreak`, which resets on another energizer, frightened expiry, life loss or a new level. When the pause ends the captured ghost becomes returning eyes. There is no extra 12,000-point bonus.

The game-over result remains visible for `GameOverTicks` or until a new match starts. State methods are independent of assets, rendering, wall-clock time, and Ebitengine. Fields are exposed for rendering; callers should use transition methods to mutate gameplay state.

### Attract loop

In `PhaseAttract` the game cycles through two scenes of equal standing, both silent, both ended by `Enter`: the arcade presentation (`AttractCycleTicks`, 1,170 ticks) and then the controls legend (`ControlsCycleTicks`, 780 ticks), after which the presentation starts again. The controls scene is a project addition with no arcade counterpart. Its screen shows, from tick 0, the presentation's top HUD (`1UP` with `LastScore`, `HIGH SCORE`, `2UP`), the `CREDIT  0` line and a centered white `CONTROLS` heading on row 5. The six `ControlsEntries` (`↑ / W`, `↓ / S`, `← / A`, `→ / D`, `P`, `M` with `MOVE UP`, `MOVE DOWN`, `MOVE LEFT`, `MOVE RIGHT`, `PAUSE`, `MUTE`) occupy rows 9, 12, 15, 18, 21 and 24, key labels in Pac-Man yellow at column 6 and descriptions in white at column 13. They are revealed with the cast table's cadence: key labels at 60, 150, 240, 330, 420 and 510 and descriptions 60 ticks later (120 … 570). `PRESS ENTER TO START` is centered in yellow on row 28, four rows below the last entry, matching the gap between the heading and first entry. It flashes from tick 600 for three seconds, alternating 30 ticks visible and 30 hidden. The arrow glyphs come from the embedded Press Start 2P font.

The presentation follows `doc/Pac-Man Attract Mode (Introductory Screen).md`, whose coordinates use the arcade's 8-pixel tiles and 16-pixel sprites. `attract.go` keeps the source values (`attractPacManStartX = 236`, speeds 1.35/1.42/0.65 px per frame, 18-pixel ghost spacing) and converts them with `attractSourceScale = CellSize / 8`. The source's sprite positions are read as top-left corners of 16-pixel sprites and converted to centers (`attractSourceCenter`), so the chase actors enter at logical x 488 (Pac-Man) and 532, 568, 604, 640 (ghosts). The chase runs along the center of zero-based row 20, the row of the decoy energizer in tile (3, 20); this keeps Pac-Man's center on the decoy's center rather than reproducing the source's 4-pixel vertical offset.

Timeline (ticks at 60 Hz): the HUD, `CHARACTER / NICKNAME` and `CREDIT  0` are drawn from tick 0; ghost sprites appear at 60, 150, 240, 330 and their character/nickname texts at 120, 210, 300, 390; the point legend (`10 PTS`, `50 PTS`) and its energizer appear at 450; from 510 the copyright and corridor decoy energizer appear, both energizers flash 10 ticks on, 10 off, and the actors enter moving left. Pac-Man eats the decoy when his center reaches the decoy's center (tick 670 with these speeds); the decoy disappears, Pac-Man and the ghosts reverse to the right, and the ghosts become blue and slow. Each time Pac-Man's center reaches an uncaptured ghost's center, that ghost is captured: everything freezes for 45 ticks while Pac-Man and the ghost are hidden and 200, 400, 800 or 1600 is drawn in light blue at the ghost's position; the ghost never reappears. Capture ticks are emergent from the speeds (686, 757, 827 and 898), not the source's approximate 720/810/900/990. Pac-Man then leaves the screen to the right (tick 1001) and the static screen remains until tick 1170, when the presentation is finished and the loop continues with the controls scene. There is no demo play, no `PUSH START BUTTON` screen and no two-player option; the attract screen is silent like the arcade's.

Rendering choices in `sceneattract.go`: the top HUD matches the play screen (`1UP` at column 3, `HIGH SCORE` at 9, `2UP` at 22, scores right-aligned at columns 6 and 15) and `1UP` shows `LastScore` (`00` at launch). Table ghosts are drawn facing right with the normal animation. Their visible centers (logical x 80 and y `row*CellSize+6` for text rows 7, 10, 13, 16) and the text's visual top at `row*CellSize` align each icon with its character and nickname in the companion PNG reference. Character descriptions start at column 7 with a space after the hyphen; quoted nicknames start at column 18, and both are drawn in the ghost's color (red, `#FFB8FF`, cyan, `#FFB852`). The legend reuses the in-game 2-pixel coin and the energizer sprite centered on column 10 of rows 24 and 26, values at column 12, and `PTS` at column 15 in an 8-pixel face aligned to the value's baseline. The copyright `© 1980 MIDWAY MFG.CO.` is pink at (4, 31) and `CREDIT  0` is white at (2, 35); there is no coin counter. Chase sprites are not duplicated across screen edges. `Enter` ends the attract screen by starting a match without a coin-insert sound, by project decision.

### Level parameters

`levelParameters` in `game/level.go` encodes all 21 rows of Table A.1 in source-column order. Go arrays cannot be constants, so this package-private table is treated as read-only. `NewLevelState` selects the appropriate row through `ParametersForLevel` and copies it into `LevelState.Parameters`. Levels 21 and above share the final row while retaining their actual level number; level numbers below 1 are invalid and rejected with a panic. New matches and level advancement load parameters automatically; losing a life preserves them.

`LevelParameters` retains every column: bonus symbol and points; Pac-Man normal and dot-eating speeds; ghost normal and tunnel speeds; both Elroy dot thresholds and speeds; frightened Pac-Man, dot-eating and ghost speeds; frightened duration and flash count. Speed fields are integer percentages, not pixel velocities. Dot-eating percentages retain the source's rounded approximations (71, 79, 83, and 87), but `PacManDotsSpeedPercent` and `FrightPacManDotsSpeedPercent` are not used by the simulation: the dot-eating slowdown is modeled exactly as the source describes it, through the one-tick (dot, fruit) and three-tick (energizer) `EatingPauseTicks` stall, so these averaged columns are kept only for table completeness. `BaseSpeedPixelsPerSecond` stores the source's canonical 100% speed, 75.75757625 pixels/second, measured at the original 8-pixel tile scale; movement converts it to the renderer's 16-pixel cells.

Source dashes are encoded as zero for all five frightened fields. This disables vulnerability on levels 17, 19, 20, and 21+, while level 18 retains its one-second window and three flashes. The simulation converts speeds and frightened durations to 60 Hz values.

The lives HUD displays reserve lives (`Lives - 1`, excluding the active life) centered on the penultimate grid row (zero-based row 34) with an additional half-row downward offset (8 logical pixels). Icons are centered on zero-based columns 3 and 5 from the left with a half-column leftward offset (8 logical pixels) for a new three-life match; further reserve lives would continue every two columns. The image is loaded once from `classic-tileset.png`, using the 26 x 26 sprite at row 0, column 4 (source x = 4, y = 0). No icons are drawn without a match or when there are no reserve lives.

`LevelState.BonusSymbols()` derives up to seven symbols in current-to-oldest order from the current level and the level parameter table. No mutable history is needed for sequential progression. The renderer places the current symbol on the far right, mirrored from the leftmost reserve-life icon, with previous levels spaced two columns to the left. For 26 x 26 sprites the current symbol's top-left is (387, 547), while the first reserve life is at (35, 547). Advancing inserts the new symbol on the right and shifts the previous six left; repeated symbols are not deduplicated. Losing lives preserves the strip and a new match resets it to one cherry. The counter is independent of fruit consumption, never exceeds seven icons, and does not emulate the level-256 overflow bug.

The application displays all four ghosts alongside Pac-Man. Blinky's chase targets Pac-Man's tile; scatter targets the upper-right corner, except during Elroy. Pinky's chase targets four tiles ahead, including the arcade's four-tile leftward offset when Pac-Man faces up; scatter targets the upper-left corner. Inky's chase doubles the vector from Blinky's current tile to a point two tiles ahead of Pac-Man, including the two-tile leftward offset when Pac-Man faces up; scatter targets tile (27, 35) in the bottom-right corner. Clyde targets Pac-Man at distances of at least eight tiles and his scatter target (0, 35) when closer; the comparison uses squared Euclidean tile distance. The shared path selector looks one tile ahead, excludes reversal and restricted upturns, and breaks ties up, left, down, right. A forced reversal replans the return turn for the preceding tile rather than retaining the obsolete forward decision. Frightened turns use a shared seeded generator; the single scatter/chase schedule pauses during fright. Energizers signal reversal even when the level has zero fright duration. Collision uses discrete tile checks after actors move, preserving tile-swap pass-through. A captured ghost's eyes navigate toward tile (13, 14) at 200% base speed on every level, without tunnel slowdown, then revive in the house. Inky returns to his original left slot and Clyde to his right slot; their existing house-release rules then determine when they leave. At the original 8-pixel scale the eyes' speed is 151.5151525 pixels/second; the game's 16-pixel tiles double the logical pixel distance while preserving tiles per second.

Pinky starts at the center of the house, Inky two tiles left and Clyde two tiles right; Inky and Clyde align with the center before exiting. After returning as eyes, Inky and Clyde revive in their original left and right slots, then align with the center when released. While waiting, ghosts bounce vertically between logical y positions `17*CellSize` and `18*CellSize`, at a fixed 40% of base speed regardless of level or mode; the speed and bounds are approximations rather than verified arcade-exact values. Release begins immediately from the current height. Personal counters prioritize Pinky, then Inky, then Clyde. Pinky's limit is always zero; Inky's is 30 dots on level 1 and zero thereafter; Clyde's is 60 on level 1, 50 on level 2 and zero thereafter. Counts are retained when a ghost leaves and resume if it returns before reaching its limit. After a death, personal counts are retained but disabled; a global counter releases Pinky at exactly seven and Inky at exactly 17 consumed dots or energizers. At exactly 32, if Clyde is still waiting, the global counter resets and personal counts resume; Clyde leaves only when his personal limit and priority permit. If Clyde already started exiting, the global counter remains active indefinitely, preserving the arcade exploit. The no-dot timer forces the highest-priority waiting ghost out after four seconds, reduced to three from level 5; consuming a collectible resets it. A new level resets all counts. Blinky always leaves immediately after revival. A ghost turns right on exit if a schedule change occurred while inside; otherwise it turns left. An exiting ghost becomes collidable on reaching a maze tile; if captured there, its eyes descend directly into the house. The original exit-capture wall-traversal bug is not reproduced.

`GhostMovementEnabled = false` freezes all ghosts' positions, including returning eyes, but leaves mode and house timers, animation, energizers and collisions active. Pause freezes everything. Tunnel and frightened speeds and Elroy thresholds use the level table. After a death Elroy is suspended until Clyde starts exiting, whether released by a personal counter or the no-dot timer.

### Controls

`ScenePlay.Update` reads the keyboard directly. Arrow keys or WASD set Pac-Man's requested direction while held; `P` toggles pause during a match; `M` toggles mute; `Enter` starts a new match from either attract scene or from `GAME OVER`, replacing the arcade's coin insert and start button. The application launches on the attract presentation; the controls scene that follows it lists these keys. There is no gamepad or touch input.

### Not implemented

The following are known gaps between the current code and the arcade or the design dossier; none is partially present in the code:

- Audio gaps: no record-beaten jingle, no intermission music, and a single siren recording pitched up by resampling instead of the arcade's distinct siren recordings. The coin-insert effect is intentionally omitted by project decision.
- Attract demo play (Phase D of the attract specification) and the `PUSH START BUTTON` / player-count screen: the attract loop alternates the presentation and the controls scene instead, and `Enter` starts the single-player match directly. `PhaseWaiting` exists only as the zero value before `game.New` enters the attract screen.
- No score-awarded extra lives (project decision; see §5.6 of the design dossier); score cap or overflow behavior is also not implemented.
- Arcade intermissions (cutscenes), the level-256 overflow, and the hard-difficulty jumper.
- Exit-capture wall-traversal bug: not reproduced.
- Configurable window scale or fullscreen; mobile packaging and lifecycle hosts (see `highscore/` above).

Natural ownership boundaries for this work are: new scenes or phases belong in `state.go` (phase and transitions), with dedicated simulation files such as `attract.go` or `controls.go` when they carry their own state, and in `sceneplay.go` or a companion drawing file such as `sceneattract.go` or `scenecontrols.go` (input and drawing); new one-shot sounds are `SoundCue` values recorded by the simulation and mapped to a player in `audio.go`, while new background tracks extend `selectBackgroundLoop`; and new per-level rules belong in the `level.go` parameter table.

## Build and Verification

The README documents Go 1.22 or later and Ebitengine's platform requirements. On Linux, a C compiler and native graphics/audio development libraries may be required. Typical commands from the repository root are:

```sh
go run .
go build -o pacman .
```

The state tests cover new-match initialization, session record retention, life loss and game over, pause toggling and freezing, collectible idempotence, level completion and progression, the opening and later intro durations, the recorded sound cues and their draining, and rejected operations outside active play. They also compare all 21 parameter rows directly with the local dossier and verify parameter loading through progression, new-match reset, and the 21+ fallback without mutating the shared table. Attract tests cover the reveal timeline, energizer flashing, entrance positions, the leftward chase and reversal at the decoy, the ordered captures with doubling scores and frozen 45-tick pauses, the presentation's completion, the attract-loop scene sequencing, `EnterAttract`, and the timed return from game over. Controls-scene tests cover the key and description reveal ticks, their match with the cast cadence, the flashing prompt and the stop at `ControlsCycleTicks`. Bonus-strip tests cover current-first ordering, duplicated symbols, the seven-icon limit, insertion and shifting on advancement, preservation after death, and reset on a new match. Bonus-fruit tests use seeded generators to cover the 70/170 triggers, the 540–600 tick range and its randomness, expiry, the half-cell overlap rule, level-dependent scoring, the popup duration, the one-tick pause, and removal on life loss, level clear, and advancement. Pac-Man and ghost tests cover the movement, targeting, mode schedule, fright, red-zone, tunnel, pass-through, capture, eyes, Elroy and ghost-house rules listed in the design dossier's §8 checklist. The dossier is read only by tests, not by the runtime. Simulation tests can run without a display or Ebitengine initialization by naming the non-rendering files (`game.go`, `sceneplay.go`, `sceneattract.go`, `scenecontrols.go`, `spritepacman.go` and `audio.go` import Ebitengine and are excluded):

```sh
go test game/state.go game/sound.go game/fruit.go game/pacman.go game/ghost.go game/attract.go game/controls.go game/constants.go game/constsantsmaze.go game/level.go game/state_test.go game/level_test.go game/pacman_test.go game/ghost_test.go game/attract_test.go game/controls_test.go
```

Persistence unit tests also run without a display: `go test ./highscore`. All filesystem tests use temporary directories, not the user's high-score file.

`go test ./...` also tests bottom-HUD symmetry, spacing, screen bounds, loading of every bonus sprite, sound decoding and silence trimming, background-loop selection, the attract loop at launch and read-only drawing of both attract scenes, and compiles the graphics and audio integration (the audio device itself is only opened by Ebitengine's running game loop, so these tests do not need a sound card). It requires Ebitengine's native dependencies and a display on Linux (or a virtual display such as `xvfb-run -a go test ./...`). Actual window rendering is not covered by these tests. There is no continuous-integration workflow in the repository; tests are run locally.
