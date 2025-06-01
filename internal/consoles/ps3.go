package consoles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NeilGraham/rom-organizer/internal/common"
	"github.com/NeilGraham/rom-organizer/internal/parsers"
)

// PS3Handler handles PlayStation 3 games
type PS3Handler struct{}

// NewPS3Handler creates a new PS3 handler
func NewPS3Handler() *PS3Handler {
	return &PS3Handler{}
}

// GetConsoleDisplayName returns the human-readable console name
func (h *PS3Handler) GetConsoleDisplayName() string {
	return "PlayStation 3"
}

// GetGameDirectoryPattern returns the expected directory pattern for PS3 games
func (h *PS3Handler) GetGameDirectoryPattern() string {
	return "PS3_GAME"
}

// ValidateGameStructure checks if the source path contains a valid PS3 game structure
func (h *PS3Handler) ValidateGameStructure(sourcePath string) error {
	_, _, err := h.findPS3GameRecursively(sourcePath, false)
	return err
}

// ExtractGameInfo extracts game information from a PS3 source path
func (h *PS3Handler) ExtractGameInfo(sourcePath string, verbose bool) (*common.GameInfo, error) {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("source path does not exist: %w", err)
	}

	var paramSFOPath string
	var gameRootPath string

	if sourceInfo.IsDir() {
		// Source is a directory - search for PS3_GAME recursively
		foundGameRoot, foundParamSFO, err := h.findPS3GameRecursively(sourcePath, verbose)
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
		tempDir, err := os.MkdirTemp("", "game-extract-*")
		if err != nil {
			return nil, fmt.Errorf("creating temporary directory: %w", err)
		}

		if verbose {
			fmt.Printf("Extracting archive to temporary directory: %s\n", tempDir)
		}

		if err := common.ExtractZip(sourcePath, tempDir); err != nil {
			os.RemoveAll(tempDir)
			return nil, fmt.Errorf("extracting archive: %w", err)
		}

		// Search for PS3_GAME recursively in extracted archive
		foundGameRoot, foundParamSFO, err := h.findPS3GameRecursively(tempDir, verbose)
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

	// Extract additional metadata
	version := paramSFO.GetString("APP_VER")
	category := paramSFO.GetString("CATEGORY")

	return &common.GameInfo{
		Title:    title,
		GameID:   titleID,
		Console:  h.GetConsoleDisplayName(),
		Version:  version,
		Category: category,
		Source:   gameRootPath,
	}, nil
}

// findPS3GameRecursively searches for PS3_GAME/PARAM.SFO recursively in a directory
func (h *PS3Handler) findPS3GameRecursively(rootPath string, verbose bool) (gameRoot, paramSFOPath string, err error) {
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
