# Advanced Categorization Tags Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add advanced DICOM tags for categorization: institution, physicians, body part, protocol, and priority.

**Architecture:** Extend existing metadata generation with new lookup tables and generator functions. New tags are added directly in `generator.go` where other metadata is created. CLI flags control whether values are varied across studies or fixed.

**Tech Stack:** Go 1.24+, github.com/suyashkumar/dicom, existing test patterns

---

## Task 1: Add Institution Data Lookup Tables

**Files:**
- Create: `internal/util/institutions.go`
- Test: `internal/util/institutions_test.go`

**Step 1: Write the failing test**

```go
// internal/util/institutions_test.go
package util

import (
	"math/rand/v2"
	"strings"
	"testing"
)

func TestGenerateInstitution_ReturnsValidData(t *testing.T) {
	inst := GenerateInstitution(nil)

	if inst.Name == "" {
		t.Error("Institution name should not be empty")
	}
	if inst.Address == "" {
		t.Error("Institution address should not be empty")
	}
	if inst.Department == "" {
		t.Error("Department should not be empty")
	}
}

func TestGenerateInstitution_Deterministic(t *testing.T) {
	rng1 := rand.New(rand.NewPCG(42, 42))
	inst1 := GenerateInstitution(rng1)

	rng2 := rand.New(rand.NewPCG(42, 42))
	inst2 := GenerateInstitution(rng2)

	if inst1.Name != inst2.Name {
		t.Errorf("Same seed should produce same institution: %s != %s", inst1.Name, inst2.Name)
	}
}

func TestGenerateStationName_Format(t *testing.T) {
	station := GenerateStationName("MR", "HEAD", nil)

	if station == "" {
		t.Error("Station name should not be empty")
	}
	if !strings.Contains(station, "_") {
		t.Errorf("Station name should contain underscore: %s", station)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/util/... -run TestGenerateInstitution -v`
Expected: FAIL with "undefined: GenerateInstitution"

**Step 3: Write minimal implementation**

```go
// internal/util/institutions.go
package util

import "math/rand/v2"

// Institution holds generated institution data
type Institution struct {
	Name       string
	Address    string
	Department string
}

var (
	// Hospitals is the list of realistic hospital names
	Hospitals = []struct {
		Name    string
		Address string
	}{
		{"CHU Bordeaux", "Place Amelie Raba-Leon, 33000 Bordeaux"},
		{"Hopital Saint-Louis", "1 Avenue Claude Vellefaux, 75010 Paris"},
		{"CHU Toulouse", "2 Rue Viguerie, 31000 Toulouse"},
		{"Clinique du Parc", "155 Boulevard Stalingrad, 69006 Lyon"},
		{"Hopital Europeen Georges-Pompidou", "20 Rue Leblanc, 75015 Paris"},
		{"CHU Nantes", "1 Place Alexis-Ricordeau, 44000 Nantes"},
		{"CHU Lille", "2 Avenue Oscar Lambret, 59000 Lille"},
		{"Hopital de la Pitie-Salpetriere", "47-83 Boulevard de l'Hopital, 75013 Paris"},
		{"CHU Montpellier", "191 Avenue du Doyen Gaston Giraud, 34090 Montpellier"},
		{"Hopital Cochin", "27 Rue du Faubourg Saint-Jacques, 75014 Paris"},
		{"Massachusetts General Hospital", "55 Fruit Street, Boston, MA 02114"},
		{"Johns Hopkins Hospital", "1800 Orleans Street, Baltimore, MD 21287"},
		{"Cleveland Clinic", "9500 Euclid Avenue, Cleveland, OH 44195"},
		{"Mayo Clinic", "200 First Street SW, Rochester, MN 55905"},
		{"UCLA Medical Center", "757 Westwood Plaza, Los Angeles, CA 90095"},
	}

	// Departments is the list of medical departments
	Departments = []string{
		"Radiologie",
		"Imagerie Medicale",
		"Neuroradiologie",
		"Radiologie Interventionnelle",
		"Urgences",
		"Cardiologie",
		"Neurologie",
		"Oncologie",
		"Pediatrie",
		"Orthopedie",
	}
)

// GenerateInstitution generates a random institution with address and department.
// If rng is nil, uses shared default RNG.
func GenerateInstitution(rng *rand.Rand) Institution {
	if rng == nil {
		rng = defaultRNG
	}

	hospital := Hospitals[rng.IntN(len(Hospitals))]
	department := Departments[rng.IntN(len(Departments))]

	return Institution{
		Name:       hospital.Name,
		Address:    hospital.Address,
		Department: department,
	}
}

// GenerateStationName generates a station name based on modality and body part.
// Format: MODALITY_BODYPART_NN (e.g., "MR_HEAD_01", "CT_CHEST_03")
func GenerateStationName(modality, bodyPart string, rng *rand.Rand) string {
	if rng == nil {
		rng = defaultRNG
	}

	num := rng.IntN(10) + 1
	return modality + "_" + bodyPart + "_" + fmt.Sprintf("%02d", num)
}
```

