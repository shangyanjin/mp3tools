# Writer Module

ID3v2.4 tag writer with UTF-8 encoding support.

## Purpose

The `writer` module is responsible for writing audio file metadata tags. It uses `github.com/bogem/id3v2` to write ID3v2.4 tags with UTF-8 encoding, ensuring proper character encoding for international characters.

## Design Philosophy

- **Write-only**: This module only writes tags, does not read or parse them
- **UTF-8 only**: All tags are written in UTF-8 encoding (ID3v2.4 standard)
- **Separation of concerns**: Works with `tagger` module (read-only) for complete tag management

## Usage

### Basic Usage

```go
import "mp3tools/internal/writer"

// Create a new writer
w, err := writer.New("path/to/file.mp3")
if err != nil {
    log.Fatal(err)
}
defer w.Close()

// Set individual tags
w.SetTitle("Song Title")
w.SetArtist("Artist Name")
w.SetAlbum("Album Name")
w.SetYear("2025")

// Save to original file
if err := w.Save(); err != nil {
    log.Fatal(err)
}
```

### Set All Tags at Once

```go
data := &writer.TagData{
    Title:   "01 歌曲名",
    Artist:  "艺术家",
    Album:   "专辑名称",
    Year:    "2025",
    Genre:   "Pop",
    Comment: "Processed by mp3tools",
}

w.SetAllTags(data)
w.Save()
```

### Convenience Functions

```go
// Write tags to file in one call
err := writer.WriteTagsToFile("file.mp3", &writer.TagData{
    Title:  "Title",
    Artist: "Artist",
})

// Write tags to a new file (copy + tag)
err := writer.WriteTagsToNewFile("source.mp3", "dest.mp3", &writer.TagData{
    Title:  "Title",
    Artist: "Artist",
})
```

### Save to Different File

```go
w, _ := writer.New("original.mp3")
w.SetTitle("New Title")

// Save to a different file (preserves original)
err := w.SaveTo("output/modified.mp3")
```

## API Reference

### Types

#### `TagWriter`
Main writer struct for handling tag operations.

#### `TagData`
Struct containing all tag fields:
- `Title` - Song title
- `Artist` - Artist name
- `Album` - Album name
- `Year` - Release year
- `Genre` - Music genre
- `Track` - Track number
- `Comment` - Comment text

### Functions

#### `New(filePath string) (*TagWriter, error)`
Creates a new TagWriter for the specified file.

#### `(w *TagWriter) SetTitle(title string)`
Sets the title tag.

#### `(w *TagWriter) SetArtist(artist string)`
Sets the artist tag.

#### `(w *TagWriter) SetAlbum(album string)`
Sets the album tag.

#### `(w *TagWriter) SetYear(year string)`
Sets the year tag.

#### `(w *TagWriter) SetGenre(genre string)`
Sets the genre tag.

#### `(w *TagWriter) SetComment(comment string)`
Sets the comment tag.

#### `(w *TagWriter) SetAllTags(data *TagData)`
Sets all tags at once from a TagData struct.

#### `(w *TagWriter) Save() error`
Saves tags to the original file.

#### `(w *TagWriter) SaveTo(destPath string) error`
Saves tags to a new file (copies original file first).

#### `(w *TagWriter) Close() error`
Closes the tag file handle.

#### `WriteTagsToFile(filePath string, data *TagData) error`
Convenience function to write tags in one call.

#### `WriteTagsToNewFile(srcPath, destPath string, data *TagData) error`
Convenience function to write tags to a new file.

## Integration with Other Modules

### With Tagger Module

```go
// Read tags with tagger (read-only)
tags, err := tagger.ReadTags("file.mp3")

// Process/modify tags
newTitle := processTitle(tags.Title)

// Write with writer (write-only)
w, _ := writer.New("file.mp3")
w.SetTitle(newTitle)
w.Save()
```

### With Encoder Module

```go
// Detect and fix encoding
fixedTitle := encoder.ConvertToUTF8(originalTitle)

// Write fixed tags
w, _ := writer.New("file.mp3")
w.SetTitle(fixedTitle)
w.Save()
```

## Notes

- Always call `Close()` or use `defer w.Close()` to properly close file handles
- The module automatically creates ID3v2.4 tags if they don't exist
- All text is written in UTF-8 encoding
- Empty strings are ignored (won't overwrite existing tags)

