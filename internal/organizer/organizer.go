package organizer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	if isOrganizedGameDir(sourcePath) {
		return organizeExistingGameDir(sourcePath, opts)
	}

	// Extract game information
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

// isOrganizedGameDir checks if the source is already an organized game directory
func isOrganizedGameDir(sourcePath string) bool {
	// Check if this looks like an organized game directory
	// Format: "{Game Name} [{Game ID}]/"
	sourceName := filepath.Base(sourcePath)

	// Look for pattern ending with [XXXX#####] where X is letter and # is digit
	if strings.Contains(sourceName, "[") && strings.Contains(sourceName, "]") {
		// Check if it has the expected subdirectories
		gameFile := filepath.Join(sourcePath, "game.7z")
		gameDir := filepath.Join(sourcePath, "game")
		updatesDir := filepath.Join(sourcePath, "_updates")
		dlcDir := filepath.Join(sourcePath, "_dlc")

		// Must have either game.7z or game/ directory, plus _updates and _dlc
		hasGame := false
		if _, err := os.Stat(gameFile); err == nil {
			hasGame = true
		}
		if _, err := os.Stat(gameDir); err == nil {
			hasGame = true
		}

		hasUpdates := false
		if _, err := os.Stat(updatesDir); err == nil {
			hasUpdates = true
		}

		hasDLC := false
		if _, err := os.Stat(dlcDir); err == nil {
			hasDLC = true
		}

		return hasGame && hasUpdates && hasDLC
	}

	return false
}

// organizeExistingGameDir reorganizes an already organized game directory
func organizeExistingGameDir(sourcePath string, opts OrganizeOptions) error {
	if opts.Verbose {
		fmt.Printf("Source appears to be an already organized game directory\n")
	}

	// Simply copy the entire organized structure
	sourceName := filepath.Base(sourcePath)
	targetPath := filepath.Join(opts.OutputDir, sourceName)

	if opts.Verbose {
		fmt.Printf("Copying organized game directory to: %s\n", targetPath)
	}

	// Check if target already exists
	if _, err := os.Stat(targetPath); err == nil && !opts.Force {
		return fmt.Errorf("target directory already exists: %s (use --force to overwrite)", targetPath)
	}

	// Copy the entire directory structure
	if err := common.CopyDir(sourcePath, targetPath); err != nil {
		return fmt.Errorf("copying organized game directory: %w", err)
	}

	// Determine format from contents
	format := "Unknown"
	if _, err := os.Stat(filepath.Join(targetPath, "game.7z")); err == nil {
		format = "Compressed (game.7z)"
	} else if _, err := os.Stat(filepath.Join(targetPath, "game")); err == nil {
		format = "Decompressed (game/ folder)"
	}

	fmt.Printf("Successfully organized PS3 game:\n")
	fmt.Printf("  Directory: %s\n", sourceName)
	fmt.Printf("  Format: %s\n", format)
	fmt.Printf("  Output: %s\n", targetPath)

	return nil
}
