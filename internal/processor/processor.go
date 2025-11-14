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
	ForceAll       bool   // Force update all tags (overwrite existing tags)
	UpdateEncoding bool   // Fix encoding only (for tag command)
	OutDir         string // Output directory (empty means update in place)
	Threads        int    // Number of worker threads
}

// Processor handles batch processing of audio files
type Processor struct {
	options      ProcessOptions
	stats        Statistics
	mu           sync.Mutex
	currentIndex int
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
	// Increment current index
	p.mu.Lock()
	p.currentIndex++
	p.mu.Unlock()

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

// getCurrentIndex returns the current processing index
func (p *Processor) getCurrentIndex() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.currentIndex
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
	meta, err := tagger.ReadTags(file.Path)
	if err != nil {
		return fmt.Errorf("failed to read tags from %s: %w", file.Path, err)
	}

	// Print output in same format as fix/tag commands
	fileName := convertPathToUTF8(filepath.Base(file.Path))
	fmt.Printf("[%d/%d] Processing: %s → Title: %q, Artist: %q, Album: %q\n",
		p.getCurrentIndex(), p.stats.Total, fileName, meta.Title, meta.Artist, meta.Album)

	return nil
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

	// Track changes for output
	var changes []string
	encodingFixed := 0
	autoAlbum := false
	autoTitle := false

	// Process metadata and track changes
	newMeta := &tagger.Metadata{
		Title:  meta.Title,
		Artist: meta.Artist,
		Album:  meta.Album,
		Year:   meta.Year,
		Genre:  meta.Genre,
		Track:  meta.Track,
	}

	// Step 1: Fix encoding first (priority)
	if newMeta.Title != "" {
		fixed, charset, changed := encoder.FixEncoding(newMeta.Title)
		if changed {
			changes = append(changes, fmt.Sprintf("Title: %s -> UTF-8", charset))
			newMeta.Title = fixed
			encodingFixed++
		}
		// Check if still garbled after encoding fix
		if encoder.IsGarbled(newMeta.Title) {
			changes = append(changes, "Title: still garbled after encoding fix")
		}
	}

	if newMeta.Artist != "" {
		fixed, charset, changed := encoder.FixEncoding(newMeta.Artist)
		if changed {
			changes = append(changes, fmt.Sprintf("Artist: %s -> UTF-8", charset))
			newMeta.Artist = fixed
			encodingFixed++
		}
		// Check if still garbled after encoding fix
		if encoder.IsGarbled(newMeta.Artist) {
			changes = append(changes, "Artist: still garbled after encoding fix")
		}
	}

	if newMeta.Album != "" {
		fixed, charset, changed := encoder.FixEncoding(newMeta.Album)
		if changed {
			changes = append(changes, fmt.Sprintf("Album: %s -> UTF-8", charset))
			newMeta.Album = fixed
			encodingFixed++
		}
		// Check if still garbled after encoding fix
		if encoder.IsGarbled(newMeta.Album) {
			changes = append(changes, "Album: still garbled after encoding fix")
		}
	}

	// Step 1.5: Clean up domains and file extensions
	if newMeta.Title != "" {
		cleaned := cleanTagText(newMeta.Title)
		if cleaned != newMeta.Title {
			newMeta.Title = cleaned
			changes = append(changes, "Title cleaned (removed domains/extensions)")
		}
		// If cleaned title is empty (e.g., CD default title), treat as empty for fallback
		if newMeta.Title == "" {
			changes = append(changes, "Title is empty after cleaning, will use filename")
		}
	}
	if newMeta.Artist != "" {
		cleaned := cleanTagText(newMeta.Artist)
		if cleaned != newMeta.Artist {
			newMeta.Artist = cleaned
			changes = append(changes, "Artist cleaned (removed domains/extensions)")
		}
	}
	if newMeta.Album != "" {
		cleaned := cleanTagText(newMeta.Album)
		if cleaned != newMeta.Album {
			newMeta.Album = cleaned
			changes = append(changes, "Album cleaned (removed domains/extensions)")
		}
		// If cleaned album is empty (e.g., URL only), treat as empty for fallback
		if newMeta.Album == "" {
			changes = append(changes, "Album is empty after cleaning, will use directory")
		}
	}

	// Step 2: Auto-format title with zero-padding
	if newMeta.Title != "" {
		formatted := formatTitle(newMeta.Title)
		if formatted != newMeta.Title {
			newMeta.Title = formatted
			autoTitle = true
			changes = append(changes, "Title zero-padded")
		}
	}

	// Step 3: Fill from filename/directory if empty or garbled (fallback)
	// Always use fallback if field is empty or garbled (even if UpdateEncoding is true)
	fileName := convertPathToUTF8(filepath.Base(file.Path))
	fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	dirName := convertPathToUTF8(filepath.Base(filepath.Dir(file.Path)))

	// Fill Title: empty or garbled (Force allows overwrite even if not garbled)
	shouldFillTitle := newMeta.Title == "" || encoder.IsGarbled(newMeta.Title) || (p.options.Force && p.options.ForceAll)
	if shouldFillTitle && fileName != "" {
		formattedTitle := formatTitleFromFilename(fileName)
		newMeta.Title = formattedTitle
		autoTitle = true
		changes = append(changes, fmt.Sprintf("Title=%q (from filename, fallback)", formattedTitle))
	}

	// Fill Album: empty or garbled (Force allows overwrite even if not garbled)
	shouldFillAlbum := newMeta.Album == "" || encoder.IsGarbled(newMeta.Album) || (p.options.Force && p.options.ForceAll)
	if shouldFillAlbum && dirName != "" && dirName != "." {
		newMeta.Album = dirName
		autoAlbum = true
		changes = append(changes, fmt.Sprintf("Album=%q (from directory, fallback)", dirName))
	}

	// Fill Artist: empty or garbled (Force allows overwrite even if not garbled)
	shouldFillArtist := newMeta.Artist == "" || encoder.IsGarbled(newMeta.Artist) || (p.options.Force && p.options.ForceAll)
	if shouldFillArtist && dirName != "" && dirName != "." {
		// Extract artist from directory name (before underscore)
		if strings.Contains(dirName, "_") {
			parts := strings.SplitN(dirName, "_", 2)
			if len(parts) >= 1 && parts[0] != "" {
				newMeta.Artist = parts[0]
				changes = append(changes, fmt.Sprintf("Artist=%q (from directory, fallback)", parts[0]))
			}
		} else {
			newMeta.Artist = dirName
			changes = append(changes, fmt.Sprintf("Artist=%q (from directory, fallback)", dirName))
		}
	}

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

	// Update statistics
	p.mu.Lock()
	p.stats.TagsUpdated++
	p.stats.EncodingFixed += encodingFixed
	if autoAlbum {
		p.stats.AutoAlbums++
	}
	if autoTitle {
		p.stats.AutoTitles++
	}
	p.mu.Unlock()

	// Print output
	fileNameForDisplay := convertPathToUTF8(filepath.Base(file.Path))
	fmt.Printf("[%d/%d] Processing: %s → Title: %q, Artist: %q, Album: %q\n",
		p.getCurrentIndex(), p.stats.Total, fileNameForDisplay, newMeta.Title, newMeta.Artist, newMeta.Album)

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

	// Before writing, check if any field is garbled and fill from filename/directory if needed
	fileNameForCheck := convertPathToUTF8(filepath.Base(file.Path))
	fileNameForCheck = strings.TrimSuffix(fileNameForCheck, filepath.Ext(fileNameForCheck))
	dirNameForCheck := convertPathToUTF8(filepath.Base(filepath.Dir(file.Path)))

	// Check and fix Title before writing
	if encoder.IsGarbled(newMeta.Title) && fileNameForCheck != "" {
		formattedTitle := formatTitleFromFilename(fileNameForCheck)
		newMeta.Title = formattedTitle
		p.mu.Lock()
		p.stats.AutoTitles++
		p.mu.Unlock()
	}

	// Check and fix Album before writing
	if encoder.IsGarbled(newMeta.Album) && dirNameForCheck != "" && dirNameForCheck != "." {
		newMeta.Album = dirNameForCheck
		p.mu.Lock()
		p.stats.AutoAlbums++
		p.mu.Unlock()
	}

	// Check and fix Artist before writing
	if encoder.IsGarbled(newMeta.Artist) && dirNameForCheck != "" && dirNameForCheck != "." {
		if strings.Contains(dirNameForCheck, "_") {
			parts := strings.SplitN(dirNameForCheck, "_", 2)
			if len(parts) >= 1 && parts[0] != "" {
				newMeta.Artist = parts[0]
			}
		} else {
			newMeta.Artist = dirNameForCheck
		}
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

	// Print output
	fileNameForDisplay := convertPathToUTF8(filepath.Base(file.Path))
	fmt.Printf("[%d/%d] Processing: %s → Title: %q, Artist: %q, Album: %q\n",
		p.getCurrentIndex(), p.stats.Total, fileNameForDisplay, newMeta.Title, newMeta.Artist, newMeta.Album)

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

	// Step 1: Fix encoding first (priority)
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

	// Step 1.5: Clean up domains and file extensions
	if newMeta.Title != "" {
		cleaned := cleanTagText(newMeta.Title)
		if cleaned != newMeta.Title {
			newMeta.Title = cleaned
		}
		// If cleaned title is empty (e.g., CD default title), treat as empty for fallback
	}
	if newMeta.Artist != "" {
		cleaned := cleanTagText(newMeta.Artist)
		if cleaned != newMeta.Artist {
			newMeta.Artist = cleaned
		}
	}
	if newMeta.Album != "" {
		cleaned := cleanTagText(newMeta.Album)
		if cleaned != newMeta.Album {
			newMeta.Album = cleaned
		}
		// If cleaned album is empty (e.g., URL only), treat as empty for fallback
	}

	// Step 2: Auto-format title with zero-padding
	if newMeta.Title != "" {
		formatted := formatTitle(newMeta.Title)
		if formatted != newMeta.Title {
			newMeta.Title = formatted
			p.mu.Lock()
			p.stats.AutoTitles++
			p.mu.Unlock()
		}
	}

	// Step 3: Fill from filename/directory if empty or garbled (fallback)
	// Always use fallback if field is empty or garbled (even if UpdateEncoding is true)
	fileNameForFallback := convertPathToUTF8(filepath.Base(file.Path))
	fileNameForFallback = strings.TrimSuffix(fileNameForFallback, filepath.Ext(fileNameForFallback))
	dirNameForFallback := convertPathToUTF8(filepath.Base(filepath.Dir(file.Path)))

	// Fill Title: empty or garbled (Force allows overwrite even if not garbled)
	shouldFillTitle := newMeta.Title == "" || encoder.IsGarbled(newMeta.Title) || (p.options.Force && p.options.ForceAll)
	if shouldFillTitle && fileNameForFallback != "" {
		formattedTitle := formatTitleFromFilename(fileNameForFallback)
		newMeta.Title = formattedTitle
		p.mu.Lock()
		p.stats.AutoTitles++
		p.mu.Unlock()
	}

	// Fill Album: empty or garbled (Force allows overwrite even if not garbled)
	shouldFillAlbum := newMeta.Album == "" || encoder.IsGarbled(newMeta.Album) || (p.options.Force && p.options.ForceAll)
	if shouldFillAlbum && dirNameForFallback != "" && dirNameForFallback != "." {
		newMeta.Album = dirNameForFallback
		p.mu.Lock()
		p.stats.AutoAlbums++
		p.mu.Unlock()
	}

	// Fill Artist: empty or garbled (Force allows overwrite even if not garbled)
	shouldFillArtist := newMeta.Artist == "" || encoder.IsGarbled(newMeta.Artist) || (p.options.Force && p.options.ForceAll)
	if shouldFillArtist && dirNameForFallback != "" && dirNameForFallback != "." {
		// Extract artist from directory name (before underscore)
		if strings.Contains(dirNameForFallback, "_") {
			parts := strings.SplitN(dirNameForFallback, "_", 2)
			if len(parts) >= 1 && parts[0] != "" {
				newMeta.Artist = parts[0]
			}
		} else {
			newMeta.Artist = dirNameForFallback
		}
	}

	// If UpdateEncoding is true and Force is false, only fix encoding, don't derive tags
	// But we already did fallback above, so this is just for early return
	if p.options.UpdateEncoding && !p.options.Force {
		// Already processed fallback above, so just return
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

// formatTitleFromFilename extracts number from end of filename and moves it to front with zero-padding
// Example: "康熙大帝（第二卷）35" -> "35 康熙大帝（第二卷）"
// Example: "康熙大帝5" -> "05 康熙大帝"
// Example: "002" -> "002" (pure number, return as is)
func formatTitleFromFilename(fileName string) string {
	// Check if filename is pure number
	if matched, _ := regexp.MatchString(`^\d+$`, fileName); matched {
		// Pure number, pad with zero if single digit
		if len(fileName) == 1 {
			return "0" + fileName
		}
		return fileName
	}

	// Match pattern: extract trailing digits from filename
	// Pattern: "text数字" or "text 数字"
	re := regexp.MustCompile(`^(.+?)(\d+)$`)
	matches := re.FindStringSubmatch(fileName)

	if len(matches) == 3 {
		text := matches[1]
		number := matches[2]

		// Remove trailing spaces from text
		text = strings.TrimRight(text, " ")

		// If text is empty after trimming, just return the number
		if text == "" {
			if len(number) == 1 {
				return "0" + number
			}
			return number
		}

		// Pad number with zero if single digit
		if len(number) == 1 {
			number = "0" + number
		}

		return number + " " + text
	}

	// If no number found at end, return original filename
	return fileName
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

// convertPathToUTF8 converts a file path component to UTF-8 encoding
func convertPathToUTF8(path string) string {
	if path == "" {
		return path
	}
	// Try to convert encoding, if it fails, return original
	utf8Path, _, err := encoder.ConvertStringToUTF8(path)
	if err != nil {
		return path
	}
	return utf8Path
}

// cleanTagText removes domains, file extensions, and other unwanted content from tag text
func cleanTagText(text string) string {
	if text == "" {
		return text
	}

	cleaned := text

	// Remove common CD default titles first
	// Match patterns like "CD Digital Audio, Track#30", "CD Digital Audio Track 30", etc.
	cdPattern := regexp.MustCompile(`(?i)^CD\s+Digital\s+Audio\s*,?\s*Track#?\s*\d+.*$`)
	if cdPattern.MatchString(cleaned) {
		return ""
	}
	// Also match variations like "CD Digital Audio, Track 30", "CDDA Track#30", "CD Track 30"
	cdPattern2 := regexp.MustCompile(`(?i)^CD\s*(Digital\s+Audio|DA)\s*,?\s*Track#?\s*\d+.*$`)
	if cdPattern2.MatchString(cleaned) {
		return ""
	}
	// Match simple patterns like "CD Track 30", "Track 30", "Track#30" (if starts with these)
	cdPattern3 := regexp.MustCompile(`(?i)^(CD\s*)?Track#?\s*\d+.*$`)
	if cdPattern3.MatchString(cleaned) && len(cleaned) < 50 { // Only if short, to avoid false positives
		return ""
	}

	// Remove URLs (http://, https://, www.)
	urlPattern := regexp.MustCompile(`(?i)(https?://[^\s]+|www\.[^\s]+)`)
	cleaned = urlPattern.ReplaceAllString(cleaned, "")

	// Remove domain patterns in brackets like [bbs.bbxpp.cn], [www.example.com]
	// Match [anything.domain] where domain is a common TLD
	domainPattern := regexp.MustCompile(`\[[^\]]*\.(com|cn|net|org|edu|gov|io|co|uk|de|fr|jp|ru|au|ca|br|in|it|es|nl|se|no|dk|fi|pl|cz|hu|gr|pt|ie|at|ch|be|tr|kr|tw|hk|sg|my|th|vn|id|ph|nz|za|mx|ar|cl|pe|eg|sa|ae|il|pk|bd|lk|np|mm|kh|la|mn|kz|uz|az|ge|am|by|ua|md|ro|bg|rs|hr|si|sk|lt|lv|ee|is|mt|cy|lu|mc|ad|li|sm|va|me|ba|mk|al|xk)[^\]]*\]`)
	cleaned = domainPattern.ReplaceAllString(cleaned, "")

	// Remove file extensions (.MP3, .mp3, .WAV, .wav, etc.)
	// Match .ext followed by space, end of string, or another dot
	extPattern := regexp.MustCompile(`(?i)\.(mp3|wav|flac|m4a|aac|ogg|wma|ape|wv|tta|tak|ofr|ofs|off|rka|shn|aa3|gsm|3gp|amr|awb|au|snd|ra|rm|ram|dct|vox|sln)(\s|$|\.)`)
	cleaned = extPattern.ReplaceAllString(cleaned, "")

	// Remove trailing/leading separators and spaces
	cleaned = strings.Trim(cleaned, " \t\n\r")

	// Remove trailing dashes and separators (but keep content before them)
	cleaned = strings.TrimRight(cleaned, "---")

	// Remove multiple consecutive spaces
	spacePattern := regexp.MustCompile(`\s+`)
	cleaned = spacePattern.ReplaceAllString(cleaned, " ")

	// Final trim
	cleaned = strings.TrimSpace(cleaned)

	// If result is empty or only contains separators, return empty
	if cleaned == "" || cleaned == "---" || cleaned == "[]" {
		return ""
	}

	return cleaned
}
