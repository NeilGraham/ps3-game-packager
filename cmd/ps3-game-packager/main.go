package main

import (
	"fmt"
	"os"

	"ps3-game-packager/internal/parsers"

	"github.com/spf13/cobra"
)

var (
	verbose    bool
	jsonOutput bool
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

func init() {
	// Add subcommands to root
	rootCmd.AddCommand(parseParamSFOCmd)

	// Add flags to parse-param-sfo command
	parseParamSFOCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	parseParamSFOCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
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
