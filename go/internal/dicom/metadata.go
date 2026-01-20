package dicom

import (
	"fmt"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// MetadataOptions contains all parameters needed to generate DICOM metadata
type MetadataOptions struct {
	NumImages      int
	Width          int
	Height         int
	InstanceNumber int

	// Shared across series
	StudyUID         string
	SeriesUID        string
	PatientID        string
	PatientName      string
	PatientBirthDate string
	PatientSex       string
	StudyDate        string
	StudyTime        string
	StudyID          string
	StudyDescription string
	AccessionNumber  string
	SeriesNumber     int

	// MRI parameters (shared across series)
	PixelSpacing         float64
	SliceThickness       float64
	SpacingBetweenSlices float64
	EchoTime             float64
	RepetitionTime       float64
	FlipAngle            float64
	SequenceName         string
	Manufacturer         string
	Model                string
	FieldStrength        float64
}

// formatFloat converts a float64 to a string suitable for DICOM numeric tags
func formatFloat(f float64) string {
	return fmt.Sprintf("%f", f)
}

// mustNewElement creates a DICOM element or panics on error.
// This simplifies element creation for test/development code.
// Panics if the tag or data is invalid (should only happen with programming errors).
func mustNewElement(t tag.Tag, data any) *dicom.Element {
	elem, err := dicom.NewElement(t, data)
	if err != nil {
		panic(err)
	}
	return elem
}

// GenerateMetadata creates a DICOM dataset with realistic MRI metadata.
// Panics if invalid tag data is provided (should only happen with programming errors).
func GenerateMetadata(opts MetadataOptions) *dicom.Dataset {
	// Create new dataset
	ds := &dicom.Dataset{
		Elements: []*dicom.Element{},
	}

	// Patient Information Module
	ds.Elements = append(ds.Elements, mustNewElement(tag.PatientName, []string{opts.PatientName}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.PatientID, []string{opts.PatientID}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.PatientBirthDate, []string{opts.PatientBirthDate}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.PatientSex, []string{opts.PatientSex}))

	// Study Information Module
	ds.Elements = append(ds.Elements, mustNewElement(tag.StudyInstanceUID, []string{opts.StudyUID}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.StudyDate, []string{opts.StudyDate}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.StudyTime, []string{opts.StudyTime}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.StudyID, []string{opts.StudyID}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.StudyDescription, []string{opts.StudyDescription}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.AccessionNumber, []string{opts.AccessionNumber}))

	// Series Information Module
	ds.Elements = append(ds.Elements, mustNewElement(tag.SeriesInstanceUID, []string{opts.SeriesUID}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.SeriesNumber, []string{fmt.Sprintf("%d", opts.SeriesNumber)}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.SeriesDescription, []string{"MRI Scan"}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.Modality, []string{"MR"}))

	// Instance Information Module
	ds.Elements = append(ds.Elements, mustNewElement(tag.InstanceNumber, []string{fmt.Sprintf("%d", opts.InstanceNumber)}))
	// SOP Class UID for MR Image Storage
	ds.Elements = append(ds.Elements, mustNewElement(tag.SOPClassUID, []string{"1.2.840.10008.5.1.4.1.1.4"}))

	// Image Pixel Module
	ds.Elements = append(ds.Elements, mustNewElement(tag.Rows, []int{opts.Height}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.Columns, []int{opts.Width}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.BitsAllocated, []int{16}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.BitsStored, []int{16}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.HighBit, []int{15}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.PixelRepresentation, []int{0}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.SamplesPerPixel, []int{1}))
	ds.Elements = append(ds.Elements, mustNewElement(tag.PhotometricInterpretation, []string{"MONOCHROME2"}))

	// MRI-specific tags (if manufacturer is set)
	if opts.Manufacturer != "" {
		ds.Elements = append(ds.Elements, mustNewElement(tag.Manufacturer, []string{opts.Manufacturer}))
	}
	if opts.Model != "" {
		ds.Elements = append(ds.Elements, mustNewElement(tag.ManufacturerModelName, []string{opts.Model}))
	}

	// MRI acquisition parameters (clinically significant)
	if opts.PixelSpacing != 0 {
		// PixelSpacing is stored as [row spacing, column spacing]
		ds.Elements = append(ds.Elements, mustNewElement(tag.PixelSpacing, []string{
			formatFloat(opts.PixelSpacing),
			formatFloat(opts.PixelSpacing),
		}))
	}
	if opts.SliceThickness != 0 {
		ds.Elements = append(ds.Elements, mustNewElement(tag.SliceThickness, []string{formatFloat(opts.SliceThickness)}))
	}
	if opts.SpacingBetweenSlices != 0 {
		ds.Elements = append(ds.Elements, mustNewElement(tag.SpacingBetweenSlices, []string{formatFloat(opts.SpacingBetweenSlices)}))
	}
	if opts.EchoTime != 0 {
		ds.Elements = append(ds.Elements, mustNewElement(tag.EchoTime, []string{formatFloat(opts.EchoTime)}))
	}
	if opts.RepetitionTime != 0 {
		ds.Elements = append(ds.Elements, mustNewElement(tag.RepetitionTime, []string{formatFloat(opts.RepetitionTime)}))
	}
	if opts.FlipAngle != 0 {
		ds.Elements = append(ds.Elements, mustNewElement(tag.FlipAngle, []string{formatFloat(opts.FlipAngle)}))
	}
	if opts.SequenceName != "" {
		ds.Elements = append(ds.Elements, mustNewElement(tag.SequenceName, []string{opts.SequenceName}))
	}
	if opts.FieldStrength != 0 {
		ds.Elements = append(ds.Elements, mustNewElement(tag.MagneticFieldStrength, []string{formatFloat(opts.FieldStrength)}))
	}

	return ds
}
