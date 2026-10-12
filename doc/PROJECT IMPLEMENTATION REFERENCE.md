# Project Implementation Reference

## Project Overview

Pacman is a Go desktop application built with Ebitengine v2.8.8. The current implementation establishes the application entry point, embeds and loads visual assets, and renders the maze at a fixed logical resolution. The game loop is wired up, but gameplay systems such as player movement, ghosts, collisions, scoring, pellets, and input handling are not implemented yet.

## Project Structure

```text
.
|-- main.go
|-- go.mod
|-- go.sum
|-- README.md
|-- assets/
|   |-- assets.go
|   |-- classic-tileset.png
|   |-- walls.png
|   |-- icon_16.png
|   |-- icon_32.png
|   |-- icon_64.png
|   |-- icon_128.png
|   `-- PressStart2P-Regular.ttf
|-- game/
|   |-- game.go
|   |-- constants.go
|   `-- debug.go
`-- doc/
    |-- PROJECT IMPLEMENTATION REFERENCE.md
    |-- ReferenceMap.png
    |-- TilesWalls_sketch.pdf
    `-- The Pac-Man Dossier.html
```

### Root files

- `main.go` is the executable entry point. It calculates the initial window dimensions from the game's logical dimensions and scale, sets the title and window icons, constructs the game, and starts Ebitengine's event loop. Initialization and run errors are fatal and reported through the standard logger.
- `go.mod` declares module `pacman-ai`, Go 1.22.0, and Ebitengine v2.8.8. `go.sum` records dependency checksums.
- `README.md` contains user-facing requirements, run/build instructions, and a short project structure overview.

### `game/`

- `game.go` owns the Ebitengine `Game` implementation and implements the `Update`, `Draw`, and `Layout` callbacks. It stores an `assets.GameImages` bundle and draws its loaded images.
- `constants.go` defines logical display/grid dimensions, scale and debug switches, plus the reference map's allowed-path cell coordinates.
- `debug.go` creates and draws cached debug overlays for allowed-path cells and the grid.

### `assets/`

- `assets.go` embeds PNG and TTF files at build time and exposes helpers to read bytes, decode and crop PNG images, load the window icon sizes, and construct the `GameImages` bundle.
- `classic-tileset.png` is the sprite atlas used for the energizer image. `LoadGameImages` crops the configured 26 x 26 atlas cell and converts it, along with the wall layer, into Ebitengine images. It also creates one reusable 4 x 4 px coin image in RGB (255, 181, 148). The expected wall dimensions are validated during loading.
- `walls.png` is the transparent maze-wall layer used by the game. It is expected to match the logical screen dimensions exactly. Its documented preparation preserves blue wall pixels from the reference map and restores wall segments obscured by the grid; transparent areas include the ghost-house doorway.
- The four `icon_*.png` files provide window-manager icon choices at multiple sizes.
- `PressStart2P-Regular.ttf` is embedded, but no current code loads or draws the font.

### `doc/`

This directory contains source/reference material: the reference map image, a wall-tile sketch PDF, and a local copy of The Pac-Man Dossier. They are not read by the runtime; the runtime maze is `assets/walls.png` and the path coordinates are compiled into `game/constants.go`.

## Runtime and Rendering Flow

1. `main` derives the initial window size by multiplying `ScreenWidth` and `ScreenHeight` by `ScreenScale` and rounding each dimension to an integer.
2. It configures the title and loads all four window icons through the `assets` package.
3. `game.New` calls `assets.LoadGameImages` with the expected logical screen size. The loader validates `walls.png`, crops the energizer from `classic-tileset.png`, and returns both as Ebitengine images. A missing, undecodable, or incorrectly sized image prevents startup with an error.
4. `ebiten.RunGame` invokes the game callbacks. `Update` currently returns immediately, so there is no simulation state or input processing.
5. `Draw` composes layers in this order: optional allowed-path background, permanent maze walls, coins, energizers, optional dashed grid. Coin positions come from `CoinCoords` and are rendered centered in their grid cells; they are visual-only and are not yet collectible. Debug images are generated lazily by `game` and cached in the `assets.GameImages` bundle. Both debug flags are currently enabled.
6. `Layout` always returns the logical dimensions. Ebitengine scales that fixed-resolution image to the window; resizing does not change the game's grid or logical coordinates.

## Implementation Decisions

### Logical coordinate system

The display is represented as a 28-column by 36-row grid with 16 logical pixels per cell. This gives a fixed internal resolution of 448 x 576 pixels. Grid coordinates are zero-based, with the origin at the top-left. To get a cell's top-left logical pixel position, multiply its column and row by `CellSize`.

`ScreenScale` controls only the initial native window dimensions. It does not alter the internal rendering resolution or the cell-coordinate system. The current scale is 1.0.

### Embedded asset loading

The `assets` package uses `embed.FS` with `//go:embed *.png *.ttf`, so runtime assets are shipped in the executable rather than loaded from external paths. `ReadFile` adds filename context to read errors. `LoadImage` decodes PNG bytes without changing pixel dimensions or colors and adds context to decode errors. `LoadWindowIcons` requests all expected icon sizes in order and fails if any one is unavailable or invalid.

`assets.LoadGameImages` loads and validates the permanent wall layer, crops the energizer sprite from the atlas, and creates the reusable coin image once during construction. `game.Game` retains the returned bundle; no asset file access occurs in the frame loop.

### Debug overlays

`DebugAllowedPathEnabled` and `DebugGridEnabled` control independent overlays and are both true in the current configuration.

- `AllowedPath` contains 300 zero-based `{column, row}` pairs extracted from the reference map. It includes tunnel and energizer cells while excluding walls, HUD regions, and the black ghost-house interior. The array is a variable because Go does not support constant arrays; code treats it as read-only.
- The allowed-path overlay paints each listed 16 x 16 cell in opaque RGB (79, 79, 79). It is created lazily on first use and then cached in `GameImages`. Its cached pixels assume `AllowedPath` remains unchanged.
- The grid overlay is a translucent gray image generated lazily on first use and cached in `GameImages`. Both horizontal and vertical lines use equal dash and gap lengths controlled by `DebugGridDashSize`. The outer right and bottom lines are clamped to the last drawable pixel to stay within image bounds. The dash size must be positive; its current value is one logical pixel.

The overlays are composed around, rather than baked into, the permanent wall image. The path layer is drawn before the walls and the grid after them, allowing the grid to remain visible over the maze.

## Current Scope and Extension Points

The implemented application is currently a rendering foundation: it opens a window and displays the maze, static coin and energizer sprites, and map-debug overlays. The coins are not yet collectible. There is no player/ghost model, tile collision graph, movement or turn logic, input mapping, collectible state, scoring, lives, level state, audio, or HUD rendering in the current code.

Natural ownership boundaries for future work are:

- Add simulation state and update behavior to `game.Game` or to types owned by the `game` package; keep `Update` responsible for advancing it.
- Add input handling in the game update path and render actors/pellets in `Draw` with an explicit layer order.
- Use the grid and `AllowedPath` as reference data, but validate or replace the latter with a gameplay-oriented maze representation before using it as authoritative collision data.
- Load and render the embedded font through a text-rendering dependency when HUD text is introduced; embedding the TTF alone does not make it available on screen.

## Build and Verification

The README documents Go 1.22 or later and Ebitengine's platform requirements. On Linux, a C compiler and native graphics/audio development libraries may be required. Typical commands from the repository root are:

```sh
go run .
go build -o pacman .
```

`go test ./...` currently compiles all packages successfully, but reports that there are no test files. Therefore this check verifies compilation only, not rendering behavior or gameplay correctness.
