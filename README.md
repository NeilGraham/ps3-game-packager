# ROM Organizer

> **Note**: This repository was originally focused on PlayStation 3 (PS3) games but has been expanded to support multiple console types. All existing PS3 functionality is preserved while providing a foundation for adding more console support.

A collection of tools for working with ROM game files, providing utilities for organizing and optimizing ROM game collections from various consoles.

**Currently Supported Consoles:**
- PlayStation 3 (PS3)

**Planned Support:**
- PlayStation 2 (PS2)
- PlayStation 1 (PSX)
- Xbox
- Xbox 360
- Nintendo GameCube
- Nintendo Wii
- And more...

## Features

- **Organize**: Organize games while preserving their existing format (compressed/decompressed)
- **Compress**: Compress games into 7z archives with organized directory structure
- **Decompress**: Organize games into decompressed format with standardized structure  
- **Metadata**: Extract metadata from ROM files (currently supports PS3 PARAM.SFO)

## Installation

1. Clone the repository
2. Build the application:
   ```bash
   go build ./cmd/rom-organizer
   ```

## Project Structure

```
rom-organizer/
├── cmd/rom-organizer/          # Main application entry point
│   └── main.go
├── internal/                   # Internal packages
│   ├── common/                 # Shared utilities and interfaces
│   │   └── utils.go           # File operations, game info structures
│   ├── consoles/              # Console-specific handlers
│   │   ├── registry.go        # Console handler registry
│   │   └── ps3.go            # PlayStation 3 handler
│   ├── detect/                # Console detection logic
│   │   ├── detect.go         # Main detection algorithm
│   │   ├── indicators.go     # Console-specific indicators
│   │   └── types.go          # Detection types and results
│   ├── organizer/             # Organization logic
│   │   └── organizer.go      # Organize command implementation
│   └── parsers/               # File parsers organized by console
│       └── ps3.go            # PS3 PARAM.SFO parser
├── go.mod
├── go.sum
└── README.md
```

## Commands

### Compress Command

Compresses games into **compressed format** with 7z archives:

```bash
rom-organizer compress <source> [flags]
```

**Output Structure:**
```
{Game Name} [{Game ID}]/
├── game.7z          # Compressed game files
├── _updates/        # Updates folder (empty)
└── _dlc/           # DLC folder (empty)
```

**Examples:**
```bash
rom-organizer compress /path/to/game_folder
rom-organizer compress --output /target/dir /path/to/game.zip
rom-organizer compress --force /path/to/game_folder
```

### Decompress Command

Organizes games into **decompressed format** with raw files:

```bash
rom-organizer decompress <source> [flags]
```

**Output Structure:**
```
{Game Name} [{Game ID}]/
├── game/            # Raw game files (uncompressed)
├── _updates/        # Updates folder (empty)
└── _dlc/           # DLC folder (empty)
```

**Examples:**
```bash
rom-organizer decompress /path/to/game_folder
rom-organizer decompress --output /target/dir /path/to/game.zip
rom-organizer decompress --force /path/to/game_folder
```

### Organize Command

Organizes games while **preserving existing format**:

```bash
rom-organizer organize <source> [flags]
```

**Output Structure:**
```
{Game Name} [{Game ID}]/
├── game.7z OR game/ # Keeps original format
├── _updates/        # Updates folder (empty)
└── _dlc/           # DLC folder (empty)
```

This command is useful for:
- Organizing games already in your preferred format
- Moving already organized game directories
- Maintaining existing compression choices

**Examples:**
```bash
rom-organizer organize /path/to/game_folder
rom-organizer organize --output /target/dir /path/to/game_folder
rom-organizer organize --force /path/to/existing_organized_game
```

### Metadata Command

Extract metadata from ROM files:

```bash
rom-organizer metadata <file> [flags]
```

Currently supports:
- **PS3 PARAM.SFO files**: Extract title, title ID, version, and other game attributes

More ROM formats will be supported in future versions.

**Examples:**
```bash
rom-organizer metadata PARAM.SFO
rom-organizer metadata --verbose PARAM.SFO
rom-organizer metadata --json PARAM.SFO
```

## Flags

All packaging commands support these flags:

- `-o, --output string`: Output directory (default: current directory)
- `-f, --force`: Overwrite existing output directory
- `-v, --verbose`: Show detailed information
- `-h, --help`: Show help for the command

The metadata command supports:
- `-v, --verbose`: Show detailed file structure information
- `-j, --json`: Output metadata in JSON format

## Requirements

- **7-Zip**: Required for the `compress` command to create 7z archives
  - **Windows**: Download from [7-zip.org](https://www.7-zip.org/) or install via `choco install 7zip`
  - **macOS**: Install via `brew install p7zip`
  - **Linux**: Install via package manager (e.g., `sudo apt install p7zip-full`)

## Supported Input Formats

### PlayStation 3 (PS3)
- **Game Folders**: Decrypted PS3 ISO folder containing `PS3_GAME/PARAM.SFO`
- **ZIP Archives**: Archive files containing PS3 game folders
- **Organized Directories**: Already organized game directories (for organize command)
- **PARAM.SFO files**: For metadata extraction

### Future Console Support
The application is designed to easily support additional consoles. Each console will have:
- Specific file structure detection
- Metadata extraction capabilities
- Standardized organization output

## Game Information Extraction

The tools automatically extract game information from console-specific metadata files:

### PlayStation 3
- **Source**: `PS3_GAME/PARAM.SFO` files
- **Extracted Data**: Game Title, Title ID (e.g., BLUS30490), App Version, Category

This information is used to create standardized directory names in the format: `{Game Name} [{Game ID}]`

## Console Detection

The application uses an intelligent detection system to automatically identify console types:

1. **File Structure Analysis**: Looks for console-specific directories and files
2. **Metadata File Detection**: Identifies characteristic metadata files
3. **Confidence Scoring**: Provides confidence levels for detections
4. **Ambiguous File Handling**: Manages files that could belong to multiple consoles

## Error Handling

The application provides detailed error messages for common issues:
- Missing metadata files
- Invalid source paths
- Missing 7z installation
- File permission issues
- Disk space problems
- Unsupported file formats
- Unsupported console types

## Adding New Console Support

The codebase is structured to make adding new console support straightforward:

1. **Create Console Handler**: Implement the `ConsoleHandler` interface in `internal/consoles/`
2. **Add Detection Rules**: Update `internal/detect/indicators.go` with console-specific indicators
3. **Register Handler**: Add the new handler to the registry in `internal/consoles/registry.go`
4. **Add Parser**: If needed, create console-specific parsers in `internal/parsers/`

## Version

Current version: 1.0.0 