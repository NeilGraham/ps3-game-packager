// Package parsers provides console-specific file format parsers
// This file contains parsers for PlayStation 3 (PS3) specific file formats
package parsers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// Data format constants for PARAM.SFO entries
const (
	FMT_UTF8_SPECIAL = 0x0004 // UTF-8 string (special case)
	FMT_UTF8         = 0x0204 // UTF-8 string
	FMT_INT32        = 0x0404 // 32-bit integer
)

// ParamSFOHeader represents the header of a PARAM.SFO file
type ParamSFOHeader struct {
	Version         uint32
	KeyTableOffset  uint32
	DataTableOffset uint32
	EntryCount      uint32
}

// ParamSFOEntry represents a single entry in the PARAM.SFO file
type ParamSFOEntry struct {
	Key     string
	Value   interface{}
	DataFmt uint16
	DataLen uint32
	DataMax uint32
	DataOff uint32
}

// ParamSFO represents a parsed PARAM.SFO file
type ParamSFO struct {
	Header  ParamSFOHeader
	Entries []ParamSFOEntry
}

// GetTitle returns the game title from the PARAM.SFO data
func (p *ParamSFO) GetTitle() string {
	for _, entry := range p.Entries {
		if entry.Key == "TITLE" {
			if str, ok := entry.Value.(string); ok {
				return str
			}
		}
	}
	return ""
}

// GetTitleID returns the title ID from the PARAM.SFO data
func (p *ParamSFO) GetTitleID() string {
	for _, entry := range p.Entries {
		if entry.Key == "TITLE_ID" {
			if str, ok := entry.Value.(string); ok {
				return str
			}
		}
	}
	return ""
}

// GetEntry returns a specific entry by key name
func (p *ParamSFO) GetEntry(key string) (ParamSFOEntry, bool) {
	for _, entry := range p.Entries {
		if entry.Key == key {
			return entry, true
		}
	}
	return ParamSFOEntry{}, false
}

// GetString returns a string value for the given key
func (p *ParamSFO) GetString(key string) string {
	if entry, found := p.GetEntry(key); found {
		if str, ok := entry.Value.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt returns an integer value for the given key
func (p *ParamSFO) GetInt(key string) uint32 {
	if entry, found := p.GetEntry(key); found {
		if num, ok := entry.Value.(uint32); ok {
			return num
		}
	}
	return 0
}

// rawEntry represents the binary structure of a PARAM.SFO entry (16 bytes)
type rawEntry struct {
	KeyOffset uint16 // Offset in key table
	DataFmt   uint16 // Data format
	DataLen   uint32 // Length of data
	DataMax   uint32 // Maximum length of data
	DataOff   uint32 // Offset in data table
}

// ParseParamSFO parses a PARAM.SFO file from raw bytes
func ParseParamSFO(data []byte) (*ParamSFO, error) {
	// Check magic header
	if len(data) < 4 || string(data[:4]) != "\x00PSF" {
		return nil, fmt.Errorf("not a valid PARAM.SFO file: invalid magic header")
	}

	if len(data) < 20 {
		return nil, fmt.Errorf("file too small to contain valid header")
	}

	// Parse header
	version := binary.LittleEndian.Uint32(data[4:8])
	keyTableOffset := binary.LittleEndian.Uint32(data[8:12])
	dataTableOffset := binary.LittleEndian.Uint32(data[12:16])
	entryCount := binary.LittleEndian.Uint32(data[16:20])

	// Validate offsets
	if keyTableOffset >= uint32(len(data)) || dataTableOffset >= uint32(len(data)) {
		return nil, fmt.Errorf("invalid table offsets")
	}

	header := ParamSFOHeader{
		Version:         version,
		KeyTableOffset:  keyTableOffset,
		DataTableOffset: dataTableOffset,
		EntryCount:      entryCount,
	}

	// Parse raw entries
	rawEntries := make([]rawEntry, entryCount)
	for i := uint32(0); i < entryCount; i++ {
		entryOffset := 20 + i*16
		if entryOffset+16 > uint32(len(data)) {
			return nil, fmt.Errorf("entry %d extends beyond file", i)
		}

		buf := bytes.NewReader(data[entryOffset : entryOffset+16])
		if err := binary.Read(buf, binary.LittleEndian, &rawEntries[i]); err != nil {
			return nil, fmt.Errorf("error parsing entry %d: %v", i, err)
		}
	}

	// Convert raw entries to structured entries
	entries := make([]ParamSFOEntry, 0, entryCount)
	for i, raw := range rawEntries {
		// Parse key
		keyStart := int(keyTableOffset) + int(raw.KeyOffset)
		if keyStart >= len(data) {
			return nil, fmt.Errorf("entry %d: invalid key offset", i)
		}

		keyEnd := bytes.IndexByte(data[keyStart:], 0)
		if keyEnd == -1 {
			return nil, fmt.Errorf("entry %d: key not null-terminated", i)
		}

		key := string(data[keyStart : keyStart+keyEnd])

		// Parse value
		valStart := int(dataTableOffset) + int(raw.DataOff)
		valEnd := valStart + int(raw.DataLen)

		if valStart < 0 || valEnd > len(data) {
			return nil, fmt.Errorf("entry %d (%s): value out of bounds", i, key)
		}

		val := data[valStart:valEnd]

		// Format value based on data format
		var formattedValue interface{}
		switch raw.DataFmt {
		case FMT_UTF8_SPECIAL, FMT_UTF8:
			// UTF-8 string, remove null terminator if present
			str := string(val)
			if nullIdx := strings.IndexByte(str, 0); nullIdx != -1 {
				str = str[:nullIdx]
			}
			formattedValue = str

		case FMT_INT32:
			if len(val) >= 4 {
				num := binary.LittleEndian.Uint32(val)
				formattedValue = num
			} else {
				return nil, fmt.Errorf("entry %d (%s): invalid integer data", i, key)
			}

		default:
			// Store as raw bytes for unsupported formats
			formattedValue = val
		}

		entry := ParamSFOEntry{
			Key:     key,
			Value:   formattedValue,
			DataFmt: raw.DataFmt,
			DataLen: raw.DataLen,
			DataMax: raw.DataMax,
			DataOff: raw.DataOff,
		}

		entries = append(entries, entry)
	}

	return &ParamSFO{
		Header:  header,
		Entries: entries,
	}, nil
}
