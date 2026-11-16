package assets

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// wallTestGame runs asset tests inside Ebitengine so GPU pixels can be read.
type wallTestGame struct {
	// tests is the package test runner.
	tests *testing.M
	// code is the test process exit status.
	code int
}

// Update executes the tests once and stops the graphical loop.
func (game *wallTestGame) Update() error {
	game.code = game.tests.Run()
	return ebiten.Termination
}

// Draw leaves the test window empty.
func (*wallTestGame) Draw(*ebiten.Image) {}

// Layout supplies a minimal logical test surface.
func (*wallTestGame) Layout(int, int) (int, int) { return 1, 1 }

// TestMain enables GPU readback for exact generated-image comparisons.
func TestMain(tests *testing.M) {
	game := &wallTestGame{tests: tests, code: 1}
	if err := ebiten.RunGame(game); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(game.code)
}

// TestWallImagesPixelPerfect checks every generated pixel against the PNG.
func TestWallImagesPixelPerfect(t *testing.T) {
	source, err := LoadImage("walls.png")
	if err != nil {
		t.Fatal(err)
	}
	size := image.Pt(448, 576)
	if source.Bounds() != (image.Rectangle{Max: size}) {
		t.Fatalf("unexpected reference bounds: %v", source.Bounds())
	}
	images, err := LoadGameImages(size)
	if err != nil {
		t.Fatal(err)
	}
	blue := color.RGBA{R: 33, G: 33, B: 255, A: 255}
	pink := color.RGBA{R: 255, G: 184, B: 255, A: 255}
	counts := make(map[color.RGBA]int)
	for _, variant := range []struct {
		// name identifies the generated variant.
		name string
		// image contains the actual Ebitengine-rendered pixels.
		image *ebiten.Image
		// white selects the blue-to-white replacement.
		white bool
	}{
		{name: "blue", image: images.Walls},
		{name: "white", image: images.WallsWhite, white: true},
	} {
		t.Run(variant.name, func(t *testing.T) {
			if variant.image == nil || variant.image.Bounds() != source.Bounds() {
				t.Fatal("generated walls must have the reference bounds")
			}
			pixels := make([]byte, 4*size.X*size.Y)
			variant.image.ReadPixels(pixels)
			for y := 0; y < size.Y; y++ {
				for x := 0; x < size.X; x++ {
					want := color.RGBAModel.Convert(source.At(x, y)).(color.RGBA)
					if !variant.white {
						counts[want]++
					}
					if variant.white && want == blue {
						want = color.RGBA{R: 255, G: 255, B: 255, A: 255}
					}
					offset := 4 * (y*size.X + x)
					got := color.RGBA{R: pixels[offset], G: pixels[offset+1], B: pixels[offset+2], A: pixels[offset+3]}
					if got != want {
						t.Fatalf("pixel (%d, %d): got %v, want %v", x, y, got, want)
					}
				}
			}
		})
	}
	if len(counts) != 3 || counts[blue] != 19168 || counts[pink] != 128 || counts[color.RGBA{}] != 238752 {
		t.Fatalf("reference palette or geometry changed: %v", counts)
	}
}

// TestLoadGameImagesRejectsWallSize checks that wall-size validation is preserved.
func TestLoadGameImagesRejectsWallSize(t *testing.T) {
	for _, size := range []image.Point{{}, {447, 576}, {448, 575}} {
		if _, err := LoadGameImages(size); err == nil {
			t.Errorf("expected a wall-size error for %v", size)
		}
	}
}

// TestGhostColorFrames checks directional frames come from the correct atlas rows.
func TestGhostColorFrames(t *testing.T) {
	source, err := LoadImage("classic-tileset.png")
	if err != nil {
		t.Fatal(err)
	}
	images, err := LoadGameImages(image.Pt(448, 576))
	if err != nil {
		t.Fatal(err)
	}
	for _, ghost := range []struct {
		// name identifies the ghost being checked.
		name string
		// row is the atlas row containing the ghost.
		row int
		// frames are the loaded directional sprites.
		frames [ghostSpriteCount]*ebiten.Image
	}{
		{"Pinky", 2, images.PinkyFrames},
		{"Inky", 3, images.InkyFrames},
	} {
		for column, frame := range ghost.frames {
			if frame == nil || frame.Bounds().Size() != image.Pt(classicTilesetTileSize, classicTilesetTileSize) {
				t.Fatalf("%s frame %d has invalid dimensions", ghost.name, column)
			}
			pixels := make([]byte, 4*classicTilesetTileSize*classicTilesetTileSize)
			frame.ReadPixels(pixels)
			for y := 0; y < classicTilesetTileSize; y++ {
				for x := 0; x < classicTilesetTileSize; x++ {
					want := color.RGBAModel.Convert(source.At(column*classicTilesetTileSize+x, ghost.row*classicTilesetTileSize+y)).(color.RGBA)
					offset := 4 * (y*classicTilesetTileSize + x)
					got := color.RGBA{R: pixels[offset], G: pixels[offset+1], B: pixels[offset+2], A: pixels[offset+3]}
					if got != want {
						t.Fatalf("%s frame %d pixel (%d,%d): got %v, want %v", ghost.name, column, x, y, got, want)
					}
				}
			}
		}
	}
}

// TestDeathFrames checks the opening implosion and the distinct final spark.
func TestDeathFrames(t *testing.T) {
	previous := classicTilesetTileSize * classicTilesetTileSize
	for index := 0; index < deathSpriteCount-1; index++ {
		frame := newDeathFrame(index)
		colored := 0
		for y := 0; y < classicTilesetTileSize; y++ {
			for x := 0; x < classicTilesetTileSize; x++ {
				if frame.RGBAAt(x, y).A != 0 {
					colored++
				}
			}
		}
		if colored == 0 || colored >= previous {
			t.Fatalf("death frame %d should shrink visibly: %d pixels after %d", index, colored, previous)
		}
		previous = colored
	}
	spark := newDeathFrame(deathSpriteCount - 1)
	if spark.RGBAAt(13, 13).A != 0 || spark.RGBAAt(13, 3).A == 0 || spark.RGBAAt(5, 5).A == 0 {
		t.Fatal("final frame should be a hollow radial spark")
	}
}
