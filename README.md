# PS3 Game Packager

A collection of tools for working with PS3 game files, starting with PARAM.SFO parsing.

## Features

- **PARAM.SFO Parser**: Extract game title, title ID, and other metadata from PS3 PARAM.SFO files
- **Clean API**: Reusable parser package for integration into other tools
- **Multiple Output Formats**: Text and JSON output support
- **Robust Error Handling**: Graceful handling of malformed files

## Installation

### From Source

```bash
git clone https://github.com/NeilGraham/ps3-game-packager.git
cd ps3-game-packager
go build -o ps3-game-packager ./cmd/ps3-game-packager
```

### Using Go Install

```bash
go install github.com/NeilGraham/ps3-game-packager/cmd/ps3-game-packager@latest
```

## Usage

### Parse PARAM.SFO Files

#### Basic usage:
```bash
./ps3-game-packager parse-param-sfo PARAM.SFO
```

Output:
```
Summary:
========
Game Title:  3D DOT GAME HEROES
Title ID:    BLUS30490
App Version: 01.00
Category:    DG
```

#### Verbose output:
```bash
./ps3-game-packager parse-param-sfo --verbose PARAM.SFO
```

#### JSON output:
```bash
./ps3-game-packager parse-param-sfo --json PARAM.SFO
```

#### Flexible flag positioning:
```bash
./ps3-game-packager parse-param-sfo PARAM.SFO --verbose --json
./ps3-game-packager parse-param-sfo --json --verbose PARAM.SFO
```

#### Get help:
```bash
./ps3-game-packager parse-param-sfo -h
./ps3-game-packager --help
```

#### Shell Completion:
```bash
# Generate bash completion
./ps3-game-packager completion bash > ps3-game-packager-completion.bash
source ps3-game-packager-completion.bash

# Generate PowerShell completion
./ps3-game-packager completion powershell > ps3-game-packager-completion.ps1
```

## Project Structure

```
ps3-game-packager/
├── cmd/
│   └── ps3-game-packager/
│       └── main.go              # CLI application entry point
├── internal/
│   └── parsers/
│       └── param_sfo.go         # PARAM.SFO parsing logic
├── go.mod                       # Go module definition
├── README.md                    # This file
└── PARAM.SFO                    # Example file for testing
```

## API Usage

The parser can also be used as a library:

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/NeilGraham/ps3-game-packager/internal/parsers"
)

func main() {
    data, err := os.ReadFile("PARAM.SFO")
    if err != nil {
        panic(err)
    }
    
    paramSFO, err := parsers.ParseParamSFO(data)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Title: %s\n", paramSFO.GetTitle())
    fmt.Printf("Title ID: %s\n", paramSFO.GetTitleID())
    fmt.Printf("App Version: %s\n", paramSFO.GetString("APP_VER"))
}
```

## PARAM.SFO Format

PARAM.SFO files contain metadata about PS3 games. The format includes:

- **Header**: File version, table offsets, and entry count
- **Index Table**: Entry definitions with key offsets and data formats
- **Key Table**: Null-terminated key names
- **Data Table**: Actual values (strings, integers, etc.)

### Supported Data Types

- `0x0004`: UTF-8 strings (special case)
- `0x0204`: UTF-8 strings (standard)
- `0x0404`: 32-bit integers

### Common Keys

- `TITLE`: Game title
- `TITLE_ID`: Unique game identifier
- `APP_VER`: Application version
- `CATEGORY`: Game category (DG for disc games)
- `ATTRIBUTE`: Game attributes
- `BOOTABLE`: Whether the game is bootable
- `LICENSE`: License information
- `PARENTAL_LEVEL`: Age rating level
- `PS3_SYSTEM_VER`: Required PS3 system version
- `RESOLUTION`: Supported resolutions
- `SOUND_FORMAT`: Supported audio formats

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o ps3-game-packager ./cmd/ps3-game-packager
```

### Adding New Parsers

1. Create a new parser in `internal/parsers/`
2. Add the command logic to `cmd/ps3-game-packager/main.go`
3. Update this README with usage examples

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- PARAM.SFO format documentation from PS3 homebrew community
- Go team for excellent tooling and documentation 