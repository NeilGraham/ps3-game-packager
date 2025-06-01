// Package main provides a development script for generating fake PS3 games for testing.
// This script creates realistic but completely fictional game data to avoid any copyright issues.
//
// Usage: go run scripts/dev/generate-test-games.go [options]
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Fake game data - completely fictional titles and IDs
var fakeGames = []struct {
	title    string
	titleID  string
	category string
	appVer   string
}{
	{"Galactic Warriors: Return of the Void", "BLUS12345", "DG", "01.00"},
	{"Crystal Quest: Legends of Mystara", "BLES67890", "DG", "01.20"},
	{"Neon Racers: Future Streets", "BCUS11111", "DG", "02.01"},
	{"Adventure Island: Lost Treasures", "BCES22222", "DG", "01.10"},
	{"Space Marines: Infinite War", "BLUS33333", "DG", "03.00"},
	{"Magic Kingdom: Dragon's Crown", "BLES44444", "DG", "01.05"},
	{"Cyber Punk: Digital Revolution", "BCUS55555", "DG", "01.30"},
	{"Fantasy Quest: Ancient Realms", "BCES66666", "DG", "02.15"},
	{"Metal Storm: Apocalypse Rising", "BLUS77777", "DG", "01.00"},
	{"Ocean Adventure: Deep Waters", "BLES88888", "DG", "01.25"},
	{"Desert Combat: Sand Warriors", "BCUS99999", "DG", "01.40"},
	{"Forest Guardian: Nature's Call", "BCES10101", "DG", "01.00"},
	{"City Builder: Metropolis Dreams", "BLUS20202", "DG", "02.30"},
	{"Puzzle Master: Mind Bender", "BLES30303", "DG", "01.15"},
	{"Racing Thunder: Speed Demons", "BCUS40404", "DG", "01.50"},
}

// PARAM.SFO structure constants
const (
	paramSFOMagic    = "\x00PSF"
	paramSFOVersion  = 0x00000101
	FMT_UTF8_SPECIAL = 0x0004
	FMT_UTF8         = 0x0204
	FMT_INT32        = 0x0404
)

// Entry represents a PARAM.SFO entry
type Entry struct {
	Key     string
	Value   interface{}
	DataFmt uint16
}

// generateParamSFO creates a fake but valid PARAM.SFO file
func generateParamSFO(game struct {
	title    string
	titleID  string
	category string
	appVer   string
}) ([]byte, error) {
	// Define all the entries we want to include
	entries := []Entry{
		{"APP_VER", game.appVer, FMT_UTF8},
		{"ATTRIBUTE", uint32(0), FMT_INT32},
		{"BOOTABLE", uint32(1), FMT_INT32},
		{"CATEGORY", game.category, FMT_UTF8},
		{"LICENSE", "This is a fake test game for development purposes only.", FMT_UTF8},
		{"NP_COMMUNICATION_ID", fmt.Sprintf("NPWR%05d_00", rand.Intn(99999)), FMT_UTF8},
		{"PARENTAL_LEVEL", uint32(1), FMT_INT32},
		{"PS3_SYSTEM_VER", "03.5500", FMT_UTF8},
		{"RESOLUTION", uint32(63), FMT_INT32},
		{"SOUND_FORMAT", uint32(279), FMT_INT32},
		{"TITLE", game.title, FMT_UTF8},
		{"TITLE_ID", game.titleID, FMT_UTF8},
		{"VERSION", game.appVer, FMT_UTF8},
	}

	// Calculate offsets
	headerSize := 20
	entryTableSize := len(entries) * 16
	keyTableOffset := uint32(headerSize + entryTableSize)

	// Build key table and calculate data table offset
	keyTable := bytes.Buffer{}
	for _, entry := range entries {
		keyTable.WriteString(entry.Key)
		keyTable.WriteByte(0) // null terminator
	}
	// Align to 4 bytes
	for keyTable.Len()%4 != 0 {
		keyTable.WriteByte(0)
	}

	dataTableOffset := keyTableOffset + uint32(keyTable.Len())

	// Build data table
	dataTable := bytes.Buffer{}
	var rawEntries []struct {
		KeyOffset uint16
		DataFmt   uint16
		DataLen   uint32
		DataMax   uint32
		DataOff   uint32
	}

	keyOffset := uint16(0)
	for _, entry := range entries {
		dataOffset := uint32(dataTable.Len())
		var dataLen uint32

		switch v := entry.Value.(type) {
		case string:
			data := []byte(v)
			dataTable.Write(data)
			dataTable.WriteByte(0) // null terminator
			dataLen = uint32(len(data) + 1)
			// Align to 4 bytes
			for dataTable.Len()%4 != 0 {
				dataTable.WriteByte(0)
			}
		case uint32:
			binary.Write(&dataTable, binary.LittleEndian, v)
			dataLen = 4
		}

		rawEntries = append(rawEntries, struct {
			KeyOffset uint16
			DataFmt   uint16
			DataLen   uint32
			DataMax   uint32
			DataOff   uint32
		}{
			KeyOffset: keyOffset,
			DataFmt:   entry.DataFmt,
			DataLen:   dataLen,
			DataMax:   dataLen + 16, // Some padding
			DataOff:   dataOffset,
		})

		keyOffset += uint16(len(entry.Key) + 1)
	}

	// Build the final file
	result := bytes.Buffer{}

	// Write header
	result.WriteString(paramSFOMagic)
	binary.Write(&result, binary.LittleEndian, uint32(paramSFOVersion))
	binary.Write(&result, binary.LittleEndian, keyTableOffset)
	binary.Write(&result, binary.LittleEndian, dataTableOffset)
	binary.Write(&result, binary.LittleEndian, uint32(len(entries)))

	// Write entry table
	for _, entry := range rawEntries {
		binary.Write(&result, binary.LittleEndian, entry)
	}

	// Write key table
	result.Write(keyTable.Bytes())

	// Write data table
	result.Write(dataTable.Bytes())

	return result.Bytes(), nil
}

