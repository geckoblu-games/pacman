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
	// Energizer is the reusable power-pellet sprite.
	Energizer *ebiten.Image
	// Life is the reserve-life sprite displayed in the bottom HUD.
	Life *ebiten.Image
	// PacManFrames contains eight directional mouth frames and one closed-mouth frame.
	PacManFrames [pacManSpriteCount]*ebiten.Image
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

	return GameImages{
		Walls:        ebiten.NewImageFromImage(walls),
		Energizer:    ebiten.NewImageFromImage(energizer),
		Life:         ebiten.NewImageFromImage(life),
		PacManFrames: pacManFrames,
		BonusSymbols: bonusSymbols,
		Coin:         coin,
	}, nil
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
