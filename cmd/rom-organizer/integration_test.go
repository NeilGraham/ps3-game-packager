package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const (
	testGamesDir        = "../../tests/test-games"
	testOrganizedDir    = "../../tests/test-organized"
	testCompressedDir   = "../../tests/test-compressed"
	testDecompressedDir = "../../tests/test-decompressed"
	testGameCount       = 5
)

// getBinaryPath returns the correct path to the ROM organizer binary
func getBinaryPath() string {
	if runtime.GOOS == "windows" {
		return ".\\rom-organizer-dev.exe"
	}
	return "./rom-organizer-dev"
}

// TestIntegration runs a comprehensive integration test of all ROM organizer functionality
func TestIntegration(t *testing.T) {
	// Ensure binary is built first
	t.Log("Building ROM organizer binary...")
	binaryName := "rom-organizer-dev"
	if runtime.GOOS == "windows" {
		binaryName = "rom-organizer-dev.exe"
	}
	cmd := exec.Command("go", "build", "-o", binaryName, ".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build ROM organizer binary: %v\nOutput: %s", err, output)
	}

	// Clean up any existing test directories
	t.Log("Cleaning up existing test directories...")
	cleanupTestDirs(t)

	// Generate test games
	t.Log("Generating test games...")
	generateTestGames(t)

	// Verify test games were created
	t.Log("Verifying test games were created...")
	verifyTestGamesCreated(t)

	// Test metadata extraction
	t.Log("Testing metadata extraction...")
	testMetadataExtraction(t)

	// Test organization
	t.Log("Testing organize command...")
	testOrganize(t)

	// Test compression
	t.Log("Testing compress command...")
	testCompress(t)

	// Test decompression
	t.Log("Testing decompress command...")
	testDecompress(t)

	// Test multiple path operations
	t.Log("Testing multiple path operations...")
	testMultiplePaths(t)

	// Clean up after tests (unless --keep flag was used in shell script)
	keepArtifacts := os.Getenv("KEEP_TEST_ARTIFACTS") == "true"
	if keepArtifacts {
		t.Log("🔒 Keeping test artifacts (KEEP_TEST_ARTIFACTS environment variable set)")
	} else {
		t.Log("Cleaning up test directories...")
		cleanupTestDirs(t)
	}

	t.Log("✅ All integration tests passed!")
}

// cleanupTestDirs removes all test directories
func cleanupTestDirs(t *testing.T) {
	dirs := []string{testGamesDir, testOrganizedDir, testCompressedDir, testDecompressedDir}
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: could not remove %s: %v", dir, err)
		}
	}
}

// generateTestGames creates fake test games using the development script
func generateTestGames(t *testing.T) {
	cmd := exec.Command("go", "run", "../../tests/generate-test-games.go",
		"-count", fmt.Sprintf("%d", testGameCount),
		"-output", testGamesDir,
		"-clean")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to generate test games: %v\nOutput: %s", err, output)
	}

	t.Logf("Generated %d test games successfully", testGameCount)
}

// verifyTestGamesCreated checks that the expected number of test games were created
func verifyTestGamesCreated(t *testing.T) {
	entries, err := os.ReadDir(testGamesDir)
	if err != nil {
		t.Fatalf("Failed to read test games directory: %v", err)
	}

	gameCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			gameCount++

			// Verify each game has the required structure
			gameDir := filepath.Join(testGamesDir, entry.Name())
			ps3GameDir := filepath.Join(gameDir, "PS3_GAME")
			paramSFOPath := filepath.Join(ps3GameDir, "PARAM.SFO")
			sfbPath := filepath.Join(gameDir, "PS3_DISC.SFB")

			if _, err := os.Stat(ps3GameDir); os.IsNotExist(err) {
				t.Errorf("Game %s missing PS3_GAME directory", entry.Name())
			}

			if _, err := os.Stat(paramSFOPath); os.IsNotExist(err) {
				t.Errorf("Game %s missing PARAM.SFO file", entry.Name())
			}

			if _, err := os.Stat(sfbPath); os.IsNotExist(err) {
				t.Errorf("Game %s missing PS3_DISC.SFB file", entry.Name())
			}
		}
	}

	if gameCount != testGameCount {
		t.Fatalf("Expected %d games, found %d", testGameCount, gameCount)
	}

	t.Logf("✅ Verified %d test games with correct structure", gameCount)
}

