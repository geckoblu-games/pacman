// Package assets loads embedded images and fonts for the desktop game.
package assets

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	// classicTilesetTileSize is the atlas cell width and height in pixels.
	classicTilesetTileSize = 26
	// energizerSourceX is the energizer's zero-based atlas column.
	energizerSourceX = 8
	// energizerSourceY is the energizer's zero-based atlas row.
	energizerSourceY = 1
	// lifeSourceX is the reserve-life icon's zero-based atlas column.
	lifeSourceX = 4
	// lifeSourceY is the reserve-life icon's zero-based atlas row.
	lifeSourceY = 0
	// pacManSpriteCount is the number of Pac-Man frames in the atlas's first row.
	pacManSpriteCount = 9
	// ghostSpriteCount counts directional ghost animation cells.
	ghostSpriteCount = 8
	// deathSpriteCount counts the implosion frames and final spark.
	deathSpriteCount = 10
	// bonusSourceY is the zero-based atlas row containing all eight bonus symbols.
	bonusSourceY = 6
	// coinSize is the generated square coin's side length in logical pixels.
	coinSize = 2
)

// GameImages contains the images used by the game renderer.
// Debug overlays are generated lazily by the game package.
type GameImages struct {
	// Walls is the transparent, full-screen permanent maze layer.
	Walls *ebiten.Image
	// WallsWhite is the same maze with blue wall pixels replaced by white.
	WallsWhite *ebiten.Image
	// Energizer is the reusable power-pellet sprite.
	Energizer *ebiten.Image
	// Life is the reserve-life sprite displayed in the bottom HUD.
	Life *ebiten.Image
	// PacManFrames contains eight directional mouth frames and one closed-mouth frame.
	PacManFrames [pacManSpriteCount]*ebiten.Image
	// PacManDeathFrames contains the generated implosion and spark sprites.
	PacManDeathFrames [deathSpriteCount]*ebiten.Image
	// BlinkyFrames contains two frames for each eye direction.
	BlinkyFrames [ghostSpriteCount]*ebiten.Image
	// PinkyFrames contains two frames for each eye direction.
	PinkyFrames [ghostSpriteCount]*ebiten.Image
	// FrightenedFrames contains two blue and two flashing white frames.
	FrightenedFrames [4]*ebiten.Image
	// GhostEyes contains the four directional returning-eye sprites.
	GhostEyes [4]*ebiten.Image
	// BonusSymbols maps dossier symbol names to their atlas sprites.
	BonusSymbols map[string]*ebiten.Image
	// Coin is the generated white sprite for ordinary dots.
	Coin *ebiten.Image
	// DebugGrid caches the dashed grid overlay after its first draw.
	DebugGrid *ebiten.Image
	// DebugAllowedPath caches the gray traversable-cell overlay.
	DebugAllowedPath *ebiten.Image
}

// files embeds PNG images and TTF fonts so runtime loading needs no external files.
//
//go:embed *.png *.ttf
var files embed.FS

// ReadFile reads an embedded asset by filename, without the assets/ prefix.
func ReadFile(name string) ([]byte, error) {
	data, err := files.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("read asset %q: %w", name, err)
	}
	return data, nil
}

// LoadImage decodes an embedded PNG without changing its dimensions or colors.
func LoadImage(name string) (image.Image, error) {
	data, err := ReadFile(name)
	if err != nil {
		return nil, err
	}
	decoded, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image %q: %w", name, err)
	}
	return decoded, nil
}

// LoadImageRegion loads an embedded PNG and returns the requested rectangle.
func LoadImageRegion(name string, region image.Rectangle) (image.Image, error) {
	source, err := LoadImage(name)
	if err != nil {
		return nil, err
	}
	if region.Empty() || !region.In(source.Bounds()) {
		return nil, fmt.Errorf("image region %v is outside bounds %v for asset %q", region, source.Bounds(), name)
	}

	cropped := image.NewRGBA(image.Rect(0, 0, region.Dx(), region.Dy()))
	draw.Draw(cropped, cropped.Bounds(), source, region.Min, draw.Src)
	return cropped, nil
}

// LoadTextFace loads an embedded OpenType or TrueType font at the requested size.
func LoadTextFace(name string, size float64) (*text.GoTextFace, error) {
	if size <= 0 {
		return nil, fmt.Errorf("font size must be positive, got %v", size)
	}
	data, err := ReadFile(name)
	if err != nil {
		return nil, err
	}
	source, err := text.NewGoTextFaceSource(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parse font %q: %w", name, err)
	}
	return &text.GoTextFace{Source: source, Size: size}, nil
}

