package assets

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/png"
)

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
