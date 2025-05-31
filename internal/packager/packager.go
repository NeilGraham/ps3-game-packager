package packager

import (
	"fmt"
	"path/filepath"

	"github.com/NeilGraham/ps3-game-packager/internal/common"
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

// UnpackageGame unpacks a PS3 game into decompressed format (game/ folder)
func UnpackageGame(sourcePath string, opts PackageOptions) error {
	if opts.Verbose {
		fmt.Printf("Unpackaging PS3 game from: %s\n", sourcePath)
		fmt.Printf("Output directory: %s\n", opts.OutputDir)
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
