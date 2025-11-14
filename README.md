# MP3 Tools

A command-line utility for batch processing audio file metadata tags, fixing encoding issues, and automating metadata operations.

## Features

- **Directory Scanning**: Recursively scan directories and subdirectories for audio files (MP3, FLAC, M4A, AAC, OGG, WMA)
- **Tag Reading**: Read metadata tags (ID3, Vorbis, etc.) including title, artist, album, year, and genre
- **Encoding Fix**: Automatically detect and convert tag encodings (UTF-8, GBK, GB2312, etc.)
- **Auto Tagging**: Automatically fill missing metadata tags
  - Auto-derive album name from directory name
  - Auto-format titles with zero-padding (01, 02, etc.)
  - Derive tags from filename/directory: Title = Number + Album, Artist = Album (with `-f` flag)
- **Batch Processing**: Multi-threaded concurrent processing for improved performance
- **Progress Display**: Real-time progress display with worker status
- **Output Directory**: Option to output processed files to a specified directory while preserving directory structure

## Installation

### Build from Source

```bash
git clone https://github.com/mp3tools/mp3tools.git
cd mp3tools
go build -o mp3tools ./cmd/mp3tools
```

## Usage

```bash
mp3tools <command> [path] [options]
```

### Commands

- `scan <path>` - Scan directory and display audio file tags
- `fix <path>` - Fix encoding issues in audio file tags
- `tag <path>` - Auto-fill missing metadata tags
- `test <path>` - Preview changes with parameters (simulation only, no file modification)
- `check <path>` - Display current tags (display only, no parameters)

### Options

- `-f, --force` - Derive tags from filename and directory name
  - For `tag` and `fix` commands: Title = Number + Album, Artist = Album
  - Default: `false` for `tag` command
- `-n, --threads <number>` - Number of worker threads (default: 5)
- `-u, --update` - Fix encoding only (for `tag` command, default: `true`) or update original files (for other commands)
- `-o, --outdir <directory>` - Output directory, preserve directory structure (default: update original files)

## Examples

### Scan audio files

```bash
mp3tools scan ./music
```

### Fix encoding issues

```bash
# Fix encoding only
mp3tools fix ./music

# Derive and update tags from filename/directory
mp3tools fix ./music -f
```

### Auto-fill tags

```bash
# Auto-fill missing tags and fix encoding (default)
mp3tools tag ./music

# Derive tags from filename/directory (Title = Number + Album, Artist = Album)
mp3tools tag ./music -f

# Only fix encoding, don't derive tags
mp3tools tag ./music -u
```

### Use custom thread count

```bash
mp3tools fix ./music -n 8
```

### Update original files

```bash
mp3tools tag ./music -u
mp3tools fix ./music -u -n 5
```

### Output to custom directory

```bash
mp3tools tag ./music -o ./custom
mp3tools fix ./music -o ./fixed -n 5

# Fix with tag derivation and update to output directory
mp3tools fix ./music -f -o ./fixed
```

### Preview changes (test mode)

```bash
# Preview with default settings
mp3tools test ./music

# Preview with tag derivation
mp3tools test ./music -f

# Preview with encoding fix only
mp3tools test ./music -u
```

### Check current tags

```bash
# Display current tags only (no parameters supported)
mp3tools check ./music
```

## Example Output

### Test Command Output

```
$ mp3tools test ./music

Scanning directory: ./music
Found 15 audio files

[1/15] Processing: 01 歌曲名.mp3 | Current tags: Title="1 歌曲名", Album="", Artist="" | Encoding: GBK -> UTF-8 | Auto-derived: Album="music" (from directory name) | Auto-formatted: Title="01 歌曲名" (zero-padded format) | Updated: ✓

---

Statistics:
  Total files: 15
  Successfully processed: 15
  Failed: 0
  Encoding fixed: 12
  Tags updated: 15
  Auto-derived albums: 8
  Auto-formatted titles: 10
```

## Technical Details

### Dependencies

- `github.com/dhowden/tag` - Audio tag reading
- `github.com/bogem/id3v2` - MP3 tag writing
- `github.com/saintfish/chardet` - Character encoding detection
- `golang.org/x/text` - Text encoding conversion
- `github.com/spf13/cobra` - CLI framework

### Architecture

- **Scanner**: Recursive directory traversal and audio file detection
- **Tagger**: Unified interface for reading/writing tags across formats
- **Encoder**: Encoding detection and conversion utilities
- **Processor**: Batch processing with worker pool pattern
- **Display**: Real-time progress display and statistics

## License

MIT License

Copyright (c) 2025 YANJIN