// testMetadataExtraction tests the metadata command on generated games
func testMetadataExtraction(t *testing.T) {
	// Get first test game
	entries, err := os.ReadDir(testGamesDir)
	if err != nil {
		t.Fatalf("Failed to read test games directory: %v", err)
	}

	var firstGamePath string
	for _, entry := range entries {
		if entry.IsDir() {
			firstGamePath = filepath.Join(testGamesDir, entry.Name())
			break
		}
	}

	if firstGamePath == "" {
		t.Fatal("No test games found for metadata testing")
	}

	// Test regular metadata output
	t.Run("regular_output", func(t *testing.T) {
		cmd := exec.Command(getBinaryPath(), "metadata", firstGamePath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Metadata command failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Summary:") {
			t.Error("Metadata output missing Summary section")
		}
		if !strings.Contains(outputStr, "Game Title:") {
			t.Error("Metadata output missing Game Title")
		}
		if !strings.Contains(outputStr, "Game ID:") {
			t.Error("Metadata output missing Game ID")
		}
	})

	// Test JSON metadata output
	t.Run("json_output", func(t *testing.T) {
		cmd := exec.Command(getBinaryPath(), "metadata", "--json", firstGamePath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("JSON metadata command failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "\"summary\"") {
			t.Error("JSON metadata output missing summary section")
		}
		if !strings.Contains(outputStr, "\"title\"") {
			t.Error("JSON metadata output missing title field")
		}
		if !strings.Contains(outputStr, "\"gameId\"") {
			t.Error("JSON metadata output missing gameId field")
		}
	})

	// Test verbose metadata output
	t.Run("verbose_output", func(t *testing.T) {
		cmd := exec.Command(getBinaryPath(), "metadata", "--verbose", firstGamePath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Verbose metadata command failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Console Detection Results:") {
			t.Error("Verbose metadata output missing console detection section")
		}
		if !strings.Contains(outputStr, "ROM Metadata Parser") {
			t.Error("Verbose metadata output missing parser section")
		}
	})

	t.Log("✅ Metadata extraction tests passed")
}

// testOrganize tests the organize command
func testOrganize(t *testing.T) {
	// Get all test games
	entries, err := os.ReadDir(testGamesDir)
	if err != nil {
		t.Fatalf("Failed to read test games directory: %v", err)
	}

	var gamePaths []string
	for _, entry := range entries {
		if entry.IsDir() {
			gamePaths = append(gamePaths, filepath.Join(testGamesDir, entry.Name()))
		}
	}

	if len(gamePaths) == 0 {
		t.Fatal("No test games found for organize testing")
	}

	// Test organize command
	args := []string{"organize", "--output", testOrganizedDir}
	args = append(args, gamePaths...)

	cmd := exec.Command(getBinaryPath(), args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Organize command failed: %v\nOutput: %s", err, output)
	}

	// Verify organized games were created
	organizedEntries, err := os.ReadDir(testOrganizedDir)
	if err != nil {
		t.Fatalf("Failed to read organized directory: %v", err)
	}

	organizedCount := 0
	for _, entry := range organizedEntries {
		if entry.IsDir() {
			organizedCount++

			// Verify organized structure (should have game/ folder)
			gameDir := filepath.Join(testOrganizedDir, entry.Name(), "game")
			if _, err := os.Stat(gameDir); os.IsNotExist(err) {
				t.Errorf("Organized game %s missing game/ directory", entry.Name())
			}
		}
	}

	if organizedCount != len(gamePaths) {
		t.Fatalf("Expected %d organized games, found %d", len(gamePaths), organizedCount)
	}

	t.Logf("✅ Organized %d games successfully", organizedCount)
}

// testCompress tests the compress command
func testCompress(t *testing.T) {
	// Get all test games
	entries, err := os.ReadDir(testGamesDir)
	if err != nil {
		t.Fatalf("Failed to read test games directory: %v", err)
	}

	var gamePaths []string
	for _, entry := range entries {
		if entry.IsDir() {
			gamePaths = append(gamePaths, filepath.Join(testGamesDir, entry.Name()))
		}
	}

	if len(gamePaths) == 0 {
		t.Fatal("No test games found for compress testing")
	}

	// Test compress command
	args := []string{"compress", "--output", testCompressedDir}
	args = append(args, gamePaths...)

	cmd := exec.Command(getBinaryPath(), args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Compress command failed: %v\nOutput: %s", err, output)
	}

	// Verify compressed games were created
	compressedEntries, err := os.ReadDir(testCompressedDir)
	if err != nil {
		t.Fatalf("Failed to read compressed directory: %v", err)
	}

	compressedCount := 0
	for _, entry := range compressedEntries {
		if entry.IsDir() {
			compressedCount++

			// Verify compressed structure (should have game.7z file)
			game7zPath := filepath.Join(testCompressedDir, entry.Name(), "game.7z")
			if _, err := os.Stat(game7zPath); os.IsNotExist(err) {
				t.Errorf("Compressed game %s missing game.7z file", entry.Name())
			}
		}
	}

	if compressedCount != len(gamePaths) {
		t.Fatalf("Expected %d compressed games, found %d", len(gamePaths), compressedCount)
	}

	t.Logf("✅ Compressed %d games successfully", compressedCount)
}

// testDecompress tests the decompress command
func testDecompress(t *testing.T) {
	// Get all test games
	entries, err := os.ReadDir(testGamesDir)
	if err != nil {
		t.Fatalf("Failed to read test games directory: %v", err)
	}

	var gamePaths []string
	for _, entry := range entries {
		if entry.IsDir() {
			gamePaths = append(gamePaths, filepath.Join(testGamesDir, entry.Name()))
		}
	}

	if len(gamePaths) == 0 {
		t.Fatal("No test games found for decompress testing")
	}

	// Test decompress command
	args := []string{"decompress", "--output", testDecompressedDir}
	args = append(args, gamePaths...)

	cmd := exec.Command(getBinaryPath(), args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Decompress command failed: %v\nOutput: %s", err, output)
	}

	// Verify decompressed games were created
	decompressedEntries, err := os.ReadDir(testDecompressedDir)
	if err != nil {
		t.Fatalf("Failed to read decompressed directory: %v", err)
	}

	decompressedCount := 0
	for _, entry := range decompressedEntries {
		if entry.IsDir() {
			decompressedCount++

			// Verify decompressed structure (should have game/ folder)
			gameDir := filepath.Join(testDecompressedDir, entry.Name(), "game")
			if _, err := os.Stat(gameDir); os.IsNotExist(err) {
				t.Errorf("Decompressed game %s missing game/ directory", entry.Name())
			}
		}
	}

	if decompressedCount != len(gamePaths) {
		t.Fatalf("Expected %d decompressed games, found %d", len(gamePaths), decompressedCount)
	}

	t.Logf("✅ Decompressed %d games successfully", decompressedCount)
}

// testMultiplePaths tests multiple path operations with metadata command
func testMultiplePaths(t *testing.T) {
	// Get first two test games
	entries, err := os.ReadDir(testGamesDir)
	if err != nil {
		t.Fatalf("Failed to read test games directory: %v", err)
	}

	var gamePaths []string
	for _, entry := range entries {
		if entry.IsDir() && len(gamePaths) < 2 {
			gamePaths = append(gamePaths, filepath.Join(testGamesDir, entry.Name()))
		}
	}

	if len(gamePaths) < 2 {
		t.Skip("Need at least 2 test games for multiple path testing")
	}

	// Test metadata command with multiple paths
	args := []string{"metadata"}
	args = append(args, gamePaths...)

	cmd := exec.Command(getBinaryPath(), args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Multiple path metadata command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Should contain metadata headers for each game
	expectedHeaders := 0
	for _, gamePath := range gamePaths {
		if strings.Contains(outputStr, fmt.Sprintf("=== Metadata for: %s ===", gamePath)) {
			expectedHeaders++
		}
	}

	if expectedHeaders != len(gamePaths) {
		t.Errorf("Expected %d metadata headers, found %d", len(gamePaths), expectedHeaders)
	}

	// Should contain multiple Summary sections
	summaryCount := strings.Count(outputStr, "Summary:")
	if summaryCount != len(gamePaths) {
		t.Errorf("Expected %d Summary sections, found %d", len(gamePaths), summaryCount)
	}

	t.Logf("✅ Multiple path metadata test passed for %d games", len(gamePaths))
}

// TestBuildBinary ensures the ROM organizer binary is built before running integration tests
func TestBuildBinary(t *testing.T) {
	t.Log("Building ROM organizer binary...")

	binaryName := "rom-organizer-dev"
	if runtime.GOOS == "windows" {
		binaryName = "rom-organizer-dev.exe"
	}
	cmd := exec.Command("go", "build", "-o", binaryName, ".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build ROM organizer binary: %v\nOutput: %s", err, output)
	}

	// Verify binary exists and is executable
	if _, err := os.Stat(binaryName); os.IsNotExist(err) {
		t.Fatal("ROM organizer binary was not created")
	}

	t.Log("✅ ROM organizer binary built successfully")
}

// TestGenerateTestGamesScript tests the test game generation script independently
func TestGenerateTestGamesScript(t *testing.T) {
	tempDir := "../../tests/temp-test-games"

	// Clean up temp directory (unless --keep flag was used)
	keepArtifacts := os.Getenv("KEEP_TEST_ARTIFACTS") == "true"
	defer func() {
		if !keepArtifacts {
			if err := os.RemoveAll(tempDir); err != nil {
				t.Logf("Warning: could not clean up temp directory: %v", err)
			}
		} else {
			t.Logf("🔒 Keeping temp test games directory: %s", tempDir)
		}
	}()

	t.Log("Testing test game generation script...")

	cmd := exec.Command("go", "run", "../../tests/generate-test-games.go",
		"-count", "3",
		"-output", tempDir,
		"-seed", "12345") // Use fixed seed for reproducible tests

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Test game generation failed: %v\nOutput: %s", err, output)
	}

	// Verify games were created
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp test games directory: %v", err)
	}

	gameCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			gameCount++
		}
	}

	if gameCount != 3 {
		t.Fatalf("Expected 3 test games, found %d", gameCount)
	}

	t.Log("✅ Test game generation script works correctly")
}