// createTestGame creates a fake PS3 game directory structure
func createTestGame(outputDir string, game struct {
	title    string
	titleID  string
	category string
	appVer   string
}) error {
	// Sanitize title for directory name
	safeName := game.title
	unsafeChars := []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*"}
	for _, char := range unsafeChars {
		safeName = strings.ReplaceAll(safeName, char, "_")
	}

	// Create game directory
	gameDir := filepath.Join(outputDir, fmt.Sprintf("%s [%s]", safeName, game.titleID))
	ps3GameDir := filepath.Join(gameDir, "PS3_GAME")

	if err := os.MkdirAll(ps3GameDir, 0755); err != nil {
		return fmt.Errorf("creating PS3_GAME directory: %w", err)
	}

	// Generate PARAM.SFO
	paramSFOData, err := generateParamSFO(game)
	if err != nil {
		return fmt.Errorf("generating PARAM.SFO: %w", err)
	}

	// Write PARAM.SFO file
	paramSFOPath := filepath.Join(ps3GameDir, "PARAM.SFO")
	if err := os.WriteFile(paramSFOPath, paramSFOData, 0644); err != nil {
		return fmt.Errorf("writing PARAM.SFO: %w", err)
	}

	// Create a fake SFB file for realism
	sfbPath := filepath.Join(gameDir, "PS3_DISC.SFB")
	fakeDiscData := []byte("FAKE_PS3_DISC_DATA_FOR_TESTING_ONLY")
	if err := os.WriteFile(sfbPath, fakeDiscData, 0644); err != nil {
		return fmt.Errorf("writing PS3_DISC.SFB: %w", err)
	}

	fmt.Printf("Created test game: %s [%s]\n", game.title, game.titleID)
	return nil
}

func main() {
	var (
		outputDir = flag.String("output", "test-games", "Output directory for test games")
		count     = flag.Int("count", 5, "Number of test games to generate")
		seed      = flag.Int64("seed", 0, "Random seed (0 for current time)")
		clean     = flag.Bool("clean", false, "Clean output directory before generating")
	)
	flag.Parse()

	// Set random seed
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}
	rand.Seed(*seed)
	fmt.Printf("Using random seed: %d\n", *seed)

	// Clean output directory if requested
	if *clean {
		if err := os.RemoveAll(*outputDir); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error cleaning output directory: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Cleaned output directory: %s\n", *outputDir)
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Generate test games
	fmt.Printf("Generating %d test games in %s\n", *count, *outputDir)
	fmt.Println("==================================================")

	// Track used title IDs to avoid duplicates
	usedTitleIDs := make(map[string]int)

	for i := 0; i < *count; i++ {
		// Pick a random game from our fake list
		gameIndex := rand.Intn(len(fakeGames))
		game := fakeGames[gameIndex]

		// Handle duplicate title IDs by appending a counter
		originalTitleID := game.titleID
		if count, exists := usedTitleIDs[originalTitleID]; exists {
			// Generate a unique variant: BLUS12345 -> BLUS12346, BLUS12347, etc.
			baseID := originalTitleID[:len(originalTitleID)-1] // Remove last digit

			// Try incrementing the last digit, then use suffix if needed
			if lastDigitInt := int(originalTitleID[len(originalTitleID)-1] - '0'); lastDigitInt < 9 {
				game.titleID = fmt.Sprintf("%s%d", baseID, lastDigitInt+count)
			} else {
				// If last digit is 9, use suffix approach
				game.titleID = fmt.Sprintf("%s_%02d", originalTitleID, count)
			}
			usedTitleIDs[originalTitleID] = count + 1
		} else {
			usedTitleIDs[originalTitleID] = 1
		}

		if err := createTestGame(*outputDir, game); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating test game %d: %v\n", i+1, err)
			continue
		}
	}

	fmt.Println("==================================================")
	fmt.Printf("Test game generation complete!\n")
	fmt.Printf("Generated games are in: %s\n", *outputDir)
	fmt.Printf("\nYou can now test the rom-organizer with:\n")
	fmt.Printf("  ./rom-organizer metadata %s/*/\n", *outputDir)
	fmt.Printf("  ./rom-organizer organize %s/*/ --output organized-test-games\n", *outputDir)
	fmt.Printf("  ./rom-organizer compress %s/*/ --output compressed-test-games\n", *outputDir)
}
