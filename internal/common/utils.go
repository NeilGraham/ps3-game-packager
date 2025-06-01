package common

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/NeilGraham/ps3-game-packager/internal/parsers"
)

// GameInfo holds information about a PS3 game
type GameInfo struct {
	Title   string
	TitleID string
	Source  string
}

// OrganizedDirInfo holds information about an organized game directory
type OrganizedDirInfo struct {
	IsOrganized     bool
	HasCompressed   bool // has game.7z
	HasDecompressed bool // has game/ folder
	GameInfo        *GameInfo
}

// SanitizeFilename removes or replaces characters that are not safe for filenames
func SanitizeFilename(filename string) string {
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

// ExtractGameInfo extracts game information from a source path
func ExtractGameInfo(sourcePath string, verbose bool) (*GameInfo, error) {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("source path does not exist: %w", err)
	}

	var paramSFOPath string
	var gameRootPath string

	if sourceInfo.IsDir() {
		// Source is a directory - search for PS3_GAME recursively
		foundGameRoot, foundParamSFO, err := findPS3GameRecursively(sourcePath, verbose)
		if err != nil {
			return nil, err
		}
		gameRootPath = foundGameRoot
		paramSFOPath = foundParamSFO
	} else {
		// Source is likely an archive file
		ext := strings.ToLower(filepath.Ext(sourcePath))
		if ext != ".zip" {
			return nil, fmt.Errorf("archive format %s not supported yet, please use .zip or extract to a folder", ext)
		}

		// Extract archive to temporary directory
		tempDir, err := os.MkdirTemp("", "ps3-extract-*")
		if err != nil {
			return nil, fmt.Errorf("creating temporary directory: %w", err)
		}

		if verbose {
			fmt.Printf("Extracting archive to temporary directory: %s\n", tempDir)
		}

		if err := ExtractZip(sourcePath, tempDir); err != nil {
			os.RemoveAll(tempDir)
			return nil, fmt.Errorf("extracting archive: %w", err)
		}

		// Search for PS3_GAME recursively in extracted archive
		foundGameRoot, foundParamSFO, err := findPS3GameRecursively(tempDir, verbose)
		if err != nil {
			os.RemoveAll(tempDir)
			return nil, err
		}
		gameRootPath = foundGameRoot
		paramSFOPath = foundParamSFO
	}

	// Parse PARAM.SFO to get game information
	if verbose {
		fmt.Printf("Reading game information from: %s\n", paramSFOPath)
	}

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

	return &GameInfo{
		Title:   title,
		TitleID: titleID,
		Source:  gameRootPath,
	}, nil
}

// findPS3GameRecursively searches for PS3_GAME/PARAM.SFO recursively in a directory
func findPS3GameRecursively(rootPath string, verbose bool) (gameRoot, paramSFOPath string, err error) {
	// First check if PS3_GAME exists at the root level (common case)
	paramSFOPath = filepath.Join(rootPath, "PS3_GAME", "PARAM.SFO")
	if _, err := os.Stat(paramSFOPath); err == nil {
		if verbose {
			fmt.Printf("Found PS3_GAME at root level: %s\n", rootPath)
		}
		return rootPath, paramSFOPath, nil
	}

	if verbose {
		fmt.Printf("PS3_GAME not found at root level, searching recursively in: %s\n", rootPath)
	}

	// Recursively search for PS3_GAME/PARAM.SFO
	var foundPath string
	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if this is a PARAM.SFO file
		if info.Name() == "PARAM.SFO" {
			// Check if it's in a PS3_GAME directory
			parentDir := filepath.Dir(path)
			if filepath.Base(parentDir) == "PS3_GAME" {
				// Found PS3_GAME/PARAM.SFO - the game root is the parent of PS3_GAME
				foundPath = filepath.Dir(parentDir)
				if verbose {
					fmt.Printf("Found PS3_GAME in nested directory: %s\n", foundPath)
				}
				return filepath.SkipDir // Stop searching
			}
		}

		return nil
	})

	if err != nil {
		return "", "", fmt.Errorf("error searching for PS3_GAME: %w", err)
	}

	if foundPath == "" {
		return "", "", fmt.Errorf("PS3_GAME/PARAM.SFO not found in %s: ensure the source contains a PS3_GAME folder with PARAM.SFO", rootPath)
	}

	paramSFOPath = filepath.Join(foundPath, "PS3_GAME", "PARAM.SFO")
	return foundPath, paramSFOPath, nil
}

