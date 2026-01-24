package modalities

import (
	"math/rand/v2"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// USGenerator generates US (Ultrasound) specific metadata.
type USGenerator struct{}

// Modality returns the US modality type.
func (g *USGenerator) Modality() Modality {
	return US
}

// SOPClassUID returns the Ultrasound Image Storage SOP Class UID.
func (g *USGenerator) SOPClassUID() string {
	return "1.2.840.10008.5.1.4.1.1.6.1"
}

// Scanners returns available US equipment configurations.
func (g *USGenerator) Scanners() []Scanner {
	return []Scanner{
		{Manufacturer: "GE MEDICAL SYSTEMS", Model: "LOGIQ E10"},
		{Manufacturer: "PHILIPS", Model: "EPIQ Elite"},
		{Manufacturer: "SIEMENS", Model: "ACUSON Sequoia"},
		{Manufacturer: "CANON", Model: "Aplio i800"},
		{Manufacturer: "SAMSUNG", Model: "RS85 Prestige"},
		{Manufacturer: "HITACHI", Model: "ARIETTA 850"},
	}
}

// GenerateSeriesParams generates US-specific parameters for a series.
func (g *USGenerator) GenerateSeriesParams(scanner Scanner, rng *rand.Rand) SeriesParams {
	// Transducer types
	transducerTypes := []string{"LINEAR", "CONVEX", "PHASED"}
	transducerType := transducerTypes[rng.IntN(len(transducerTypes))]

	// Transducer frequency based on type
	var transducerFrequency float64
	switch transducerType {
	case "LINEAR":
		transducerFrequency = 7.0 + rng.Float64()*8.0 // 7-15 MHz (superficial)
	case "CONVEX":
		transducerFrequency = 2.0 + rng.Float64()*4.0 // 2-6 MHz (abdominal)
	case "PHASED":
		transducerFrequency = 2.0 + rng.Float64()*3.0 // 2-5 MHz (cardiac)
	}

	// Pixel spacing (varies with depth and frequency)
	pixelSpacing := 0.2 + rng.Float64()*0.3 // 0.2-0.5 mm

	// Window settings for ultrasound
	windowCenter := 128.0
	windowWidth := 256.0

	params := SeriesParams{
		Modality:            US,
		Scanner:             scanner,
		PixelSpacing:        pixelSpacing,
		SliceThickness:      0, // Not applicable for US
		TransducerType:      transducerType,
		TransducerFrequency: transducerFrequency,
		WindowCenter:        windowCenter,
		WindowWidth:         windowWidth,
	}

	return params
}

// PixelConfig returns US pixel data configuration.
func (g *USGenerator) PixelConfig() PixelConfig {
	return PixelConfig{
		BitsAllocated:       8,
		BitsStored:          8,
		HighBit:             7,
		PixelRepresentation: 0, // Unsigned
		MinValue:            0,
		MaxValue:            255,
		BaseValue:           128,
	}
}

// AppendModalityElements appends US-specific DICOM elements to a dataset.
func (g *USGenerator) AppendModalityElements(ds *dicom.Dataset, params SeriesParams) error {
	// Convert transducer frequency from MHz to Hz (UL tag expects integer Hz)
	transducerFreqHz := int(params.TransducerFrequency * 1000000)

	elements := []*dicom.Element{
		mustNewElement(tag.TransducerType, []string{params.TransducerType}),
		mustNewElement(tag.TransducerFrequency, []int{transducerFreqHz}),
		// Number of frames (single frame for now)
		mustNewElement(tag.NumberOfFrames, []string{"1"}),
	}

	ds.Elements = append(ds.Elements, elements...)
	return nil
}

// WindowPresets returns US window presets.
func (g *USGenerator) WindowPresets() []WindowPreset {
	return []WindowPreset{
		{Name: "DEFAULT", Center: 128, Width: 256},
		{Name: "BRIGHT", Center: 100, Width: 200},
		{Name: "CONTRAST", Center: 150, Width: 300},
	}
}
