# Pacman
A recreation of the arcade game "Pacman" written in GO.

Implemented: the full maze with 244 collectibles, Pac-Man movement with cornering,
the four ghosts with their arcade targeting, scatter/chase schedule, frightened mode
and ghost-house release rules, Cruise Elroy, bonus fruit, the death and level-clear
sequences, pause, and a persistent high score. Sound is not implemented yet.

## 🔧 Powered By

- [Go](https://go.dev/)
- [Ebiten](https://ebitengine.org/)

## 🎮 Run It

If you already have Go installed ([instructions here](https://go.dev/doc/install)), run the game with:

```bash
go run https://github.com/geckoblu-games/pacman-go
```

## Main resources
- the great pacman dossier: (multiple links, because this beautiful article gets nuked frequently)
  - [The Pac-Man Dossier](https://pacman.holenet.info)
  - [The Pac-Man Dossier (Archive)](https://web.archive.org/web/20220118071921/https://www.gamedeveloper.com/design/the-pac-man-dossier)
  - [Pac-Man Ghost AI Explained](https://www.youtube.com/watch?v=ataGotQ7ir8)
- Videos of the game
  - [the game in action](https://www.youtube.com/watch?v=-CbyAk3Sn9I)
  - [longplay until the game glitches out](https://www.youtube.com/watch?v=AuoH0vz3Mqk)
- [Making Games in Go for Absolute Beginners](https://threedots.tech/post/making-games-in-go/)

### Font
- [Press Start 2P Font](https://fonts.google.com/specimen/Press+Start+2P)

## Requirements

- Go 1.22 or later.
- A desktop environment with OpenGL graphics support.
- On Linux: a C compiler and the development libraries required by Ebitengine.

On Debian/Ubuntu, install the native dependencies with:

```sh
sudo apt install build-essential libasound2-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libgl1-mesa-dev libxxf86vm-dev pkg-config
```

For other platforms, see the
[installation guide](https://ebitengine.org/en/documents/install.html).

## Running

From the project directory:

```sh
go mod tidy
go run .
```

The first command downloads dependencies and generates `go.sum`.
A match starts automatically with the `READY!` intro. To exit, close the window.

### Controls

| Key | Action |
| --- | --- |
| Arrow keys or `W` `A` `S` `D` | Steer Pac-Man (held direction is buffered until the turn is possible) |
| `P` | Pause / resume |
| `Enter` | Start a new match after `GAME OVER` |

### Debugging

To keep all ghosts stationary, set `GhostMovementEnabled` to `false` in
`game/constants.go`; their modes, animation, house timers, and collisions remain
active. `DebugGridEnabled` and `DebugAllowedPathEnabled` in the same file draw the
logical grid and the walkable cells.

## Testing

The simulation tests run without a display:

```sh
go test ./highscore
go test game/state.go game/fruit.go game/pacman.go game/ghost.go game/constants.go game/constsantsmaze.go game/level.go game/state_test.go game/level_test.go game/pacman_test.go game/ghost_test.go
```

`go test ./...` also exercises the rendering integration and needs a display
(for example `xvfb-run -a go test ./...` on Linux).

## Persistent high score

The high score is restored on startup and saved at level completion, game over,
and normal window closure, only when the record has increased.
Default file locations:

| Platform | File |
| --- | --- |
| Linux | `$XDG_DATA_HOME/geckoblu/pacman/highscore` if `XDG_DATA_HOME` is absolute, otherwise `~/.local/share/geckoblu/pacman/highscore` |
| Windows | `%LOCALAPPDATA%\geckoblu\pacman\highscore` |
| macOS | `~/Library/Application Support/geckoblu/pacman/highscore` |
| iOS | `$HOME/Library/Application Support/geckoblu/pacman/highscore`, inside the app sandbox |
| Android | An explicit file inside the app's private persistent files directory, supplied by the mobile host |

The directory is created automatically. Read/write errors are logged without
stopping play; an unsaved record is retried at the next checkpoint.
An abrupt termination can lose progress since the last save. Concurrent game
instances do not coordinate writes.

### Mobile module integration

The persistence module supports Android and iOS, but this repository still has
only a desktop entry point: mobile packaging and lifecycle hooks are not included.
On Android, the native host must obtain `Context.getFilesDir()` and pass a file
inside that directory to the store, for example:

```go
store := highscore.New(filepath.Join(appPrivateFilesDir, "highscore"))
```

`DefaultPath` returns an explicit error on Android rather than guessing a sandbox
path or using shared storage. On iOS, `DefaultPath` requires an absolute sandbox
`HOME`; a native host can alternatively supply a file under its Application Support
directory through `New`. Do not use cache or temporary directories for the record.
Mobile hosts must connect loading, checkpoints, and background/lifecycle saves;
mobile process termination does not guarantee a final save callback.

## Building

```sh
go build -o pacman .
./pacman
```