// GenerateTargetPath creates the target directory path for a game
func GenerateTargetPath(gameInfo *GameInfo, outputDir string) string {
	sanitizedTitle := SanitizeFilename(gameInfo.Title)
	targetDirName := fmt.Sprintf("%s [%s]", sanitizedTitle, gameInfo.TitleID)
	return filepath.Join(outputDir, targetDirName)
}

// CreateTargetStructure creates the base directory structure for a packed game
func CreateTargetStructure(targetPath string, force bool) error {
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

	return nil
}

// ExtractZip extracts a ZIP archive to the specified destination
func ExtractZip(src, dest string) error {
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

// Create7zArchive creates a 7z archive from the source directory
func Create7zArchive(sourceDir, archivePath string) error {
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

	// Convert to absolute paths to avoid issues with directory changes
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return fmt.Errorf("getting absolute path for source directory: %w", err)
	}

	absArchivePath, err := filepath.Abs(archivePath)
	if err != nil {
		return fmt.Errorf("getting absolute path for archive: %w", err)
	}

	// Build command arguments for maximum compression
	// We use "." to archive everything in the current directory (after cd)
	args := []string{
		"a",            // add to archive
		"-t7z",         // archive type 7z
		"-mx=9",        // maximum compression level
		"-mfb=64",      // number of fast bytes for LZMA
		"-md=32m",      // dictionary size
		"-ms=on",       // solid archive for better compression
		absArchivePath, // output archive path (absolute)
		".",            // source files (current directory contents)
	}

	execCmd := exec.Command(cmd, args...)

	// Change working directory to source directory
	// This ensures only the contents are archived, not the directory name
	execCmd.Dir = absSourceDir

	// Capture output for debugging
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf(`7z command failed: %w

Command: %s %s
Working Directory: %s
Stdout: %s
Stderr: %s

This usually indicates:
1. The source directory is empty or doesn't exist
2. Permission issues with the source or destination
3. Insufficient disk space for the archive`,
			err, cmd, strings.Join(args, " "), absSourceDir, stdout.String(), stderr.String())
	}

	return nil
}

// CopyDir copies the contents of one directory to another
func CopyDir(src, dest string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("reading source directory %s: %w", src, err)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("creating destination directory %s: %w", dest, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", destPath, err)
			}
			if err := CopyDir(srcPath, destPath); err != nil {
				return fmt.Errorf("copying directory from %s to %s: %w", srcPath, destPath, err)
			}
		} else {
			if err := CopyFile(srcPath, destPath); err != nil {
				return fmt.Errorf("copying file from %s to %s: %w", srcPath, destPath, err)
			}
		}
	}

	return nil
}

// CopyFile copies a single file from source to destination
func CopyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source file %s: %w", src, err)
	}
	defer srcFile.Close()

	// Ensure destination directory exists
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("creating destination directory %s: %w", destDir, err)
	}

	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("creating destination file %s: %w", dest, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("copying data from %s to %s: %w", src, dest, err)
	}

	return nil
}

