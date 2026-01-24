package modalities

import (
	"math/rand/v2"
	"testing"
)

func TestGetGenerator_MR(t *testing.T) {
	gen := GetGenerator(MR)
	if gen.Modality() != MR {
		t.Errorf("Expected MR modality, got %v", gen.Modality())
	}
	if gen.SOPClassUID() != "1.2.840.10008.5.1.4.1.1.4" {
		t.Errorf("Unexpected MR SOP Class UID: %s", gen.SOPClassUID())
	}
}

func TestGetGenerator_CT(t *testing.T) {
	gen := GetGenerator(CT)
	if gen.Modality() != CT {
		t.Errorf("Expected CT modality, got %v", gen.Modality())
	}
	if gen.SOPClassUID() != "1.2.840.10008.5.1.4.1.1.2" {
		t.Errorf("Unexpected CT SOP Class UID: %s", gen.SOPClassUID())
	}
}

func TestGetGenerator_Default(t *testing.T) {
	gen := GetGenerator(Modality("UNKNOWN"))
	if gen.Modality() != MR {
		t.Errorf("Unknown modality should default to MR, got %v", gen.Modality())
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"MR", true},
		{"CT", true},
		{"CR", true},
		{"DX", true},
		{"US", true},
		{"MG", true},
		{"mr", false}, // case sensitive
		{"ct", false},
		{"UNKNOWN", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsValid(tt.input)
			if got != tt.want {
				t.Errorf("IsValid(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAllModalities(t *testing.T) {
	mods := AllModalities()
	if len(mods) != 6 {
		t.Errorf("Expected 6 modalities, got %d", len(mods))
	}

	// Verify all modalities are present
	expected := map[Modality]bool{MR: false, CT: false, CR: false, DX: false, US: false, MG: false}
	for _, m := range mods {
		if _, ok := expected[m]; ok {
			expected[m] = true
		}
	}

	for mod, found := range expected {
		if !found {
			t.Errorf("%s modality not found", mod)
		}
	}
}

func TestMRGenerator_Scanners(t *testing.T) {
	gen := &MRGenerator{}
	scanners := gen.Scanners()

	if len(scanners) == 0 {
		t.Fatal("Expected at least one MR scanner")
	}

	// Verify all scanners have required fields
	for i, s := range scanners {
		if s.Manufacturer == "" {
			t.Errorf("Scanner %d has empty manufacturer", i)
		}
		if s.Model == "" {
			t.Errorf("Scanner %d has empty model", i)
		}
		if s.FieldStrength <= 0 {
			t.Errorf("Scanner %d has invalid field strength: %f", i, s.FieldStrength)
		}
	}
}

func TestMRGenerator_GenerateSeriesParams(t *testing.T) {
	gen := &MRGenerator{}
	rng := rand.New(rand.NewPCG(42, 42))
	scanner := Scanner{Manufacturer: "SIEMENS", Model: "Test", FieldStrength: 1.5}

	params := gen.GenerateSeriesParams(scanner, rng)

	// Verify MR-specific params are set
	if params.Modality != MR {
		t.Errorf("Expected MR modality, got %v", params.Modality)
	}
	if params.PixelSpacing <= 0 {
		t.Errorf("Invalid PixelSpacing: %f", params.PixelSpacing)
	}
	if params.SliceThickness <= 0 {
		t.Errorf("Invalid SliceThickness: %f", params.SliceThickness)
	}
	if params.EchoTime <= 0 {
		t.Errorf("Invalid EchoTime: %f", params.EchoTime)
	}
	if params.RepetitionTime <= 0 {
		t.Errorf("Invalid RepetitionTime: %f", params.RepetitionTime)
	}
	if params.FlipAngle <= 0 {
		t.Errorf("Invalid FlipAngle: %f", params.FlipAngle)
	}
	if params.MagneticFieldStrength != 1.5 {
		t.Errorf("Expected 1.5T field strength, got %f", params.MagneticFieldStrength)
	}
}

func TestMRGenerator_PixelConfig(t *testing.T) {
	gen := &MRGenerator{}
	cfg := gen.PixelConfig()

	if cfg.BitsAllocated != 16 {
		t.Errorf("Expected 16 bits allocated, got %d", cfg.BitsAllocated)
	}
	if cfg.BitsStored != 12 {
		t.Errorf("Expected 12 bits stored, got %d", cfg.BitsStored)
	}
	if cfg.HighBit != 11 {
		t.Errorf("Expected high bit 11, got %d", cfg.HighBit)
	}
	if cfg.PixelRepresentation != 0 {
		t.Errorf("MR should use unsigned pixels, got %d", cfg.PixelRepresentation)
	}
}

func TestCTGenerator_Scanners(t *testing.T) {
	gen := &CTGenerator{}
	scanners := gen.Scanners()

	if len(scanners) == 0 {
		t.Fatal("Expected at least one CT scanner")
	}

	// Verify all scanners have required fields
	for i, s := range scanners {
		if s.Manufacturer == "" {
			t.Errorf("Scanner %d has empty manufacturer", i)
		}
		if s.Model == "" {
			t.Errorf("Scanner %d has empty model", i)
		}
		if s.DetectorRows <= 0 {
			t.Errorf("Scanner %d has invalid detector rows: %d", i, s.DetectorRows)
		}
	}
}

func TestCTGenerator_GenerateSeriesParams(t *testing.T) {
	gen := &CTGenerator{}
	rng := rand.New(rand.NewPCG(42, 42))
	scanner := Scanner{Manufacturer: "SIEMENS", Model: "Test", DetectorRows: 128}

	params := gen.GenerateSeriesParams(scanner, rng)

	// Verify CT-specific params are set
	if params.Modality != CT {
		t.Errorf("Expected CT modality, got %v", params.Modality)
	}
	if params.PixelSpacing <= 0 {
		t.Errorf("Invalid PixelSpacing: %f", params.PixelSpacing)
	}
	if params.SliceThickness <= 0 {
		t.Errorf("Invalid SliceThickness: %f", params.SliceThickness)
	}

	// CT-specific params
	validKVP := []float64{80, 100, 120, 140}
	found := false
	for _, v := range validKVP {
		if params.KVP == v {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Invalid KVP: %f", params.KVP)
	}

	if params.XRayTubeCurrent < 100 || params.XRayTubeCurrent > 400 {
		t.Errorf("Invalid XRayTubeCurrent: %d", params.XRayTubeCurrent)
	}

	if params.RescaleIntercept != -1024 {
		t.Errorf("Expected CT RescaleIntercept -1024, got %f", params.RescaleIntercept)
	}

	if params.RescaleSlope != 1 {
		t.Errorf("Expected CT RescaleSlope 1, got %f", params.RescaleSlope)
	}
}

func TestCTGenerator_PixelConfig(t *testing.T) {
	gen := &CTGenerator{}
	cfg := gen.PixelConfig()

	if cfg.BitsAllocated != 16 {
		t.Errorf("Expected 16 bits allocated, got %d", cfg.BitsAllocated)
	}
	if cfg.BitsStored != 16 {
		t.Errorf("Expected 16 bits stored for CT, got %d", cfg.BitsStored)
	}
	if cfg.HighBit != 15 {
		t.Errorf("Expected high bit 15, got %d", cfg.HighBit)
	}
	if cfg.PixelRepresentation != 1 {
		t.Errorf("CT should use signed pixels for HU, got %d", cfg.PixelRepresentation)
	}
	if cfg.MinValue != -1024 {
		t.Errorf("CT MinValue should be -1024, got %d", cfg.MinValue)
	}
}

func TestCTGenerator_WindowPresets(t *testing.T) {
	gen := &CTGenerator{}
	presets := gen.WindowPresets()

	if len(presets) == 0 {
		t.Fatal("Expected at least one window preset")
	}

	// Check for common CT presets
	presetNames := make(map[string]bool)
	for _, p := range presets {
		presetNames[p.Name] = true
		if p.Width <= 0 {
			t.Errorf("Preset %s has invalid width: %f", p.Name, p.Width)
		}
	}

	expectedPresets := []string{"BRAIN", "BONE", "LUNG"}
	for _, name := range expectedPresets {
		if !presetNames[name] {
			t.Errorf("Expected preset %s not found", name)
		}
	}
}

func TestMRGenerator_WindowPresets(t *testing.T) {
	gen := &MRGenerator{}
	presets := gen.WindowPresets()

	if len(presets) == 0 {
		t.Fatal("Expected at least one window preset")
	}

	for _, p := range presets {
		if p.Width <= 0 {
			t.Errorf("Preset %s has invalid width: %f", p.Name, p.Width)
		}
	}
}

// CR Generator Tests
func TestGetGenerator_CR(t *testing.T) {
	gen := GetGenerator(CR)
	if gen.Modality() != CR {
		t.Errorf("Expected CR modality, got %v", gen.Modality())
	}
	if gen.SOPClassUID() != "1.2.840.10008.5.1.4.1.1.1" {
		t.Errorf("Unexpected CR SOP Class UID: %s", gen.SOPClassUID())
	}
}

func TestCRGenerator_Scanners(t *testing.T) {
	gen := &CRGenerator{}
	scanners := gen.Scanners()

	if len(scanners) == 0 {
		t.Fatal("Expected at least one CR scanner")
	}

	for i, s := range scanners {
		if s.Manufacturer == "" {
			t.Errorf("Scanner %d has empty manufacturer", i)
		}
		if s.Model == "" {
			t.Errorf("Scanner %d has empty model", i)
		}
	}
}

func TestCRGenerator_GenerateSeriesParams(t *testing.T) {
	gen := &CRGenerator{}
	rng := rand.New(rand.NewPCG(42, 42))
	scanner := Scanner{Manufacturer: "FUJIFILM", Model: "FCR Profect CS"}

	params := gen.GenerateSeriesParams(scanner, rng)

	if params.Modality != CR {
		t.Errorf("Expected CR modality, got %v", params.Modality)
	}
	if params.PixelSpacing <= 0 {
		t.Errorf("Invalid PixelSpacing: %f", params.PixelSpacing)
	}
	if params.ViewPosition == "" {
		t.Error("ViewPosition should be set")
	}
	if params.ImagerPixelSpacing <= 0 {
		t.Errorf("Invalid ImagerPixelSpacing: %f", params.ImagerPixelSpacing)
	}
	if params.DistanceSourceToDetector <= 0 {
		t.Errorf("Invalid DistanceSourceToDetector: %f", params.DistanceSourceToDetector)
	}
	if params.Exposure <= 0 {
		t.Errorf("Invalid Exposure: %d", params.Exposure)
	}
}

func TestCRGenerator_PixelConfig(t *testing.T) {
	gen := &CRGenerator{}
	cfg := gen.PixelConfig()

	if cfg.BitsAllocated != 16 {
		t.Errorf("Expected 16 bits allocated, got %d", cfg.BitsAllocated)
	}
	if cfg.BitsStored != 12 {
		t.Errorf("Expected 12 bits stored for CR, got %d", cfg.BitsStored)
	}
	if cfg.HighBit != 11 {
		t.Errorf("Expected high bit 11, got %d", cfg.HighBit)
	}
	if cfg.PixelRepresentation != 0 {
		t.Errorf("CR should use unsigned pixels, got %d", cfg.PixelRepresentation)
	}
}

// DX Generator Tests
func TestGetGenerator_DX(t *testing.T) {
	gen := GetGenerator(DX)
	if gen.Modality() != DX {
		t.Errorf("Expected DX modality, got %v", gen.Modality())
	}
	if gen.SOPClassUID() != "1.2.840.10008.5.1.4.1.1.1.1" {
		t.Errorf("Unexpected DX SOP Class UID: %s", gen.SOPClassUID())
	}
}

func TestDXGenerator_Scanners(t *testing.T) {
	gen := &DXGenerator{}
	scanners := gen.Scanners()

	if len(scanners) == 0 {
		t.Fatal("Expected at least one DX scanner")
	}

	for i, s := range scanners {
		if s.Manufacturer == "" {
			t.Errorf("Scanner %d has empty manufacturer", i)
		}
		if s.Model == "" {
			t.Errorf("Scanner %d has empty model", i)
		}
	}
}

func TestDXGenerator_GenerateSeriesParams(t *testing.T) {
	gen := &DXGenerator{}
	rng := rand.New(rand.NewPCG(42, 42))
	scanner := Scanner{Manufacturer: "SIEMENS", Model: "Ysio Max"}

	params := gen.GenerateSeriesParams(scanner, rng)

	if params.Modality != DX {
		t.Errorf("Expected DX modality, got %v", params.Modality)
	}
	if params.PixelSpacing <= 0 {
		t.Errorf("Invalid PixelSpacing: %f", params.PixelSpacing)
	}
	if params.ViewPosition == "" {
		t.Error("ViewPosition should be set")
	}
	if params.KVP < 60 || params.KVP > 140 {
		t.Errorf("Invalid KVP: %f", params.KVP)
	}
	if params.ExposureTime <= 0 {
		t.Errorf("Invalid ExposureTime: %d", params.ExposureTime)
	}
}

func TestDXGenerator_PixelConfig(t *testing.T) {
	gen := &DXGenerator{}
	cfg := gen.PixelConfig()

	if cfg.BitsAllocated != 16 {
		t.Errorf("Expected 16 bits allocated, got %d", cfg.BitsAllocated)
	}
	if cfg.BitsStored != 14 {
		t.Errorf("Expected 14 bits stored for DX, got %d", cfg.BitsStored)
	}
	if cfg.HighBit != 13 {
		t.Errorf("Expected high bit 13, got %d", cfg.HighBit)
	}
	if cfg.PixelRepresentation != 0 {
		t.Errorf("DX should use unsigned pixels, got %d", cfg.PixelRepresentation)
	}
}

// US Generator Tests
func TestGetGenerator_US(t *testing.T) {
	gen := GetGenerator(US)
	if gen.Modality() != US {
		t.Errorf("Expected US modality, got %v", gen.Modality())
	}
	if gen.SOPClassUID() != "1.2.840.10008.5.1.4.1.1.6.1" {
		t.Errorf("Unexpected US SOP Class UID: %s", gen.SOPClassUID())
	}
}

func TestUSGenerator_Scanners(t *testing.T) {
	gen := &USGenerator{}
	scanners := gen.Scanners()

	if len(scanners) == 0 {
		t.Fatal("Expected at least one US scanner")
	}

	for i, s := range scanners {
		if s.Manufacturer == "" {
			t.Errorf("Scanner %d has empty manufacturer", i)
		}
		if s.Model == "" {
			t.Errorf("Scanner %d has empty model", i)
		}
	}
}

func TestUSGenerator_GenerateSeriesParams(t *testing.T) {
	gen := &USGenerator{}
	rng := rand.New(rand.NewPCG(42, 42))
	scanner := Scanner{Manufacturer: "GE MEDICAL SYSTEMS", Model: "LOGIQ E10"}

	params := gen.GenerateSeriesParams(scanner, rng)

	if params.Modality != US {
		t.Errorf("Expected US modality, got %v", params.Modality)
	}
	if params.PixelSpacing <= 0 {
		t.Errorf("Invalid PixelSpacing: %f", params.PixelSpacing)
	}
	if params.TransducerType == "" {
		t.Error("TransducerType should be set")
	}
	validTypes := map[string]bool{"LINEAR": true, "CONVEX": true, "PHASED": true}
	if !validTypes[params.TransducerType] {
		t.Errorf("Invalid TransducerType: %s", params.TransducerType)
	}
	if params.TransducerFrequency <= 0 {
		t.Errorf("Invalid TransducerFrequency: %f", params.TransducerFrequency)
	}
}

func TestUSGenerator_PixelConfig(t *testing.T) {
	gen := &USGenerator{}
	cfg := gen.PixelConfig()

	if cfg.BitsAllocated != 8 {
		t.Errorf("Expected 8 bits allocated for US, got %d", cfg.BitsAllocated)
	}
	if cfg.BitsStored != 8 {
		t.Errorf("Expected 8 bits stored for US, got %d", cfg.BitsStored)
	}
	if cfg.HighBit != 7 {
		t.Errorf("Expected high bit 7, got %d", cfg.HighBit)
	}
	if cfg.PixelRepresentation != 0 {
		t.Errorf("US should use unsigned pixels, got %d", cfg.PixelRepresentation)
	}
	if cfg.MaxValue != 255 {
		t.Errorf("US MaxValue should be 255, got %d", cfg.MaxValue)
	}
}

// MG Generator Tests
func TestGetGenerator_MG(t *testing.T) {
	gen := GetGenerator(MG)
	if gen.Modality() != MG {
		t.Errorf("Expected MG modality, got %v", gen.Modality())
	}
	if gen.SOPClassUID() != "1.2.840.10008.5.1.4.1.1.1.2" {
		t.Errorf("Unexpected MG SOP Class UID: %s", gen.SOPClassUID())
	}
}

func TestMGGenerator_Scanners(t *testing.T) {
	gen := &MGGenerator{}
	scanners := gen.Scanners()

	if len(scanners) == 0 {
		t.Fatal("Expected at least one MG scanner")
	}

	for i, s := range scanners {
		if s.Manufacturer == "" {
			t.Errorf("Scanner %d has empty manufacturer", i)
		}
		if s.Model == "" {
			t.Errorf("Scanner %d has empty model", i)
		}
	}
}

func TestMGGenerator_GenerateSeriesParams(t *testing.T) {
	gen := &MGGenerator{}
	rng := rand.New(rand.NewPCG(42, 42))
	scanner := Scanner{Manufacturer: "HOLOGIC", Model: "Selenia Dimensions"}

	params := gen.GenerateSeriesParams(scanner, rng)

	if params.Modality != MG {
		t.Errorf("Expected MG modality, got %v", params.Modality)
	}
	if params.PixelSpacing <= 0 {
		t.Errorf("Invalid PixelSpacing: %f", params.PixelSpacing)
	}
	if params.ImageLaterality != "L" && params.ImageLaterality != "R" {
		t.Errorf("Invalid ImageLaterality: %s", params.ImageLaterality)
	}
	validViews := map[string]bool{"CC": true, "MLO": true, "ML": true, "LM": true}
	if !validViews[params.ViewPosition] {
		t.Errorf("Invalid ViewPosition: %s", params.ViewPosition)
	}
	validAnodes := map[string]bool{"MOLYBDENUM": true, "RHODIUM": true, "TUNGSTEN": true}
	if !validAnodes[params.AnodeTargetMaterial] {
		t.Errorf("Invalid AnodeTargetMaterial: %s", params.AnodeTargetMaterial)
	}
	if params.CompressionForce < 80 || params.CompressionForce > 200 {
		t.Errorf("Invalid CompressionForce: %f", params.CompressionForce)
	}
	if params.OrganDose <= 0 {
		t.Errorf("Invalid OrganDose: %f", params.OrganDose)
	}
}

func TestMGGenerator_PixelConfig(t *testing.T) {
	gen := &MGGenerator{}
	cfg := gen.PixelConfig()

	if cfg.BitsAllocated != 16 {
		t.Errorf("Expected 16 bits allocated, got %d", cfg.BitsAllocated)
	}
	if cfg.BitsStored != 14 {
		t.Errorf("Expected 14 bits stored for MG, got %d", cfg.BitsStored)
	}
	if cfg.HighBit != 13 {
		t.Errorf("Expected high bit 13, got %d", cfg.HighBit)
	}
	if cfg.PixelRepresentation != 0 {
		t.Errorf("MG should use unsigned pixels, got %d", cfg.PixelRepresentation)
	}
}

func TestMGGenerator_WindowPresets(t *testing.T) {
	gen := &MGGenerator{}
	presets := gen.WindowPresets()

	if len(presets) == 0 {
		t.Fatal("Expected at least one window preset")
	}

	for _, p := range presets {
		if p.Width <= 0 {
			t.Errorf("Preset %s has invalid width: %f", p.Name, p.Width)
		}
	}
}
