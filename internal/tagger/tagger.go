package tagger

import (
	"fmt"
	"os"

	"github.com/bogem/id3v2/v2"
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
	// Try to read using id3v2 first to get raw bytes
	id3Tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err == nil {
		defer id3Tag.Close()

		// Get raw text frames to handle encoding properly
		title := readTextFrame(id3Tag, "TIT2")
		artist := readTextFrame(id3Tag, "TPE1")
		album := readTextFrame(id3Tag, "TALB")
		genre := readTextFrame(id3Tag, "TCON")
		comment := readCommentFrame(id3Tag)

		// Get year and track
		year := 0
		if yearStr := id3Tag.Year(); yearStr != "" {
			// Try to parse year
			if len(yearStr) >= 4 {
				fmt.Sscanf(yearStr[:4], "%d", &year)
			}
		}

		track := 0
		trackFrame := id3Tag.GetTextFrame("TRCK")
		if trackFrame.Text != "" {
			fmt.Sscanf(trackFrame.Text, "%d", &track)
		}

		// Determine format
		format := tag.UnknownFormat
		if id3Tag != nil {
			format = tag.Format("MP3")
		}

		return &Metadata{
			Title:   title,
			Artist:  artist,
			Album:   album,
			Year:    year,
			Genre:   genre,
			Track:   track,
			Comment: comment,
			Format:  format,
		}, nil
	}

	// Fallback to dhowden/tag if id3v2 fails
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

// readTextFrame reads a text frame and handles encoding conversion
func readTextFrame(tag *id3v2.Tag, frameID string) string {
	textFrame := tag.GetTextFrame(frameID)
	if textFrame.Text == "" {
		return ""
	}

	// Get the text value
	text := textFrame.Text

	// If text contains invalid UTF-8 or looks like it needs encoding conversion,
	// we'll let the encoder.FixEncoding handle it later
	return text
}

// readCommentFrame reads a comment frame and handles encoding conversion
func readCommentFrame(tag *id3v2.Tag) string {
	commentFrames := tag.GetFrames(tag.CommonID("COMM"))
	if len(commentFrames) == 0 {
		return ""
	}

	// Get the first comment frame
	if comm, ok := commentFrames[0].(id3v2.CommentFrame); ok {
		return comm.Text
	}

	return ""
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
