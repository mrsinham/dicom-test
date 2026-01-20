package dicom

import (
	"testing"

	"github.com/suyashkumar/dicom/pkg/tag"
)

func TestGenerateMetadata_BasicStructure(t *testing.T) {
	opts := MetadataOptions{
		NumImages:      10,
		Width:          256,
		Height:         256,
		InstanceNumber: 1,
		PatientID:      "TEST123",
		PatientName:    "DOE^JOHN",
		StudyUID:       "1.2.3.4.5",
		SeriesUID:      "1.2.3.4.6",
	}

	ds := GenerateMetadata(opts)

	if ds == nil {
		t.Fatal("Expected non-nil dataset")
	}
}

func TestGenerateMetadata_RequiredTags(t *testing.T) {
	opts := MetadataOptions{
		NumImages:        10,
		Width:            256,
		Height:           256,
		InstanceNumber:   1,
		PatientID:        "TEST123",
		PatientName:      "DOE^JOHN",
		PatientBirthDate: "19800101",
		PatientSex:       "M",
		StudyUID:         "1.2.3.4.5",
		SeriesUID:        "1.2.3.4.6",
		StudyDate:        "20260111",
		StudyTime:        "120000",
		StudyID:          "STD001",
		StudyDescription: "Test Study",
		AccessionNumber:  "ACC001",
		SeriesNumber:     1,
	}

	ds := GenerateMetadata(opts)

	// Check patient tags exist
	patientName, err := ds.FindElementByTag(tag.PatientName)
	if err != nil {
		t.Error("PatientName tag not found")
	}
	if patientName == nil {
		t.Error("PatientName is nil")
	}

	// Check study tags exist
	studyUID, err := ds.FindElementByTag(tag.StudyInstanceUID)
	if err != nil {
		t.Error("StudyInstanceUID tag not found")
	}
	if studyUID == nil {
		t.Error("StudyInstanceUID is nil")
	}

	// Check image tags exist
	rows, err := ds.FindElementByTag(tag.Rows)
	if err != nil {
		t.Error("Rows tag not found")
	}
	if rows == nil {
		t.Error("Rows is nil")
	}
}

func TestGenerateMetadata_MRIParameters(t *testing.T) {
	opts := MetadataOptions{
		Width:                256,
		Height:               256,
		InstanceNumber:       1,
		PatientID:            "TEST123",
		PatientName:          "DOE^JOHN",
		StudyUID:             "1.2.3.4.5",
		SeriesUID:            "1.2.3.4.6",
		PixelSpacing:         0.9375,
		SliceThickness:       5.0,
		SpacingBetweenSlices: 6.0,
		EchoTime:             30.0,
		RepetitionTime:       2000.0,
		FlipAngle:            90.0,
		SequenceName:         "T1_MPRAGE",
		FieldStrength:        3.0,
	}

	ds := GenerateMetadata(opts)

	// Check MRI parameters are populated
	tests := []struct {
		name   string
		tag    tag.Tag
		exists bool
	}{
		{"PixelSpacing", tag.PixelSpacing, true},
		{"SliceThickness", tag.SliceThickness, true},
		{"SpacingBetweenSlices", tag.SpacingBetweenSlices, true},
		{"EchoTime", tag.EchoTime, true},
		{"RepetitionTime", tag.RepetitionTime, true},
		{"FlipAngle", tag.FlipAngle, true},
		{"SequenceName", tag.SequenceName, true},
		{"MagneticFieldStrength", tag.MagneticFieldStrength, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elem, err := ds.FindElementByTag(tt.tag)
			if tt.exists {
				if err != nil {
					t.Errorf("%s tag not found: %v", tt.name, err)
				}
				if elem == nil {
					t.Errorf("%s is nil", tt.name)
				}
			}
		})
	}
}

func TestGenerateMetadata_MRIParametersNotSet(t *testing.T) {
	// Test that MRI parameters are not added when not set (zero values)
	opts := MetadataOptions{
		Width:          256,
		Height:         256,
		InstanceNumber: 1,
		PatientID:      "TEST123",
		PatientName:    "DOE^JOHN",
		StudyUID:       "1.2.3.4.5",
		SeriesUID:      "1.2.3.4.6",
		// All MRI parameters intentionally left as zero values
	}

	ds := GenerateMetadata(opts)

	// These tags should not be present when values are zero/empty
	tags := []struct {
		name string
		tag  tag.Tag
	}{
		{"PixelSpacing", tag.PixelSpacing},
		{"SliceThickness", tag.SliceThickness},
		{"SpacingBetweenSlices", tag.SpacingBetweenSlices},
		{"EchoTime", tag.EchoTime},
		{"RepetitionTime", tag.RepetitionTime},
		{"FlipAngle", tag.FlipAngle},
		{"SequenceName", tag.SequenceName},
		{"MagneticFieldStrength", tag.MagneticFieldStrength},
	}

	for _, tt := range tags {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ds.FindElementByTag(tt.tag)
			if err == nil {
				t.Errorf("%s should not be present when not set", tt.name)
			}
		})
	}
}
