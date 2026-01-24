package modalities

import (
	"math/rand/v2"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// MGGenerator generates MG (Mammography) specific metadata.
type MGGenerator struct{}

// Modality returns the MG modality type.
func (g *MGGenerator) Modality() Modality {
	return MG
}

// SOPClassUID returns the Digital Mammography X-Ray Image Storage SOP Class UID.
func (g *MGGenerator) SOPClassUID() string {
	return "1.2.840.10008.5.1.4.1.1.1.2"
}

// Scanners returns available MG equipment configurations.
func (g *MGGenerator) Scanners() []Scanner {
	return []Scanner{
		{Manufacturer: "HOLOGIC", Model: "Selenia Dimensions"},
		{Manufacturer: "GE MEDICAL SYSTEMS", Model: "Senographe Pristina"},
		{Manufacturer: "SIEMENS", Model: "MAMMOMAT Revelation"},
		{Manufacturer: "FUJIFILM", Model: "AMULET Innovality"},
		{Manufacturer: "PHILIPS", Model: "MicroDose SI"},
		{Manufacturer: "IMS GIOTTO", Model: "Class"},
	}
}

// GenerateSeriesParams generates MG-specific parameters for a series.
func (g *MGGenerator) GenerateSeriesParams(scanner Scanner, rng *rand.Rand) SeriesParams {
	// Image laterality (left or right breast)
	lateralities := []string{"L", "R"}
	imageLaterality := lateralities[rng.IntN(len(lateralities))]

	// View positions for mammography
	viewPositions := []string{"CC", "MLO", "ML", "LM"}
	viewPosition := viewPositions[rng.IntN(len(viewPositions))]

	// Anode target materials
	anodeMaterials := []string{"MOLYBDENUM", "RHODIUM", "TUNGSTEN"}
	anodeTargetMaterial := anodeMaterials[rng.IntN(len(anodeMaterials))]

	// Filter materials (often paired with anode)
	filterMaterials := []string{"MOLYBDENUM", "RHODIUM", "SILVER", "ALUMINUM"}
	filterMaterial := filterMaterials[rng.IntN(len(filterMaterials))]

	// Compression force (typically 80-200 N)
	compressionForce := 80.0 + rng.Float64()*120.0

	// Organ dose (typically 1-3 mGy)
	organDose := 1.0 + rng.Float64()*2.0

	// Pixel spacing (very high resolution, typically 0.05-0.1 mm)
	pixelSpacing := 0.05 + rng.Float64()*0.05

	// Exposure parameters
	kvp := float64(25 + rng.IntN(10)) // 25-34 kVp (lower than general radiography)
	exposure := 50 + rng.IntN(150)    // 50-200 mAs

	// Window settings for mammography
	windowCenter := 3000.0 + rng.Float64()*1000.0 // 3000-4000
	windowWidth := 6000.0 + rng.Float64()*2000.0  // 6000-8000

	params := SeriesParams{
		Modality:            MG,
		Scanner:             scanner,
		PixelSpacing:        pixelSpacing,
		SliceThickness:      0, // Not applicable for MG
		ImageLaterality:     imageLaterality,
		ViewPosition:        viewPosition,
		AnodeTargetMaterial: anodeTargetMaterial,
		FilterMaterial:      filterMaterial,
		CompressionForce:    compressionForce,
		OrganDose:           organDose,
		KVP:                 kvp,
		Exposure:            exposure,
		WindowCenter:        windowCenter,
		WindowWidth:         windowWidth,
	}

	return params
}

// PixelConfig returns MG pixel data configuration.
func (g *MGGenerator) PixelConfig() PixelConfig {
	return PixelConfig{
		BitsAllocated:       16,
		BitsStored:          14, // High resolution for mammography
		HighBit:             13,
		PixelRepresentation: 0, // Unsigned
		MinValue:            0,
		MaxValue:            16383,
		BaseValue:           8192,
	}
}

// AppendModalityElements appends MG-specific DICOM elements to a dataset.
func (g *MGGenerator) AppendModalityElements(ds *dicom.Dataset, params SeriesParams) error {
	elements := []*dicom.Element{
		mustNewElement(tag.ImageLaterality, []string{params.ImageLaterality}),
		mustNewElement(tag.ViewPosition, []string{params.ViewPosition}),
		mustNewElement(tag.AnodeTargetMaterial, []string{params.AnodeTargetMaterial}),
		mustNewElement(tag.FilterMaterial, []string{params.FilterMaterial}),
		mustNewElement(tag.CompressionForce, []string{floatToDS(params.CompressionForce)}),
		mustNewElement(tag.OrganDose, []string{floatToDS(params.OrganDose)}),
		mustNewElement(tag.KVP, []string{floatToDS(params.KVP)}),
		mustNewElement(tag.Exposure, []string{intToIS(params.Exposure)}),
		// Photometric interpretation for mammography (typically MONOCHROME1)
		mustNewElement(tag.PhotometricInterpretation, []string{"MONOCHROME1"}),
	}

	ds.Elements = append(ds.Elements, elements...)
	return nil
}

// WindowPresets returns MG window presets.
func (g *MGGenerator) WindowPresets() []WindowPreset {
	return []WindowPreset{
		{Name: "DEFAULT", Center: 8192, Width: 16383},
		{Name: "DENSE", Center: 6000, Width: 10000},
		{Name: "FATTY", Center: 10000, Width: 12000},
		{Name: "CALCIFICATION", Center: 12000, Width: 8000},
	}
}
