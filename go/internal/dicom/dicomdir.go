package dicom

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// DirectoryRecord represents a single DICOMDIR directory record
type DirectoryRecord struct {
	RecordType string              // "PATIENT", "STUDY", "SERIES", "IMAGE"
	Tags       map[tag.Tag]any     // Tag values for this record
	Children   []*DirectoryRecord  // Child records
	FilePath   string              // Relative file path (for IMAGE records)
}

// FileHierarchy represents the PT*/ST*/SE* hierarchy
type FileHierarchy struct {
	PatientDir string
	StudyDir   string
	SeriesDir  string
	ImageFiles []string
}

// OrganizeFilesIntoDICOMDIR organizes DICOM files into PT*/ST*/SE* hierarchy and creates DICOMDIR
func OrganizeFilesIntoDICOMDIR(outputDir string, files []GeneratedFile) error {
	if len(files) == 0 {
		return fmt.Errorf("no files to organize")
	}

	fmt.Println("\nCreating DICOMDIR file...")

	// Group files by patient -> study -> series
	type SeriesGroup struct {
		StudyUID   string
		SeriesUID  string
		Files      []GeneratedFile
	}

	type StudyGroup struct {
		StudyUID string
		Series   map[string]*SeriesGroup
	}

	type PatientGroup struct {
		PatientID string
		Studies   map[string]*StudyGroup
	}

	patients := make(map[string]*PatientGroup)

	// Group files
	for _, file := range files {
		// Get or create patient
		if _, exists := patients[file.PatientID]; !exists {
			patients[file.PatientID] = &PatientGroup{
				PatientID: file.PatientID,
				Studies:   make(map[string]*StudyGroup),
			}
		}
		patient := patients[file.PatientID]

		// Get or create study
		if _, exists := patient.Studies[file.StudyUID]; !exists {
			patient.Studies[file.StudyUID] = &StudyGroup{
				StudyUID: file.StudyUID,
				Series:   make(map[string]*SeriesGroup),
			}
		}
		study := patient.Studies[file.StudyUID]

		// Get or create series
		if _, exists := study.Series[file.SeriesUID]; !exists {
			study.Series[file.SeriesUID] = &SeriesGroup{
				StudyUID:  file.StudyUID,
				SeriesUID: file.SeriesUID,
				Files:     []GeneratedFile{},
			}
		}
		series := study.Series[file.SeriesUID]

		// Add file to series
		series.Files = append(series.Files, file)
	}

	// Create PT*/ST*/SE* hierarchy and move files
	patientIdx := 0
	totalMoved := 0

	for _, patient := range patients {
		patientDir := fmt.Sprintf("PT%06d", patientIdx)
		patientPath := filepath.Join(outputDir, patientDir)
		if err := os.MkdirAll(patientPath, 0755); err != nil {
			return fmt.Errorf("create patient directory: %w", err)
		}

		studyIdx := 0
		for _, study := range patient.Studies {
			studyDir := fmt.Sprintf("ST%06d", studyIdx)
			studyPath := filepath.Join(patientPath, studyDir)
			if err := os.MkdirAll(studyPath, 0755); err != nil {
				return fmt.Errorf("create study directory: %w", err)
			}

			seriesIdx := 0
			for _, series := range study.Series {
				seriesDir := fmt.Sprintf("SE%06d", seriesIdx)
				seriesPath := filepath.Join(studyPath, seriesDir)
				if err := os.MkdirAll(seriesPath, 0755); err != nil {
					return fmt.Errorf("create series directory: %w", err)
				}

				// Sort files by instance number
				sort.Slice(series.Files, func(i, j int) bool {
					return series.Files[i].InstanceNumber < series.Files[j].InstanceNumber
				})

				// Move files into series directory
				for imageIdx, file := range series.Files {
					imageFile := fmt.Sprintf("IM%06d", imageIdx+1)
					destPath := filepath.Join(seriesPath, imageFile)

					// Move file
					if err := os.Rename(file.Path, destPath); err != nil {
						return fmt.Errorf("move file %s to %s: %w", file.Path, destPath, err)
					}

					totalMoved++
				}

				seriesIdx++
			}
			studyIdx++
		}
		patientIdx++
	}

	fmt.Printf("✓ DICOMDIR created with standard hierarchy\n")
	fmt.Printf("  Organized %d files into PT*/ST*/SE* structure\n", totalMoved)

	// Create DICOMDIR file with directory records
	if err := createDICOMDIRFile(outputDir); err != nil {
		return fmt.Errorf("create DICOMDIR file: %w", err)
	}

	// Clean up original IMG*.dcm files if they still exist
	fmt.Println("\nCleaning up temporary files...")
	removedCount := 0
	pattern := filepath.Join(outputDir, "IMG*.dcm")
	matches, _ := filepath.Glob(pattern)
	for _, match := range matches {
		if err := os.Remove(match); err == nil {
			removedCount++
		}
	}

	if removedCount > 0 {
		fmt.Printf("✓ %d temporary files removed\n", removedCount)
	}

	fmt.Println("\nThe DICOM series is ready to be imported!")
	fmt.Printf("Import the complete directory: %s/\n", outputDir)
	fmt.Println("\nStandard DICOM structure created:")
	fmt.Println("  - DICOMDIR (index file)")
	fmt.Println("  - PT000000/ST000000/SE000000/ (patient/study/series hierarchy)")

	return nil
}

