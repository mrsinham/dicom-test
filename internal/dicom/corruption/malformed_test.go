package corruption

import (
	"testing"
)

func TestGenerateMalformedPlaceholders(t *testing.T) {
	elements := generateMalformedPlaceholders()

	if len(elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(elements))
	}

	// Verify OW placeholder
	if elements[0].Tag.Group != 0x0069 || elements[0].Tag.Element != 0x0010 {
		t.Errorf("first element should be (0069,0010), got %v", elements[0].Tag)
	}
	if elements[0].RawValueRepresentation != "OW" {
		t.Errorf("first element should have OW VR, got %s", elements[0].RawValueRepresentation)
	}

	// Verify FL placeholder
	if elements[1].Tag.Group != 0x0071 || elements[1].Tag.Element != 0x0010 {
		t.Errorf("second element should be (0071,0010), got %v", elements[1].Tag)
	}
	if elements[1].RawValueRepresentation != "FL" {
		t.Errorf("second element should have FL VR, got %s", elements[1].RawValueRepresentation)
	}
}

func TestPatchTagValueLength(t *testing.T) {
	// Build a minimal Explicit VR LE data segment with a short-form VR tag
	// Layout: Group(2) | Element(2) | VR(2) | VL(2) | Data
	data := []byte{
		0x71, 0x00, // Group 0x0071 (LE)
		0x10, 0x00, // Element 0x0010 (LE)
		'F', 'L', // VR = "FL"
		0x08, 0x00, // VL = 8 (valid)
		0x00, 0x00, 0x80, 0x3F, // 1.0f
		0x00, 0x00, 0x00, 0x40, // 2.0f
	}

	patched := patchTagValueLength(data, 0x0071, 0x0010, 7)
	if !patched {
		t.Fatal("expected patchTagValueLength to return true")
	}

	// Verify VL was changed to 7
	vl := uint16(data[6]) | uint16(data[7])<<8
	if vl != 7 {
		t.Errorf("expected VL=7, got %d", vl)
	}
}

func TestPatchTagValueLength_LongForm(t *testing.T) {
	// Build an Explicit VR LE data segment with a long-form VR tag (OW)
	// Layout: Group(2) | Element(2) | VR(2) | Reserved(2) | VL(4) | Data
	data := []byte{
		0x69, 0x00, // Group 0x0069 (LE)
		0x10, 0x00, // Element 0x0010 (LE)
		'O', 'W', // VR = "OW"
		0x00, 0x00, // Reserved
		0x08, 0x00, 0x00, 0x00, // VL = 8 (valid)
		0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04, 0x00, // Data
	}

	patched := patchTagValueLength(data, 0x0069, 0x0010, 7)
	if !patched {
		t.Fatal("expected patchTagValueLength to return true")
	}

	// Verify VL was changed to 7
	vl := uint32(data[8]) | uint32(data[9])<<8 | uint32(data[10])<<16 | uint32(data[11])<<24
	if vl != 7 {
		t.Errorf("expected VL=7, got %d", vl)
	}
}

func TestPatchTagValueLength_NotFound(t *testing.T) {
	data := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	patched := patchTagValueLength(data, 0x0070, 0x0253, 7)
	if patched {
		t.Error("expected patchTagValueLength to return false for missing tag")
	}
}
