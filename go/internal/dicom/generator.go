package dicom

import (
	"fmt"
	"hash/fnv"
	"math"
	randv2 "math/rand/v2"
	"os"
	"path/filepath"

	"github.com/julien/dicom-test/go/internal/util"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// writeDatasetToFile writes a DICOM dataset to a file
func writeDatasetToFile(filename string, ds dicom.Dataset) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return dicom.Write(f, ds)
}

// GeneratorOptions contains all parameters needed to generate a DICOM series
type GeneratorOptions struct {
	NumImages  int
	TotalSize  string
	OutputDir  string
	Seed       int64
	NumStudies int
}

// GeneratedFile contains information about a generated DICOM file
type GeneratedFile struct {
	Path           string
	StudyUID       string
	SeriesUID      string
	SOPInstanceUID string
	PatientID      string
	StudyID        string
	SeriesNumber   int
	InstanceNumber int
}

// CalculateDimensions calculates optimal image dimensions based on total size and number of images
func CalculateDimensions(totalBytes int64, numImages int) (width, height int, err error) {
	if totalBytes <= 0 {
		return 0, 0, fmt.Errorf("total bytes must be > 0")
	}
	if numImages <= 0 {
		return 0, 0, fmt.Errorf("number of images must be > 0")
	}

	// Subtract metadata overhead (100KB estimate)
	metadataOverhead := int64(100 * 1024)
	availableBytes := totalBytes - metadataOverhead
	if availableBytes <= 0 {
		return 0, 0, fmt.Errorf("total size too small (need at least 100KB for metadata)")
	}

	// DICOM max size check (2^32 - 10MB ≈ 4.28GB)
	maxDICOMSize := int64(math.Pow(2, 32)) - 10*1024*1024
	if availableBytes > maxDICOMSize {
		availableBytes = maxDICOMSize
	}

	// Calculate total pixels: availableBytes / 2 (uint16 = 2 bytes per pixel)
	totalPixels := availableBytes / 2

	// Pixels per frame
	pixelsPerFrame := totalPixels / int64(numImages)

	// Dimension: sqrt(pixelsPerFrame)
	dimension := int(math.Sqrt(float64(pixelsPerFrame)))

	// Round DOWN to multiple of 256 (or 128 if < 256) to ensure we don't exceed size
	if dimension >= 256 {
		width = (dimension / 256) * 256
	} else if dimension >= 128 {
		width = 128
	} else {
		width = 128 // Minimum
	}

	height = width

	// Ensure minimum dimensions
	if width < 128 {
		width = 128
		height = 128
	}

	return width, height, nil
}

