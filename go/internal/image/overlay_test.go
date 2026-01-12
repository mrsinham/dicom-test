package image

import (
	"testing"
)

func TestAddTextOverlay_Range(t *testing.T) {
	width, height := 256, 256
	pixels := GenerateSingleImage(width, height, 42)

	err := AddTextOverlay(pixels, width, height, 5, 10)
	if err != nil {
		t.Fatalf("AddTextOverlay failed: %v", err)
	}

	// Verify pixels still in valid range
	for i, pixel := range pixels {
		if pixel > 4095 {
			t.Errorf("Pixel %d value %d exceeds 12-bit max after overlay", i, pixel)
		}
	}
}

func TestAddTextOverlay_ModifiesImage(t *testing.T) {
	width, height := 256, 256
	pixels := GenerateSingleImage(width, height, 42)

	// Make a copy before overlay
	original := make([]uint16, len(pixels))
	copy(original, pixels)

	err := AddTextOverlay(pixels, width, height, 5, 10)
	if err != nil {
		t.Fatalf("AddTextOverlay failed: %v", err)
	}

	// Check that at least some pixels changed (text was drawn)
	different := false
	for i := range pixels {
		if pixels[i] != original[i] {
			different = true
			break
		}
	}

	if !different {
		t.Errorf("Expected overlay to modify pixels")
	}
}
