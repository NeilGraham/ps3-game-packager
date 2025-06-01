package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NeilGraham/ps3-game-packager/internal/organizer"
	"github.com/NeilGraham/ps3-game-packager/internal/packager"
	"github.com/NeilGraham/ps3-game-packager/internal/parsers"
)

var (
	verbose    bool
	jsonOutput bool
	outputDir  string
	force      bool
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

var packageCmd = &cobra.Command{
	Use:   "package <source> [source...]",
	Short: "Package PS3 games into compressed format",
	Long: `Package PS3 game folders or archives into compressed format.

This command takes one or more decrypted PS3 ISO game folders or archive files and packages them
into a standardized directory structure with compressed game files:

{Game Name} [{Game ID}]/
├── game.7z          (compressed game files)
├── _updates/        (updates folder - empty for now)
└── _dlc/           (DLC folder - empty for now)

The game information (title and ID) is extracted from PS3_GAME/PARAM.SFO.

Examples:
  ps3-game-packager package /path/to/game_folder
  ps3-game-packager package /path/to/game1 /path/to/game2 /path/to/game3
  ps3-game-packager package --output /target/dir /path/to/game.zip
  ps3-game-packager package --force /path/to/game_folder1 /path/to/game_folder2`,
	Args: cobra.MinimumNArgs(1),
	RunE: packageHandler,
}

var unpackageCmd = &cobra.Command{
	Use:   "unpackage <source> [source...]",
	Short: "Unpackage PS3 games into decompressed format",
	Long: `Unpackage PS3 game folders or archives into decompressed format.

This command takes one or more decrypted PS3 ISO game folders or archive files and packages them
into a standardized directory structure with decompressed game files:

{Game Name} [{Game ID}]/
├── game/            (raw game files, uncompressed)
├── _updates/        (updates folder - empty for now)
└── _dlc/           (DLC folder - empty for now)

The game information (title and ID) is extracted from PS3_GAME/PARAM.SFO.

Examples:
  ps3-game-packager unpackage /path/to/game_folder
  ps3-game-packager unpackage /path/to/game1 /path/to/game2 /path/to/game3
  ps3-game-packager unpackage --output /target/dir /path/to/game.zip
  ps3-game-packager unpackage --force /path/to/game_folder1 /path/to/game_folder2`,
	Args: cobra.MinimumNArgs(1),
	RunE: unpackageHandler,
}

var organizeCmd = &cobra.Command{
	Use:   "organize <source> [source...]",
	Short: "Organize PS3 games while keeping their existing format",
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

Examples:
  ps3-game-packager organize /path/to/game_folder
  ps3-game-packager organize /path/to/game1 /path/to/game2 /path/to/game3
  ps3-game-packager organize --output /target/dir /path/to/game_folder
  ps3-game-packager organize --force /path/to/existing_organized_game1 /path/to/game2`,
	Args: cobra.MinimumNArgs(1),
	RunE: organizeHandler,
}

func init() {
	// Add subcommands to root
	rootCmd.AddCommand(parseParamSFOCmd)
	rootCmd.AddCommand(packageCmd)
	rootCmd.AddCommand(unpackageCmd)
	rootCmd.AddCommand(organizeCmd)

	// Add flags to parse-param-sfo command
	parseParamSFOCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	parseParamSFOCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")

	// Add flags to package command
	packageCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for packaged game")
	packageCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing output directory")
	packageCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")

	// Add flags to unpackage command
	unpackageCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for unpackaged game")
	unpackageCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing output directory")
	unpackageCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")

	// Add flags to organize command
	organizeCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for organized game")
	organizeCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing output directory")
	organizeCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
}

func packageHandler(cmd *cobra.Command, args []string) error {
	opts := packager.PackageOptions{
		OutputDir: outputDir,
		Force:     force,
		Verbose:   verbose,
	}
	return packager.PackageGames(args, opts)
}

func unpackageHandler(cmd *cobra.Command, args []string) error {
	opts := packager.PackageOptions{
		OutputDir: outputDir,
		Force:     force,
		Verbose:   verbose,
	}
	return packager.UnpackageGames(args, opts)
}

func organizeHandler(cmd *cobra.Command, args []string) error {
	opts := organizer.OrganizeOptions{
		OutputDir: outputDir,
		Force:     force,
		Verbose:   verbose,
	}
	return organizer.OrganizeGames(args, opts)
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