Note: Add `"fmt"` to imports in institutions.go.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/util/... -run "TestGenerateInstitution|TestGenerateStationName" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/util/institutions.go internal/util/institutions_test.go
git commit -m "$(cat <<'EOF'
feat: add institution data generator

Add lookup tables for hospitals and departments with
GenerateInstitution() and GenerateStationName() functions.
EOF
)"
```

---

## Task 2: Add Physician Name Generator

**Files:**
- Modify: `internal/util/names.go`
- Modify: `internal/util/names_test.go`

**Step 1: Write the failing test**

Add to `internal/util/names_test.go`:

```go
func TestGeneratePhysicianName_Format(t *testing.T) {
	name := GeneratePhysicianName(nil)

	if !strings.Contains(name, "^") {
		t.Errorf("Physician name should contain '^' separator, got: %s", name)
	}

	parts := strings.Split(name, "^")
	if len(parts) < 2 {
		t.Errorf("Physician name should have at least 2 parts, got: %s", name)
	}
}

func TestGeneratePhysicianName_HasTitle(t *testing.T) {
	// Run multiple times to check title prefix appears sometimes
	hasTitle := false
	for i := 0; i < 100; i++ {
		name := GeneratePhysicianName(nil)
		if strings.HasPrefix(name, "Dr") {
			hasTitle = true
			break
		}
	}
	if !hasTitle {
		t.Error("Expected some physician names to have Dr title")
	}
}

