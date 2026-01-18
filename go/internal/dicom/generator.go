package dicom

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/julien/dicom-test/go/internal/image"
	"github.com/julien/dicom-test/go/internal/util"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

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

	rand.Seed(seed)

	// Generate shared patient info
	patientID := fmt.Sprintf("PID%06d", rand.Intn(900000)+100000)
	patientSex := []string{"M", "F"}[rand.Intn(2)]
	patientName := util.GeneratePatientName(patientSex)
	patientBirthDate := fmt.Sprintf("%04d%02d%02d",
		rand.Intn(51)+1950, // 1950-2000
		rand.Intn(12)+1,    // 1-12
		rand.Intn(28)+1)    // 1-28

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
		now := time.Now()
		studyDate := now.Format("20060102")
		studyTime := now.Format("150405")
		studyID := fmt.Sprintf("STD%04d", rand.Intn(9000)+1000)
		var studyDescription string
		if opts.NumStudies > 1 {
			studyDescription = fmt.Sprintf("Brain MRI - Study %d", studyNum)
		} else {
			studyDescription = "Brain MRI"
		}
		accessionNumber := fmt.Sprintf("ACC%06d", rand.Intn(900000)+100000)

		// Generate series-specific MRI parameters (same for all images in series)
		seriesPixelSpacing := rand.Float64()*1.5 + 0.5        // 0.5-2.0
		seriesSliceThickness := rand.Float64()*4.0 + 1.0      // 1.0-5.0
		seriesSpacingBetweenSlices := seriesSliceThickness + rand.Float64()*0.5
		seriesEchoTime := rand.Float64()*20.0 + 10.0          // 10-30
		seriesRepetitionTime := rand.Float64()*400.0 + 400.0  // 400-800
		seriesFlipAngle := rand.Float64()*30.0 + 60.0         // 60-90
		seriesSequenceName := []string{"T1_MPRAGE", "T1_SE", "T2_FSE", "T2_FLAIR"}[rand.Intn(4)]

		// Select MRI scanner
		mfr := manufacturers[rand.Intn(len(manufacturers))]

		// Calculate images for this study
		numImagesThisStudy := imagesPerStudy
		if studyNum <= remainingImages {
			numImagesThisStudy++
		}

		fmt.Printf("\nStudy %d/%d: %d images\n", studyNum, opts.NumStudies, numImagesThisStudy)
		fmt.Printf("  StudyID: %s, Description: %s\n", studyID, studyDescription)
		fmt.Printf("  Scanner: %s %s (%.1fT)\n", mfr.Name, mfr.Model, mfr.FieldStrength)
		fmt.Printf("  Parameters: PixelSpacing=%.2fmm, SliceThickness=%.2fmm\n",
			seriesPixelSpacing, seriesSliceThickness)

		// Generate each DICOM file for this study
		for instanceInStudy := 1; instanceInStudy <= numImagesThisStudy; instanceInStudy++ {
			// Generate SOP Instance UID
			sopInstanceUID := util.GenerateDeterministicUID(
				fmt.Sprintf("%s_study_%d_instance_%d", opts.OutputDir, studyNum, instanceInStudy))

			// Generate metadata
			metadata := GenerateMetadata(MetadataOptions{
				NumImages:            numImagesThisStudy,
				Width:                width,
				Height:               height,
				InstanceNumber:       instanceInStudy,
				StudyUID:             studyUID,
				SeriesUID:            seriesUID,
				PatientID:            patientID,
				PatientName:          patientName,
				PatientBirthDate:     patientBirthDate,
				PatientSex:           patientSex,
				StudyDate:            studyDate,
				StudyTime:            studyTime,
				StudyID:              studyID,
				StudyDescription:     studyDescription,
				AccessionNumber:      accessionNumber,
				SeriesNumber:         1,
				PixelSpacing:         seriesPixelSpacing,
				SliceThickness:       seriesSliceThickness,
				SpacingBetweenSlices: seriesSpacingBetweenSlices,
				EchoTime:             seriesEchoTime,
				RepetitionTime:       seriesRepetitionTime,
				FlipAngle:            seriesFlipAngle,
				SequenceName:         seriesSequenceName,
				Manufacturer:         mfr.Name,
				Model:                mfr.Model,
				FieldStrength:        mfr.FieldStrength,
			})

			// Add SOP Instance UID
			metadata.Elements = append(metadata.Elements,
				mustNewElement(tag.SOPInstanceUID, []string{sopInstanceUID}))

			// Add position information (slice location)
			slicePosition := float64(instanceInStudy-1) * seriesSpacingBetweenSlices
			metadata.Elements = append(metadata.Elements,
				mustNewElement(tag.ImagePositionPatient, []string{
					"0", "0", fmt.Sprintf("%f", slicePosition),
				}))
			metadata.Elements = append(metadata.Elements,
				mustNewElement(tag.ImageOrientationPatient, []string{
					"1", "0", "0", "0", "1", "0", // Axial orientation
				}))
			metadata.Elements = append(metadata.Elements,
				mustNewElement(tag.SliceLocation, []string{fmt.Sprintf("%f", slicePosition)}))

			// Generate pixel data with overlay
			pixelSeed := seed + int64(globalImageIndex)*1000
			pixels := image.GenerateSingleImage(width, height, pixelSeed)
			if err := image.AddTextOverlay(pixels, width, height, globalImageIndex, opts.NumImages); err != nil {
				return nil, fmt.Errorf("add text overlay: %w", err)
			}

			// Convert pixels to byte array (little endian)
			pixelData := make([]byte, len(pixels)*2)
			for i, p := range pixels {
				pixelData[i*2] = byte(p)         // Low byte
				pixelData[i*2+1] = byte(p >> 8) // High byte
			}

			// Add pixel data element
			pixelDataElem, err := dicom.NewElement(tag.PixelData, dicom.PixelDataInfo{
				IsEncapsulated: false,
				Frames: []*dicom.Frame{
					{
						Encapsulated: false,
						NativeData: dicom.NativeFrame{
							BitsPerSample: 16,
							Rows:          height,
							Cols:          width,
							Data:          pixelData,
						},
					},
				},
			})
			if err != nil {
				return nil, fmt.Errorf("create pixel data element: %w", err)
			}
			metadata.Elements = append(metadata.Elements, pixelDataElem)

			// Write DICOM file
			filename := fmt.Sprintf("IMG%04d.dcm", globalImageIndex)
			filepath := filepath.Join(opts.OutputDir, filename)

			if err := dicom.WriteDatasetToFile(filepath, *metadata); err != nil {
				return nil, fmt.Errorf("write DICOM file %s: %w", filepath, err)
			}

			// Record generated file info
			generatedFiles = append(generatedFiles, GeneratedFile{
				Path:           filepath,
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
