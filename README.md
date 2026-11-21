# Pacman
A recreation of the arcade game "Pacman" written in GO.


## 🔧 Powered By

- [Go](https://go.dev/)
- [Ebiten](https://ebitengine.org/)

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

### Sounds
- Sound effects and music are taken from [Warhorst/pacman](https://github.com/Warhorst/pacman) (see its README for the original sources); they live in `assets/sounds/` and are embedded in the executable.

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

## 🎮 Run It

If you already have Go installed ([instructions here](https://go.dev/doc/install)), run the game with:

```bash
go run https://github.com/geckoblu-games/pacman-go
```

### Controls

| Key                           | Action                                                                |
| ----------------------------- | --------------------------------------------------------------------- |
| Arrow keys or `W` `A` `S` `D` | Steer Pac-Man (held direction is buffered until the turn is possible) |
| `P`                           | Pause / resume                                                        |
| `M`                           | Mute / unmute                                                         |
| `Enter`                       | Start a match from the attract screen or after `GAME OVER`            |

After the classic cast presentation, the attract screen shows a controls legend that lists these keys.

## Building

```sh
go build -o pacman .
./pacman
```

## 🌐 Play in the Browser

The game also compiles to WebAssembly and is published to GitHub Pages by the
[`pages.yml`](.github/workflows/pages.yml) workflow on every push to `main`
(the repository's Pages source must be set to **GitHub Actions**). The browser
build keeps the high score only for the current session: it is not persisted.

To build and serve the web version locally:

```sh
mkdir -p dist
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o dist/pacman.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" dist/
cp web/index.html dist/
python3 -m http.server --directory dist 8080
```

Then open <http://localhost:8080/>. Browsers unlock audio only after the first
key press or click on the page.

