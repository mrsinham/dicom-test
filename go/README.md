# DICOM MRI Generator (Go)

Go implementation of the DICOM MRI generator for testing medical platforms.

## Building

```bash
go build -o bin/generate-dicom-mri ./cmd/generate-dicom-mri
```

## Usage

```bash
./bin/generate-dicom-mri --num-images 10 --total-size 100MB --output test-series
```

## Development

Run tests:
```bash
go test ./...
```

Run with coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```
