package packager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NeilGraham/rom-organizer/internal/common"
)

// PackageOptions holds options for packaging operations
type PackageOptions struct {
	OutputDir string
	Force     bool
	Verbose   bool
}

// PackageGame packages a PS3 game into compressed format (game.7z)
func PackageGame(sourcePath string, opts PackageOptions) error {
	if opts.Verbose {
		fmt.Printf("Packaging PS3 game from: %s\n", sourcePath)
		fmt.Printf("Output directory: %s\n", opts.OutputDir)
	}

	// First check if this is an organized directory
	organizedInfo, err := common.DetectOrganizedDirectory(sourcePath, opts.Verbose)
	if err != nil {
		return fmt.Errorf("checking if directory is organized: %w", err)
	}

	if organizedInfo.IsOrganized {
		return handleOrganizedDirectoryForPackaging(sourcePath, organizedInfo, opts)
	}

	// Extract game information for unorganized directories
	gameInfo, err := common.ExtractGameInfo(sourcePath, opts.Verbose)
	if err != nil {
		return err
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

	// Create compressed game.7z archive
	game7zPath := filepath.Join(targetPath, "game.7z")

	if opts.Verbose {
		fmt.Printf("Creating game.7z archive...\n")
	}

	if err := common.Create7zArchive(gameInfo.Source, game7zPath); err != nil {
		return fmt.Errorf("creating game.7z archive: %w", err)
	}

	fmt.Printf("Successfully packaged PS3 game:\n")
	fmt.Printf("  Title: %s\n", gameInfo.Title)
	fmt.Printf("  Title ID: %s\n", gameInfo.TitleID)
	fmt.Printf("  Format: Compressed (game.7z)\n")
	fmt.Printf("  Output: %s\n", targetPath)

	return nil
}

// handleOrganizedDirectoryForPackaging handles packaging of organized directories
func handleOrganizedDirectoryForPackaging(sourcePath string, organizedInfo *common.OrganizedDirInfo, opts PackageOptions) error {
	// If it already has game.7z (compressed), check if we need to do anything
	if organizedInfo.HasCompressed && !organizedInfo.HasDecompressed {
		if opts.Verbose {
			fmt.Printf("Source is already in compressed format (game.7z)\n")
			fmt.Printf("No action needed\n")
		}
		fmt.Printf("Directory is already packaged:\n")
		fmt.Printf("  Title: %s\n", organizedInfo.GameInfo.Title)
		fmt.Printf("  Title ID: %s\n", organizedInfo.GameInfo.TitleID)
		fmt.Printf("  Format: Compressed (game.7z)\n")
		fmt.Printf("  Location: %s\n", sourcePath)
		return nil
	}

	// If it has game/ folder (decompressed), compress it to game.7z
	if organizedInfo.HasDecompressed {
		gameDir := filepath.Join(sourcePath, "game")
		game7zPath := filepath.Join(sourcePath, "game.7z")

		if opts.Verbose {
			fmt.Printf("Converting decompressed game/ folder to compressed game.7z...\n")
			fmt.Printf("Compressing contents of: %s\n", gameDir)
		}

		// Create the 7z archive from the game folder contents
		// We want to compress the contents of the game/ folder, not the folder itself
		if err := common.Create7zArchive(gameDir, game7zPath); err != nil {
			return fmt.Errorf("creating game.7z archive: %w", err)
		}

		// Remove the game/ folder if compression was successful
		if opts.Verbose {
			fmt.Printf("Removing original game/ folder...\n")
		}
		if err := os.RemoveAll(gameDir); err != nil {
			// Don't fail if we can't remove the folder, just warn
			fmt.Printf("Warning: could not remove original game/ folder: %v\n", err)
		}

		fmt.Printf("Successfully converted to compressed format:\n")
		fmt.Printf("  Title: %s\n", organizedInfo.GameInfo.Title)
		fmt.Printf("  Title ID: %s\n", organizedInfo.GameInfo.TitleID)
		fmt.Printf("  Format: Compressed (game.7z)\n")
		fmt.Printf("  Location: %s\n", sourcePath)
		return nil
	}

	return fmt.Errorf("organized directory has unexpected format")
}

// UnpackageGame unpacks a PS3 game into decompressed format (game/ folder)
func UnpackageGame(sourcePath string, opts PackageOptions) error {
	if opts.Verbose {
		fmt.Printf("Unpackaging PS3 game from: %s\n", sourcePath)
		fmt.Printf("Output directory: %s\n", opts.OutputDir)
	}

	// First check if this is an organized directory
	organizedInfo, err := common.DetectOrganizedDirectory(sourcePath, opts.Verbose)
	if err != nil {
		return fmt.Errorf("checking if directory is organized: %w", err)
	}

	if organizedInfo.IsOrganized {
		return handleOrganizedDirectoryForUnpackaging(sourcePath, organizedInfo, opts)
	}

	// Extract game information for unorganized directories
	gameInfo, err := common.ExtractGameInfo(sourcePath, opts.Verbose)
	if err != nil {
		return err
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

	// Create game folder with raw files
	gameDir := filepath.Join(targetPath, "game")

	if opts.Verbose {
		fmt.Printf("Copying game files to game/ folder...\n")
	}

	if err := common.CopyDir(gameInfo.Source, gameDir); err != nil {
		return fmt.Errorf("copying game files: %w", err)
	}

	fmt.Printf("Successfully unpackaged PS3 game:\n")
	fmt.Printf("  Title: %s\n", gameInfo.Title)
	fmt.Printf("  Title ID: %s\n", gameInfo.TitleID)
	fmt.Printf("  Format: Decompressed (game/ folder)\n")
	fmt.Printf("  Output: %s\n", targetPath)

	return nil
}

// handleOrganizedDirectoryForUnpackaging handles unpackaging of organized directories
func handleOrganizedDirectoryForUnpackaging(sourcePath string, organizedInfo *common.OrganizedDirInfo, opts PackageOptions) error {
	// If it already has game/ folder (decompressed), check if we need to do anything
	if organizedInfo.HasDecompressed && !organizedInfo.HasCompressed {
		if opts.Verbose {
			fmt.Printf("Source is already in decompressed format (game/ folder)\n")
			fmt.Printf("No action needed\n")
		}
		fmt.Printf("Directory is already unpackaged:\n")
		fmt.Printf("  Title: %s\n", organizedInfo.GameInfo.Title)
		fmt.Printf("  Title ID: %s\n", organizedInfo.GameInfo.TitleID)
		fmt.Printf("  Format: Decompressed (game/ folder)\n")
		fmt.Printf("  Location: %s\n", sourcePath)
		return nil
	}

	// If it has game.7z (compressed), extract it to game/ folder
	if organizedInfo.HasCompressed {
		game7zPath := filepath.Join(sourcePath, "game.7z")
		gameDir := filepath.Join(sourcePath, "game")

		if opts.Verbose {
			fmt.Printf("Extracting compressed game.7z to game/ folder...\n")
			fmt.Printf("Extracting to: %s\n", gameDir)
		}

		// Ensure the game directory doesn't exist to prevent conflicts
		if _, err := os.Stat(gameDir); err == nil {
			if opts.Verbose {
				fmt.Printf("Removing existing game/ folder before extraction...\n")
			}
			if err := os.RemoveAll(gameDir); err != nil {
				return fmt.Errorf("removing existing game/ folder: %w", err)
			}
		}

		// Create the game directory
		if err := os.MkdirAll(gameDir, 0755); err != nil {
			return fmt.Errorf("creating game/ directory: %w", err)
		}

		// Extract the 7z archive to the game folder
		if err := common.Extract7zArchive(game7zPath, gameDir); err != nil {
			return fmt.Errorf("extracting game.7z archive: %w", err)
		}

		// Remove the game.7z file if extraction was successful
		if opts.Verbose {
			fmt.Printf("Removing original game.7z file...\n")
		}
		if err := os.Remove(game7zPath); err != nil {
			// Don't fail if we can't remove the file, just warn
			fmt.Printf("Warning: could not remove original game.7z file: %v\n", err)
		}

		fmt.Printf("Successfully converted to decompressed format:\n")
		fmt.Printf("  Title: %s\n", organizedInfo.GameInfo.Title)
		fmt.Printf("  Title ID: %s\n", organizedInfo.GameInfo.TitleID)
		fmt.Printf("  Format: Decompressed (game/ folder)\n")
		fmt.Printf("  Location: %s\n", sourcePath)
		return nil
	}

	return fmt.Errorf("organized directory has unexpected format")
}

// PackageGames packages multiple PS3 games into compressed format (game.7z)
func PackageGames(sourcePaths []string, opts PackageOptions) error {
	var errors []error
	successCount := 0
	totalCount := len(sourcePaths)

	for i, sourcePath := range sourcePaths {
		if opts.Verbose {
			fmt.Printf("\n=== Processing %d/%d: %s ===\n", i+1, totalCount, sourcePath)
		}

		if err := PackageGame(sourcePath, opts); err != nil {
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

// UnpackageGames unpacks multiple PS3 games into decompressed format (game/ folder)
func UnpackageGames(sourcePaths []string, opts PackageOptions) error {
	var errors []error
	successCount := 0
	totalCount := len(sourcePaths)

	for i, sourcePath := range sourcePaths {
		if opts.Verbose {
			fmt.Printf("\n=== Processing %d/%d: %s ===\n", i+1, totalCount, sourcePath)
		}

		if err := UnpackageGame(sourcePath, opts); err != nil {
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