// DetectOrganizedDirectory checks if a directory is already organized and determines its format
func DetectOrganizedDirectory(sourcePath string, verbose bool) (*OrganizedDirInfo, error) {
	sourceName := filepath.Base(sourcePath)

	// Check if this looks like an organized game directory
	// Format: "{Game Name} [{Game ID}]/"
	if !strings.Contains(sourceName, "[") || !strings.Contains(sourceName, "]") {
		return &OrganizedDirInfo{IsOrganized: false}, nil
	}

	// Check if it has the expected subdirectories
	gameFile := filepath.Join(sourcePath, "game.7z")
	gameDir := filepath.Join(sourcePath, "game")
	updatesDir := filepath.Join(sourcePath, "_updates")
	dlcDir := filepath.Join(sourcePath, "_dlc")

	hasCompressed := false
	if _, err := os.Stat(gameFile); err == nil {
		hasCompressed = true
	}

	hasDecompressed := false
	if _, err := os.Stat(gameDir); err == nil {
		hasDecompressed = true
	}

	hasUpdates := false
	if _, err := os.Stat(updatesDir); err == nil {
		hasUpdates = true
	}

	hasDLC := false
	if _, err := os.Stat(dlcDir); err == nil {
		hasDLC = true
	}

	// Must have either game.7z or game/ directory, plus _updates and _dlc to be considered organized
	isOrganized := (hasCompressed || hasDecompressed) && hasUpdates && hasDLC

	if !isOrganized {
		return &OrganizedDirInfo{IsOrganized: false}, nil
	}

	if verbose {
		fmt.Printf("Detected organized game directory: %s\n", sourcePath)
		if hasCompressed {
			fmt.Printf("  Format: Compressed (game.7z)\n")
		}
		if hasDecompressed {
			fmt.Printf("  Format: Decompressed (game/ folder)\n")
		}
	}

	// Try to extract game info from the directory name
	// Format: "{Game Name} [{Game ID}]"
	titleID := ""
	title := ""

	if start := strings.LastIndex(sourceName, "["); start != -1 {
		if end := strings.LastIndex(sourceName, "]"); end != -1 && end > start {
			titleID = sourceName[start+1 : end]
			title = strings.TrimSpace(sourceName[:start])
		}
	}

	// If we couldn't parse from directory name, try to read from PARAM.SFO
	if title == "" || titleID == "" {
		var paramSFOPath string
		if hasDecompressed {
			paramSFOPath = filepath.Join(gameDir, "PS3_GAME", "PARAM.SFO")
		} else if hasCompressed {
			// For compressed format, we can't easily read PARAM.SFO without extracting
			// Use parsed values from directory name
		}

		if paramSFOPath != "" {
			if _, err := os.Stat(paramSFOPath); err == nil {
				if gameInfo, err := extractGameInfoFromParamSFO(paramSFOPath); err == nil {
					title = gameInfo.Title
					titleID = gameInfo.TitleID
				}
			}
		}
	}

	gameInfo := &GameInfo{
		Title:   title,
		TitleID: titleID,
		Source:  sourcePath,
	}

	return &OrganizedDirInfo{
		IsOrganized:     true,
		HasCompressed:   hasCompressed,
		HasDecompressed: hasDecompressed,
		GameInfo:        gameInfo,
	}, nil
}

// extractGameInfoFromParamSFO extracts game info from a PARAM.SFO file
func extractGameInfoFromParamSFO(paramSFOPath string) (*GameInfo, error) {
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

	return &GameInfo{
		Title:   title,
		TitleID: titleID,
		Source:  filepath.Dir(filepath.Dir(paramSFOPath)), // parent of PS3_GAME
	}, nil
}

// Extract7zArchive extracts a 7z archive to the specified destination
func Extract7zArchive(archivePath, destDir string) error {
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

	// Build command arguments for extraction
	args := []string{
		"x",            // extract with full paths
		archivePath,    // source archive
		"-o" + destDir, // output directory (note: no space between -o and path)
		"-y",           // assume yes for all prompts
	}

	execCmd := exec.Command(cmd, args...)

	// Capture output for debugging
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf(`7z extraction failed: %w

Command: %s %s
Stdout: %s
Stderr: %s

This usually indicates:
1. The archive file is corrupted or doesn't exist
2. Permission issues with the source or destination
3. Insufficient disk space for extraction`,
			err, cmd, strings.Join(args, " "), stdout.String(), stderr.String())
	}

	return nil
}
