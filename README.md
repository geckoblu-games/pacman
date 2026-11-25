# Pacman
A recreation of the arcade game Pac-Man written in Go.

You can play it online [here](https://geckoblu-games.github.io/pacman/).

Or you can build it yourself for a desktop version (see [Build from source](#build-from-source)).

## Controls

| Key                           | Action                                                                |
| ----------------------------- | --------------------------------------------------------------------- |
| Arrow keys or `W` `A` `S` `D` | Steer Pac-Man (held direction is buffered until the turn is possible) |
| `P`                           | Pause / resume                                                        |
| `M`                           | Mute / unmute                                                         |
| `Enter`                       | Start a match from the attract screen or after `GAME OVER`            |

## Powered By

- [Go](https://go.dev/)
- [Ebiten](https://ebitengine.org/)

## Main resources
- the great pacman dossier: (a copy in  ```./doc```)
  - [The Pac-Man Dossier](https://pacman.holenet.info)
  - [Pac-Man Ghost AI Explained](https://www.youtube.com/watch?v=ataGotQ7ir8)
- Videos of the game
  - [the game in action](https://www.youtube.com/watch?v=-CbyAk3Sn9I)
  - [longplay until the game glitches out](https://www.youtube.com/watch?v=AuoH0vz3Mqk)
- [Making Games in Go for Absolute Beginners](https://threedots.tech/post/making-games-in-go/)

### Font
- [Press Start 2P Font](https://fonts.google.com/specimen/Press+Start+2P)

### Sounds
- Sound effects and music are taken from [Warhorst/pacman](https://github.com/Warhorst/pacman) (see its README for the original sources).

## Build from source

Install [Go](https://go.dev/doc/install) 1.27 or later.

The game is built with [Ebitengine](https://ebitengine.org/); make sure its system dependencies are installed, see the [installation guide](https://ebitengine.org/en/documents/install.html).

Clone the repository:
```
git clone https://github.com/geckoblu-games/pacman.git

cd pacman
```

Compile the game:

```
go build .
```

The executable will be created in the project directory.

### Run

On Linux and macOS:

```
./pacman
```

On Windows:

```
.\pacman.exe
```

### Install on Linux

With Go and Ebitengine's Linux system dependencies installed, run:

```
./install-linux.sh
```

The script builds the game and adds Pac-Man to the current user's application
menu without requiring administrator privileges. 

