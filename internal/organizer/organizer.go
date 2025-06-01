package organizer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NeilGraham/ps3-game-packager/internal/common"
)

// OrganizeOptions holds options for organizing operations
type OrganizeOptions struct {
	OutputDir string
	Force     bool
	Verbose   bool
}

// OrganizeGame organizes a PS3 game while keeping it in its existing format
func OrganizeGame(sourcePath string, opts OrganizeOptions) error {
	if opts.Verbose {
		fmt.Printf("Organizing PS3 game from: %s\n", sourcePath)
		fmt.Printf("Output directory: %s\n", opts.OutputDir)
	}

	// Check if this is an existing organized game directory
	organizedInfo, err := common.DetectOrganizedDirectory(sourcePath, opts.Verbose)
	if err != nil {
		return fmt.Errorf("checking if directory is organized: %w", err)
	}

	if organizedInfo.IsOrganized {
		if opts.Verbose {
			fmt.Printf("Source is already an organized game directory\n")
			fmt.Printf("No action needed - directory is already in the correct format\n")
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

	// Copy game in its original format
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("checking source: %w", err)
	}

	var format string
	if sourceInfo.IsDir() {
		// Copy as decompressed game folder
		gameDir := filepath.Join(targetPath, "game")
		if opts.Verbose {
			fmt.Printf("Copying game files to game/ folder (keeping decompressed format)...\n")
		}
		if err := common.CopyDir(gameInfo.Source, gameDir); err != nil {
			return fmt.Errorf("copying game files: %w", err)
		}
		format = "Decompressed (game/ folder)"
	} else {
		// This shouldn't happen since ExtractGameInfo handles archives
		// But if it does, we'll treat it as a compressed archive
		return fmt.Errorf("archive files should be handled by package/unpackage commands")
	}

	fmt.Printf("Successfully organized PS3 game:\n")
	fmt.Printf("  Title: %s\n", gameInfo.Title)
	fmt.Printf("  Title ID: %s\n", gameInfo.TitleID)
	fmt.Printf("  Format: %s\n", format)
	fmt.Printf("  Output: %s\n", targetPath)

	return nil
}
