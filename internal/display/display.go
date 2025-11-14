package display

import (
	"fmt"
	"sync"
	"time"
)

type ProgressDisplay struct {
	total     int
	processed int
	mu        sync.Mutex
	startTime time.Time
	workers   map[int]*WorkerStatus
}

type WorkerStatus struct {
	ID      int
	Current string
	Status  string
	mu      sync.Mutex
}

func NewProgressDisplay(total int, numWorkers int) *ProgressDisplay {
	workers := make(map[int]*WorkerStatus)
	for i := 0; i < numWorkers; i++ {
		workers[i] = &WorkerStatus{ID: i}
	}

	return &ProgressDisplay{
		total:     total,
		processed: 0,
		startTime: time.Now(),
		workers:   workers,
	}
}

func (pd *ProgressDisplay) UpdateWorker(workerID int, file string, status string) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	if worker, ok := pd.workers[workerID]; ok {
		worker.mu.Lock()
		worker.Current = file
		worker.Status = status
		worker.mu.Unlock()
	}

	pd.render()
}

func (pd *ProgressDisplay) Increment() {
	pd.mu.Lock()
	pd.processed++
	pd.mu.Unlock()
	pd.render()
}

func (pd *ProgressDisplay) render() {
	fmt.Printf("Processing Audio Files\n")
	fmt.Printf("=====================\n\n")
	fmt.Printf("Scanning directory...\n")
	fmt.Printf("Found %d audio files\n\n", pd.total)

	elapsed := time.Since(pd.startTime)
	percent := float64(pd.processed) / float64(pd.total) * 100
	speed := float64(pd.processed) / elapsed.Seconds()

	fmt.Printf("Progress: %d/%d (%.1f%%) | Speed: %.1f files/s\n\n", 
		pd.processed, pd.total, percent, speed)

	for i := 0; i < len(pd.workers); i++ {
		if worker, ok := pd.workers[i]; ok {
			worker.mu.Lock()
			if worker.Current != "" {
				fmt.Printf("[Worker %d] %s | %s\n", worker.ID+1, worker.Current, worker.Status)
			}
			worker.mu.Unlock()
		}
	}
}

func (pd *ProgressDisplay) Finish() {
	pd.mu.Lock()
	defer pd.mu.Unlock()
	pd.render()
	fmt.Print("\n")
}

func PrintScanResult(file string, metadata *Metadata) {
	fmt.Printf("File: %s\n", file)
	if metadata.Title != "" {
		fmt.Printf("  Title: %s\n", metadata.Title)
	}
	if metadata.Artist != "" {
		fmt.Printf("  Artist: %s\n", metadata.Artist)
	}
	if metadata.Album != "" {
		fmt.Printf("  Album: %s\n", metadata.Album)
	}
	if metadata.Year > 0 {
		fmt.Printf("  Year: %d\n", metadata.Year)
	}
	fmt.Println()
}

func PrintStatistics(stats *Statistics) {
	fmt.Println("Statistics:")
	fmt.Printf("  Total files: %d\n", stats.Total)
	fmt.Printf("  Successfully processed: %d\n", stats.Success)
	fmt.Printf("  Failed: %d\n", stats.Failed)
	fmt.Printf("  Encoding fixed: %d\n", stats.EncodingFixed)
	fmt.Printf("  Tags updated: %d\n", stats.TagsUpdated)
	fmt.Printf("  Auto-derived albums: %d\n", stats.AutoAlbums)
	fmt.Printf("  Auto-formatted titles: %d\n", stats.AutoTitles)
}

type Metadata struct {
	Title  string
	Artist string
	Album  string
	Year   int
}

type Statistics struct {
	Total         int
	Success       int
	Failed        int
	EncodingFixed int
	TagsUpdated   int
	AutoAlbums    int
	AutoTitles    int
}

