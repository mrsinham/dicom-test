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

	ds, err := GenerateMetadata(opts)
	if err != nil {
		t.Fatalf("GenerateMetadata failed: %v", err)
	}

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

	ds, err := GenerateMetadata(opts)
	if err != nil {
		t.Fatalf("GenerateMetadata failed: %v", err)
	}

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
