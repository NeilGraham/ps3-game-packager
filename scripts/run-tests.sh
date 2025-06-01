#!/bin/bash

# ROM Organizer Test Suite Runner
# This script runs the complete test suite including integration tests

set -e  # Exit on any error

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
cd "$script_dir/.."

echo "🧪 ROM Organizer Test Suite"
echo "=========================="
echo


# Clean up any existing test artifacts
echo "🧹 Cleaning up existing test artifacts..."
rm -rf test-games/ test-organized/ test-compressed/ test-decompressed/
rm -rf temp-test-games/ bench-games-*/ organized-test-games/ compressed-test-games/ decompressed-test-games/
rm -f rom-organizer-dev rom-organizer-dev.exe rom-organizer-bench rom-organizer-bench.exe
echo "   ✅ Cleanup complete"
echo

# Run unit tests
echo "🔬 Running unit tests..."
go test ./internal/... -v
echo "   ✅ Unit tests passed"
echo

# Build the test binary
echo "🔨 Building test binary..."
go test -c ./cmd/rom-organizer -o rom-organizer-test
echo "   ✅ Test binary built"
echo

# Run integration tests with verbose output
echo "🚀 Running integration tests..."
echo "   This will test the complete workflow:"
echo "   • Test game generation"
echo "   • Binary building"
echo "   • Metadata extraction (single & multiple paths)"
echo "   • Organize command"
echo "   • Compress command"
echo "   • Decompress command"
echo

# Change to cmd/rom-organizer directory for tests since they expect to be run from there
cd cmd/rom-organizer

# Run specific integration tests in order
echo "📋 Running TestBuildBinary..."
go test -v -run TestBuildBinary

echo "📋 Running TestGenerateTestGamesScript..."
go test -v -run TestGenerateTestGamesScript

echo "📋 Running TestIntegration..."
go test -v -run TestIntegration -timeout 10m

cd ../..

echo
echo "🎯 Running benchmarks..."
cd cmd/rom-organizer
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
echo
echo "🧹 Final cleanup..."
rm -rf test-games/ test-organized/ test-compressed/ test-decompressed/
rm -rf temp-test-games/ bench-games-*/ organized-test-games/ compressed-test-games/ decompressed-test-games/
rm -f rom-organizer-dev rom-organizer-dev.exe rom-organizer-bench rom-organizer-bench.exe rom-organizer-test
echo "   ✅ Cleanup complete" 