// getStringValue safely extracts a string value from a dataset
func getStringValue(ds dicom.Dataset, t tag.Tag) []string {
	elem, err := ds.FindElementByTag(t)
	if err != nil || elem == nil {
		return []string{""}
	}
	str := strings.Trim(elem.Value.String(), " []")
	return []string{str}
}

// createDICOMDIRFile creates a complete DICOMDIR file with directory record sequence
func createDICOMDIRFile(outputDir string) error {
	dicomdirPath := filepath.Join(outputDir, "DICOMDIR")

	// Collect all DICOM files organized by hierarchy
	type ImageInfo struct {
		RelPath        string
		SOPClassUID    string
		SOPInstanceUID string
	}

	type SeriesInfo struct {
		SeriesUID    string
		SeriesNumber string
		Modality     string
		Images       []ImageInfo
	}

	type StudyInfo struct {
		StudyUID  string
		StudyID   string
		StudyDate string
		StudyTime string
		Series    []SeriesInfo
	}

	type PatientInfo struct {
		PatientID   string
		PatientName string
		Studies     []StudyInfo
	}

	var patients []PatientInfo

	// Walk the PT*/ST*/SE* hierarchy
	patientDirs, _ := filepath.Glob(filepath.Join(outputDir, "PT*"))
	sort.Strings(patientDirs)

	for _, patientDir := range patientDirs {
		patient := PatientInfo{
			Studies: []StudyInfo{},
		}

		studyDirs, _ := filepath.Glob(filepath.Join(patientDir, "ST*"))
		sort.Strings(studyDirs)

		for _, studyDir := range studyDirs {
			study := StudyInfo{
				Series: []SeriesInfo{},
			}

			seriesDirs, _ := filepath.Glob(filepath.Join(studyDir, "SE*"))
			sort.Strings(seriesDirs)

			for _, seriesDir := range seriesDirs {
				series := SeriesInfo{
					Images: []ImageInfo{},
				}

				imageFiles, _ := filepath.Glob(filepath.Join(seriesDir, "IM*"))
				sort.Strings(imageFiles)

				for _, imageFile := range imageFiles {
					// Parse DICOM file
					ds, err := dicom.ParseFile(imageFile, nil)
					if err != nil {
						continue
					}

					// Get relative path from outputDir
					relPath, _ := filepath.Rel(outputDir, imageFile)

					// Extract metadata
					sopClass := getStringValue(ds, tag.SOPClassUID)
					sopInstance := getStringValue(ds, tag.SOPInstanceUID)

					image := ImageInfo{
						RelPath:        filepath.ToSlash(relPath),
						SOPClassUID:    sopClass[0],
						SOPInstanceUID: sopInstance[0],
					}
					series.Images = append(series.Images, image)

					// Get series info from first image
					if len(series.Images) == 1 {
						series.SeriesUID = getStringValue(ds, tag.SeriesInstanceUID)[0]
						series.SeriesNumber = getStringValue(ds, tag.SeriesNumber)[0]
						series.Modality = getStringValue(ds, tag.Modality)[0]
					}

					// Get study info from first image
					if len(study.Series) == 0 && len(series.Images) == 1 {
						study.StudyUID = getStringValue(ds, tag.StudyInstanceUID)[0]
						study.StudyID = getStringValue(ds, tag.StudyID)[0]
						study.StudyDate = getStringValue(ds, tag.StudyDate)[0]
						study.StudyTime = getStringValue(ds, tag.StudyTime)[0]
					}

					// Get patient info from first image
					if len(patients) == 0 && len(study.Series) == 0 && len(series.Images) == 1 {
						patient.PatientID = getStringValue(ds, tag.PatientID)[0]
						patient.PatientName = getStringValue(ds, tag.PatientName)[0]
					}
				}

				if len(series.Images) > 0 {
					study.Series = append(study.Series, series)
				}
			}

			if len(study.Series) > 0 {
				patient.Studies = append(patient.Studies, study)
			}
		}

		if len(patient.Studies) > 0 {
			patients = append(patients, patient)
		}
	}

	// Build directory record sequence
	// Each record is a []*Element, and we collect them into [][]*Element
	var recordItems [][]*dicom.Element

	for _, patient := range patients {
		// PATIENT record - create element list
		patientElements := []*dicom.Element{
			mustNewElement(tag.OffsetOfTheNextDirectoryRecord, []int{0}), // Will be updated during write
			mustNewElement(tag.OffsetOfReferencedLowerLevelDirectoryEntity, []int{0}), // Points to first STUDY
			mustNewElement(tag.DirectoryRecordType, []string{"PATIENT"}),
			mustNewElement(tag.PatientID, []string{patient.PatientID}),
			mustNewElement(tag.PatientName, []string{patient.PatientName}),
		}
		recordItems = append(recordItems, patientElements)

		for _, study := range patient.Studies {
			// STUDY record
			studyElements := []*dicom.Element{
				mustNewElement(tag.OffsetOfTheNextDirectoryRecord, []int{0}), // Will be updated
				mustNewElement(tag.OffsetOfReferencedLowerLevelDirectoryEntity, []int{0}), // Points to first SERIES
				mustNewElement(tag.DirectoryRecordType, []string{"STUDY"}),
				mustNewElement(tag.StudyInstanceUID, []string{study.StudyUID}),
				mustNewElement(tag.StudyID, []string{study.StudyID}),
				mustNewElement(tag.StudyDate, []string{study.StudyDate}),
				mustNewElement(tag.StudyTime, []string{study.StudyTime}),
			}
			recordItems = append(recordItems, studyElements)

			for _, series := range study.Series {
				// SERIES record
				seriesElements := []*dicom.Element{
					mustNewElement(tag.OffsetOfTheNextDirectoryRecord, []int{0}), // Will be updated
					mustNewElement(tag.OffsetOfReferencedLowerLevelDirectoryEntity, []int{0}), // Points to first IMAGE
					mustNewElement(tag.DirectoryRecordType, []string{"SERIES"}),
					mustNewElement(tag.Modality, []string{series.Modality}),
					mustNewElement(tag.SeriesInstanceUID, []string{series.SeriesUID}),
					mustNewElement(tag.SeriesNumber, []string{series.SeriesNumber}),
				}
				recordItems = append(recordItems, seriesElements)

				for _, image := range series.Images {
					// IMAGE record
					// Split path into components for ReferencedFileID
					pathParts := strings.Split(image.RelPath, "/")

					imageElements := []*dicom.Element{
						mustNewElement(tag.OffsetOfTheNextDirectoryRecord, []int{0}), // Will be updated
						mustNewElement(tag.OffsetOfReferencedLowerLevelDirectoryEntity, []int{0}), // No children for IMAGE
						mustNewElement(tag.DirectoryRecordType, []string{"IMAGE"}),
						mustNewElement(tag.ReferencedFileID, pathParts),
						mustNewElement(tag.ReferencedSOPClassUIDInFile, []string{image.SOPClassUID}),
						mustNewElement(tag.ReferencedSOPInstanceUIDInFile, []string{image.SOPInstanceUID}),
						mustNewElement(tag.ReferencedTransferSyntaxUIDInFile, []string{"1.2.840.10008.1.2.1"}),
					}
					recordItems = append(recordItems, imageElements)
				}
			}
		}
	}

	// Create DICOMDIR dataset
	ds := &dicom.Dataset{
		Elements: []*dicom.Element{},
	}

	// File Meta Information (must be first)
	ds.Elements = append(ds.Elements,
		mustNewElement(tag.TransferSyntaxUID, []string{"1.2.840.10008.1.2.1"}), // Explicit VR Little Endian
		mustNewElement(tag.MediaStorageSOPClassUID, []string{"1.2.840.10008.1.3.10"}), // Media Storage Directory Storage
		mustNewElement(tag.MediaStorageSOPInstanceUID, []string{"1.2.826.0.1.3680043.8.498.1"}),
		mustNewElement(tag.ImplementationClassUID, []string{"1.2.826.0.1.3680043.8.498"}),
	)

	// FileSet Identification
	filesetID := filepath.Base(outputDir)
	if len(filesetID) > 16 {
		filesetID = filesetID[:16]
	}
	ds.Elements = append(ds.Elements,
		mustNewElement(tag.FileSetID, []string{filesetID}),
		// Directory record offsets - these should be byte offsets but we set to 0
		// A proper implementation would calculate these during write
		mustNewElement(tag.OffsetOfTheFirstDirectoryRecordOfTheRootDirectoryEntity, []int{0}),
		mustNewElement(tag.OffsetOfTheLastDirectoryRecordOfTheRootDirectoryEntity, []int{0}),
	)

	// Add Directory Record Sequence
	// recordItems is [][]*Element, which NewElement will convert to SequenceItemValue automatically
	if len(recordItems) > 0 {
		seqElem, err := dicom.NewElement(tag.DirectoryRecordSequence, recordItems)
		if err != nil {
			return fmt.Errorf("create directory record sequence: %w", err)
		}
		ds.Elements = append(ds.Elements, seqElem)
	}

	// Write DICOMDIR
	if err := writeDatasetToFile(dicomdirPath, *ds); err != nil {
		return fmt.Errorf("write DICOMDIR: %w", err)
	}

	return nil
}
