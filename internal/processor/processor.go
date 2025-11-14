package processor

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"mp3tools/internal/encoder"
	"mp3tools/internal/scanner"
	"mp3tools/internal/tagger"
	"mp3tools/internal/writer"
)

// ProcessOptions contains options for processing files
type ProcessOptions struct {
	Force          bool   // Derive tags from filename and directory
	UpdateEncoding bool   // Fix encoding only (for tag command)
	OutDir         string // Output directory (empty means update in place)
	Threads        int    // Number of worker threads
}

// Processor handles batch processing of audio files
type Processor struct {
	options ProcessOptions
	stats   Statistics
	mu      sync.Mutex
}

// Statistics tracks processing statistics
type Statistics struct {
	Total         int
	Success       int
	Failed        int
	EncodingFixed int
	TagsUpdated   int
	AutoAlbums    int
	AutoTitles    int
}

// New creates a new Processor with the given options
func New(options ProcessOptions) *Processor {
	return &Processor{
		options: options,
		stats:   Statistics{},
	}
}

// ProcessFiles processes a list of audio files
func (p *Processor) ProcessFiles(files []scanner.AudioFile, command string, threads int) error {
	p.stats.Total = len(files)

	// Create worker pool
	jobs := make(chan scanner.AudioFile, len(files))
	results := make(chan error, len(files))

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for file := range jobs {
				err := p.processFile(file, command)
				results <- err
			}
		}(i)
	}

	// Send jobs
	for _, file := range files {
		jobs <- file
	}
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()
	close(results)

	// Collect results
	for err := range results {
		if err != nil {
			p.mu.Lock()
			p.stats.Failed++
			p.mu.Unlock()
			fmt.Printf("Error: %v\n", err)
		} else {
			p.mu.Lock()
			p.stats.Success++
			p.mu.Unlock()
		}
	}

	// Print statistics
	p.printStatistics()

	return nil
}

