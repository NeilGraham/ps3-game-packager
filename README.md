# PS3 Game Packager

A collection of tools for working with PS3 game files, providing utilities for organizing and optimizing PS3 game collections.

## Features

- **Package**: Compress PS3 games into 7z archives with organized directory structure
- **Unpackage**: Organize PS3 games into decompressed format with standardized structure  
- **Organize**: Organize PS3 games while preserving their existing format (compressed/decompressed)
- **Parse PARAM.SFO**: Extract metadata from PS3 game files

## Installation

1. Clone the repository
2. Build the application:
   ```bash
   go build ./cmd/ps3-game-packager
   ```

## Project Structure

```
ps3-game-packager/
├── cmd/ps3-game-packager/          # Main application entry point
│   └── main.go
├── internal/                       # Internal packages
│   ├── common/                     # Shared utilities
│   │   └── utils.go               # File operations, game info extraction
│   ├── organizer/                  # Organization logic
│   │   └── organizer.go           # Organize command implementation
│   ├── packager/                   # Packaging logic
│   │   └── packager.go            # Package/unpackage implementations
│   └── parsers/                    # File parsers
│       └── param_sfo.go           # PARAM.SFO parser
├── go.mod
├── go.sum
└── README.md
```

## Commands

### Package Command

Packages PS3 games into **compressed format** with 7z archives:

```bash
ps3-game-packager package <source> [flags]
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
ps3-game-packager package /path/to/game_folder
ps3-game-packager package --output /target/dir /path/to/game.zip
ps3-game-packager package --force /path/to/game_folder
```

### Unpackage Command

Packages PS3 games into **decompressed format** with raw files:

```bash
ps3-game-packager unpackage <source> [flags]
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
ps3-game-packager unpackage /path/to/game_folder
ps3-game-packager unpackage --output /target/dir /path/to/game.zip
ps3-game-packager unpackage --force /path/to/game_folder
```

### Organize Command

Organizes PS3 games while **preserving existing format**:

```bash
ps3-game-packager organize <source> [flags]
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
ps3-game-packager organize /path/to/game_folder
ps3-game-packager organize --output /target/dir /path/to/game_folder
ps3-game-packager organize --force /path/to/existing_organized_game
```

### Parse PARAM.SFO Command

Extract metadata from PS3 PARAM.SFO files:

```bash
ps3-game-packager parse-param-sfo <PARAM.SFO file> [flags]
```

**Examples:**
```bash
ps3-game-packager parse-param-sfo PARAM.SFO
ps3-game-packager parse-param-sfo --verbose PARAM.SFO
ps3-game-packager parse-param-sfo --json PARAM.SFO
```

## Flags

All packaging commands support these flags:

- `-o, --output string`: Output directory (default: current directory)
- `-f, --force`: Overwrite existing output directory
- `-v, --verbose`: Show detailed information
- `-h, --help`: Show help for the command

## Requirements

- **7-Zip**: Required for the `package` command to create 7z archives
  - **Windows**: Download from [7-zip.org](https://www.7-zip.org/) or install via `choco install 7zip`
  - **macOS**: Install via `brew install p7zip`
  - **Linux**: Install via package manager (e.g., `sudo apt install p7zip-full`)

## Input Formats

- **PS3 Game Folders**: Decrypted PS3 ISO folder containing `PS3_GAME/PARAM.SFO`
- **ZIP Archives**: Archive files containing PS3 game folders
- **Organized Directories**: Already organized game directories (for organize command)

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

## Version

Current version: 1.0.0 