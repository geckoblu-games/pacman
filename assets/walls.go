package assets

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// newWallImages draws exact source-color runs as integer-aligned rectangles.
// Only the original blue wall color changes in the white variant.
func newWallImages(source image.Image) (blue, white *ebiten.Image) {
	bounds := source.Bounds()
	blue = ebiten.NewImage(bounds.Dx(), bounds.Dy())
	white = ebiten.NewImage(bounds.Dx(), bounds.Dy())
	wallBlue := color.RGBA{R: 33, G: 33, B: 255, A: 255}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; {
			pixel := color.RGBAModel.Convert(source.At(x, y)).(color.RGBA)
			end := x + 1
			for end < bounds.Max.X && color.RGBAModel.Convert(source.At(end, y)).(color.RGBA) == pixel {
				end++
			}
			if pixel.A != 0 {
				left, top, width := float32(x-bounds.Min.X), float32(y-bounds.Min.Y), float32(end-x)
				vector.DrawFilledRect(blue, left, top, width, 1, pixel, false)
				if pixel == wallBlue {
					pixel = color.RGBA{R: 255, G: 255, B: 255, A: 255}
				}
				vector.DrawFilledRect(white, left, top, width, 1, pixel, false)
			}
			x = end
		}
	}
	return blue, white
}