func TestGeneratePhysicianName_Deterministic(t *testing.T) {
	rng1 := rand.New(rand.NewPCG(42, 42))
	name1 := GeneratePhysicianName(rng1)

	rng2 := rand.New(rand.NewPCG(42, 42))
	name2 := GeneratePhysicianName(rng2)

	if name1 != name2 {
		t.Errorf("Same seed should produce same name: %s != %s", name1, name2)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/util/... -run TestGeneratePhysicianName -v`
Expected: FAIL with "undefined: GeneratePhysicianName"

**Step 3: Write minimal implementation**

Add to `internal/util/names.go`:

```go
// GeneratePhysicianName generates a realistic physician name.
// Format: "LASTNAME^FIRSTNAME" or "LASTNAME^FIRSTNAME^^^DR" (with title)
// Uses same name pools as patients, 50% chance of "Dr" title.
// If rng is nil, uses shared default RNG.
func GeneratePhysicianName(rng *rand.Rand) string {
	if rng == nil {
		rng = defaultRNG
	}

	// Randomly pick sex for name generation
	sex := "M"
	if rng.Float64() < 0.5 {
		sex = "F"
	}

	// 20% chance of French name
	useFrench := rng.Float64() < FrenchNameProbability

	var firstName string
	var lastName string

	if useFrench {
		if sex == "M" {
			firstName = FrenchMaleFirstNames[rng.IntN(len(FrenchMaleFirstNames))]
		} else {
			firstName = FrenchFemaleFirstNames[rng.IntN(len(FrenchFemaleFirstNames))]
		}
		lastName = FrenchLastNames[rng.IntN(len(FrenchLastNames))]
	} else {
		if sex == "M" {
			firstName = EnglishMaleFirstNames[rng.IntN(len(EnglishMaleFirstNames))]
		} else {
			firstName = EnglishFemaleFirstNames[rng.IntN(len(EnglishFemaleFirstNames))]
		}
		lastName = EnglishLastNames[rng.IntN(len(EnglishLastNames))]
	}

	// DICOM PN format: LASTNAME^FIRSTNAME^^^PREFIX
	// 50% chance of Dr title
	if rng.Float64() < 0.5 {
		return "Dr " + lastName + "^" + firstName
	}
	return lastName + "^" + firstName
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/util/... -run TestGeneratePhysicianName -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/util/names.go internal/util/names_test.go
git commit -m "$(cat <<'EOF'
feat: add physician name generator

Add GeneratePhysicianName() function that generates realistic
physician names with optional "Dr" title prefix.
EOF
)"
```

---

## Task 3: Add Body Part and Protocol Lookup Tables

**Files:**
- Create: `internal/util/clinical.go`
- Test: `internal/util/clinical_test.go`

**Step 1: Write the failing test**

```go
// internal/util/clinical_test.go
package util

import (
	"math/rand/v2"
	"testing"
)

func TestGetBodyPartsForModality_MR(t *testing.T) {
	parts := GetBodyPartsForModality("MR")
	if len(parts) == 0 {
		t.Error("MR should have body parts")
	}

	// Should contain HEAD for MR
	found := false
	for _, p := range parts {
		if p == "HEAD" {
			found = true
			break
		}
	}
	if !found {
		t.Error("MR should include HEAD body part")
	}
}

func TestGetBodyPartsForModality_Unknown(t *testing.T) {
	parts := GetBodyPartsForModality("UNKNOWN")
	if len(parts) == 0 {
		t.Error("Unknown modality should return default body parts")
	}
}

func TestGenerateProtocolName_Format(t *testing.T) {
	protocol := GenerateProtocolName("MR", "HEAD", nil)
	if protocol == "" {
		t.Error("Protocol name should not be empty")
	}
}

func TestGenerateProtocolName_Deterministic(t *testing.T) {
	rng1 := rand.New(rand.NewPCG(42, 42))
	p1 := GenerateProtocolName("MR", "HEAD", rng1)

	rng2 := rand.New(rand.NewPCG(42, 42))
	p2 := GenerateProtocolName("MR", "HEAD", rng2)

	if p1 != p2 {
		t.Errorf("Same seed should produce same protocol: %s != %s", p1, p2)
	}
}

func TestGenerateClinicalIndication_NotEmpty(t *testing.T) {
	indication := GenerateClinicalIndication("MR", "HEAD", nil)
	if indication == "" {
		t.Error("Clinical indication should not be empty")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/util/... -run "TestGetBodyParts|TestGenerateProtocol|TestGenerateClinical" -v`
Expected: FAIL with "undefined: GetBodyPartsForModality"

**Step 3: Write minimal implementation**

```go
// internal/util/clinical.go
package util

import "math/rand/v2"

// BodyPartsByModality maps modalities to appropriate body parts
var BodyPartsByModality = map[string][]string{
	"MR": {"HEAD", "BRAIN", "CSPINE", "TSPINE", "LSPINE", "KNEE", "SHOULDER", "HIP", "ANKLE", "WRIST", "PELVIS", "ABDOMEN", "CHEST"},
	"CT": {"HEAD", "CHEST", "ABDOMEN", "PELVIS", "CSPINE", "TSPINE", "LSPINE", "EXTREMITY"},
	"CR": {"CHEST", "HAND", "FOOT", "KNEE", "SHOULDER", "SKULL", "SPINE", "PELVIS", "RIBS"},
	"DX": {"CHEST", "HAND", "FOOT", "KNEE", "SHOULDER", "SKULL", "SPINE", "PELVIS", "RIBS"},
	"US": {"ABDOMEN", "PELVIS", "BREAST", "THYROID", "HEART", "LIVER", "KIDNEY", "UTERUS"},
	"MG": {"BREAST"},
}

// DefaultBodyParts is used when modality is unknown
var DefaultBodyParts = []string{"HEAD", "CHEST", "ABDOMEN", "EXTREMITY"}

// ProtocolsByModalityAndBodyPart maps modality+bodypart to protocol names
var ProtocolsByModalityAndBodyPart = map[string]map[string][]string{
	"MR": {
		"HEAD":     {"BRAIN_ROUTINE", "BRAIN_WITH_CONTRAST", "BRAIN_STROKE", "BRAIN_TUMOR", "BRAIN_MS"},
		"BRAIN":    {"BRAIN_ROUTINE", "BRAIN_WITH_CONTRAST", "BRAIN_STROKE", "BRAIN_TUMOR", "BRAIN_MS"},
		"CSPINE":   {"CSPINE_ROUTINE", "CSPINE_WITH_CONTRAST"},
		"TSPINE":   {"TSPINE_ROUTINE", "TSPINE_WITH_CONTRAST"},
		"LSPINE":   {"LSPINE_ROUTINE", "LSPINE_WITH_CONTRAST", "LSPINE_DISC"},
		"KNEE":     {"KNEE_ROUTINE", "KNEE_ACL", "KNEE_MENISCUS"},
		"SHOULDER": {"SHOULDER_ROUTINE", "SHOULDER_ARTHROGRAM"},
		"HIP":      {"HIP_ROUTINE", "HIP_ARTHROGRAM"},
		"ABDOMEN":  {"ABDOMEN_ROUTINE", "MRCP", "LIVER_DYNAMIC"},
		"PELVIS":   {"PELVIS_ROUTINE", "PROSTATE_MP"},
	},
	"CT": {
		"HEAD":     {"HEAD_ROUTINE", "HEAD_TRAUMA", "HEAD_STROKE", "HEAD_SINUS"},
		"CHEST":    {"CHEST_ROUTINE", "CHEST_PE", "CHEST_HRCT", "CHEST_TRAUMA"},
		"ABDOMEN":  {"ABDOMEN_ROUTINE", "ABDOMEN_TRIPLE_PHASE", "ABDOMEN_TRAUMA"},
		"PELVIS":   {"PELVIS_ROUTINE", "PELVIS_WITH_CONTRAST"},
		"CSPINE":   {"CSPINE_TRAUMA", "CSPINE_ROUTINE"},
	},
}

// ClinicalIndications maps body parts to common clinical indications
var ClinicalIndications = map[string][]string{
	"HEAD":     {"Cephalees persistantes", "Vertiges", "Trouble de la vision", "Suspicion AVC", "Bilan tumoral"},
	"BRAIN":    {"Cephalees persistantes", "Vertiges", "Trouble de la vision", "Suspicion AVC", "Bilan tumoral"},
	"CHEST":    {"Toux chronique", "Dyspnee", "Douleur thoracique", "Bilan infectieux", "Suspicion EP"},
	"ABDOMEN":  {"Douleur abdominale", "Bilan hepatique", "Masse abdominale", "Occlusion"},
	"KNEE":     {"Douleur genou", "Traumatisme", "Suspicion rupture LCA", "Blocage articulaire"},
	"SHOULDER": {"Douleur epaule", "Limitation mobilite", "Traumatisme"},
	"LSPINE":   {"Lombalgie", "Sciatique", "Bilan hernie discale"},
	"PELVIS":   {"Douleur pelvienne", "Bilan oncologique", "Trouble urinaire"},
	"BREAST":   {"Depistage", "Masse palpable", "Bilan extension"},
}

// DefaultIndications is used when body part has no specific indications
var DefaultIndications = []string{"Bilan diagnostique", "Controle", "Suivi"}

// GetBodyPartsForModality returns the list of body parts for a modality
func GetBodyPartsForModality(modality string) []string {
	if parts, ok := BodyPartsByModality[modality]; ok {
		return parts
	}
	return DefaultBodyParts
}

// GenerateBodyPart returns a random body part appropriate for the modality
func GenerateBodyPart(modality string, rng *rand.Rand) string {
	if rng == nil {
		rng = defaultRNG
	}
	parts := GetBodyPartsForModality(modality)
	return parts[rng.IntN(len(parts))]
}

// GenerateProtocolName generates a protocol name for the given modality and body part
func GenerateProtocolName(modality, bodyPart string, rng *rand.Rand) string {
	if rng == nil {
		rng = defaultRNG
	}

	if modalityProtocols, ok := ProtocolsByModalityAndBodyPart[modality]; ok {
		if protocols, ok := modalityProtocols[bodyPart]; ok {
			return protocols[rng.IntN(len(protocols))]
		}
	}

	// Default protocol name
	return modality + "_" + bodyPart + "_ROUTINE"
}

// GenerateClinicalIndication generates a clinical indication for the body part
func GenerateClinicalIndication(modality, bodyPart string, rng *rand.Rand) string {
	if rng == nil {
		rng = defaultRNG
	}

	if indications, ok := ClinicalIndications[bodyPart]; ok {
		return indications[rng.IntN(len(indications))]
	}
	return DefaultIndications[rng.IntN(len(DefaultIndications))]
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/util/... -run "TestGetBodyParts|TestGenerateProtocol|TestGenerateClinical" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/util/clinical.go internal/util/clinical_test.go
git commit -m "$(cat <<'EOF'
feat: add clinical data generators

Add body parts by modality, protocol names, and clinical
indication generators for DICOM categorization tags.
EOF
)"
```

---

## Task 4: Add Priority Type

**Files:**
- Create: `internal/util/priority.go`
- Test: `internal/util/priority_test.go`

**Step 1: Write the failing test**

```go
// internal/util/priority_test.go
package util

import (
	"math/rand/v2"
	"testing"
)

func TestParsePriority_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected Priority
	}{
		{"HIGH", PriorityHigh},
		{"high", PriorityHigh},
		{"ROUTINE", PriorityRoutine},
		{"routine", PriorityRoutine},
		{"LOW", PriorityLow},
		{"low", PriorityLow},
	}

	for _, tc := range tests {
		result, err := ParsePriority(tc.input)
		if err != nil {
			t.Errorf("ParsePriority(%q) returned error: %v", tc.input, err)
		}
		if result != tc.expected {
			t.Errorf("ParsePriority(%q) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

func TestParsePriority_Invalid(t *testing.T) {
	_, err := ParsePriority("INVALID")
	if err == nil {
		t.Error("ParsePriority(INVALID) should return error")
	}
}

func TestPriority_String(t *testing.T) {
	if PriorityHigh.String() != "HIGH" {
		t.Errorf("PriorityHigh.String() = %s, want HIGH", PriorityHigh.String())
	}
	if PriorityRoutine.String() != "ROUTINE" {
		t.Errorf("PriorityRoutine.String() = %s, want ROUTINE", PriorityRoutine.String())
	}
	if PriorityLow.String() != "LOW" {
		t.Errorf("PriorityLow.String() = %s, want LOW", PriorityLow.String())
	}
}

func TestGeneratePriority_Distribution(t *testing.T) {
	// Generate many priorities and check distribution
	counts := map[Priority]int{}
	rng := rand.New(rand.NewPCG(42, 42))

	for i := 0; i < 1000; i++ {
		p := GeneratePriority(rng)
		counts[p]++
	}

	// ROUTINE should be most common (~70%), HIGH ~20%, LOW ~10%
	if counts[PriorityRoutine] < 500 {
		t.Errorf("ROUTINE should be most common, got %d/1000", counts[PriorityRoutine])
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/util/... -run TestParsePriority -v`
Expected: FAIL with "undefined: ParsePriority"

**Step 3: Write minimal implementation**

```go
// internal/util/priority.go
package util

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// Priority represents exam priority level
type Priority int

const (
	PriorityRoutine Priority = iota
	PriorityHigh
	PriorityLow
)

// String returns the DICOM string representation of the priority
func (p Priority) String() string {
	switch p {
	case PriorityHigh:
		return "HIGH"
	case PriorityLow:
		return "LOW"
	default:
		return "ROUTINE"
	}
}

// ParsePriority parses a string into a Priority
func ParsePriority(s string) (Priority, error) {
	switch strings.ToUpper(s) {
	case "HIGH":
		return PriorityHigh, nil
	case "ROUTINE":
		return PriorityRoutine, nil
	case "LOW":
		return PriorityLow, nil
	default:
		return PriorityRoutine, fmt.Errorf("invalid priority: %s (valid: HIGH, ROUTINE, LOW)", s)
	}
}

// GeneratePriority generates a random priority with realistic distribution.
// Distribution: 70% ROUTINE, 20% HIGH, 10% LOW
func GeneratePriority(rng *rand.Rand) Priority {
	if rng == nil {
		rng = defaultRNG
	}

	r := rng.Float64()
	if r < 0.70 {
		return PriorityRoutine
	} else if r < 0.90 {
		return PriorityHigh
	}
	return PriorityLow
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/util/... -run "TestParsePriority|TestPriority_String|TestGeneratePriority" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/util/priority.go internal/util/priority_test.go
git commit -m "$(cat <<'EOF'
feat: add priority type for exam categorization

Add Priority type with HIGH/ROUTINE/LOW values, parser,
and realistic distribution generator (70/20/10).
EOF
)"
```

---

## Task 5: Add CLI Flags for Categorization Options

**Files:**
- Modify: `cmd/dicomforge/main.go`

**Step 1: Write the failing test** (manual test via CLI)

This is a CLI change, so we'll test manually after implementation.

**Step 2: Add new flags to main.go**

In `cmd/dicomforge/main.go`, after existing flag definitions (around line 24):

```go
	// Categorization options
	institution := flag.String("institution", "", "Institution name (random if not specified)")
	department := flag.String("department", "", "Department name (random if not specified)")
	bodyPart := flag.String("body-part", "", "Body part examined (random per modality if not specified)")
	priority := flag.String("priority", "ROUTINE", "Exam priority: HIGH, ROUTINE, LOW")
	variedMetadata := flag.Bool("varied-metadata", false, "Generate varied institutions/physicians across studies")
```

Update `GeneratorOptions` struct usage (around line 77):

```go
	// Parse priority
	parsedPriority, err := util.ParsePriority(*priority)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create generator options
	opts := dicom.GeneratorOptions{
		NumImages:      *numImages,
		TotalSize:      *totalSize,
		OutputDir:      *outputDir,
		Seed:           *seed,
		NumStudies:     *numStudies,
		NumPatients:    *numPatients,
		Workers:        *workers,
		Institution:    *institution,
		Department:     *department,
		BodyPart:       *bodyPart,
		Priority:       parsedPriority,
		VariedMetadata: *variedMetadata,
	}
```

Add import for util package at the top:

```go
import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/mrsinham/dicomforge/internal/dicom"
	"github.com/mrsinham/dicomforge/internal/util"
)
```

Update `printHelp()` to document new flags.

**Step 3: Run to verify compilation**

Run: `go build ./cmd/dicomforge`
Expected: FAIL (GeneratorOptions doesn't have new fields yet)

**Step 4: Commit (partial - flags only)**

We'll commit after Task 6 when GeneratorOptions is updated.

---

## Task 6: Update GeneratorOptions and Add New Tags

**Files:**
- Modify: `internal/dicom/generator.go`

**Step 1: Update GeneratorOptions struct**

Add new fields to `GeneratorOptions` (around line 154):

```go
// GeneratorOptions contains all parameters needed to generate a DICOM series
type GeneratorOptions struct {
	NumImages   int
	TotalSize   string
	OutputDir   string
	Seed        int64
	NumStudies  int
	NumPatients int
	Workers     int

	// Categorization options
	Institution    string        // Fixed institution name (empty = random)
	Department     string        // Fixed department name (empty = random)
	BodyPart       string        // Fixed body part (empty = random per modality)
	Priority       util.Priority // Exam priority
	VariedMetadata bool          // Generate varied institutions/physicians per study
}
```

Add import for util at the top of generator.go:

```go
import (
	// ... existing imports ...
	"github.com/mrsinham/dicomforge/internal/util"
)
```

**Step 2: Generate categorization metadata in GenerateDICOMSeries**

In `GenerateDICOMSeries`, after generating patient info (around line 375), add institution generation:

```go
	// Generate institution info (shared or varied per study)
	var defaultInstitution util.Institution
	if !opts.VariedMetadata {
		if opts.Institution != "" {
			defaultInstitution = util.Institution{
				Name:       opts.Institution,
				Address:    "",
				Department: opts.Department,
			}
			if defaultInstitution.Department == "" {
				defaultInstitution.Department = util.Departments[rng.IntN(len(util.Departments))]
			}
		} else {
			defaultInstitution = util.GenerateInstitution(rng)
			if opts.Department != "" {
				defaultInstitution.Department = opts.Department
			}
		}
	}

	// Generate body part (if fixed)
	bodyPart := opts.BodyPart
	if bodyPart == "" {
		bodyPart = util.GenerateBodyPart("MR", rng)
	}
```

**Step 3: Add new tags to metadata slice**

In the study loop (around line 510), after existing metadata elements, add:

```go
	// Categorization metadata
	var studyInstitution util.Institution
	if opts.VariedMetadata {
		studyInstitution = util.GenerateInstitution(rng)
	} else {
		studyInstitution = defaultInstitution
	}

	referringPhysician := util.GeneratePhysicianName(rng)
	performingPhysician := util.GeneratePhysicianName(rng)
	operatorName := util.GeneratePhysicianName(rng)
	protocolName := util.GenerateProtocolName("MR", bodyPart, rng)
	clinicalIndication := util.GenerateClinicalIndication("MR", bodyPart, rng)
	stationName := util.GenerateStationName("MR", bodyPart, rng)
```

Add these elements to the metadata slice (after line 554, before closing `}`):

```go
				// Categorization tags
				mustNewElement(tag.InstitutionName, []string{studyInstitution.Name}),
				mustNewElement(tag.InstitutionalDepartmentName, []string{studyInstitution.Department}),
				mustNewElement(tag.StationName, []string{stationName}),
				mustNewElement(tag.ReferringPhysicianName, []string{referringPhysician}),
				mustNewElement(tag.PerformingPhysicianName, []string{performingPhysician}),
				mustNewElement(tag.OperatorsName, []string{operatorName}),
				mustNewElement(tag.BodyPartExamined, []string{bodyPart}),
				mustNewElement(tag.ProtocolName, []string{protocolName}),
				mustNewElement(tag.RequestedProcedureDescription, []string{clinicalIndication}),
				mustNewElement(tag.Priority, []string{opts.Priority.String()}),
```

Note: Some tags like `InstitutionAddress` and `RequestingPhysician` can be added if the DICOM library supports them. Check available tags first.

**Step 4: Build and test**

Run: `go build ./cmd/dicomforge && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/dicomforge/main.go internal/dicom/generator.go
git commit -m "$(cat <<'EOF'
feat: add categorization tags to DICOM generation

Add new CLI flags: --institution, --department, --body-part,
--priority, --varied-metadata.

Generate tags: InstitutionName, InstitutionalDepartmentName,
StationName, ReferringPhysicianName, PerformingPhysicianName,
OperatorsName, BodyPartExamined, ProtocolName,
RequestedProcedureDescription, Priority.
EOF
)"
```

---

## Task 7: Add Integration Test for Categorization Tags

**Files:**
- Modify: `tests/integration_test.go`

**Step 1: Write the failing test**

Add to `tests/integration_test.go`:

```go
func TestCategorizationTags(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "dicom_categorization_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Generate DICOM with categorization options
	opts := dicom.GeneratorOptions{
		NumImages:      2,
		TotalSize:      "1MB",
		OutputDir:      tmpDir,
		Seed:           12345,
		NumStudies:     1,
		NumPatients:    1,
		Institution:    "Test Hospital",
		Department:     "Radiology",
		BodyPart:       "HEAD",
		Priority:       util.PriorityHigh,
		VariedMetadata: false,
	}

	files, err := dicom.GenerateDICOMSeries(opts)
	if err != nil {
		t.Fatalf("Failed to generate DICOM: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(files))
	}

	// Read first file and verify tags
	f, err := os.Open(files[0].Path)
	if err != nil {
		t.Fatalf("Failed to open DICOM file: %v", err)
	}
	defer f.Close()

	ds, err := dicomlib.Parse(f, nil)
	if err != nil {
		t.Fatalf("Failed to parse DICOM: %v", err)
	}

	// Check InstitutionName
	elem, err := ds.FindElementByTag(tag.InstitutionName)
	if err != nil {
		t.Error("InstitutionName tag not found")
	} else {
		val := elem.Value.GetValue().([]string)[0]
		if val != "Test Hospital" {
			t.Errorf("InstitutionName = %s, want Test Hospital", val)
		}
	}

	// Check BodyPartExamined
	elem, err = ds.FindElementByTag(tag.BodyPartExamined)
	if err != nil {
		t.Error("BodyPartExamined tag not found")
	} else {
		val := elem.Value.GetValue().([]string)[0]
		if val != "HEAD" {
			t.Errorf("BodyPartExamined = %s, want HEAD", val)
		}
	}

	// Check Priority
	elem, err = ds.FindElementByTag(tag.Priority)
	if err != nil {
		t.Error("Priority tag not found")
	} else {
		val := elem.Value.GetValue().([]string)[0]
		if val != "HIGH" {
			t.Errorf("Priority = %s, want HIGH", val)
		}
	}

	// Check ReferringPhysicianName exists
	_, err = ds.FindElementByTag(tag.ReferringPhysicianName)
	if err != nil {
		t.Error("ReferringPhysicianName tag not found")
	}

	// Check ProtocolName exists
	_, err = ds.FindElementByTag(tag.ProtocolName)
	if err != nil {
		t.Error("ProtocolName tag not found")
	}
}
```

Add imports at the top of the test file:

```go
import (
	// ... existing imports ...
	"github.com/mrsinham/dicomforge/internal/util"
	dicomlib "github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)
```

**Step 2: Run test to verify it passes**

Run: `go test ./tests/... -run TestCategorizationTags -v`
Expected: PASS

**Step 3: Commit**

```bash
git add tests/integration_test.go
git commit -m "$(cat <<'EOF'
test: add integration test for categorization tags

Verify InstitutionName, BodyPartExamined, Priority,
ReferringPhysicianName, and ProtocolName tags are
correctly generated in DICOM files.
EOF
)"
```

---

## Task 8: Update Help Text and Documentation

**Files:**
- Modify: `cmd/dicomforge/main.go` (printHelp function)

**Step 1: Update printHelp()**

Replace the `printHelp()` function with updated documentation:

```go
func printHelp() {
	fmt.Println("dicomforge")
	fmt.Println("==========")
	fmt.Println()
	fmt.Println("Generate valid DICOM multi-file MRI series for testing medical platforms.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dicomforge --num-images <N> --total-size <SIZE> [options]")
	fmt.Println()
	fmt.Println("Required arguments:")
	fmt.Println("  --num-images <N>      Number of DICOM images/slices to generate")
	fmt.Println("  --total-size <SIZE>   Total size (e.g., '100MB', '1GB', '4.5GB')")
	fmt.Println()
	fmt.Println("Optional arguments:")
	fmt.Println("  --output <DIR>        Output directory (default: 'dicom_series')")
	fmt.Println("  --seed <N>            Seed for reproducibility (auto-generated if not specified)")
	fmt.Println("  --num-studies <N>     Number of studies to generate (default: 1)")
	fmt.Println("  --num-patients <N>    Number of patients (default: 1)")
	fmt.Printf("  --workers <N>         Number of parallel workers (default: %d = CPU cores)\n", runtime.NumCPU())
	fmt.Println()
	fmt.Println("Categorization options:")
	fmt.Println("  --institution <NAME>  Institution name (random if not specified)")
	fmt.Println("  --department <NAME>   Department name (random if not specified)")
	fmt.Println("  --body-part <PART>    Body part: HEAD, CHEST, ABDOMEN, KNEE, etc. (random if not specified)")
	fmt.Println("  --priority <LEVEL>    Priority: HIGH, ROUTINE, LOW (default: ROUTINE)")
	fmt.Println("  --varied-metadata     Generate varied institutions/physicians per study")
	fmt.Println()
	fmt.Println("  --help                Show this help message")
	fmt.Println("  --version             Show version")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Generate 10 images, 100MB total")
	fmt.Println("  dicomforge --num-images 10 --total-size 100MB")
	fmt.Println()
	fmt.Println("  # Generate with specific institution and body part")
	fmt.Println("  dicomforge --num-images 50 --total-size 500MB \\")
	fmt.Println("    --institution \"CHU Bordeaux\" --body-part HEAD --priority HIGH")
	fmt.Println()
	fmt.Println("  # Generate varied metadata across studies")
	fmt.Println("  dicomforge --num-images 100 --total-size 1GB \\")
	fmt.Println("    --num-studies 5 --varied-metadata")
	fmt.Println()
	fmt.Println("Generated DICOM tags include:")
	fmt.Println("  - Patient: PatientName, PatientID, PatientBirthDate, PatientSex")
	fmt.Println("  - Study: StudyInstanceUID, StudyDate, StudyDescription")
	fmt.Println("  - Series: SeriesInstanceUID, SeriesDescription, Modality")
	fmt.Println("  - Institution: InstitutionName, InstitutionalDepartmentName, StationName")
	fmt.Println("  - Physicians: ReferringPhysicianName, PerformingPhysicianName, OperatorsName")
	fmt.Println("  - Clinical: BodyPartExamined, ProtocolName, Priority")
	fmt.Println("  - MRI: Manufacturer, SequenceName, MagneticFieldStrength, etc.")
}
```

**Step 2: Build and verify**

Run: `go build ./cmd/dicomforge && ./dicomforge --help`
Expected: Updated help text displayed

**Step 3: Commit**

```bash
git add cmd/dicomforge/main.go
git commit -m "$(cat <<'EOF'
docs: update CLI help with categorization options

Document new flags: --institution, --department, --body-part,
--priority, --varied-metadata. List all generated DICOM tags.
EOF
)"
```

---

## Task 9: Run Full Test Suite and Final Verification

**Step 1: Run all tests**

Run: `go test ./... -v`
Expected: All tests PASS

**Step 2: Build and manual test**

```bash
go build ./cmd/dicomforge

# Test with default categorization
./dicomforge --num-images 5 --total-size 10MB --output test_default

# Test with custom categorization
./dicomforge --num-images 5 --total-size 10MB --output test_custom \
  --institution "Test Hospital" --body-part HEAD --priority HIGH

# Test with varied metadata
./dicomforge --num-images 10 --total-size 20MB --output test_varied \
  --num-studies 3 --varied-metadata

# Verify with dcmdump or similar tool
dcmdump test_custom/PT000000/ST000000/SE000000/IM000001 | grep -E "(Institution|BodyPart|Priority|Physician|Protocol)"
```

**Step 3: Clean up test directories**

```bash
rm -rf test_default test_custom test_varied
```

**Step 4: Final commit (if any fixes needed)**

```bash
git status
# If clean, no commit needed
```

---

## Summary

This plan implements Feature 2 (Advanced Categorization Tags) in 9 tasks:

1. **Institution data** — Lookup tables and generator
2. **Physician names** — Name generator with titles
3. **Clinical data** — Body parts, protocols, indications
4. **Priority type** — HIGH/ROUTINE/LOW with parser
5. **CLI flags** — New command-line options
6. **Generator update** — Add tags to DICOM metadata
7. **Integration test** — Verify tags in generated files
8. **Documentation** — Update help text
9. **Verification** — Full test suite and manual testing

Total new tags added:
- `InstitutionName`
- `InstitutionalDepartmentName`
- `StationName`
- `ReferringPhysicianName`
- `PerformingPhysicianName`
- `OperatorsName`
- `BodyPartExamined`
- `ProtocolName`
- `RequestedProcedureDescription`
- `Priority`
