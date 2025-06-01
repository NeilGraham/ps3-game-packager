# Development Scripts

This directory contains development tools and scripts used for testing and development of the ROM organizer. **These are not intended for end users** and should only be used by developers working on the project.

## Scripts

### `generate-test-games.go`

Generates fake PS3 game directories with valid PARAM.SFO files for testing purposes. All game data is completely fictional to avoid any copyright concerns.

#### Usage

```bash
# Generate 5 test games in test-games/ directory
go run scripts/dev/generate-test-games.go

# Generate 10 test games in a custom directory
go run scripts/dev/generate-test-games.go -count 10 -output my-test-games

# Clean existing test games and generate new ones
go run scripts/dev/generate-test-games.go -clean -count 8

# Use a specific random seed for reproducible tests
go run scripts/dev/generate-test-games.go -seed 12345
```

#### Options

- `-output <dir>`: Output directory for test games (default: "test-games")
- `-count <n>`: Number of test games to generate (default: 5)
- `-seed <n>`: Random seed for reproducible generation (default: current time)
- `-clean`: Clean output directory before generating new games

#### Generated Structure

Each test game creates the following structure:
```
{Game Name} [{Game ID}]/
├── PS3_GAME/
│   └── PARAM.SFO    # Valid binary PARAM.SFO file
└── PS3_DISC.SFB     # Fake disc metadata
```

#### Testing with Generated Games

After generating test games, you can test the ROM organizer:

```bash
# Test metadata extraction
./rom-organizer metadata test-games/*/

# Test organization
./rom-organizer organize test-games/*/ --output organized-test-games

# Test compression
./rom-organizer compress test-games/*/ --output compressed-test-games

# Test decompression
./rom-organizer decompress compressed-test-games/*/ --output decompressed-test-games
```

## Fake Game Data

The script includes 15 completely fictional game titles with realistic but fake title IDs:

- Galactic Warriors: Return of the Void [BLUS12345]
- Crystal Quest: Legends of Mystara [BLES67890]
- Neon Racers: Future Streets [BCUS11111]
- And more...

All game data is designed to:
- Have realistic PS3 title ID formats (BLUS/BLES/BCUS/BCES + 5 digits)
- Include common PARAM.SFO fields found in real games
- Be obviously fictional to avoid any legal concerns
- Test various edge cases in game titles (special characters, long names, etc.)

## Legal Notice

All generated content is completely fictional and created solely for testing purposes. No actual game content or copyrighted material is used or referenced. 