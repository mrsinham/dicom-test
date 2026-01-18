#!/bin/bash
# Manual test - generates a small DICOM series to verify functionality
# This doesn't require Go test framework, just the compiled binary

set -e

echo "======================================"
echo "Manual DICOM Generation Test"
echo "======================================"
echo ""

# Check if binary exists
BINARY="../bin/generate-dicom-mri"

if [ ! -f "$BINARY" ]; then
    echo "Binary not found. Building..."
    cd ..
    go build -o bin/generate-dicom-mri ./cmd/generate-dicom-mri
    cd tests
    echo "✓ Binary built"
    echo ""
fi

# Clean up previous test
TEST_DIR="manual-test-output"
if [ -d "$TEST_DIR" ]; then
    echo "Cleaning up previous test..."
    rm -rf "$TEST_DIR"
fi

echo "Generating test DICOM series..."
echo "  - 5 images"
echo "  - 10MB total"
echo "  - Seed: 42"
echo ""

# Generate
$BINARY --num-images 5 --total-size 10MB --output "$TEST_DIR" --seed 42

echo ""
echo "======================================"
echo "Verification"
echo "======================================"

# Check DICOMDIR exists
if [ -f "$TEST_DIR/DICOMDIR" ]; then
    echo "✓ DICOMDIR exists"
else
    echo "✗ DICOMDIR missing"
    exit 1
fi

# Check PT/ST/SE hierarchy
if [ -d "$TEST_DIR/PT000000/ST000000/SE000000" ]; then
    echo "✓ PT000000/ST000000/SE000000/ hierarchy exists"
else
    echo "✗ Hierarchy missing"
    exit 1
fi

# Count images
IMAGE_COUNT=$(find "$TEST_DIR" -name "IM*" | wc -l)
if [ "$IMAGE_COUNT" -eq 5 ]; then
    echo "✓ 5 image files found"
else
    echo "✗ Expected 5 images, found $IMAGE_COUNT"
    exit 1
fi

# Check no temporary files
TEMP_COUNT=$(find "$TEST_DIR" -name "IMG*.dcm" | wc -l)
if [ "$TEMP_COUNT" -eq 0 ]; then
    echo "✓ No temporary IMG*.dcm files"
else
    echo "⚠ Found $TEMP_COUNT temporary files (should be cleaned up)"
fi

# Calculate total size
TOTAL_SIZE=$(du -sh "$TEST_DIR" | cut -f1)
echo "✓ Total size: $TOTAL_SIZE"

echo ""
echo "Directory structure:"
tree "$TEST_DIR" 2>/dev/null || find "$TEST_DIR" -type f | head -10

echo ""
echo "======================================"
echo "✓ Manual test passed!"
echo "======================================"
echo ""
echo "Output directory: $TEST_DIR"
echo ""
echo "To validate with Python (if pydicom installed):"
echo "  python3 -c \"import pydicom; ds = pydicom.dcmread('$TEST_DIR/PT000000/ST000000/SE000000/IM000001'); print(f'{ds.PatientName} - {ds.Modality} - {ds.Rows}x{ds.Columns}')\""
echo ""
echo "To clean up:"
echo "  rm -rf $TEST_DIR"
