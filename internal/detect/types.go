package detect

// ConsoleType represents the different console types we can detect
type ConsoleType int

const (
	Unknown ConsoleType = iota
	PS3
	// Future console types can be added here
	// PS2
	// Xbox
	// Xbox360
	// etc.
)

// String returns the string representation of the console type
func (c ConsoleType) String() string {
	switch c {
	case PS3:
		return "PlayStation 3"
	default:
		return "Unknown"
	}
}

// DetectionResult holds the result of console detection
type DetectionResult struct {
	ConsoleType    ConsoleType // The detected console type
	GamePath       string      // Path to the game directory (parent of indicator)
	Confidence     float64     // Confidence level (0.0 to 1.0)
	IndicatorFound string      // The specific indicator that was found
	AmbiguousFiles []string    // Files that need secondary analysis
	SearchDepth    int         // How deep we searched to find this
}

// IsValid returns true if the detection result is valid
func (r DetectionResult) IsValid() bool {
	return r.ConsoleType != Unknown && r.Confidence > 0.0
}

// IsHighConfidence returns true if we're very confident about the detection
func (r DetectionResult) IsHighConfidence() bool {
	return r.Confidence >= 0.8
}
