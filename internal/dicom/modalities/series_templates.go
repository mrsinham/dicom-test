// internal/dicom/modalities/series_templates.go
package modalities

import (
	"math/rand/v2"
)

// SeriesTemplate defines a template for a series within a study
type SeriesTemplate struct {
	SequenceName      string  // MR sequence name (e.g., "T1_SE", "T2_FSE")
	SeriesDescription string  // Human-readable description
	Orientation       string  // SAG, AX, COR
	HasContrast       bool    // Whether this series uses contrast
	ContrastAgent     string  // Contrast agent name if HasContrast
	WindowCenter      float64 // Series-specific window center (0 = use default)
	WindowWidth       float64 // Series-specific window width (0 = use default)
}

// Orientation values
const (
	OrientationAxial    = "AX"
	OrientationSagittal = "SAG"
	OrientationCoronal  = "COR"
)

// ImageOrientationPatient returns the DICOM ImageOrientationPatient values
// Format: row direction cosines followed by column direction cosines
func (t SeriesTemplate) ImageOrientationPatient() []float64 {
	switch t.Orientation {
	case OrientationAxial:
		return []float64{1, 0, 0, 0, 1, 0}
	case OrientationSagittal:
		return []float64{0, 1, 0, 0, 0, -1}
	case OrientationCoronal:
		return []float64{1, 0, 0, 0, 0, -1}
	default:
		return []float64{1, 0, 0, 0, 1, 0} // Default to axial
	}
}

// MR Brain series templates
var mrBrainTemplates = []SeriesTemplate{
	{SequenceName: "T1_SE", SeriesDescription: "T1 SAG", Orientation: OrientationSagittal},
	{SequenceName: "T2_FSE", SeriesDescription: "T2 AX", Orientation: OrientationAxial},
	{SequenceName: "T2_FLAIR", SeriesDescription: "FLAIR AX", Orientation: OrientationAxial},
	{SequenceName: "T1_MPRAGE", SeriesDescription: "T1 SAG +C", Orientation: OrientationSagittal, HasContrast: true, ContrastAgent: "GADOVIST"},
	{SequenceName: "DWI", SeriesDescription: "DWI AX", Orientation: OrientationAxial},
	{SequenceName: "T2_STAR", SeriesDescription: "T2* GRE", Orientation: OrientationAxial},
}

// MR Knee series templates
var mrKneeTemplates = []SeriesTemplate{
	{SequenceName: "T1_SE", SeriesDescription: "T1 SAG", Orientation: OrientationSagittal},
	{SequenceName: "T2_FSE", SeriesDescription: "T2 SAG FAT-SAT", Orientation: OrientationSagittal},
	{SequenceName: "PD_FSE", SeriesDescription: "PD COR", Orientation: OrientationCoronal},
	{SequenceName: "T2_FSE", SeriesDescription: "T2 AX", Orientation: OrientationAxial},
	{SequenceName: "T1_SE", SeriesDescription: "T1 COR", Orientation: OrientationCoronal},
}

// MR Spine series templates
var mrSpineTemplates = []SeriesTemplate{
	{SequenceName: "T1_SE", SeriesDescription: "T1 SAG", Orientation: OrientationSagittal},
	{SequenceName: "T2_FSE", SeriesDescription: "T2 SAG", Orientation: OrientationSagittal},
	{SequenceName: "STIR", SeriesDescription: "STIR SAG", Orientation: OrientationSagittal},
	{SequenceName: "T2_FSE", SeriesDescription: "T2 AX", Orientation: OrientationAxial},
}

// MR Abdomen series templates
var mrAbdomenTemplates = []SeriesTemplate{
	{SequenceName: "T2_SSFSE", SeriesDescription: "T2 COR SSFSE", Orientation: OrientationCoronal},
	{SequenceName: "T2_FSE", SeriesDescription: "T2 AX FAT-SAT", Orientation: OrientationAxial},
	{SequenceName: "T1_VIBE", SeriesDescription: "T1 AX PRE", Orientation: OrientationAxial},
	{SequenceName: "T1_VIBE", SeriesDescription: "T1 AX +C ART", Orientation: OrientationAxial, HasContrast: true, ContrastAgent: "GADOVIST"},
	{SequenceName: "T1_VIBE", SeriesDescription: "T1 AX +C PORT", Orientation: OrientationAxial, HasContrast: true, ContrastAgent: "GADOVIST"},
}

