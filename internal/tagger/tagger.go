package tagger

import (
	"fmt"
	"os"

	"github.com/dhowden/tag"
)

// Metadata represents audio file metadata
type Metadata struct {
	Title   string
	Artist  string
	Album   string
	Year    int
	Genre   string
	Track   int
	Comment string
	Format  tag.Format
}

// ReadTags reads metadata tags from an audio file
func ReadTags(filePath string) (*Metadata, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	meta, err := tag.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read tags: %w", err)
	}

	track, _ := meta.Track()
	year := 0
	if meta.Year() != 0 {
		year = meta.Year()
	}

	return &Metadata{
		Title:   meta.Title(),
		Artist:  meta.Artist(),
		Album:   meta.Album(),
		Year:    year,
		Genre:   meta.Genre(),
		Track:   track,
		Comment: meta.Comment(),
		Format:  meta.Format(),
	}, nil
}

// HasTag checks if a specific tag field has a value
func (m *Metadata) HasTag(field string) bool {
	switch field {
	case "title":
		return m.Title != ""
	case "artist":
		return m.Artist != ""
	case "album":
		return m.Album != ""
	case "year":
		return m.Year != 0
	case "genre":
		return m.Genre != ""
	case "track":
		return m.Track != 0
	case "comment":
		return m.Comment != ""
	default:
		return false
	}
}

// IsEmpty checks if all tags are empty
func (m *Metadata) IsEmpty() bool {
	return m.Title == "" &&
		m.Artist == "" &&
		m.Album == "" &&
		m.Year == 0 &&
		m.Genre == "" &&
		m.Track == 0 &&
		m.Comment == ""
}

// GetRawBytes returns raw bytes of a tag field for encoding detection
func GetRawBytes(filePath string, field string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	meta, err := tag.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read tags: %w", err)
	}

	var value string
	switch field {
	case "title":
		value = meta.Title()
	case "artist":
		value = meta.Artist()
	case "album":
		value = meta.Album()
	case "genre":
		value = meta.Genre()
	case "comment":
		value = meta.Comment()
	default:
		return nil, fmt.Errorf("unsupported field: %s", field)
	}

	return []byte(value), nil
}

