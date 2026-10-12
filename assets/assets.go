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
	classicTilesetTileSize = 26
	energizerSourceX       = 8
	energizerSourceY       = 1
	coinSize               = 2
)

// GameImages contains the images used by the game renderer.
// Debug overlays are generated lazily by the game package.
type GameImages struct {
	Walls            *ebiten.Image
	Energizer        *ebiten.Image
	Coin             *ebiten.Image
	DebugGrid        *ebiten.Image
	DebugAllowedPath *ebiten.Image
}

// PNG images and TTF fonts are embedded so runtime asset loading needs no external files.
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

// LoadGameImages loads and validates all image assets required by the game.
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
	coin := ebiten.NewImage(coinSize, coinSize)
	// coin.Fill(color.RGBA{R: 255, G: 181, B: 148, A: 255})
	coin.Fill(color.RGBA{R: 255, G: 255, B: 255, A: 255})

	return GameImages{
		Walls:     ebiten.NewImageFromImage(walls),
		Energizer: ebiten.NewImageFromImage(energizer),
		Coin:      coin,
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