// LoadGameImages validates the maze size and loads sprites; debug caches start nil.
func LoadGameImages(wallSize image.Point) (GameImages, error) {
	walls, err := LoadImage("walls.png")
	if err != nil {
		return GameImages{}, err
	}
	if bounds := walls.Bounds(); bounds.Dx() != wallSize.X || bounds.Dy() != wallSize.Y {
		return GameImages{}, fmt.Errorf("wall image size is %dx%d, expected %dx%d", bounds.Dx(), bounds.Dy(), wallSize.X, wallSize.Y)
	}

	energizerRegion := image.Rect(
		energizerSourceX*classicTilesetTileSize,
		energizerSourceY*classicTilesetTileSize,
		(energizerSourceX+1)*classicTilesetTileSize,
		(energizerSourceY+1)*classicTilesetTileSize,
	)
	energizer, err := LoadImageRegion("classic-tileset.png", energizerRegion)
	if err != nil {
		return GameImages{}, err
	}
	lifeRegion := image.Rect(
		lifeSourceX*classicTilesetTileSize,
		lifeSourceY*classicTilesetTileSize,
		(lifeSourceX+1)*classicTilesetTileSize,
		(lifeSourceY+1)*classicTilesetTileSize,
	)
	life, err := LoadImageRegion("classic-tileset.png", lifeRegion)
	if err != nil {
		return GameImages{}, err
	}
	var pacManFrames [pacManSpriteCount]*ebiten.Image
	for column := range pacManFrames {
		region := image.Rect(
			column*classicTilesetTileSize,
			0,
			(column+1)*classicTilesetTileSize,
			classicTilesetTileSize,
		)
		frame, err := LoadImageRegion("classic-tileset.png", region)
		if err != nil {
			return GameImages{}, fmt.Errorf("load Pac-Man frame %d: %w", column, err)
		}
		pacManFrames[column] = ebiten.NewImageFromImage(frame)
	}
	var deathFrames [deathSpriteCount]*ebiten.Image
	for index := range deathFrames {
		deathFrames[index] = ebiten.NewImageFromImage(newDeathFrame(index))
	}
	var blinkyFrames [ghostSpriteCount]*ebiten.Image
	var pinkyFrames [ghostSpriteCount]*ebiten.Image
	var frightenedFrames [4]*ebiten.Image
	var ghostEyes [4]*ebiten.Image
	for column := 0; column < ghostSpriteCount; column++ {
		for _, row := range []int{1, 2, 5} {
			region := image.Rect(column*classicTilesetTileSize, row*classicTilesetTileSize,
				(column+1)*classicTilesetTileSize, (row+1)*classicTilesetTileSize)
			frame, err := LoadImageRegion("classic-tileset.png", region)
			if err != nil {
				return GameImages{}, fmt.Errorf("load ghost frame at %d,%d: %w", column, row, err)
			}
			sprite := ebiten.NewImageFromImage(frame)
			switch {
			case row == 1:
				blinkyFrames[column] = sprite
			case row == 2:
				pinkyFrames[column] = sprite
			case column < 4:
				frightenedFrames[column] = sprite
			default:
				ghostEyes[column-4] = sprite
			}
		}
	}
	bonusSymbols := make(map[string]*ebiten.Image, 8)
	for column, symbol := range []string{"Cherries", "Strawberry", "Peach", "Apple", "Grapes", "Galaxian", "Bell", "Key"} {
		region := image.Rect(
			column*classicTilesetTileSize,
			bonusSourceY*classicTilesetTileSize,
			(column+1)*classicTilesetTileSize,
			(bonusSourceY+1)*classicTilesetTileSize,
		)
		sprite, err := LoadImageRegion("classic-tileset.png", region)
		if err != nil {
			return GameImages{}, fmt.Errorf("load bonus symbol %q: %w", symbol, err)
		}
		bonusSymbols[symbol] = ebiten.NewImageFromImage(sprite)
	}
	coin := ebiten.NewImage(coinSize, coinSize)
	coin.Fill(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	wallsBlue, wallsWhite := newWallImages(walls)

	return GameImages{
		Walls:             wallsBlue,
		WallsWhite:        wallsWhite,
		Energizer:         ebiten.NewImageFromImage(energizer),
		Life:              ebiten.NewImageFromImage(life),
		PacManFrames:      pacManFrames,
		PacManDeathFrames: deathFrames,
		BlinkyFrames:      blinkyFrames,
		PinkyFrames:       pinkyFrames,
		FrightenedFrames:  frightenedFrames,
		GhostEyes:         ghostEyes,
		BonusSymbols:      bonusSymbols,
		Coin:              coin,
	}, nil
}

// newDeathFrame generates a yellow upward-opening implosion frame or final spark.
func newDeathFrame(index int) *image.RGBA {
	frame := image.NewRGBA(image.Rect(0, 0, classicTilesetTileSize, classicTilesetTileSize))
	center := float64(classicTilesetTileSize-1) / 2
	yellow := color.RGBA{R: 255, G: 255, A: 255}
	for y := 0; y < classicTilesetTileSize; y++ {
		for x := 0; x < classicTilesetTileSize; x++ {
			dx, dy := float64(x)-center, float64(y)-center
			distance := math.Hypot(dx, dy)
			if index == deathSpriteCount-1 {
				ray := math.Abs(dx) <= 0.5 || math.Abs(dy) <= 0.5 || math.Abs(math.Abs(dx)-math.Abs(dy)) <= 0.5
				if ray && distance >= 4 && distance <= 12 {
					frame.SetRGBA(x, y, yellow)
				}
			} else if distance <= 12 && math.Abs(math.Atan2(dx, -dy)) >= float64(index+1)*math.Pi/10 {
				frame.SetRGBA(x, y, yellow)
			}
		}
	}
	return frame
}

// LoadWindowIcons supplies all available icon sizes for the window manager to choose from.
func LoadWindowIcons() ([]image.Image, error) {
	icons := make([]image.Image, 0, 4)
	for _, name := range []string{"icon_16.png", "icon_32.png", "icon_64.png", "icon_128.png"} {
		icon, err := LoadImage(name)
		if err != nil {
			return nil, err
		}
		icons = append(icons, icon)
	}
	return icons, nil
}
