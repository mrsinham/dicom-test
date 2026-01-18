# Integration Tests for Go DICOM Generator

This directory contains integration tests for the DICOM MRI generator.

## Test Coverage

### TestGenerateSeries_Basic
Tests basic DICOM series generation with 5 images:
- Verifies correct number of files generated
- Checks that all files exist
- Validates UIDs and patient IDs are set

### TestOrganizeFiles_DICOMDIRStructure
Tests DICOMDIR organization and file hierarchy:
- Verifies DICOMDIR file exists
- Checks PT000000/ST000000/SE000000/ hierarchy
- Validates image files are moved correctly
- Ensures temporary IMG*.dcm files are cleaned up

### TestValidation_RequiredTags
Validates DICOM file contents:
- Checks all required DICOM tags are present
- Verifies tag values (Modality=MR, BitsAllocated=16, etc.)
- Parses generated files with suyashkumar/dicom library

### TestMultiStudy
Tests multi-study generation:
- Generates 15 images across 3 studies
- Verifies correct directory structure for each study
- Validates image distribution

### TestReproducibility_SameSeed
Tests reproducibility:
- Generates two series with same seed
- Verifies PatientID is identical
- Checks deterministic behavior

### TestCalculateDimensions
Tests dimension calculation algorithm:
- Validates dimensions for various size/image combinations
- Checks dimensions are within expected ranges

## Running the Tests

### Prerequisites

1. **Build environment**: Requires Go 1.21+
2. **Network connection**: First run needs to download dependencies
3. **Disk space**: Tests create temporary DICOM files (cleaned up automatically)

### Run All Tests

```bash
cd /home/user/dicom-test/go
go test ./tests -v
```

### Run Specific Test

```bash
# Run only basic generation test
go test ./tests -v -run TestGenerateSeries_Basic

# Run only DICOMDIR structure test
go test ./tests -v -run TestOrganizeFiles_DICOMDIRStructure

# Run only validation test
go test ./tests -v -run TestValidation_RequiredTags
```

### Run with Coverage

```bash
go test ./tests -v -cover
```

### Run Short Tests Only

```bash
go test ./tests -v -short
```

## Expected Output

Successful test run:
```
=== RUN   TestGenerateSeries_Basic
    integration_test.go:25: Generating DICOM series in: /tmp/TestGenerateSeries_Basic...
    integration_test.go:45: Generated file 1: /tmp/.../IMG0001.dcm
    integration_test.go:45: Generated file 2: /tmp/.../IMG0002.dcm
    ...
    integration_test.go:61: ✓ Basic generation test passed
--- PASS: TestGenerateSeries_Basic (0.50s)

=== RUN   TestOrganizeFiles_DICOMDIRStructure
    integration_test.go:79: Generated 5 files, organizing into DICOMDIR...
    integration_test.go:94: ✓ DICOMDIR exists: /tmp/.../DICOMDIR
    integration_test.go:102: ✓ Patient directory exists: /tmp/.../PT000000
    ...
    integration_test.go:133: ✓ DICOMDIR structure test passed
--- PASS: TestOrganizeFiles_DICOMDIRStructure (0.60s)

...

PASS
ok      github.com/julien/dicom-test/go/tests   3.456s
```

## Test Data

All tests use `t.TempDir()` which:
- Creates unique temporary directories for each test
- Automatically cleans up after test completion
- Prevents conflicts between parallel tests

## Troubleshooting

### Network Errors

If you see errors about downloading dependencies:
```
go: downloading github.com/suyashkumar/dicom v1.1.0
dial tcp: lookup proxy.golang.org: no such host
```

**Solution**: Ensure you have internet connection for first run, or use vendored dependencies:
```bash
go mod vendor
go test ./tests -mod=vendor -v
```

### Module Errors

If you see:
```
go: updates to go.mod needed; to update it:
    go mod tidy
```

**Solution**: Run from the go directory and update modules:
```bash
cd /home/user/dicom-test/go
go mod tidy
go test ./tests -v
```

### Permission Errors

If tests fail creating files:
```
permission denied: /tmp/...
```

**Solution**: Ensure /tmp is writable or set TMPDIR:
```bash
export TMPDIR=/path/to/writable/dir
go test ./tests -v
```

## Adding New Tests

To add a new integration test:

1. Create a new test function in `integration_test.go`:
```go
func TestMyNewFeature(t *testing.T) {
    outputDir := t.TempDir()

    // Test setup
    opts := internaldicom.GeneratorOptions{
        NumImages: 5,
        TotalSize: "10MB",
        OutputDir: outputDir,
        Seed: 42,
        NumStudies: 1,
    }

    // Execute
    files, err := internaldicom.GenerateDICOMSeries(opts)
    if err != nil {
        t.Fatalf("Failed: %v", err)
    }

    // Verify
    // ... your assertions here ...

    t.Logf("✓ My test passed")
}
```

2. Run your new test:
```bash
go test ./tests -v -run TestMyNewFeature
```

## CI/CD Integration

For automated testing in CI pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run integration tests
  run: |
    cd go
    go test ./tests -v -timeout 5m
```

## Performance

Expected test execution times (approximate):
- TestGenerateSeries_Basic: 0.3-0.5s
- TestOrganizeFiles_DICOMDIRStructure: 0.4-0.7s
- TestValidation_RequiredTags: 0.3-0.5s
- TestMultiStudy: 0.8-1.2s
- TestReproducibility_SameSeed: 0.5-0.8s
- TestCalculateDimensions: <0.1s

**Total**: ~3-5 seconds for all tests

## Next Steps

See `docs/plans/2026-01-18-go-integration-tests.md` for the complete testing plan including:
- Additional test suites (validation, compatibility, performance)
- Python comparison scripts
- Benchmark tests
- Extended validation with pydicom
