// internal/dicom/modalities/series_templates_test.go
package modalities

import (
	"math/rand/v2"
	"testing"
)

func TestSeriesTemplate_ImageOrientationPatient(t *testing.T) {
	tests := []struct {
		orientation string
		want        []float64
	}{
		{OrientationAxial, []float64{1, 0, 0, 0, 1, 0}},
		{OrientationSagittal, []float64{0, 1, 0, 0, 0, -1}},
		{OrientationCoronal, []float64{1, 0, 0, 0, 0, -1}},
		{"", []float64{1, 0, 0, 0, 1, 0}}, // Default to axial
	}

	for _, tt := range tests {
		t.Run(tt.orientation, func(t *testing.T) {
			tmpl := SeriesTemplate{Orientation: tt.orientation}
			got := tmpl.ImageOrientationPatient()
			if len(got) != 6 {
				t.Fatalf("ImageOrientationPatient() returned %d values, want 6", len(got))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ImageOrientationPatient()[%d] = %f, want %f", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGetSeriesTemplates_MR(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))

	tests := []struct {
		bodyPart string
		count    int
	}{
		{"HEAD", 3},
		{"BRAIN", 4},
		{"KNEE", 3},
		{"LSPINE", 2},
		{"ABDOMEN", 5},
		{"UNKNOWN", 3}, // Should use default (brain)
	}

	for _, tt := range tests {
		t.Run(tt.bodyPart, func(t *testing.T) {
			templates := GetSeriesTemplates(MR, tt.bodyPart, tt.count, rng)
			if len(templates) != tt.count {
				t.Errorf("GetSeriesTemplates(MR, %q, %d) returned %d templates, want %d",
					tt.bodyPart, tt.count, len(templates), tt.count)
			}
		})
	}
}

func TestGetSeriesTemplates_CT(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))

	templates := GetSeriesTemplates(CT, "CHEST", 3, rng)
	if len(templates) != 3 {
		t.Errorf("GetSeriesTemplates(CT, CHEST, 3) returned %d templates, want 3", len(templates))
	}
}

func TestGetSeriesTemplates_CR(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))

	templates := GetSeriesTemplates(CR, "CHEST", 2, rng)
	if len(templates) != 2 {
		t.Errorf("GetSeriesTemplates(CR, CHEST, 2) returned %d templates, want 2", len(templates))
	}
}

func TestGetSeriesTemplates_US(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))

	templates := GetSeriesTemplates(US, "", 2, rng)
	if len(templates) != 2 {
		t.Errorf("GetSeriesTemplates(US, \"\", 2) returned %d templates, want 2", len(templates))
	}
}

func TestGetSeriesTemplates_MG(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))

	templates := GetSeriesTemplates(MG, "BREAST", 4, rng)
	if len(templates) == 0 {
		t.Error("GetSeriesTemplates(MG, BREAST, 4) returned no templates")
	}
}

func TestGetSeriesTemplates_MoreThanAvailable(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))

	// Request more templates than available
	templates := GetSeriesTemplates(US, "", 100, rng)
	if len(templates) != len(usTemplates) {
		t.Errorf("GetSeriesTemplates with count > pool size should return all templates, got %d, want %d",
			len(templates), len(usTemplates))
	}
}

func TestGetDefaultSeriesCount(t *testing.T) {
	tests := []struct {
		modality Modality
		want     int
	}{
		{MR, 4},
		{CT, 3},
		{CR, 2},
		{DX, 2},
		{US, 2},
		{MG, 4},
	}

	for _, tt := range tests {
		t.Run(string(tt.modality), func(t *testing.T) {
			got := GetDefaultSeriesCount(tt.modality)
			if got != tt.want {
				t.Errorf("GetDefaultSeriesCount(%s) = %d, want %d", tt.modality, got, tt.want)
			}
		})
	}
}

func TestSeriesTemplatesHaveRequiredFields(t *testing.T) {
	allTemplates := map[string][]SeriesTemplate{
		"mrBrain":           mrBrainTemplates,
		"mrKnee":            mrKneeTemplates,
		"mrSpine":           mrSpineTemplates,
		"mrAbdomen":         mrAbdomenTemplates,
		"ctWithContrast":    ctWithContrastTemplates,
		"ctWithoutContrast": ctWithoutContrastTemplates,
		"crDX":              crDXTemplates,
		"us":                usTemplates,
		"mg":                mgTemplates,
	}

	for name, templates := range allTemplates {
		t.Run(name, func(t *testing.T) {
			for i, tmpl := range templates {
				if tmpl.SeriesDescription == "" {
					t.Errorf("%s[%d] has empty SeriesDescription", name, i)
				}
				if tmpl.Orientation == "" {
					t.Errorf("%s[%d] has empty Orientation", name, i)
				}
				if tmpl.HasContrast && tmpl.ContrastAgent == "" {
					t.Errorf("%s[%d] has HasContrast=true but no ContrastAgent", name, i)
				}
			}
		})
	}
}
