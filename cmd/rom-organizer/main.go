package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NeilGraham/rom-organizer/internal/detect"
	"github.com/NeilGraham/rom-organizer/internal/organizer"
	"github.com/NeilGraham/rom-organizer/internal/packager"
	"github.com/NeilGraham/rom-organizer/internal/parsers"
)

var (
	verbose    bool
	jsonOutput bool
	outputDir  string
	force      bool
	moveSource bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "rom-organizer",
	Short: "Tools for working with ROM game files",
	Long: `ROM Organizer - A collection of tools for working with ROM game files.

This toolkit provides utilities for organizing and optimizing ROM game files from various consoles.`,
	Version: "1.0.0",
}

var metadataCmd = &cobra.Command{
	Use:   "metadata <file>",
	Short: "Parse metadata from ROM files",
	Long: `Parse metadata from ROM files and extract game information.

This command can extract metadata from various ROM file formats including:
- PS3 PARAM.SFO files (contains title, title ID, version, and other game attributes)

More ROM formats will be supported in future versions.

Examples:
  rom-organizer metadata PARAM.SFO
  rom-organizer metadata --verbose PARAM.SFO
  rom-organizer metadata PARAM.SFO --json
  rom-organizer metadata --json --verbose PARAM.SFO`,
	Args: cobra.ExactArgs(1),
	RunE: metadataHandler,
}

var compressCmd = &cobra.Command{
	Use:     "compress <source> [source...]",
	Aliases: []string{"c"},
	Short:   "Compress PS3 games into 7z format",
	Long: `Compress PS3 game folders or archives into 7z format.

This command takes one or more decrypted PS3 ISO game folders or archive files and organizes them
into a standardized directory structure with compressed game files:

{Game Name} [{Game ID}]/
├── game.7z          (compressed game files)
├── _updates/        (updates folder - empty for now)
└── _dlc/           (DLC folder - empty for now)

The game information (title and ID) is extracted from PS3_GAME/PARAM.SFO.

Examples:
  rom-organizer compress /path/to/game_folder
  rom-organizer c /path/to/game1 /path/to/game2 /path/to/game3
  rom-organizer compress --output /target/dir /path/to/game.zip
  rom-organizer c --force /path/to/game_folder1 /path/to/game_folder2`,
	Args: cobra.MinimumNArgs(1),
	RunE: compressHandler,
}

var decompressCmd = &cobra.Command{
	Use:     "decompress <source> [source...]",
	Aliases: []string{"d"},
	Short:   "Decompress PS3 games into folder format",
	Long: `Decompress PS3 game folders or archives into folder format.

This command takes one or more decrypted PS3 ISO game folders or archive files and organizes them
into a standardized directory structure with decompressed game files:

{Game Name} [{Game ID}]/
├── game/            (raw game files, uncompressed)
├── _updates/        (updates folder - empty for now)
└── _dlc/           (DLC folder - empty for now)

The game information (title and ID) is extracted from PS3_GAME/PARAM.SFO.

Examples:
  rom-organizer decompress /path/to/game_folder
  rom-organizer d /path/to/game1 /path/to/game2 /path/to/game3
  rom-organizer decompress --output /target/dir /path/to/game.zip
  rom-organizer d --force /path/to/game_folder1 /path/to/game_folder2`,
	Args: cobra.MinimumNArgs(1),
	RunE: decompressHandler,
}

var organizeCmd = &cobra.Command{
	Use:     "organize <source> [source...]",
	Aliases: []string{"o"},
	Short:   "Organize PS3 games while keeping their existing format",
	Long: `Organize PS3 game folders into the standard structure while keeping their existing format.

This command takes one or more PS3 game folders (or already organized game directories) and 
organizes them into the standardized directory structure while preserving the 
original format (compressed or decompressed):

{Game Name} [{Game ID}]/
├── game.7z OR game/ (keeps original format)
├── _updates/        (updates folder - empty for now)
└── _dlc/           (DLC folder - empty for now)

This is useful for organizing games that are already in your preferred format
without changing their compression state.

By default, files are copied to preserve the original directory. Use --move to 
move files instead (faster, saves disk space, but removes the original).
The --move flag only works with unorganized directories for safety.

Examples:
  rom-organizer organize /path/to/game_folder
  rom-organizer o /path/to/game1 /path/to/game2 /path/to/game3
  rom-organizer organize --output /target/dir /path/to/game_folder
  rom-organizer o --move /path/to/game_folder1 /path/to/game_folder2
  rom-organizer organize --force /path/to/existing_organized_game1 /path/to/game2`,
	Args: cobra.MinimumNArgs(1),
	RunE: organizeHandler,
}

