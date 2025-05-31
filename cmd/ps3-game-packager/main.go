package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NeilGraham/ps3-game-packager/internal/parsers"
)

var (
	verbose      bool
	jsonOutput   bool
	outputDir    string
	force        bool
	decompressed bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "ps3-game-packager",
	Short: "Tools for working with PS3 game files",
	Long: `PS3 Game Packager - A collection of tools for working with PS3 game files.

This toolkit provides utilities for organizing and optimizing PS3 game files.`,
	Version: "1.0.0",
}

var parseParamSFOCmd = &cobra.Command{
	Use:   "parse-param-sfo <PARAM.SFO file>",
	Short: "Parse a PS3 PARAM.SFO file and extract game information",
	Long: `Parse a PS3 PARAM.SFO file and extract game information.

PARAM.SFO files contain metadata about PS3 games including the title, 
title ID, version information, and other game attributes.

Examples:
  ps3-game-packager parse-param-sfo PARAM.SFO
  ps3-game-packager parse-param-sfo --verbose PARAM.SFO
  ps3-game-packager parse-param-sfo PARAM.SFO --json
  ps3-game-packager parse-param-sfo --json --verbose PARAM.SFO`,
	Args: cobra.ExactArgs(1),
	RunE: parseParamSFOHandler,
}

var packCmd = &cobra.Command{
	Use:   "pack <source>",
	Short: "Package a PS3 game folder or archive into organized format",
	Long: `Pack a PS3 game folder or archive into the standardized format.

This command takes a decrypted PS3 ISO game folder or archive file and packages it
into a standardized directory structure:

Default (compressed):
{Game Name} [{Game ID}]/
├── game.7z          (compressed game files)
├── _updates/        (updates folder - empty for now)
└── _dlc/           (DLC folder - empty for now)

With --decompressed flag:
{Game Name} [{Game ID}]/
├── game/            (raw game files, uncompressed)
├── _updates/        (updates folder - empty for now)
└── _dlc/           (DLC folder - empty for now)

The game information (title and ID) is extracted from PS3_GAME/PARAM.SFO.

Examples:
  ps3-game-packager pack /path/to/game_folder
  ps3-game-packager pack --output /target/dir /path/to/game.zip
  ps3-game-packager pack --decompressed /path/to/game_folder
  ps3-game-packager pack --force /path/to/game_folder`,
	Args: cobra.ExactArgs(1),
	RunE: packHandler,
}

func init() {
	// Add subcommands to root
	rootCmd.AddCommand(parseParamSFOCmd)
	rootCmd.AddCommand(packCmd)

	// Add flags to parse-param-sfo command
	parseParamSFOCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	parseParamSFOCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")

	// Add flags to pack command
	packCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for packed game")
	packCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing output directory")
	packCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	packCmd.Flags().BoolVarP(&decompressed, "decompressed", "d", false, "Pack the game in decompressed format")
}

func parseParamSFOHandler(cmd *cobra.Command, args []string) error {
	filename := args[0]

	// Read and parse the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	paramSFO, err := parsers.ParseParamSFO(data)
	if err != nil {
		return fmt.Errorf("parsing PARAM.SFO: %w", err)
	}

	// Output based on format preference
	if jsonOutput {
		outputJSON(paramSFO)
	} else {
		outputText(paramSFO, verbose)
	}

	return nil
}

func outputText(paramSFO *parsers.ParamSFO, verbose bool) {
	if verbose {
		fmt.Printf("PARAM.SFO Parser\n")
		fmt.Printf("================\n")
		fmt.Printf("Version:         %d.%d\n",
			paramSFO.Header.Version&0xFF,
			(paramSFO.Header.Version>>8)&0xFF)
		fmt.Printf("Key Table:       offset %d\n", paramSFO.Header.KeyTableOffset)
		fmt.Printf("Data Table:      offset %d\n", paramSFO.Header.DataTableOffset)
		fmt.Printf("Entry Count:     %d\n\n", paramSFO.Header.EntryCount)

		fmt.Println("Entries:")
		fmt.Println("--------")

		for _, entry := range paramSFO.Entries {
			var valueStr string
			switch v := entry.Value.(type) {
			case string:
				valueStr = v
			case uint32:
				valueStr = fmt.Sprintf("%d", v)
			case []byte:
				valueStr = fmt.Sprintf("[unsupported format 0x%04X]", entry.DataFmt)
			default:
				valueStr = fmt.Sprintf("%v", v)
			}
			fmt.Printf("%-20s %s\n", entry.Key+":", valueStr)
		}
		fmt.Println()
	}

	// Always show summary
	fmt.Println("Summary:")
	fmt.Println("========")

	title := paramSFO.GetTitle()
	titleID := paramSFO.GetTitleID()

	if title != "" {
		fmt.Printf("Game Title:  %s\n", title)
	} else {
		fmt.Println("Game Title:  [not found]")
	}

	if titleID != "" {
		fmt.Printf("Title ID:    %s\n", titleID)
	} else {
		fmt.Println("Title ID:    [not found]")
	}

	// Show some additional useful info
	if appVer := paramSFO.GetString("APP_VER"); appVer != "" {
		fmt.Printf("App Version: %s\n", appVer)
	}
	if category := paramSFO.GetString("CATEGORY"); category != "" {
		fmt.Printf("Category:    %s\n", category)
	}
}

