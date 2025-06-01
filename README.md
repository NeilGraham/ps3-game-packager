# ROM Organizer

At the moment the repository is focused on PS3 games, but it will be extended to support other consoles in the future.

A collection of tools for working with ROM game files, providing utilities for organizing and optimizing ROM game collections.

## Features

- **Organize**: Organize PS3 games while preserving their existing format (compressed/decompressed)
- **Compress**: Compress PS3 games into 7z archives with organized directory structure
- **Decompress**: Organize PS3 games into decompressed format with standardized structure  
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
├── internal/                       # Internal packages
│   ├── common/                     # Shared utilities
│   │   └── utils.go               # File operations, game info extraction
│   ├── organizer/                  # Organization logic
│   │   └── organizer.go           # Organize command implementation
│   ├── packager/                   # Packaging logic
│   │   └── packager.go            # Package/unpackage implementations
│   └── parsers/                    # File parsers organized by console
│       └── ps3.go                 # PS3 PARAM.SFO parser
├── go.mod
├── go.sum
└── README.md
```

## Commands

### Compress Command

Compresses PS3 games into **compressed format** with 7z archives:

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

Organizes PS3 games into **decompressed format** with raw files:

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

Organizes PS3 games while **preserving existing format**:

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

## Input Formats

- **PS3 Game Folders**: Decrypted PS3 ISO folder containing `PS3_GAME/PARAM.SFO`
- **ZIP Archives**: Archive files containing PS3 game folders
- **Organized Directories**: Already organized game directories (for organize command)
- **PS3 PARAM.SFO files**: For metadata extraction

## Game Information Extraction

The tools automatically extract game information from `PS3_GAME/PARAM.SFO` files:
- Game Title
- Title ID (e.g., BLUS30490)
- App Version
- Category

This information is used to create standardized directory names in the format: `{Game Name} [{Game ID}]`

## Error Handling

The application provides detailed error messages for common issues:
- Missing PARAM.SFO files
- Invalid source paths
- Missing 7z installation
- File permission issues
- Disk space problems
- Unsupported file formats (for metadata command)

## Version

Current version: 1.0.0 