func init() {
	// Add subcommands to root
	rootCmd.AddCommand(metadataCmd)
	rootCmd.AddCommand(compressCmd)
	rootCmd.AddCommand(decompressCmd)
	rootCmd.AddCommand(organizeCmd)

	// Add flags to metadata command
	metadataCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	metadataCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")

	// Add flags to compress command
	compressCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for compressed game")
	compressCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing output directory")
	compressCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")

	// Add flags to decompress command
	decompressCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for decompressed game")
	decompressCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing output directory")
	decompressCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")

	// Add flags to organize command
	organizeCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for organized game")
	organizeCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing output directory")
	organizeCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	organizeCmd.Flags().BoolVarP(&moveSource, "move", "m", false, "Move files instead of copying (deletes source directory - only works with unorganized directories)")
}

func compressHandler(cmd *cobra.Command, args []string) error {
	opts := packager.PackageOptions{
		OutputDir: outputDir,
		Force:     force,
		Verbose:   verbose,
	}
	return packager.PackageGames(args, opts)
}

func decompressHandler(cmd *cobra.Command, args []string) error {
	opts := packager.PackageOptions{
		OutputDir: outputDir,
		Force:     force,
		Verbose:   verbose,
	}
	return packager.UnpackageGames(args, opts)
}

func organizeHandler(cmd *cobra.Command, args []string) error {
	opts := organizer.OrganizeOptions{
		OutputDir:  outputDir,
		Force:      force,
		Verbose:    verbose,
		MoveSource: moveSource,
	}
	return organizer.OrganizeGames(args, opts)
}

func metadataHandler(cmd *cobra.Command, args []string) error {
	path := args[0]

	// First, auto-detect the console type
	detection, err := detect.DetectConsole(path)
	if err != nil {
		return fmt.Errorf("error detecting console type: %w", err)
	}

	if verbose {
		fmt.Printf("Console Detection Results:\n")
		fmt.Printf("=========================\n")
		fmt.Printf("Console Type:    %s\n", detection.ConsoleType.String())
		fmt.Printf("Confidence:      %.2f\n", detection.Confidence)
		fmt.Printf("Game Path:       %s\n", detection.GamePath)
		fmt.Printf("Indicator:       %s\n", detection.IndicatorFound)
		fmt.Printf("Search Depth:    %d\n", detection.SearchDepth)
		if len(detection.AmbiguousFiles) > 0 {
			fmt.Printf("Ambiguous Files: %d found\n", len(detection.AmbiguousFiles))
			for _, file := range detection.AmbiguousFiles {
				fmt.Printf("  - %s\n", file)
			}
		}
		fmt.Println()
	}

	// Handle different console types
	switch detection.ConsoleType {
	case detect.PS3:
		return handlePS3Metadata(path, detection)
	case detect.Unknown:
		if len(detection.AmbiguousFiles) > 0 {
			fmt.Printf("Found %d ambiguous files that need further analysis:\n", len(detection.AmbiguousFiles))
			for _, file := range detection.AmbiguousFiles {
				fmt.Printf("  - %s\n", file)
			}
			return fmt.Errorf("ambiguous file types detected - specific console type analysis not yet implemented")
		}
		return fmt.Errorf("unable to determine console type for: %s", path)
	default:
		return fmt.Errorf("metadata extraction for %s is not yet implemented", detection.ConsoleType.String())
	}
}

// handlePS3Metadata handles metadata extraction for PS3 games
func handlePS3Metadata(originalPath string, detection *detect.DetectionResult) error {
	// For PS3, we need to find the PARAM.SFO file
	var paramSFOPath string

	// If the detection found PS3_GAME directory, look for PARAM.SFO inside it
	if detection.IndicatorFound == "PS3_GAME" {
		paramSFOPath = fmt.Sprintf("%s/PS3_GAME/PARAM.SFO", detection.GamePath)
	} else if detection.IndicatorFound == "PARAM.SFO" {
		// If PARAM.SFO was found directly, use that path
		paramSFOPath = fmt.Sprintf("%s/PARAM.SFO", detection.GamePath)
	} else {
		return fmt.Errorf("unable to locate PARAM.SFO file for PS3 game")
	}

	// Read and parse the PARAM.SFO file
	data, err := os.ReadFile(paramSFOPath)
	if err != nil {
		return fmt.Errorf("reading PARAM.SFO file: %w", err)
	}

	paramSFO, err := parsers.ParseParamSFO(data)
	if err != nil {
		return fmt.Errorf("parsing PS3 PARAM.SFO: %w", err)
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
		fmt.Printf("ROM Metadata Parser\n")
		fmt.Printf("===================\n")
		fmt.Printf("File Type:       PS3 PARAM.SFO\n")
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
