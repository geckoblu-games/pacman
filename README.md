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
| `Enter`                       | Start a match from the attract screen or after `GAME OVER`            |

## Building

```sh
go build -o pacman .
./pacman
```