// processFile processes a single audio file
func (p *Processor) processFile(file scanner.AudioFile, command string) error {
	switch command {
	case "scan":
		return p.scanFile(file)
	case "fix":
		return p.fixFile(file)
	case "tag":
		return p.tagFile(file)
	case "test":
		return p.testFile(file)
	case "check":
		return p.checkFile(file)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// scanFile scans and displays file tags
func (p *Processor) scanFile(file scanner.AudioFile) error {
	meta, err := tagger.ReadTags(file.Path)
	if err != nil {
		return fmt.Errorf("failed to read tags from %s: %w", file.Path, err)
	}

	fmt.Printf("File: %s\n", file.RelPath)
	fmt.Printf("  Title: %s\n", meta.Title)
	fmt.Printf("  Artist: %s\n", meta.Artist)
	fmt.Printf("  Album: %s\n", meta.Album)
	if meta.Year > 0 {
		fmt.Printf("  Year: %d\n", meta.Year)
	}
	if meta.Genre != "" {
		fmt.Printf("  Genre: %s\n", meta.Genre)
	}
	fmt.Println()

	return nil
}

// checkFile displays current tags only
func (p *Processor) checkFile(file scanner.AudioFile) error {
	return p.scanFile(file)
}

// testFile simulates processing without modifying files
func (p *Processor) testFile(file scanner.AudioFile) error {
	meta, err := tagger.ReadTags(file.Path)
	if err != nil {
		return fmt.Errorf("failed to read tags from %s: %w", file.Path, err)
	}

	// Simulate processing
	newMeta := p.processMetadata(meta, file)

	// Display what would be changed
	fmt.Printf("File: %s\n", file.RelPath)
	fmt.Printf("  Current: Title=%q, Artist=%q, Album=%q\n", meta.Title, meta.Artist, meta.Album)
	fmt.Printf("  New:     Title=%q, Artist=%q, Album=%q\n", newMeta.Title, newMeta.Artist, newMeta.Album)
	fmt.Println()

	return nil
}

// fixFile fixes encoding issues
func (p *Processor) fixFile(file scanner.AudioFile) error {
	meta, err := tagger.ReadTags(file.Path)
	if err != nil {
		return fmt.Errorf("failed to read tags from %s: %w", file.Path, err)
	}

	// Process metadata
	newMeta := p.processMetadata(meta, file)

	// Determine output path
	outPath := file.Path
	if p.options.OutDir != "" {
		outPath = filepath.Join(p.options.OutDir, file.RelPath)
	}

	// Write tags
	data := &writer.TagData{
		Title:  newMeta.Title,
		Artist: newMeta.Artist,
		Album:  newMeta.Album,
		Year:   strconv.Itoa(newMeta.Year),
		Genre:  newMeta.Genre,
	}

	if outPath == file.Path {
		// Update in place
		if err := writer.WriteTagsToFile(outPath, data); err != nil {
			return fmt.Errorf("failed to write tags to %s: %w", outPath, err)
		}
	} else {
		// Write to new file
		if err := writer.WriteTagsToNewFile(file.Path, outPath, data); err != nil {
			return fmt.Errorf("failed to write tags to %s: %w", outPath, err)
		}
	}

	p.mu.Lock()
	p.stats.TagsUpdated++
	p.mu.Unlock()

	return nil
}

// tagFile auto-fills missing metadata tags
func (p *Processor) tagFile(file scanner.AudioFile) error {
	meta, err := tagger.ReadTags(file.Path)
	if err != nil {
		return fmt.Errorf("failed to read tags from %s: %w", file.Path, err)
	}

	// Process metadata
	newMeta := p.processMetadata(meta, file)

	// Determine output path
	outPath := file.Path
	if p.options.OutDir != "" {
		outPath = filepath.Join(p.options.OutDir, file.RelPath)
	}

	// Write tags
	data := &writer.TagData{
		Title:  newMeta.Title,
		Artist: newMeta.Artist,
		Album:  newMeta.Album,
		Year:   strconv.Itoa(newMeta.Year),
		Genre:  newMeta.Genre,
	}

	if outPath == file.Path {
		// Update in place
		if err := writer.WriteTagsToFile(outPath, data); err != nil {
			return fmt.Errorf("failed to write tags to %s: %w", outPath, err)
		}
	} else {
		// Write to new file
		if err := writer.WriteTagsToNewFile(file.Path, outPath, data); err != nil {
			return fmt.Errorf("failed to write tags to %s: %w", outPath, err)
		}
	}

	p.mu.Lock()
	p.stats.TagsUpdated++
	p.mu.Unlock()

	return nil
}

// processMetadata processes metadata according to options
func (p *Processor) processMetadata(meta *tagger.Metadata, file scanner.AudioFile) *tagger.Metadata {
	newMeta := &tagger.Metadata{
		Title:  meta.Title,
		Artist: meta.Artist,
		Album:  meta.Album,
		Year:   meta.Year,
		Genre:  meta.Genre,
		Track:  meta.Track,
	}

	// Fix encoding
	if newMeta.Title != "" {
		fixed, _, changed := encoder.FixEncoding(newMeta.Title)
		if changed {
			newMeta.Title = fixed
			p.mu.Lock()
			p.stats.EncodingFixed++
			p.mu.Unlock()
		}
	}

	if newMeta.Artist != "" {
		fixed, _, changed := encoder.FixEncoding(newMeta.Artist)
		if changed {
			newMeta.Artist = fixed
			p.mu.Lock()
			p.stats.EncodingFixed++
			p.mu.Unlock()
		}
	}

	if newMeta.Album != "" {
		fixed, _, changed := encoder.FixEncoding(newMeta.Album)
		if changed {
			newMeta.Album = fixed
			p.mu.Lock()
			p.stats.EncodingFixed++
			p.mu.Unlock()
		}
	}

	// If UpdateEncoding is true, only fix encoding, don't derive tags
	if p.options.UpdateEncoding {
		return newMeta
	}

	// Auto-derive album from directory name if empty
	if newMeta.Album == "" {
		dirName := filepath.Base(filepath.Dir(file.Path))
		if dirName != "" && dirName != "." {
			newMeta.Album = dirName
			p.mu.Lock()
			p.stats.AutoAlbums++
			p.mu.Unlock()
		}
	}

	// Auto-format title with zero-padding
	if newMeta.Title != "" {
		formatted := formatTitle(newMeta.Title)
		if formatted != newMeta.Title {
			newMeta.Title = formatted
			p.mu.Lock()
			p.stats.AutoTitles++
			p.mu.Unlock()
		}
	}

	// If Force is true, derive tags from filename and directory
	if p.options.Force {
		fileName := filepath.Base(file.Path)
		fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))

		// Extract number from filename (e.g., "01", "1", etc.)
		re := regexp.MustCompile(`^(\d+)`)
		matches := re.FindStringSubmatch(fileName)

		if len(matches) >= 2 {
			number := matches[1]
			// Pad with zero if single digit
			if len(number) == 1 {
				number = "0" + number
			}

			// Get album from directory name
			dirName := filepath.Base(filepath.Dir(file.Path))
			album := dirName

			// Format: Title = Number + Album
			newMeta.Title = number + " " + album
			// Artist = Album
			newMeta.Artist = album
			newMeta.Album = album
		}
	}

	return newMeta
}

// formatTitle formats title with zero-padding (e.g., "1 Title" -> "01 Title")
func formatTitle(title string) string {
	// Match pattern: "number space title"
	re := regexp.MustCompile(`^(\d+)\s+(.+)$`)
	matches := re.FindStringSubmatch(title)

	if len(matches) == 3 {
		number := matches[1]
		rest := matches[2]

		// If number is single digit, pad with zero
		if len(number) == 1 {
			return "0" + number + " " + rest
		}
	}

	return title
}

// extractNumberAndTitle extracts number and title from filename
func extractNumberAndTitle(fileName string) []string {
	re := regexp.MustCompile(`^(\d+)\s+(.+)$`)
	matches := re.FindStringSubmatch(fileName)
	if len(matches) == 3 {
		// Pad number with zero if single digit
		number := matches[1]
		if len(number) == 1 {
			number = "0" + number
		}
		return []string{"", number, matches[2]}
	}
	return nil
}

// printStatistics prints processing statistics
func (p *Processor) printStatistics() {
	fmt.Println("\n---")
	fmt.Println("\nStatistics:")
	fmt.Printf("  Total files: %d\n", p.stats.Total)
	fmt.Printf("  Successfully processed: %d\n", p.stats.Success)
	fmt.Printf("  Failed: %d\n", p.stats.Failed)
	fmt.Printf("  Encoding fixed: %d\n", p.stats.EncodingFixed)
	fmt.Printf("  Tags updated: %d\n", p.stats.TagsUpdated)
	fmt.Printf("  Auto-derived albums: %d\n", p.stats.AutoAlbums)
	fmt.Printf("  Auto-formatted titles: %d\n", p.stats.AutoTitles)
	fmt.Println()
}
