package organizer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NeilGraham/rom-organizer/internal/common"
)

// OrganizeOptions holds options for organizing operations
type OrganizeOptions struct {
	OutputDir  string
	Force      bool
	Verbose    bool
	MoveSource bool
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
		if opts.MoveSource {
			fmt.Printf("⚠️  WARNING: --move flag ignored for already organized directories (safety measure)\n")
		}

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

	// Copy game in its original format
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("checking source: %w", err)
	}

	var format string
	if sourceInfo.IsDir() {
		// Process as decompressed game folder
		gameDir := filepath.Join(targetPath, "game")

		if opts.MoveSource {
			if opts.Verbose {
				fmt.Printf("Moving game files to game/ folder (keeping decompressed format)...\n")
				fmt.Printf("⚠️  WARNING: Original directory will be deleted after move\n")
			}

			// Move the game files first
			if err := common.CopyDir(gameInfo.Source, gameDir); err != nil {
				return fmt.Errorf("moving game files: %w", err)
			}

			// Handle cleanup based on whether user specified the exact game directory or a parent
			if sourcePath == gameInfo.Source {
				// User specified the exact game directory - use standard cleanup
				if err := common.MoveDirWithCleanup(gameInfo.Source, gameDir, opts.Force, opts.Verbose); err != nil {
					return fmt.Errorf("moving game files: %w", err)
				}
			} else {
				// User specified a parent directory - clean up the user-specified path
				// First remove the game source directory since we copied its contents
				if err := os.RemoveAll(gameInfo.Source); err != nil {
					return fmt.Errorf("removing game source directory: %w", err)
				}

				// Then check if the user-specified directory is effectively empty
				isEmpty, err := common.IsDirEffectivelyEmpty(sourcePath)
				if err != nil {
					return fmt.Errorf("checking if source directory is empty: %w", err)
				}

				if isEmpty {
					// Safe to remove - directory contains no files
					if opts.Verbose {
						fmt.Printf("Removing empty source directory: %s\n", sourcePath)
					}
					if err := os.RemoveAll(sourcePath); err != nil {
						return fmt.Errorf("removing empty source directory: %w", err)
					}
				} else {
					// Directory contains files - check if force is enabled
					if opts.Force {
						if opts.Verbose {
							fmt.Printf("⚠️  Forcefully removing source directory with remaining files: %s\n", sourcePath)
						}
						if err := os.RemoveAll(sourcePath); err != nil {
							return fmt.Errorf("forcefully removing source directory: %w", err)
						}
					} else {
						fmt.Printf("⚠️  WARNING: Source directory contains remaining files and was not deleted: %s\n", sourcePath)
						fmt.Printf("    Use --force to delete the source directory even with remaining files\n")
					}
				}
			}

			// If both --move and --force are specified, delete the entire source directory
			if opts.Force && sourcePath != gameInfo.Source {
				if opts.Verbose {
					fmt.Printf("⚠️  Deleting entire source directory (--force --move): %s\n", sourcePath)
				}
				if err := os.RemoveAll(sourcePath); err != nil {
					fmt.Printf("⚠️  WARNING: Could not delete entire source directory %s: %v\n", sourcePath, err)
				} else if opts.Verbose {
					fmt.Printf("Successfully deleted entire source directory: %s\n", sourcePath)
				}
			}
		} else {
			if opts.Verbose {
				fmt.Printf("Copying game files to game/ folder (keeping decompressed format)...\n")
			}
			if err := common.CopyDir(gameInfo.Source, gameDir); err != nil {
				return fmt.Errorf("copying game files: %w", err)
			}
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

// OrganizeGames organizes multiple PS3 games while keeping them in their existing format
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
