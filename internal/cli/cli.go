package cli

import (
	"fmt"
	"os"

	"mp3tools/internal/processor"
	"mp3tools/internal/scanner"

	"github.com/spf13/cobra"
)

var (
	force    bool
	forceAll bool
	threads  int
	outdir   string
	update   bool
)

var rootCmd = &cobra.Command{
	Use:   "mp3tools",
	Short: "MP3 Tools",
	Long: `Usage:
  mp3tools <command> [path] [options]

Commands:
  scan <path>    Scan directory and display audio file tags
  fix <path>     Fix encoding issues in audio file tags
  tag <path>     Auto-fill missing metadata tags
  test <path>    Preview changes with parameters (simulation only, no file modification)
  check <path>   Display current tags (display only, no parameters)

Options:
  -f, --force    Derive tags from filename and directory name (for tag command)
  -a, --all      Force update all tags (overwrite existing tags)
  -n, --threads  Number of worker threads (default: 5)
  -u, --update   Fix encoding only (for tag command, default: true) or update original files (for other commands)
  -o, --outdir   Output directory, preserve directory structure (default: update original files)

Examples:
  mp3tools scan ./music
  mp3tools fix ./music -u
  mp3tools tag ./music -f
  mp3tools check ./music -u`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan directory and display audio file tags",
	Args:  cobra.ExactArgs(1),
	Run:   runScan,
}

var fixCmd = &cobra.Command{
	Use:   "fix [path]",
	Short: "Fix encoding issues in audio file tags",
	Args:  cobra.ExactArgs(1),
	Run:   runFix,
}

var tagCmd = &cobra.Command{
	Use:   "tag [path]",
	Short: "Auto-fill missing metadata tags",
	Args:  cobra.ExactArgs(1),
	Run:   runTag,
}

var testCmd = &cobra.Command{
	Use:   "test [path]",
	Short: "Preview changes (simulation only)",
	Args:  cobra.ExactArgs(1),
	Run:   runTest,
}

var checkCmd = &cobra.Command{
	Use:   "check [path]",
	Short: "Verify and apply changes",
	Args:  cobra.ExactArgs(1),
	Run:   runCheck,
}

func init() {
	rootCmd.AddCommand(scanCmd, fixCmd, tagCmd, testCmd, checkCmd)

	// Custom help template to remove duplicate sections
	rootCmd.SetHelpTemplate(`{{.Long}}`)
	rootCmd.SetUsageTemplate(`{{.Long}}`)

	scanCmd.Flags().BoolVarP(&force, "force", "f", false, "Force overwrite existing tags")
	scanCmd.Flags().IntVarP(&threads, "threads", "n", 5, "Number of worker threads")
	scanCmd.Flags().StringVarP(&outdir, "outdir", "o", "", "Output directory, preserve directory structure (default: update original files)")
	scanCmd.Flags().BoolVarP(&update, "update", "u", false, "Update original MP3 files (overwrite)")

	fixCmd.Flags().BoolVarP(&force, "force", "f", false, "Derive tags from filename and directory name")
	fixCmd.Flags().BoolVarP(&forceAll, "all", "a", false, "Force update all tags (overwrite existing tags)")
	fixCmd.Flags().IntVarP(&threads, "threads", "n", 5, "Number of worker threads")
	fixCmd.Flags().StringVarP(&outdir, "outdir", "o", "output", "Output directory, preserve directory structure (default: output)")
	fixCmd.Flags().BoolVarP(&update, "update", "u", false, "Update original MP3 files (overwrite)")

	tagCmd.Flags().BoolVarP(&force, "force", "f", false, "Derive tags from filename and directory name")
	tagCmd.Flags().BoolVarP(&forceAll, "all", "a", false, "Force update all tags (overwrite existing tags)")
	tagCmd.Flags().IntVarP(&threads, "threads", "n", 5, "Number of worker threads")
	tagCmd.Flags().StringVarP(&outdir, "outdir", "o", "output", "Output directory, preserve directory structure (default: output)")
	tagCmd.Flags().BoolVarP(&update, "update", "u", true, "Fix encoding only (default: true)")

	testCmd.Flags().BoolVarP(&force, "force", "f", false, "Derive tags from filename and directory name")
	testCmd.Flags().BoolVarP(&forceAll, "all", "a", false, "Force update all tags (overwrite existing tags)")
	testCmd.Flags().IntVarP(&threads, "threads", "n", 5, "Number of worker threads")
	testCmd.Flags().BoolVarP(&update, "update", "u", true, "Fix encoding only (default: true)")

	// check command has no flags - display only
}

func Execute() error {
	return rootCmd.Execute()
}

func runScan(cmd *cobra.Command, args []string) {
	path := args[0]
	files, err := scanner.ScanDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No audio files found")
		return
	}

	// Default: update original files (unless -o is specified)
	outputDir := outdir
	if update {
		outputDir = ""
	}

	proc := processor.New(processor.ProcessOptions{
		Force:    force,
		ForceAll: forceAll,
		OutDir:   outputDir,
		Threads:  threads,
	})

	if err := proc.ProcessFiles(files, "scan", threads); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing files: %v\n", err)
		os.Exit(1)
	}
}

func runFix(cmd *cobra.Command, args []string) {
	path := args[0]
	files, err := scanner.ScanDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No audio files found")
		return
	}

	// Default: update original files (unless -o is specified)
	outputDir := outdir
	if update {
		outputDir = ""
	}

	proc := processor.New(processor.ProcessOptions{
		Force:          force,
		ForceAll:       forceAll,
		UpdateEncoding: false,
		OutDir:         outputDir,
		Threads:        threads,
	})

	if err := proc.ProcessFiles(files, "fix", threads); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing files: %v\n", err)
		os.Exit(1)
	}
}

func runTag(cmd *cobra.Command, args []string) {
	path := args[0]
	files, err := scanner.ScanDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No audio files found")
		return
	}

	// Default: update original files (unless -o is specified)
	outputDir := outdir
	if outdir == "" {
		outputDir = ""
	}

	proc := processor.New(processor.ProcessOptions{
		Force:          force,
		ForceAll:       forceAll,
		UpdateEncoding: update,
		OutDir:         outputDir,
		Threads:        threads,
	})

	if err := proc.ProcessFiles(files, "tag", threads); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing files: %v\n", err)
		os.Exit(1)
	}
}

func runTest(cmd *cobra.Command, args []string) {
	path := args[0]
	files, err := scanner.ScanDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No audio files found")
		return
	}

	fmt.Printf("Preview Mode - No changes will be made\n")
	fmt.Printf("Scanning directory: %s\n", path)
	fmt.Printf("Found %d audio files\n\n", len(files))

	proc := processor.New(processor.ProcessOptions{
		Force:          force,
		UpdateEncoding: update,
		OutDir:         "",
		Threads:        threads,
	})

	if err := proc.ProcessFiles(files, "test", threads); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing files: %v\n", err)
		os.Exit(1)
	}
}

func runCheck(cmd *cobra.Command, args []string) {
	path := args[0]
	files, err := scanner.ScanDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No audio files found")
		return
	}

	// check command: display only, no parameters
	proc := processor.New(processor.ProcessOptions{
		Force:          false,
		UpdateEncoding: false,
		OutDir:         "",
		Threads:        5,
	})

	if err := proc.ProcessFiles(files, "check", 5); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing files: %v\n", err)
		os.Exit(1)
	}
}