func outputJSON(paramSFO *parsers.ParamSFO) {
	fmt.Printf("{\n")
	fmt.Printf("  \"header\": {\n")
	fmt.Printf("    \"version\": \"%d.%d\",\n",
		paramSFO.Header.Version&0xFF,
		(paramSFO.Header.Version>>8)&0xFF)
	fmt.Printf("    \"keyTableOffset\": %d,\n", paramSFO.Header.KeyTableOffset)
	fmt.Printf("    \"dataTableOffset\": %d,\n", paramSFO.Header.DataTableOffset)
	fmt.Printf("    \"entryCount\": %d\n", paramSFO.Header.EntryCount)
	fmt.Printf("  },\n")
	fmt.Printf("  \"entries\": {\n")

	for i, entry := range paramSFO.Entries {
		fmt.Printf("    \"%s\": ", entry.Key)
		switch v := entry.Value.(type) {
		case string:
			fmt.Printf("\"%s\"", v)
		case uint32:
			fmt.Printf("%d", v)
		case []byte:
			fmt.Printf("null")
		default:
			fmt.Printf("\"%v\"", v)
		}
		if i < len(paramSFO.Entries)-1 {
			fmt.Printf(",")
		}
		fmt.Printf("\n")
	}

	fmt.Printf("  },\n")
	fmt.Printf("  \"summary\": {\n")
	fmt.Printf("    \"title\": \"%s\",\n", paramSFO.GetTitle())
	fmt.Printf("    \"titleId\": \"%s\",\n", paramSFO.GetTitleID())
	fmt.Printf("    \"appVersion\": \"%s\",\n", paramSFO.GetString("APP_VER"))
	fmt.Printf("    \"category\": \"%s\"\n", paramSFO.GetString("CATEGORY"))
	fmt.Printf("  }\n")
	fmt.Printf("}\n")
}

func packHandler(cmd *cobra.Command, args []string) error {
	source := args[0]

	if verbose {
		fmt.Printf("Packing PS3 game from: %s\n", source)
		fmt.Printf("Output directory: %s\n", outputDir)
	}

	// Check if source exists
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("source path does not exist: %w", err)
	}

	// Determine source type and locate PS3_GAME/PARAM.SFO
	var paramSFOPath string
	var gameRootPath string

	if sourceInfo.IsDir() {
		// Source is a directory
		gameRootPath = source
		paramSFOPath = filepath.Join(source, "PS3_GAME", "PARAM.SFO")
	} else {
		// Source is likely an archive file

		// For now, we'll support extracting from ZIP archives
		// Later we can add support for other archive types
		ext := strings.ToLower(filepath.Ext(source))
		if ext != ".zip" {
			return fmt.Errorf("archive format %s not supported yet, please use .zip or extract to a folder", ext)
		}

		// Extract archive to temporary directory
		tempDir, err := os.MkdirTemp("", "ps3-pack-*")
		if err != nil {
			return fmt.Errorf("creating temporary directory: %w", err)
		}
		defer os.RemoveAll(tempDir)

		if verbose {
			fmt.Printf("Extracting archive to temporary directory: %s\n", tempDir)
		}

		if err := extractZip(source, tempDir); err != nil {
			return fmt.Errorf("extracting archive: %w", err)
		}

		gameRootPath = tempDir
		paramSFOPath = filepath.Join(tempDir, "PS3_GAME", "PARAM.SFO")
	}

	// Check if PARAM.SFO exists
	if _, err := os.Stat(paramSFOPath); err != nil {
		return fmt.Errorf("PARAM.SFO not found at %s: ensure the source contains a PS3_GAME folder with PARAM.SFO", paramSFOPath)
	}

	// Parse PARAM.SFO to get game information
	if verbose {
		fmt.Printf("Reading game information from: %s\n", paramSFOPath)
	}

	paramSFOData, err := os.ReadFile(paramSFOPath)
	if err != nil {
		return fmt.Errorf("reading PARAM.SFO: %w", err)
	}

	paramSFO, err := parsers.ParseParamSFO(paramSFOData)
	if err != nil {
		return fmt.Errorf("parsing PARAM.SFO: %w", err)
	}

	title := paramSFO.GetTitle()
	titleID := paramSFO.GetTitleID()

	if title == "" {
		return fmt.Errorf("game title not found in PARAM.SFO")
	}
	if titleID == "" {
		return fmt.Errorf("title ID not found in PARAM.SFO")
	}

	// Sanitize title for filesystem
	sanitizedTitle := sanitizeFilename(title)

	// Create target directory name: {Game Name} [{Game ID}]
	targetDirName := fmt.Sprintf("%s [%s]", sanitizedTitle, titleID)
	targetPath := filepath.Join(outputDir, targetDirName)

	if verbose {
		fmt.Printf("Game Title: %s\n", title)
		fmt.Printf("Title ID: %s\n", titleID)
		fmt.Printf("Target directory: %s\n", targetPath)
	}

	// Check if target directory already exists
	if _, err := os.Stat(targetPath); err == nil && !force {
		return fmt.Errorf("target directory already exists: %s (use --force to overwrite)", targetPath)
	}

	// Create target directory structure
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("creating target directory: %w", err)
	}

	// Create _updates and _dlc subdirectories
	updatesDir := filepath.Join(targetPath, "_updates")
	dlcDir := filepath.Join(targetPath, "_dlc")

	if err := os.MkdirAll(updatesDir, 0755); err != nil {
		return fmt.Errorf("creating _updates directory: %w", err)
	}

	if err := os.MkdirAll(dlcDir, 0755); err != nil {
		return fmt.Errorf("creating _dlc directory: %w", err)
	}

	// Create game archive or folder based on decompressed flag
	if decompressed {
		// Create game folder with raw files
		gameDir := filepath.Join(targetPath, "game")

		if verbose {
			fmt.Printf("Copying game files to game/ folder...\n")
		}

		if err := copyDir(gameRootPath, gameDir); err != nil {
			return fmt.Errorf("copying game files: %w", err)
		}
	} else {
		// Create compressed game.7z archive
		game7zPath := filepath.Join(targetPath, "game.7z")

		if verbose {
			fmt.Printf("Creating game.7z archive...\n")
		}

		if err := create7zArchive(gameRootPath, game7zPath); err != nil {
			return fmt.Errorf("creating game.7z archive: %w", err)
		}
	}

	fmt.Printf("Successfully packed PS3 game:\n")
	fmt.Printf("  Title: %s\n", title)
	fmt.Printf("  Title ID: %s\n", titleID)
	if decompressed {
		fmt.Printf("  Format: Decompressed (game/ folder)\n")
	} else {
		fmt.Printf("  Format: Compressed (game.7z)\n")
	}
	fmt.Printf("  Output: %s\n", targetPath)

	return nil
}

