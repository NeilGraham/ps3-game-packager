package detect

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// MaxSearchDepth limits how deep we'll search to prevent infinite loops
	MaxSearchDepth = 8
)

// DetectConsole analyzes a path and attempts to determine what console type it contains
// It performs a breadth-first search, stopping early when definitive indicators are found
func DetectConsole(rootPath string) (*DetectionResult, error) {
	// Verify the path exists
	if _, err := os.Stat(rootPath); err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}

	result := &DetectionResult{
		ConsoleType:    Unknown,
		GamePath:       rootPath,
		Confidence:     0.0,
		AmbiguousFiles: make([]string, 0),
	}

	// Start recursive search
	err := searchDirectory(rootPath, rootPath, 0, result)
	if err != nil {
		return nil, err
	}

	// If we found a definitive indicator, we're done
	if result.IsValid() {
		return result, nil
	}

	// TODO: In future iterations, analyze ambiguous files here
	if len(result.AmbiguousFiles) > 0 {
		// For now, just set low confidence if we found ambiguous files
		result.Confidence = 0.3
		result.IndicatorFound = fmt.Sprintf("Found %d ambiguous files", len(result.AmbiguousFiles))
	}

	return result, nil
}

// searchDirectory recursively searches a directory for console indicators
func searchDirectory(currentPath, rootPath string, depth int, result *DetectionResult) error {
	// Prevent infinite recursion
	if depth > MaxSearchDepth {
		return nil
	}

	// If we already found a high-confidence match, stop searching
	if result.IsHighConfidence() {
		return nil
	}

	entries, err := os.ReadDir(currentPath)
	if err != nil {
		// Don't fail the entire search if we can't read one directory
		return nil
	}

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(currentPath, name)

		// Skip hidden files and directories
		if name[0] == '.' {
			continue
		}

		// Check for definitive indicators
		if IsDefinitiveIndicator(name) {
			console := GetConsoleFromIndicator(name)

			// For directory indicators (like PS3_GAME), the parent is the game path
			// For file indicators (like PARAM.SFO), use the current directory
			var gamePath string
			if entry.IsDir() {
				gamePath = currentPath // Parent of the indicator directory
			} else {
				gamePath = currentPath // Directory containing the indicator file
			}

			result.ConsoleType = console
			result.GamePath = gamePath
			result.Confidence = 0.95 // High confidence for definitive indicators
			result.IndicatorFound = name
			result.SearchDepth = depth

			return nil // Stop searching once we find a definitive indicator
		}

		// Check for ambiguous files
		if !entry.IsDir() && IsAmbiguousFile(name) {
			result.AmbiguousFiles = append(result.AmbiguousFiles, fullPath)
		}

		// Recursively search subdirectories
		if entry.IsDir() {
			err := searchDirectory(fullPath, rootPath, depth+1, result)
			if err != nil {
				continue // Continue searching other directories
			}

			// If we found a definitive indicator in a subdirectory, stop
			if result.IsHighConfidence() {
				return nil
			}
		}
	}

	return nil
}

// DetectConsoleFromFile analyzes a single file and attempts to determine its console type
// This is useful for analyzing individual ROM files
func DetectConsoleFromFile(filePath string) (*DetectionResult, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("file does not exist: %w", err)
	}

	result := &DetectionResult{
		GamePath:       filePath,
		AmbiguousFiles: make([]string, 0),
	}

	filename := filepath.Base(filePath)

	// Check if it's a definitive indicator file
	if IsDefinitiveIndicator(filename) {
		result.ConsoleType = GetConsoleFromIndicator(filename)
		result.Confidence = 0.9
		result.IndicatorFound = filename
		return result, nil
	}

	// Check if it's an ambiguous file
	if IsAmbiguousFile(filename) {
		result.AmbiguousFiles = append(result.AmbiguousFiles, filePath)
		result.Confidence = 0.3
		result.IndicatorFound = fmt.Sprintf("Ambiguous file: %s", filename)
	}

	return result, nil
}
