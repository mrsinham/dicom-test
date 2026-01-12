package image

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// AddTextOverlay adds text "File X/Y" to the image pixels.
//
// Modifies pixels in place. Text is drawn with white color and black outline
// for visibility against varying backgrounds. Uses basicfont for simplicity;
// full TrueType font rendering can be added later using golang.org/x/image/font/opentype.
//
// The function converts the uint16 pixel data to an image, draws the text,
// and converts back to uint16, ensuring all values remain in the valid 12-bit
// range (0-4095).
func AddTextOverlay(pixels []uint16, width, height, imageNum, totalImages int) error {
	if len(pixels) != width*height {
		return fmt.Errorf("pixel slice length %d does not match dimensions %dx%d", len(pixels), width, height)
	}

	// Convert uint16 pixels to Gray16 image for manipulation
	img := image.NewGray16(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			// Scale from 12-bit (0-4095) to 16-bit (0-65535) for better contrast
			val := uint32(pixels[idx]) * 16
			if val > 65535 {
				val = 65535
			}
			img.SetGray16(x, y, color.Gray16{Y: uint16(val)})
		}
	}

	// Create RGBA image for drawing (easier to draw text with colors)
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	// Prepare text
	text := fmt.Sprintf("File %d/%d", imageNum, totalImages)

	// Use basicfont.Face7x13 - a simple fixed-width font
	face := basicfont.Face7x13

	// Calculate text position: centered horizontally, near top (5% from top)
	paddingTop := int(float64(height) * 0.05)

	// Measure text width
	textWidth := font.MeasureString(face, text).Ceil()
	x := (width - textWidth) / 2

	// For basicfont, metrics are available from the font.Metrics method
	metrics := face.Metrics()
	y := paddingTop + metrics.Ascent.Ceil()

	// Create a drawer
	drawer := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(color.Black),
		Face: face,
		Dot:  fixed.P(x, y),
	}

	// Draw black outline for visibility (thick outline)
	outlineThickness := 2
	for dx := -outlineThickness; dx <= outlineThickness; dx++ {
		for dy := -outlineThickness; dy <= outlineThickness; dy++ {
			if dx != 0 || dy != 0 { // Skip center
				drawer.Dot = fixed.P(x+dx, y+dy)
				drawer.DrawString(text)
			}
		}
	}

	// Draw main text in white
	drawer.Src = image.NewUniform(color.White)
	drawer.Dot = fixed.P(x, y)
	drawer.DrawString(text)

	// Convert back to grayscale
	gray := image.NewGray(rgba.Bounds())
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			gray.Set(x, y, rgba.At(x, y))
		}
	}

	// Convert back to uint16 pixels and scale back to 12-bit range (0-4095)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			grayColor := gray.GrayAt(x, y)
			// Scale from 8-bit (0-255) to 12-bit (0-4095)
			val := uint32(grayColor.Y) * 16
			if val > 4095 {
				val = 4095
			}
			pixels[idx] = uint16(val)
		}
	}

	return nil
}
