package modalities

import (
	"math/rand/v2"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// CRGenerator generates CR (Computed Radiography) specific metadata.
type CRGenerator struct{}

// Modality returns the CR modality type.
func (g *CRGenerator) Modality() Modality {
	return CR
}

// SOPClassUID returns the Computed Radiography Image Storage SOP Class UID.
func (g *CRGenerator) SOPClassUID() string {
	return "1.2.840.10008.5.1.4.1.1.1"
}

// Scanners returns available CR equipment configurations.
func (g *CRGenerator) Scanners() []Scanner {
	return []Scanner{
		{Manufacturer: "FUJIFILM", Model: "FCR Profect CS"},
		{Manufacturer: "CARESTREAM", Model: "DRX-Revolution"},
		{Manufacturer: "AGFA", Model: "CR 30-X"},
		{Manufacturer: "KONICA MINOLTA", Model: "Regius Model 110"},
		{Manufacturer: "PHILIPS", Model: "PCR Eleva S"},
	}
}

// GenerateSeriesParams generates CR-specific parameters for a series.
func (g *CRGenerator) GenerateSeriesParams(scanner Scanner, rng *rand.Rand) SeriesParams {
	// View positions for radiography
	viewPositions := []string{"AP", "PA", "LAT", "LL", "RL"}
	viewPosition := viewPositions[rng.IntN(len(viewPositions))]

	// Imager pixel spacing (detector resolution)
	imagerPixelSpacing := 0.1 + rng.Float64()*0.1 // 0.1-0.2 mm

	// Distances
	distanceSourceToDetector := 1000.0 + rng.Float64()*800.0  // 1000-1800 mm
	distanceSourceToPatient := 800.0 + rng.Float64()*700.0    // 800-1500 mm

	// Exposure
	exposure := 1 + rng.IntN(50) // 1-50 mAs

	// Window settings for radiography
	windowCenter := 2048.0 + rng.Float64()*1000.0 // 2048-3048
	windowWidth := 4096.0 + rng.Float64()*2000.0  // 4096-6096

	params := SeriesParams{
		Modality:                 CR,
		Scanner:                  scanner,
		PixelSpacing:             imagerPixelSpacing,
		SliceThickness:           0, // Not applicable for CR
		ViewPosition:             viewPosition,
		ImagerPixelSpacing:       imagerPixelSpacing,
		DistanceSourceToDetector: distanceSourceToDetector,
		DistanceSourceToPatient:  distanceSourceToPatient,
		Exposure:                 exposure,
		WindowCenter:             windowCenter,
		WindowWidth:              windowWidth,
	}

	return params
}

// PixelConfig returns CR pixel data configuration.
func (g *CRGenerator) PixelConfig() PixelConfig {
	return PixelConfig{
		BitsAllocated:       16,
		BitsStored:          12,
		HighBit:             11,
		PixelRepresentation: 0, // Unsigned
		MinValue:            0,
		MaxValue:            4095,
		BaseValue:           2048,
	}
}

// AppendModalityElements appends CR-specific DICOM elements to a dataset.
func (g *CRGenerator) AppendModalityElements(ds *dicom.Dataset, params SeriesParams) error {
	elements := []*dicom.Element{
		mustNewElement(tag.ViewPosition, []string{params.ViewPosition}),
		mustNewElement(tag.ImagerPixelSpacing, []string{
			floatToDS(params.ImagerPixelSpacing),
			floatToDS(params.ImagerPixelSpacing),
		}),
		mustNewElement(tag.DistanceSourceToDetector, []string{floatToDS(params.DistanceSourceToDetector)}),
		mustNewElement(tag.DistanceSourceToPatient, []string{floatToDS(params.DistanceSourceToPatient)}),
		mustNewElement(tag.Exposure, []string{intToIS(params.Exposure)}),
		// Plate ID for CR
		mustNewElement(tag.PlateID, []string{"PLATE001"}),
	}

	ds.Elements = append(ds.Elements, elements...)
	return nil
}

// WindowPresets returns CR window presets.
func (g *CRGenerator) WindowPresets() []WindowPreset {
	return []WindowPreset{
		{Name: "DEFAULT", Center: 2048, Width: 4096},
		{Name: "SOFT_TISSUE", Center: 1500, Width: 3000},
		{Name: "BONE", Center: 3000, Width: 2000},
	}
}
