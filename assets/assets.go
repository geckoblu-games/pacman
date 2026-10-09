package assets

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/png"
)

//go:embed *.png *.ttf
var files embed.FS

func ReadFile(name string) ([]byte, error) {
	data, err := files.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("read asset %q: %w", name, err)
	}
	return data, nil
}

func LoadWindowIcons() ([]image.Image, error) {
	icons := make([]image.Image, 0, 4)
	for _, name := range []string{"icon_16.png", "icon_32.png", "icon_64.png", "icon_128.png"} {
		data, err := ReadFile(name)
		if err != nil {
			return nil, err
		}
		icon, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("decode icon %q: %w", name, err)
		}
		icons = append(icons, icon)
	}
	return icons, nil
}
