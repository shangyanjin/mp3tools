# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- Initial implementation of MP3 tools
- `scan` command: Scan directory and display audio file tags
- `fix` command: Fix encoding issues in audio file tags
- `tag` command: Auto-fill missing metadata tags
- `test` command: Preview changes with parameters (simulation only)
- `check` command: Display current tags (display only, no parameters)
- Multi-threaded concurrent processing with configurable worker threads (default: 5)
- Real-time progress display with file counter `[current/total]`
- Encoding detection and conversion (UTF-8, GBK, GB2312, etc.)
- Double encoding fix support (UTF-8 bytes misinterpreted as ISO-8859-1)
- Auto-derive album name from directory name
- Auto-format titles with zero-padding support (e.g., "1 Title" -> "01 Title")
- Output directory support with directory structure preservation
- Tag derivation from filename/directory: Title = Number + Album, Artist = Album (with `-f` flag)
- ID3v2.4 tag writing with UTF-8 encoding
- Unified output format across all commands

### Changed
- **BREAKING**: Only MP3 files are supported (removed FLAC, M4A, AAC, OGG, WMA support)
- Default behavior: Update original files (no output directory by default)
- `tag` command: `-f` flag default is `false`, `-u` flag default is `true`
- `fix` command: Supports `-f` flag to derive tags before fixing encoding
- `test` command: Supports `-f` and `-u` flags for simulation
- `check` command: Unified output format with fix/tag commands
- Output format: Shows `old -> new` changes for each modified field
- Output format: Shows `No changes` when no modifications are made

### Added Features
- `-u, --update` flag: Fix encoding only (for `tag` command, default: `true`) or update original files (for other commands)
- `-f, --force` flag: Derive tags from filename and directory name (Title = Number + Album, Artist = Album)
- `-n, --threads` flag: Configure number of worker threads (default: 5)
- `-o, --outdir` flag: Specify custom output directory (default: update original files)

### Output Format
- `fix`/`tag` commands: `[n/total] Processing: filename | Title: "old" -> "new" | Album: "old" -> "new" | Artist: "old" -> "new" | Updated: ✓`
- `check` command: `[n/total] Processing: filename | Title: "value" | Album: "value" | Artist: "value"`
- No changes: `[n/total] Processing: filename | No changes`

## [0.1.0] - 2025-11-14

### Initial Release
- Core functionality implemented
- Command-line interface with Cobra framework
- Audio tag reading and writing support
- Encoding detection and conversion
- Batch processing capabilities
- Progress display system

