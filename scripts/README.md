# Scripts Directory

This directory contains various scripts for development, testing, and maintenance of the ROM organizer.

## 🧪 Testing Scripts

### `run-tests.sh`

Comprehensive test suite runner that validates all ROM organizer functionality.

**Usage:**
```bash
# From the repository root
./scripts/run-tests.sh
```

**What it does:**
1. **Cleanup** - Removes any existing test artifacts
2. **Unit Tests** - Runs all internal package tests
3. **Build Test** - Builds the ROM organizer binary
4. **Integration Tests** - Comprehensive workflow testing:
   - Test game generation using fake PARAM.SFO files
   - Metadata extraction (single and multiple paths)
   - Organize command validation
   - Compress command validation  
   - Decompress command validation
   - File structure verification
5. **Benchmarks** - Performance testing of key operations
6. **Final Cleanup** - Removes all test artifacts

**Requirements:**
- Go 1.19+ installed
- 7-Zip available in PATH (for compression tests)
- Sufficient disk space for temporary test files

### Integration Test Suite (`cmd/rom-organizer/integration_test.go`)

Comprehensive Go test suite that can be run independently:

```bash
# Run all integration tests
cd cmd/rom-organizer
go test -v -run TestIntegration

# Run specific test components
go test -v -run TestBuildBinary
go test -v -run TestGenerateTestGamesScript
go test -v -run TestMetadataExtraction

# Run benchmarks
go test -bench=BenchmarkFullWorkflow -benchtime=5x
```

**Test Coverage:**
- ✅ Binary building and validation
- ✅ Test game generation with fake PARAM.SFO files
- ✅ Metadata extraction (regular, JSON, verbose output)
- ✅ Multiple path operations
- ✅ Organize, compress, and decompress workflows
- ✅ File structure validation
- ✅ Error handling and edge cases

## 🛠️ Development Scripts

### `dev/generate-test-games.go`

Generates fake PS3 games for testing without any copyright concerns.

**Features:**
- Creates realistic but completely fictional game data
- Generates valid binary PARAM.SFO files
- Includes 15 pre-defined fake game titles
- Supports custom output directories and game counts
- Reproducible generation with seed support

**Usage:**
```bash
# Generate 5 test games (default)
go run scripts/dev/generate-test-games.go

# Generate 10 games in custom directory
go run scripts/dev/generate-test-games.go -count 10 -output my-test-games

# Clean and regenerate with specific seed
go run scripts/dev/generate-test-games.go -clean -seed 12345
```

See `dev/README.md` for detailed usage information.

## 🚀 Running Tests

### Quick Test
```bash
# Just run the test suite
./scripts/run-tests.sh
```

### Manual Testing
```bash
# Generate test games manually
go run scripts/dev/generate-test-games.go -count 3

# Build and test manually
go build -o rom-organizer-test ./cmd/rom-organizer
./rom-organizer-test metadata test-games/*/
./rom-organizer-test organize test-games/*/ --output organized
```

### CI/CD Integration

The test suite is designed for CI/CD environments:

```yaml
# Example GitHub Actions usage
- name: Run ROM Organizer Tests
  run: |
    chmod +x scripts/run-tests.sh
    ./scripts/run-tests.sh
```

## 📊 Test Output

The test suite provides detailed progress reporting:

```
🧪 ROM Organizer Test Suite
==========================

🧹 Cleaning up existing test artifacts...
   ✅ Cleanup complete

🔬 Running unit tests...
   ✅ Unit tests passed

🚀 Running integration tests...
   • Test game generation
   • Binary building  
   • Metadata extraction (single & multiple paths)
   • Organize command
   • Compress command
   • Decompress command

📋 Running TestIntegration...
   ✅ Generated 5 test games successfully
   ✅ Verified 5 test games with correct structure
   ✅ Metadata extraction tests passed
   ✅ Organized 5 games successfully
   ✅ Compressed 5 games successfully
   ✅ Decompressed 5 games successfully

✅ All tests completed successfully!
```

## 🔧 Troubleshooting

### Common Issues

**"7z command not found"**
- Install 7-Zip and ensure it's in your PATH
- On Windows: Download from 7-zip.org and add to PATH
- On Linux: `sudo apt install p7zip-full`
- On macOS: `brew install p7zip`

**"Permission denied"** 
- Make sure scripts are executable: `chmod +x scripts/run-tests.sh`
- Check that Go has write permissions in the test directory

**Tests fail on Windows**
- Ensure you're using Git Bash or WSL for shell script execution
- Path separators may need adjustment for Windows-specific testing

### Test Data

All test data is automatically generated and cleaned up. No external dependencies on real game files are required.

The fake games generated include:
- Valid PARAM.SFO binary files with realistic structure
- Fictional game titles to avoid copyright issues
- PS3-style title IDs (BLUS/BLES/BCUS/BCES format)
- Complete directory structures matching real PS3 games 