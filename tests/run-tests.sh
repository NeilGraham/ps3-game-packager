#!/bin/bash

# ROM Organizer Test Suite Runner
# This script runs the complete test suite including integration tests

set -e  # Exit on any error

# Parse command line arguments
KEEP_ARTIFACTS=false
for arg in "$@"; do
    case $arg in
        --keep)
            KEEP_ARTIFACTS=true
            shift
            ;;
        *)
            # Unknown option
            echo "Usage: $0 [--keep]"
            echo "  --keep    Keep test artifacts after completion (useful for debugging)"
            exit 1
            ;;
    esac
done

# Get the directory of this script and navigate to project root
script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
project_root="$(dirname "$script_dir")"
cd "$project_root"

echo "🧪 ROM Organizer Test Suite"
echo "=========================="
echo "Running from: $(pwd)"
if [ "$KEEP_ARTIFACTS" = true ]; then
    echo "🔒 Keeping test artifacts (--keep flag enabled)"
fi
echo

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Could not find go.mod in project root"
    echo "   Please ensure this script is in the tests/ directory of the ROM organizer project"
    exit 1
fi

# Clean up any existing test artifacts
if [ "$KEEP_ARTIFACTS" = false ]; then
    echo "🧹 Cleaning up existing test artifacts..."
    # Clean up legacy artifacts that might be in wrong locations
    rm -rf test-games/ test-organized/ test-compressed/ test-decompressed/
    rm -rf temp-test-games/ bench-games-*/ organized-test-games/ compressed-test-games/ decompressed-test-games/
    rm -f rom-organizer-dev rom-organizer-dev.exe rom-organizer-bench rom-organizer-bench.exe
    # Clean up artifacts in cmd/rom-organizer directory (legacy)
    rm -f cmd/rom-organizer/rom-organizer-dev cmd/rom-organizer/rom-organizer-dev.exe
    rm -f cmd/rom-organizer/rom-organizer-bench cmd/rom-organizer/rom-organizer-bench.exe
    rm -rf cmd/rom-organizer/test-games/ cmd/rom-organizer/test-organized/ cmd/rom-organizer/test-compressed/ cmd/rom-organizer/test-decompressed/
    rm -rf cmd/rom-organizer/temp-test-games/ cmd/rom-organizer/bench-games-*/
    # Clean up test artifacts in tests directory (current location)
    rm -rf tests/test-games/ tests/test-organized/ tests/test-compressed/ tests/test-decompressed/
    rm -rf tests/temp-test-games/ tests/bench-games-*/
    echo "   ✅ Cleanup complete"
else
    echo "🔒 Skipping initial cleanup (--keep flag enabled)"
fi
echo

# Run unit tests
echo "🔬 Running unit tests..."
go test ./internal/... -v
echo "   ✅ Unit tests passed"
echo

# Run integration tests with verbose output (from cmd/rom-organizer directory)
echo "🚀 Running integration tests..."
echo "   This will test the complete workflow:"
echo "   • Test game generation using tests/generate-test-games.go → tests/test-games/"
echo "   • Binary building in cmd/rom-organizer/"
echo "   • Metadata extraction (single & multiple paths)"
echo "   • Organize command → tests/test-organized/"
echo "   • Compress command → tests/test-compressed/"
echo "   • Decompress command → tests/test-decompressed/"
echo

# Set environment variable for Go tests to know about --keep flag
if [ "$KEEP_ARTIFACTS" = true ]; then
    export KEEP_TEST_ARTIFACTS=true
fi

# Change to cmd/rom-organizer directory for integration tests
cd cmd/rom-organizer

# Run specific integration tests in order
echo "📋 Running TestBuildBinary..."
go test -v -run TestBuildBinary

echo "📋 Running TestGenerateTestGamesScript..."
go test -v -run TestGenerateTestGamesScript

echo "📋 Running TestIntegration..."
go test -v -run TestIntegration -timeout 10m

echo
echo "🎯 Running benchmarks..."
go test -bench=BenchmarkFullWorkflow -benchtime=3x -timeout 5m

cd ../..

echo
echo "✅ All tests completed successfully!"
echo
echo "📊 Test Summary:"
echo "   • Unit tests: ✅ Passed"
echo "   • Integration tests: ✅ Passed"
echo "   • Benchmarks: ✅ Completed"
echo
echo "🎉 ROM Organizer is working correctly!"

# Clean up test artifacts
if [ "$KEEP_ARTIFACTS" = false ]; then
    echo
    echo "🧹 Final cleanup..."
    # Clean up legacy artifacts that might be in wrong locations
    rm -rf test-games/ test-organized/ test-compressed/ test-decompressed/
    rm -rf temp-test-games/ bench-games-*/ organized-test-games/ compressed-test-games/ decompressed-test-games/
    rm -f rom-organizer-dev rom-organizer-dev.exe rom-organizer-bench rom-organizer-bench.exe
    # Clean up artifacts in cmd/rom-organizer directory (legacy + current binaries)
    rm -f cmd/rom-organizer/rom-organizer-dev cmd/rom-organizer/rom-organizer-dev.exe
    rm -f cmd/rom-organizer/rom-organizer-bench cmd/rom-organizer/rom-organizer-bench.exe
    rm -rf cmd/rom-organizer/test-games/ cmd/rom-organizer/test-organized/ cmd/rom-organizer/test-compressed/ cmd/rom-organizer/test-decompressed/
    rm -rf cmd/rom-organizer/temp-test-games/ cmd/rom-organizer/bench-games-*/
    # Clean up test artifacts in tests directory (current location)
    rm -rf tests/test-games/ tests/test-organized/ tests/test-compressed/ tests/test-decompressed/
    rm -rf tests/temp-test-games/ tests/bench-games-*/
    echo "   ✅ Cleanup complete"
else
    echo
    echo "🔒 Keeping test artifacts for inspection:"
    echo "   • Test games: tests/test-games/"
    echo "   • Organized games: tests/test-organized/"
    echo "   • Compressed games: tests/test-compressed/"
    echo "   • Decompressed games: tests/test-decompressed/"
    echo "   • Test binaries: cmd/rom-organizer/rom-organizer-dev*"
    echo "   • Temp test games: tests/temp-test-games/"
    echo "   💡 To clean up manually: rm -rf tests/test-* tests/temp-test-games tests/bench-games-* cmd/rom-organizer/rom-organizer-*"
fi 