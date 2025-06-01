package organizer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NeilGraham/rom-organizer/internal/common"
	"github.com/NeilGraham/rom-organizer/internal/detect"
	"github.com/NeilGraham/rom-organizer/internal/parsers"
)

// OrganizeOptions holds options for organizing operations
type OrganizeOptions struct {
	OutputDir  string
	Force      bool
	Verbose    bool
	MoveSource bool
}

// OrganizeGame organizes a ROM game while keeping it in its existing format
func OrganizeGame(sourcePath string, opts OrganizeOptions) error {
	if opts.Verbose {
		fmt.Printf("Organizing ROM game from: %s\n", sourcePath)
		fmt.Printf("Output directory: %s\n", opts.OutputDir)
	}

	// Use the new detection system to identify console type and game location
	detection, err := detect.DetectConsole(sourcePath)
	if err != nil {
		return fmt.Errorf("detecting console type: %w", err)
	}

	if opts.Verbose {
		fmt.Printf("Console Detection Results:\n")
		fmt.Printf("Console Type: %s (confidence: %.2f)\n", detection.ConsoleType.String(), detection.Confidence)
		fmt.Printf("Game Path: %s\n", detection.GamePath)
		fmt.Printf("Indicator: %s\n", detection.IndicatorFound)
	}

	// Handle different console types
	switch detection.ConsoleType {
	case detect.PS3:
		return organizePS3Game(sourcePath, detection, opts)
	case detect.Unknown:
		if len(detection.AmbiguousFiles) > 0 {
			return fmt.Errorf("found ambiguous files but console-specific organization not yet implemented - detected %d ambiguous files", len(detection.AmbiguousFiles))
		}
		return fmt.Errorf("unable to determine console type for: %s", sourcePath)
	default:
		return fmt.Errorf("organization for %s is not yet implemented", detection.ConsoleType.String())
	}
}

// organizePS3Game handles organization of PS3 games specifically
func organizePS3Game(sourcePath string, detection *detect.DetectionResult, opts OrganizeOptions) error {
	// Check if this is an existing organized game directory first
	organizedInfo, err := common.DetectOrganizedDirectory(sourcePath, opts.Verbose)
	if err != nil {
		return fmt.Errorf("checking if directory is organized: %w", err)
	}

	if organizedInfo.IsOrganized {
		if opts.MoveSource {
			fmt.Printf("⚠️  WARNING: --move flag ignored for already organized directories (safety measure)\n")
		}

		if opts.Verbose {
			fmt.Printf("Source is already an organized game directory\n")
		}

		// Determine format from contents
		format := "Unknown"
		if organizedInfo.HasCompressed && organizedInfo.HasDecompressed {
			format = "Mixed (both game.7z and game/ folder)"
		} else if organizedInfo.HasCompressed {
			format = "Compressed (game.7z)"
		} else if organizedInfo.HasDecompressed {
			format = "Decompressed (game/ folder)"
		}

		fmt.Printf("Directory is already organized:\n")
		fmt.Printf("  Title: %s\n", organizedInfo.GameInfo.Title)
		fmt.Printf("  Title ID: %s\n", organizedInfo.GameInfo.TitleID)
		fmt.Printf("  Format: %s\n", format)
		fmt.Printf("  Location: %s\n", sourcePath)

		return nil
	}

	// Extract game information using the detected game path
	gameInfo, err := extractPS3GameInfo(detection)
	if err != nil {
		return fmt.Errorf("extracting PS3 game info: %w", err)
	}

	// Generate target path
	targetPath := common.GenerateTargetPath(gameInfo, opts.OutputDir)

	if opts.Verbose {
		fmt.Printf("Game Title: %s\n", gameInfo.Title)
		fmt.Printf("Title ID: %s\n", gameInfo.TitleID)
		fmt.Printf("Target directory: %s\n", targetPath)
	}

	// Create target directory structure
	if err := common.CreateTargetStructure(targetPath, opts.Force); err != nil {
		return err
	}

	// Clean up existing game files if force is enabled
	if opts.Force {
		gameFile := filepath.Join(targetPath, "game.7z")
		gameDir := filepath.Join(targetPath, "game")

		if _, err := os.Stat(gameFile); err == nil {
			if opts.Verbose {
				fmt.Printf("Removing existing game.7z file...\n")
			}
			if err := os.Remove(gameFile); err != nil {
				return fmt.Errorf("removing existing game.7z: %w", err)
			}
		}

		if _, err := os.Stat(gameDir); err == nil {
			if opts.Verbose {
				fmt.Printf("Removing existing game/ directory...\n")
			}
			if err := os.RemoveAll(gameDir); err != nil {
				return fmt.Errorf("removing existing game/ directory: %w", err)
			}
		}
	}

	// Organize the game files
	gameDir := filepath.Join(targetPath, "game")

	if opts.MoveSource {
		if opts.Verbose {
			fmt.Printf("Moving game files to game/ folder (keeping decompressed format)...\n")
		}

		// Move the detected game directory to the target
		if err := moveGameDirectory(detection.GamePath, gameDir, opts.Verbose); err != nil {
			return fmt.Errorf("moving game directory: %w", err)
		}

		// Handle cleanup of the original source directory
		if err := cleanupSourceAfterMove(sourcePath, detection.GamePath, opts); err != nil {
			return fmt.Errorf("cleaning up source directory: %w", err)
		}

	} else {
		if opts.Verbose {
			fmt.Printf("Copying game files to game/ folder (keeping decompressed format)...\n")
		}

		// Copy the detected game directory to the target
		if err := common.CopyDir(detection.GamePath, gameDir); err != nil {
			return fmt.Errorf("copying game directory: %w", err)
		}
	}

	fmt.Printf("Successfully organized PS3 game:\n")
	fmt.Printf("  Title: %s\n", gameInfo.Title)
	fmt.Printf("  Title ID: %s\n", gameInfo.TitleID)
	fmt.Printf("  Format: Decompressed (game/ folder)\n")
	fmt.Printf("  Output: %s\n", targetPath)

	return nil
}