// BenchmarkFullWorkflow benchmarks the complete workflow
func BenchmarkFullWorkflow(t *testing.B) {
	// Build binary once
	benchBinaryName := "rom-organizer-bench"
	if runtime.GOOS == "windows" {
		benchBinaryName = "rom-organizer-bench.exe"
	}
	if err := exec.Command("go", "build", "-o", benchBinaryName, ".").Run(); err != nil {
		t.Fatalf("Failed to build binary for benchmark: %v", err)
	}
	defer os.Remove(benchBinaryName)

	benchBinaryPath := benchBinaryName
	if runtime.GOOS != "windows" {
		benchBinaryPath = "./" + benchBinaryName
	}

	t.ResetTimer()

	for i := 0; i < t.N; i++ {
		benchDir := fmt.Sprintf("../../tests/bench-games-%d", i)

		// Generate test games
		if err := exec.Command("go", "run", "../../tests/generate-test-games.go",
			"-count", "2", "-output", benchDir, "-clean").Run(); err != nil {
			t.Fatalf("Benchmark game generation failed: %v", err)
		}

		// Get game paths
		entries, err := os.ReadDir(benchDir)
		if err != nil {
			t.Fatalf("Failed to read benchmark games: %v", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				gamePath := filepath.Join(benchDir, entry.Name())

				// Test metadata extraction
				exec.Command(benchBinaryPath, "metadata", gamePath).Run()
			}
		}

		// Clean up
		os.RemoveAll(benchDir)
	}
}
