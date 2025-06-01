package detect

import (
	"strings"
)

// ConsoleIndicators maps specific files/folders to console types
// These are definitive indicators that immediately identify a console
var ConsoleIndicators = map[string]ConsoleType{
	"PS3_GAME":  PS3, // PS3 decrypted ISO directory structure
	"PARAM.SFO": PS3, // PS3 metadata file (when found at appropriate level)
}

// AmbiguousExtensions are file extensions that could belong to multiple consoles
// These require secondary analysis to determine the actual console type
var AmbiguousExtensions = []string{
	".pkg", // PS3 package files (but could be other consoles in future)
	".iso", // Could be PS1, PS2, PS3, Xbox, GameCube, etc.
	".chd", // Compressed Hunks of Data - could be various consoles
}

// IsAmbiguousFile checks if a filename has an ambiguous extension
func IsAmbiguousFile(filename string) bool {
	lower := strings.ToLower(filename)
	for _, ext := range AmbiguousExtensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// GetConsoleFromIndicator returns the console type for a given indicator
// Returns Unknown if the indicator is not recognized
func GetConsoleFromIndicator(indicator string) ConsoleType {
	if console, exists := ConsoleIndicators[indicator]; exists {
		return console
	}
	return Unknown
}

// IsDefinitiveIndicator checks if a filename/dirname is a definitive console indicator
func IsDefinitiveIndicator(name string) bool {
	_, exists := ConsoleIndicators[name]
	return exists
}