// sanitizeFilename removes or replaces characters that are not safe for filenames
func sanitizeFilename(filename string) string {
	// Replace problematic characters with underscores
	unsafe := []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*"}
	result := filename

	for _, char := range unsafe {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Trim spaces and dots from the end
	result = strings.TrimRight(result, ". ")

	return result
}

// extractZip extracts a ZIP archive to the specified destination
func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		path := filepath.Join(dest, f.Name)

		// Check for directory traversal
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			rc.Close()
			return fmt.Errorf("invalid file path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.FileInfo().Mode())
			rc.Close()
			continue
		}

		// Create the directories for this file
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			rc.Close()
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// create7zArchive creates a 7z archive from the source directory
func create7zArchive(sourceDir, archivePath string) error {
	// Check for available 7z commands
	possibleCommands := []string{"7z", "7za", "7zr"}
	var cmd string

	for _, cmdName := range possibleCommands {
		if _, err := exec.LookPath(cmdName); err == nil {
			cmd = cmdName
			break
		}
	}

	if cmd == "" {
		return fmt.Errorf(`7z command not found in PATH. Please install 7-zip or p7zip:

Windows:
  - Download and install 7-Zip from https://www.7-zip.org/
  - Or install via chocolatey: choco install 7zip
  - Or install via winget: winget install 7zip.7zip

macOS:
  - Install via Homebrew: brew install p7zip
  - Or install via MacPorts: sudo port install p7zip

Linux:
  - Ubuntu/Debian: sudo apt-get install p7zip-full
  - CentOS/RHEL: sudo yum install p7zip
  - Arch Linux: sudo pacman -S p7zip`)
	}

	// Build command arguments for maximum compression
	args := []string{
		"a",                           // add to archive
		"-t7z",                        // archive type 7z
		"-mx=9",                       // maximum compression level
		"-mfb=64",                     // number of fast bytes for LZMA
		"-md=32m",                     // dictionary size
		"-ms=on",                      // solid archive for better compression
		archivePath,                   // output archive path
		filepath.Join(sourceDir, "*"), // source files (all files in source directory)
	}

	execCmd := exec.Command(cmd, args...)

	// Capture output for debugging
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf(`7z command failed: %w

Command: %s %s
Stdout: %s
Stderr: %s

This usually indicates:
1. The source directory is empty or doesn't exist
2. Permission issues with the source or destination
3. Insufficient disk space for the archive`,
			err, cmd, strings.Join(args, " "), stdout.String(), stderr.String())
	}

	return nil
}

// copyDir copies the contents of one directory to another
func copyDir(src, dest string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return err
			}
			if err := copyDir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from source to destination
func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}
