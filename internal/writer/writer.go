package writer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bogem/id3v2/v2"
)

// TagWriter handles writing ID3v2.4 tags with UTF-8 encoding
type TagWriter struct {
	filePath string
	tag      *id3v2.Tag
}

// TagData represents the metadata to be written
type TagData struct {
	Title   string
	Artist  string
	Album   string
	Year    string
	Genre   string
	Track   string
	Comment string
}

// New creates a new TagWriter for the specified file
func New(filePath string) (*TagWriter, error) {
	// Open or create ID3v2.4 tag
	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		// If file doesn't have tag, create new one
		tag = id3v2.NewEmptyTag()
	}

	// Set version to ID3v2.4
	tag.SetVersion(4)

	return &TagWriter{
		filePath: filePath,
		tag:      tag,
	}, nil
}

// SetTitle sets the title tag
func (w *TagWriter) SetTitle(title string) {
	if title != "" {
		w.tag.SetTitle(title)
	}
}

// SetArtist sets the artist tag
func (w *TagWriter) SetArtist(artist string) {
	if artist != "" {
		w.tag.SetArtist(artist)
	}
}

// SetAlbum sets the album tag
func (w *TagWriter) SetAlbum(album string) {
	if album != "" {
		w.tag.SetAlbum(album)
	}
}

// SetYear sets the year tag
func (w *TagWriter) SetYear(year string) {
	if year != "" {
		w.tag.SetYear(year)
	}
}

// SetGenre sets the genre tag
func (w *TagWriter) SetGenre(genre string) {
	if genre != "" {
		w.tag.SetGenre(genre)
	}
}

// SetComment sets the comment tag
func (w *TagWriter) SetComment(comment string) {
	if comment != "" {
		commentFrame := id3v2.CommentFrame{
			Encoding:    id3v2.EncodingUTF8,
			Language:    "eng",
			Description: "",
			Text:        comment,
		}
		w.tag.AddCommentFrame(commentFrame)
	}
}

// SetAllTags sets all tags at once
func (w *TagWriter) SetAllTags(data *TagData) {
	if data.Title != "" {
		w.SetTitle(data.Title)
	}
	if data.Artist != "" {
		w.SetArtist(data.Artist)
	}
	if data.Album != "" {
		w.SetAlbum(data.Album)
	}
	if data.Year != "" {
		w.SetYear(data.Year)
	}
	if data.Genre != "" {
		w.SetGenre(data.Genre)
	}
	if data.Comment != "" {
		w.SetComment(data.Comment)
	}
}

// Save writes the tags to the original file
func (w *TagWriter) Save() error {
	if err := w.tag.Save(); err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}
	return nil
}

// SaveTo writes the tags to a new file (copy with new tags)
func (w *TagWriter) SaveTo(destPath string) error {
	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Save current tag values before closing
	title := w.tag.Title()
	artist := w.tag.Artist()
	album := w.tag.Album()
	year := w.tag.Year()
	genre := w.tag.Genre()

	// Copy original file to destination
	if err := copyFile(w.filePath, destPath); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Open destination file and write tags
	destTag, err := id3v2.Open(destPath, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open destination file: %w", err)
	}
	defer destTag.Close()

	// Set version to ID3v2.4
	destTag.SetVersion(4)

	// Copy all frames
	destTag.SetTitle(title)
	destTag.SetArtist(artist)
	destTag.SetAlbum(album)
	destTag.SetYear(year)
	destTag.SetGenre(genre)

	// Save to destination
	if err := destTag.Save(); err != nil {
		return fmt.Errorf("failed to save destination tags: %w", err)
	}

	return nil
}

// Close closes the tag file
func (w *TagWriter) Close() error {
	if w.tag != nil {
		return w.tag.Close()
	}
	return nil
}

// GetTag returns the underlying tag for advanced operations
func (w *TagWriter) GetTag() *id3v2.Tag {
	return w.tag
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceData, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	if err := os.WriteFile(dst, sourceData, 0644); err != nil {
		return err
	}

	return nil
}

// WriteTagsToFile is a convenience function to write tags to a file in one call
func WriteTagsToFile(filePath string, data *TagData) error {
	writer, err := New(filePath)
	if err != nil {
		return err
	}
	defer writer.Close()

	writer.SetAllTags(data)
	return writer.Save()
}

// WriteTagsToNewFile is a convenience function to write tags to a new file
func WriteTagsToNewFile(srcPath, destPath string, data *TagData) error {
	writer, err := New(srcPath)
	if err != nil {
		return err
	}
	defer writer.Close()

	writer.SetAllTags(data)
	return writer.SaveTo(destPath)
}
