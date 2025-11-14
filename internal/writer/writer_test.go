package writer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.mp3")

	// Create an empty MP3 file for testing
	if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	writer, err := New(testFile)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	if writer.filePath != testFile {
		t.Errorf("Expected filePath %s, got %s", testFile, writer.filePath)
	}

	if writer.tag == nil {
		t.Error("Expected tag to be initialized")
	}
}

func TestSetAllTags(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.mp3")

	if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	writer, err := New(testFile)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	data := &TagData{
		Title:  "Test Title",
		Artist: "Test Artist",
		Album:  "Test Album",
		Year:   "2025",
		Genre:  "Test Genre",
	}

	writer.SetAllTags(data)

	// Verify tags were set
	if writer.tag.Title() != data.Title {
		t.Errorf("Expected title %s, got %s", data.Title, writer.tag.Title())
	}
	if writer.tag.Artist() != data.Artist {
		t.Errorf("Expected artist %s, got %s", data.Artist, writer.tag.Artist())
	}
	if writer.tag.Album() != data.Album {
		t.Errorf("Expected album %s, got %s", data.Album, writer.tag.Album())
	}
	if writer.tag.Year() != data.Year {
		t.Errorf("Expected year %s, got %s", data.Year, writer.tag.Year())
	}
	if writer.tag.Genre() != data.Genre {
		t.Errorf("Expected genre %s, got %s", data.Genre, writer.tag.Genre())
	}
}

func TestWriteTagsToFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.mp3")

	if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	data := &TagData{
		Title:  "测试标题",
		Artist: "测试艺术家",
		Album:  "测试专辑",
		Year:   "2025",
	}

	if err := WriteTagsToFile(testFile, data); err != nil {
		t.Fatalf("Failed to write tags: %v", err)
	}

	// Verify tags were written
	writer, err := New(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer writer.Close()

	if writer.tag.Title() != data.Title {
		t.Errorf("Expected title %s, got %s", data.Title, writer.tag.Title())
	}
}