// GenerateDICOMSeries generates a complete DICOM series with multiple studies
func GenerateDICOMSeries(opts GeneratorOptions) ([]GeneratedFile, error) {
	// Parse total size
	totalBytes, err := util.ParseSize(opts.TotalSize)
	if err != nil {
		return nil, fmt.Errorf("invalid size: %w", err)
	}

	// Calculate dimensions
	width, height, err := CalculateDimensions(totalBytes, opts.NumImages)
	if err != nil {
		return nil, fmt.Errorf("calculate dimensions: %w", err)
	}

	fmt.Printf("Resolution: %dx%d pixels per image\n", width, height)

	// Create output directory
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	// Set seed for reproducibility
	var seed int64
	if opts.Seed != 0 {
		seed = opts.Seed
		fmt.Printf("Using seed: %d\n", seed)
	} else {
		// Generate deterministic seed from output directory name
		h := fnv.New64a()
		h.Write([]byte(opts.OutputDir))
		seed = int64(h.Sum64())
		fmt.Printf("Auto-generated seed from '%s': %d\n", opts.OutputDir, seed)
		fmt.Println("  (same directory = same patient/study IDs)")
	}

	// Create RNG for patient name generation
	rng := randv2.New(randv2.NewPCG(uint64(seed), uint64(seed)))

	// Generate shared patient info
	patientID := fmt.Sprintf("PID%06d", rng.IntN(900000)+100000)
	patientSex := []string{"M", "F"}[rng.IntN(2)]
	patientName := util.GeneratePatientName(patientSex, rng)
	patientBirthDate := fmt.Sprintf("%04d%02d%02d",
		rng.IntN(51)+1950, // 1950-2000
		rng.IntN(12)+1,    // 1-12
		rng.IntN(28)+1)    // 1-28

	fmt.Printf("Generating %d DICOM files...\n", opts.NumImages)
	fmt.Printf("Patient: %s (ID: %s, DOB: %s, Sex: %s)\n",
		patientName, patientID, patientBirthDate, patientSex)
	fmt.Printf("Number of studies: %d\n", opts.NumStudies)

	// Calculate images per study
	imagesPerStudy := opts.NumImages / opts.NumStudies
	remainingImages := opts.NumImages % opts.NumStudies

	var generatedFiles []GeneratedFile
	globalImageIndex := 1

	// Manufacturer options
	manufacturers := []struct {
		Name          string
		Model         string
		FieldStrength float64
	}{
		{"SIEMENS", "Avanto", 1.5},
		{"SIEMENS", "Skyra", 3.0},
		{"GE MEDICAL SYSTEMS", "Signa HDxt", 1.5},
		{"GE MEDICAL SYSTEMS", "Discovery MR750", 3.0},
		{"PHILIPS", "Achieva", 1.5},
		{"PHILIPS", "Ingenia", 3.0},
	}

	// Generate DICOM files for each study
	for studyNum := 1; studyNum <= opts.NumStudies; studyNum++ {
		// Generate deterministic UIDs for this study
		studyUID := util.GenerateDeterministicUID(fmt.Sprintf("%s_study_%d", opts.OutputDir, studyNum))
		seriesUID := util.GenerateDeterministicUID(fmt.Sprintf("%s_study_%d_series_1", opts.OutputDir, studyNum))

		// Generate study-specific info
		studyID := fmt.Sprintf("STD%04d", rng.IntN(9000)+1000)
		var studyDescription string
		if opts.NumStudies > 1 {
			studyDescription = fmt.Sprintf("Brain MRI - Study %d", studyNum)
		} else {
			studyDescription = "Brain MRI"
		}

		// Generate series-specific MRI parameters (same for all images in series)
		_ = rng.Float64()*1.5 + 0.5        // seriesPixelSpacing
		_ = rng.Float64()*4.0 + 1.0      // seriesSliceThickness
		// seriesSpacingBetweenSlices := seriesSliceThickness + rng.Float64()*0.5
		// seriesEchoTime := rng.Float64()*20.0 + 10.0          // 10-30
		// seriesRepetitionTime := rng.Float64()*400.0 + 400.0  // 400-800
		// seriesFlipAngle := rng.Float64()*30.0 + 60.0         // 60-90
		// seriesSequenceName := []string{"T1_MPRAGE", "T1_SE", "T2_FSE", "T2_FLAIR"}[rng.IntN(4)]

		// Select MRI scanner
		mfr := manufacturers[rng.IntN(len(manufacturers))]

		// Calculate images for this study
		numImagesThisStudy := imagesPerStudy
		if studyNum <= remainingImages {
			numImagesThisStudy++
		}

		fmt.Printf("\nStudy %d/%d: %d images\n", studyNum, opts.NumStudies, numImagesThisStudy)
		fmt.Printf("  StudyID: %s, Description: %s\n", studyID, studyDescription)
		fmt.Printf("  Scanner: %s %s (%.1fT)\n", mfr.Name, mfr.Model, mfr.FieldStrength)
		// fmt.Printf("  Parameters: PixelSpacing=%.2fmm, SliceThickness=%.2fmm\n",
		// 	seriesPixelSpacing, seriesSliceThickness)

		// Generate each DICOM file for this study
		for instanceInStudy := 1; instanceInStudy <= numImagesThisStudy; instanceInStudy++ {
			// Generate SOP Instance UID
			sopInstanceUID := util.GenerateDeterministicUID(
				fmt.Sprintf("%s_study_%d_instance_%d", opts.OutputDir, studyNum, instanceInStudy))

			// Generate metadata with essential fields
			metadata := &dicom.Dataset{
				Elements: []*dicom.Element{
					// File meta information (must be first)
					mustNewElement(tag.TransferSyntaxUID, []string{"1.2.840.10008.1.2.1"}),
					// Patient module
					mustNewElement(tag.PatientName, []string{patientName}),
					mustNewElement(tag.PatientID, []string{patientID}),
					mustNewElement(tag.PatientBirthDate, []string{patientBirthDate}),
					mustNewElement(tag.PatientSex, []string{patientSex}),
					// Study module
					mustNewElement(tag.StudyInstanceUID, []string{studyUID}),
					// mustNewElement(tag.StudyID, []string{studyID}),
					// mustNewElement(tag.StudyDescription, []string{studyDescription}),
					// Series module
					mustNewElement(tag.SeriesInstanceUID, []string{seriesUID}),
					mustNewElement(tag.SeriesNumber, []string{fmt.Sprintf("%d", 1)}),
					mustNewElement(tag.Modality, []string{"MR"}),
					// Instance module
					mustNewElement(tag.SOPInstanceUID, []string{sopInstanceUID}),
					mustNewElement(tag.SOPClassUID, []string{"1.2.840.10008.5.1.4.1.1.4"}),
					mustNewElement(tag.InstanceNumber, []string{fmt.Sprintf("%d", instanceInStudy)}),
					// Image pixel module
					mustNewElement(tag.Rows, []int{height}),
					mustNewElement(tag.Columns, []int{width}),
					mustNewElement(tag.BitsAllocated, []int{16}),
					mustNewElement(tag.BitsStored, []int{16}),
					mustNewElement(tag.HighBit, []int{15}),
					mustNewElement(tag.PixelRepresentation, []int{0}),
					mustNewElement(tag.SamplesPerPixel, []int{1}),
					mustNewElement(tag.PhotometricInterpretation, []string{"MONOCHROME2"}),
				},
			}

			// TODO: Add pixel data later - testing without it first

			// Write DICOM file
			filename := fmt.Sprintf("IMG%04d.dcm", globalImageIndex)
			filePath := filepath.Join(opts.OutputDir, filename)

			if err := writeDatasetToFile(filePath, dicom.Dataset{Elements: metadata.Elements}); err != nil {
				return nil, fmt.Errorf("write DICOM file %s: %w", filePath, err)
			}

			// Record generated file info
			generatedFiles = append(generatedFiles, GeneratedFile{
				Path:           filePath,
				StudyUID:       studyUID,
				SeriesUID:      seriesUID,
				SOPInstanceUID: sopInstanceUID,
				PatientID:      patientID,
				StudyID:        studyID,
				SeriesNumber:   1,
				InstanceNumber: instanceInStudy,
			})

			// Progress indicator
			if globalImageIndex%10 == 0 || globalImageIndex == opts.NumImages {
				progress := float64(globalImageIndex) / float64(opts.NumImages) * 100
				fmt.Printf("  Progress: %d/%d (%.0f%%)\n", globalImageIndex, opts.NumImages, progress)
			}

			globalImageIndex++
		}
	}

	fmt.Printf("\n✓ %d DICOM files created in: %s/\n", opts.NumImages, opts.OutputDir)

	return generatedFiles, nil
}
