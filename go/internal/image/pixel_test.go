package image

import (
	"testing"
)

func TestGenerateSingleImage_Size(t *testing.T) {
	width, height := 256, 256
	pixels := GenerateSingleImage(width, height, 42)

	expectedSize := width * height
	if len(pixels) != expectedSize {
		t.Errorf("Expected %d pixels, got %d", expectedSize, len(pixels))
	}
}

func TestGenerateSingleImage_Range(t *testing.T) {
	width, height := 128, 128
	pixels := GenerateSingleImage(width, height, 42)

	for i, pixel := range pixels {
		if pixel > 4095 {
			t.Errorf("Pixel %d value %d exceeds 12-bit max (4095)", i, pixel)
		}
	}
}

func TestGenerateSingleImage_Deterministic(t *testing.T) {
	width, height := 128, 128

	pixels1 := GenerateSingleImage(width, height, 42)
	pixels2 := GenerateSingleImage(width, height, 42)

	if len(pixels1) != len(pixels2) {
		t.Fatalf("Pixel slices have different lengths")
	}

	for i := range pixels1 {
		if pixels1[i] != pixels2[i] {
			t.Errorf("Pixel %d differs: %d != %d", i, pixels1[i], pixels2[i])
		}
	}
}

func TestGenerateSingleImage_Different(t *testing.T) {
	width, height := 128, 128

	pixels1 := GenerateSingleImage(width, height, 42)
	pixels2 := GenerateSingleImage(width, height, 43)

	same := true
	for i := range pixels1 {
		if pixels1[i] != pixels2[i] {
			same = false
			break
		}
	}

	if same {
		t.Errorf("Different seeds should produce different pixel data")
	}
}