// extractPS3GameInfo extracts game information from a PS3 detection result
func extractPS3GameInfo(detection *detect.DetectionResult) (*common.GameInfo, error) {
	// Find the PARAM.SFO file based on the detection result
	var paramSFOPath string

	if detection.IndicatorFound == "PS3_GAME" {
		paramSFOPath = filepath.Join(detection.GamePath, "PS3_GAME", "PARAM.SFO")
	} else if detection.IndicatorFound == "PARAM.SFO" {
		paramSFOPath = filepath.Join(detection.GamePath, "PARAM.SFO")
	} else {
		return nil, fmt.Errorf("unexpected PS3 indicator: %s", detection.IndicatorFound)
	}

	// Read and parse the PARAM.SFO file directly
	paramSFOData, err := os.ReadFile(paramSFOPath)
	if err != nil {
		return nil, fmt.Errorf("reading PARAM.SFO: %w", err)
	}

	paramSFO, err := parsers.ParseParamSFO(paramSFOData)
	if err != nil {
		return nil, fmt.Errorf("parsing PARAM.SFO: %w", err)
	}

	title := paramSFO.GetTitle()
	titleID := paramSFO.GetTitleID()

	if title == "" {
		return nil, fmt.Errorf("game title not found in PARAM.SFO")
	}
	if titleID == "" {
		return nil, fmt.Errorf("title ID not found in PARAM.SFO")
	}

	return &common.GameInfo{
		Title:   title,
		TitleID: titleID,
		Source:  detection.GamePath,
	}, nil
}

// moveGameDirectory moves a game directory from source to destination
func moveGameDirectory(src, dest string, verbose bool) error {
	if verbose {
		fmt.Printf("Moving directory: %s -> %s\n", src, dest)
	}

	// First copy the directory
	if err := common.CopyDir(src, dest); err != nil {
		return fmt.Errorf("copying directory during move: %w", err)
	}

	// Then remove the source
	if err := os.RemoveAll(src); err != nil {
		return fmt.Errorf("removing source directory after move: %w", err)
	}

	if verbose {
		fmt.Printf("Successfully moved directory\n")
	}

	return nil
}

// cleanupSourceAfterMove handles cleanup of the source directory after moving game files
func cleanupSourceAfterMove(originalSourcePath, gameSourcePath string, opts OrganizeOptions) error {
	// If the user specified the exact game directory, we're done
	if originalSourcePath == gameSourcePath {
		if opts.Verbose {
			fmt.Printf("Source directory was the exact game directory - cleanup complete\n")
		}
		return nil
	}

	// User specified a parent directory, check if it's now empty or should be cleaned up
	if opts.Verbose {
		fmt.Printf("Checking if source directory should be cleaned up: %s\n", originalSourcePath)
	}

	// Check if the directory is effectively empty
	isEmpty, err := common.IsDirEffectivelyEmpty(originalSourcePath)
	if err != nil {
		return fmt.Errorf("checking if source directory is empty: %w", err)
	}

	if isEmpty {
		// Safe to remove - directory contains no significant files
		if opts.Verbose {
			fmt.Printf("Removing empty source directory: %s\n", originalSourcePath)
		}
		if err := os.RemoveAll(originalSourcePath); err != nil {
			return fmt.Errorf("removing empty source directory: %w", err)
		}
		if opts.Verbose {
			fmt.Printf("Successfully removed empty source directory\n")
		}
	} else {
		// Directory contains files - check if force is enabled
		if opts.Force {
			if opts.Verbose {
				fmt.Printf("⚠️  Forcefully removing source directory with remaining files: %s\n", originalSourcePath)
			}
			if err := os.RemoveAll(originalSourcePath); err != nil {
				return fmt.Errorf("forcefully removing source directory: %w", err)
			}
			if opts.Verbose {
				fmt.Printf("Successfully removed source directory with force\n")
			}
		} else {
			fmt.Printf("⚠️  WARNING: Source directory contains remaining files and was not deleted: %s\n", originalSourcePath)
			fmt.Printf("    Use --force to delete the source directory even with remaining files\n")
		}
	}

	return nil
}

// OrganizeGames organizes multiple ROM games while keeping them in their existing format
func OrganizeGames(sourcePaths []string, opts OrganizeOptions) error {
	var errors []error
	successCount := 0
	totalCount := len(sourcePaths)

	for i, sourcePath := range sourcePaths {
		if opts.Verbose {
			fmt.Printf("\n=== Processing %d/%d: %s ===\n", i+1, totalCount, sourcePath)
		}

		if err := OrganizeGame(sourcePath, opts); err != nil {
			fmt.Printf("Error processing %s: %v\n", sourcePath, err)
			errors = append(errors, fmt.Errorf("%s: %w", sourcePath, err))
		} else {
			successCount++
		}
	}

	// Print summary
	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Successfully processed: %d/%d games\n", successCount, totalCount)
	if len(errors) > 0 {
		fmt.Printf("Failed: %d games\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
		return fmt.Errorf("failed to process %d out of %d games", len(errors), totalCount)
	}

	return nil
}