// CT series templates (contrast phases)
var ctWithContrastTemplates = []SeriesTemplate{
	{SeriesDescription: "Sans contraste", Orientation: OrientationAxial},
	{SeriesDescription: "Arteriel", Orientation: OrientationAxial, HasContrast: true, ContrastAgent: "IOMERON 400"},
	{SeriesDescription: "Portal", Orientation: OrientationAxial, HasContrast: true, ContrastAgent: "IOMERON 400"},
	{SeriesDescription: "Tardif", Orientation: OrientationAxial, HasContrast: true, ContrastAgent: "IOMERON 400"},
}

// CT without contrast templates
var ctWithoutContrastTemplates = []SeriesTemplate{
	{SeriesDescription: "Acquisition standard", Orientation: OrientationAxial},
	{SeriesDescription: "Reconstruction os", Orientation: OrientationAxial, WindowCenter: 400, WindowWidth: 2000},
	{SeriesDescription: "Reconstruction poumon", Orientation: OrientationAxial, WindowCenter: -600, WindowWidth: 1500},
}

// CR/DX templates - typically single series, multiple views
var crDXTemplates = []SeriesTemplate{
	{SeriesDescription: "Face", Orientation: OrientationCoronal},
	{SeriesDescription: "Profil", Orientation: OrientationSagittal},
	{SeriesDescription: "Oblique", Orientation: OrientationAxial},
}

// US templates
var usTemplates = []SeriesTemplate{
	{SeriesDescription: "Mode B", Orientation: OrientationAxial},
	{SeriesDescription: "Doppler couleur", Orientation: OrientationAxial},
	{SeriesDescription: "Mesures", Orientation: OrientationAxial},
}

// MG templates - standard mammography views
var mgTemplates = []SeriesTemplate{
	{SeriesDescription: "CC Droit", Orientation: OrientationAxial},
	{SeriesDescription: "MLO Droit", Orientation: OrientationAxial},
	{SeriesDescription: "CC Gauche", Orientation: OrientationAxial},
	{SeriesDescription: "MLO Gauche", Orientation: OrientationAxial},
}

// GetSeriesTemplates returns series templates for the given modality and body part
func GetSeriesTemplates(modality Modality, bodyPart string, count int, rng *rand.Rand) []SeriesTemplate {
	var pool []SeriesTemplate

	switch modality {
	case MR:
		switch bodyPart {
		case "HEAD", "BRAIN":
			pool = mrBrainTemplates
		case "KNEE", "ANKLE", "FOOT", "SHOULDER", "ELBOW", "WRIST", "HIP":
			pool = mrKneeTemplates
		case "CSPINE", "TSPINE", "LSPINE", "SPINE":
			pool = mrSpineTemplates
		case "ABDOMEN", "PELVIS", "LIVER":
			pool = mrAbdomenTemplates
		default:
			pool = mrBrainTemplates // Default to brain
		}
	case CT:
		// 50% chance of contrast series
		if rng.IntN(2) == 0 {
			pool = ctWithContrastTemplates
		} else {
			pool = ctWithoutContrastTemplates
		}
	case CR, DX:
		pool = crDXTemplates
	case US:
		pool = usTemplates
	case MG:
		pool = mgTemplates
	default:
		pool = mrBrainTemplates
	}

	// Select templates up to count
	if count >= len(pool) {
		return pool
	}

	// Shuffle and select
	selected := make([]SeriesTemplate, len(pool))
	copy(selected, pool)
	rng.Shuffle(len(selected), func(i, j int) {
		selected[i], selected[j] = selected[j], selected[i]
	})

	return selected[:count]
}

// GetDefaultSeriesCount returns the default number of series for a modality
func GetDefaultSeriesCount(modality Modality) int {
	switch modality {
	case MR:
		return 4
	case CT:
		return 3
	case CR, DX:
		return 2
	case US:
		return 2
	case MG:
		return 4
	default:
		return 1
	}
}
