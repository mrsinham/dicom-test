#!/bin/bash
# Run integration tests for Go DICOM generator

set -e

cd "$(dirname "$0")/.."

echo "======================================"
echo "DICOM Generator Integration Tests"
echo "======================================"
echo ""

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "Error: Must run from go/ directory"
    exit 1
fi

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

echo "Go version: $(go version)"
echo ""

# Parse arguments
VERBOSE=""
TEST_FILTER=""
COVERAGE=""
SHORT=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE="-v"
            shift
            ;;
        -c|--coverage)
            COVERAGE="-cover"
            shift
            ;;
        -s|--short)
            SHORT="-short"
            shift
            ;;
        -t|--test)
            TEST_FILTER="-run $2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  -v, --verbose     Verbose output"
            echo "  -c, --coverage    Show coverage"
            echo "  -s, --short       Run short tests only"
            echo "  -t, --test NAME   Run specific test"
            echo "  -h, --help        Show this help"
            echo ""
            echo "Examples:"
            echo "  $0                                    # Run all tests"
            echo "  $0 -v                                 # Verbose output"
            echo "  $0 -v -c                              # Verbose with coverage"
            echo "  $0 -t TestGenerateSeries_Basic        # Run specific test"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h for help"
            exit 1
            ;;
    esac
done

# Run tests
echo "Running tests..."
echo ""

# Set GOTOOLCHAIN to use local Go version
export GOTOOLCHAIN=local

# Run the tests
go test ./tests $VERBOSE $COVERAGE $SHORT $TEST_FILTER

EXIT_CODE=$?

echo ""
if [ $EXIT_CODE -eq 0 ]; then
    echo "======================================"
    echo "✓ All tests passed!"
    echo "======================================"
else
    echo "======================================"
    echo "✗ Some tests failed"
    echo "======================================"
fi

exit $EXIT_CODE
