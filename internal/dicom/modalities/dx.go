package modalities

import (
	"math/rand/v2"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// DXGenerator generates DX (Digital X-Ray) specific metadata.
type DXGenerator struct{}

// Modality returns the DX modality type.
func (g *DXGenerator) Modality() Modality {
	return DX
}

// SOPClassUID returns the Digital X-Ray Image Storage SOP Class UID.
func (g *DXGenerator) SOPClassUID() string {
	return "1.2.840.10008.5.1.4.1.1.1.1"
}

// Scanners returns available DX equipment configurations.
func (g *DXGenerator) Scanners() []Scanner {
	return []Scanner{
		{Manufacturer: "SIEMENS", Model: "Ysio Max"},
		{Manufacturer: "GE MEDICAL SYSTEMS", Model: "Discovery XR656"},
		{Manufacturer: "PHILIPS", Model: "DigitalDiagnost C90"},
		{Manufacturer: "CARESTREAM", Model: "DRX-Evolution Plus"},
		{Manufacturer: "CANON", Model: "CXDI-710C Wireless"},
		{Manufacturer: "FUJIFILM", Model: "FDR D-EVO II"},
	}
}

// GenerateSeriesParams generates DX-specific parameters for a series.
func (g *DXGenerator) GenerateSeriesParams(scanner Scanner, rng *rand.Rand) SeriesParams {
	// View positions for radiography
	viewPositions := []string{"AP", "PA", "LAT", "LL", "RL"}
	viewPosition := viewPositions[rng.IntN(len(viewPositions))]

	// Detector pixel spacing (higher resolution than CR)
	detectorPixelSpacing := 0.1 + rng.Float64()*0.05 // 0.1-0.15 mm

	// Distances
	distanceSourceToDetector := 1000.0 + rng.Float64()*800.0  // 1000-1800 mm
	distanceSourceToPatient := 800.0 + rng.Float64()*700.0    // 800-1500 mm

	// Exposure parameters
	exposure := 1 + rng.IntN(50)          // 1-50 mAs
	kvp := float64(60 + rng.IntN(81))     // 60-140 kVp
	exposureTime := 10 + rng.IntN(91)     // 10-100 ms

	// Window settings for digital radiography
	windowCenter := 2048.0 + rng.Float64()*1000.0 // 2048-3048
	windowWidth := 4096.0 + rng.Float64()*2000.0  // 4096-6096

	params := SeriesParams{
		Modality:                 DX,
		Scanner:                  scanner,
		PixelSpacing:             detectorPixelSpacing,
		SliceThickness:           0, // Not applicable for DX
		ViewPosition:             viewPosition,
		ImagerPixelSpacing:       detectorPixelSpacing,
		DistanceSourceToDetector: distanceSourceToDetector,
		DistanceSourceToPatient:  distanceSourceToPatient,
		Exposure:                 exposure,
		KVP:                      kvp,
		ExposureTime:             exposureTime,
		WindowCenter:             windowCenter,
		WindowWidth:              windowWidth,
	}

	return params
}

// PixelConfig returns DX pixel data configuration.
func (g *DXGenerator) PixelConfig() PixelConfig {
	return PixelConfig{
		BitsAllocated:       16,
		BitsStored:          14, // DX typically uses 14-bit
		HighBit:             13,
		PixelRepresentation: 0, // Unsigned
		MinValue:            0,
		MaxValue:            16383,
		BaseValue:           8192,
	}
}

// AppendModalityElements appends DX-specific DICOM elements to a dataset.
func (g *DXGenerator) AppendModalityElements(ds *dicom.Dataset, params SeriesParams) error {
	elements := []*dicom.Element{
		mustNewElement(tag.ViewPosition, []string{params.ViewPosition}),
		mustNewElement(tag.ImagerPixelSpacing, []string{
			floatToDS(params.ImagerPixelSpacing),
			floatToDS(params.ImagerPixelSpacing),
		}),
		mustNewElement(tag.DistanceSourceToDetector, []string{floatToDS(params.DistanceSourceToDetector)}),
		mustNewElement(tag.DistanceSourceToPatient, []string{floatToDS(params.DistanceSourceToPatient)}),
		mustNewElement(tag.Exposure, []string{intToIS(params.Exposure)}),
		mustNewElement(tag.KVP, []string{floatToDS(params.KVP)}),
		mustNewElement(tag.ExposureTime, []string{intToIS(params.ExposureTime)}),
		// Detector type for digital
		mustNewElement(tag.DetectorType, []string{"SCINTILLATOR"}),
	}

	ds.Elements = append(ds.Elements, elements...)
	return nil
}

// WindowPresets returns DX window presets.
func (g *DXGenerator) WindowPresets() []WindowPreset {
	return []WindowPreset{
		{Name: "DEFAULT", Center: 8192, Width: 16383},
		{Name: "CHEST", Center: 6000, Width: 12000},
		{Name: "BONE", Center: 10000, Width: 8000},
		{Name: "SOFT_TISSUE", Center: 5000, Width: 10000},
	}
}
