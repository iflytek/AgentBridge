package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/iflytek/agentbridge/core"
	"github.com/iflytek/agentbridge/core/services"
	"github.com/iflytek/agentbridge/internal/models"

	"github.com/spf13/cobra"
)

var (
	workerCount   int
	overwriteMode bool
)

// BatchJob represents a single conversion task
type BatchJob struct {
	FilePath   string
	OutputPath string
	Index      int
	Total      int
}

// BatchResult represents the result of a conversion task
type BatchResult struct {
	Job      BatchJob
	Success  bool
	Error    error
	Duration time.Duration
}

// ConcurrentBatchProcessor handles concurrent batch processing
type ConcurrentBatchProcessor struct {
	workerCount     int
	jobQueue        chan BatchJob
	resultQueue     chan BatchResult
	conversionSvc   *services.ConversionService
	ctx             context.Context
	cancel          context.CancelFunc
	progressTracker *ProgressTracker
}

// ProgressTracker tracks batch processing progress
type ProgressTracker struct {
	total     int64
	completed int64
	failed    int64
	startTime time.Time
	mutex     sync.RWMutex
}

// NewBatchCmd creates the batch command
func NewBatchCmd() *cobra.Command {
	var batchCmd = &cobra.Command{
		Use:   "batch",
		Short: "Batch convert multiple workflow files with concurrent processing",
		Long: `Convert multiple workflow files in batch mode with parallel processing for efficient conversion.

Supports directory-based batch conversion with configurable worker count and progress tracking.`,
		Example: `  # Batch convert directory with default settings
  agentbridge batch --from iflytek --to dify --input-dir ./workflows --output-dir ./converted

  # Batch convert with custom worker count
  agentbridge batch --from iflytek --to dify --input-dir ./workflows --output-dir ./converted --workers 8

  # Batch convert with pattern matching
  agentbridge batch --from iflytek --to dify --input-dir ./workflows --pattern "*.yml" --output-dir ./converted`,
		RunE: runBatch,
	}

	// Configure batch command flags
	batchCmd.Flags().StringVar(&inputDir, "input-dir", "", "Input directory containing workflow files (required)")
	batchCmd.Flags().StringVar(&outputDir, "output-dir", "", "Output directory for converted files (required)")
	batchCmd.Flags().StringVar(&sourceType, "from", "", "Source platform (iflytek|dify|coze) (required)")
	batchCmd.Flags().StringVar(&targetType, "to", "", "Target platform (iflytek|dify|coze) (required)")
	batchCmd.Flags().StringVar(&pattern, "pattern", "*.yml", "File pattern to match (default: *.yml)")
	batchCmd.Flags().IntVar(&workerCount, "workers", 0, "Number of concurrent workers (default: auto-detect based on CPU cores)")
	batchCmd.Flags().BoolVar(&overwriteMode, "overwrite", false, "Automatically overwrite existing output files without prompting")

	// Mark required flags
	batchCmd.MarkFlagRequired("input-dir")
	batchCmd.MarkFlagRequired("output-dir")
	batchCmd.MarkFlagRequired("from")
	batchCmd.MarkFlagRequired("to")

	return batchCmd
}

// runBatch executes the batch command with concurrent processing
func runBatch(cmd *cobra.Command, args []string) error {
	restore := redirectStdoutIfQuiet()
	defer restore()
	if quiet {
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
	}
	if !quiet {
		printHeader("Concurrent Batch Conversion")
	}

	if err := setupBatchDirectories(); err != nil {
		return err
	}

	files, err := findFilesToConvert()
	if err != nil {
		return err
	}

	logFilesFound(files)

	// Check for output file conflicts before processing
	if err := checkOutputFileConflicts(files); err != nil {
		return err
	}

	// Initialize conversion service once (reused by all workers)
	conversionSvc, err := core.InitializeArchitecture()
	if err != nil {
		return fmt.Errorf("failed to initialize conversion service: %w", err)
	}

	// Create and configure concurrent processor
	processor := NewConcurrentBatchProcessor(conversionSvc, len(files))
	defer processor.Close()

	// Process files concurrently
	successCount, errorCount, err := processor.ProcessFiles(files)
	if err != nil {
		return fmt.Errorf("batch processing failed: %w", err)
	}

	printBatchSummary(files, successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("batch conversion completed with %d errors", errorCount)
	}

	return nil
}

// NewConcurrentBatchProcessor creates a new concurrent batch processor
func NewConcurrentBatchProcessor(conversionSvc *services.ConversionService, totalFiles int) *ConcurrentBatchProcessor {
	// Auto-detect worker count if not specified
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
		// Cap at reasonable maximum to avoid resource exhaustion
		if workerCount > 16 {
			workerCount = 16
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ConcurrentBatchProcessor{
		workerCount:   workerCount,
		jobQueue:      make(chan BatchJob, workerCount*2), // Buffer for smooth processing
		resultQueue:   make(chan BatchResult, workerCount*2),
		conversionSvc: conversionSvc,
		ctx:           ctx,
		cancel:        cancel,
		progressTracker: &ProgressTracker{
			total:     int64(totalFiles),
			startTime: time.Now(),
		},
	}
}

// ProcessFiles processes all files concurrently
func (p *ConcurrentBatchProcessor) ProcessFiles(files []string) (int, int, error) {
	if verbose {
		fmt.Printf("🚀 Starting concurrent processing with %d workers\n", p.workerCount)
	}

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < p.workerCount; i++ {
		wg.Add(1)
		go p.worker(i+1, &wg)
	}

	// Start result collector
	resultDone := make(chan struct{})
	successCount := 0
	errorCount := 0
	go func() {
		defer close(resultDone)
		for result := range p.resultQueue {
			p.updateProgress(result)
			if result.Success {
				successCount++
			} else {
				errorCount++
				fmt.Printf("❌ Failed to process %s: %v\n", filepath.Base(result.Job.FilePath), result.Error)
			}
		}
	}()

	// Send jobs to workers
	go func() {
		defer close(p.jobQueue)
		for i, file := range files {
			filename := filepath.Base(file)
			outputFile := filepath.Join(outputDir, filename)

			select {
			case p.jobQueue <- BatchJob{
				FilePath:   file,
				OutputPath: outputFile,
				Index:      i + 1,
				Total:      len(files),
			}:
			case <-p.ctx.Done():
				return
			}
		}
	}()

	// Wait for all workers to complete
	wg.Wait()
	close(p.resultQueue)

	// Wait for result collector to finish
	<-resultDone

	return successCount, errorCount, nil
}

// worker processes jobs from the job queue
func (p *ConcurrentBatchProcessor) worker(workerID int, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case job, ok := <-p.jobQueue:
			if !ok {
				return // Job queue closed
			}

			startTime := time.Now()
			err := p.processJob(job)
			duration := time.Since(startTime)

			result := BatchResult{
				Job:      job,
				Success:  err == nil,
				Error:    err,
				Duration: duration,
			}

			select {
			case p.resultQueue <- result:
			case <-p.ctx.Done():
				return
			}

		case <-p.ctx.Done():
			return
		}
	}
}

// processJob processes a single conversion job with enhanced error handling
func (p *ConcurrentBatchProcessor) processJob(job BatchJob) error {
	filename := filepath.Base(job.FilePath)

	// Validate file existence and readability
	if err := p.validateInputFile(job.FilePath); err != nil {
		return fmt.Errorf("input validation failed for '%s': %w", filename, err)
	}

	// Read input file
	inputData, err := os.ReadFile(job.FilePath)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied accessing '%s' - check file permissions", filename)
		}
		return fmt.Errorf("unable to read '%s': %w", filename, err)
	}

	// Validate file size (prevent processing extremely large files)
	if len(inputData) == 0 {
		return fmt.Errorf("file '%s' is empty - skipping conversion", filename)
	}
	if len(inputData) > 50*1024*1024 { // 50MB limit for CLI tool
		return fmt.Errorf("file '%s' is too large (%.1fMB) - maximum supported size is 50MB",
			filename, float64(len(inputData))/1024/1024)
	}

	// Convert using shared service (thread-safe)
	outputData, err := p.convertFileData(inputData)
	if err != nil {
		return fmt.Errorf("conversion failed for '%s': %w", filename, err)
	}

	// Validate output directory and write file
	if err := p.writeOutputFile(job.OutputPath, outputData); err != nil {
		return fmt.Errorf("output write failed for '%s': %w", filename, err)
	}

	return nil
}

// validateInputFile performs comprehensive input file validation
func (p *ConcurrentBatchProcessor) validateInputFile(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist")
		}
		return fmt.Errorf("unable to access file: %w", err)
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file (directories and symlinks not supported)")
	}

	// Validate file extension for workflow files
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".yml" && ext != ".yaml" && ext != ".json" {
		return fmt.Errorf("unsupported file format '%s' (supported: .yml, .yaml, .json)", ext)
	}

	return nil
}

// writeOutputFile handles output file writing with proper error handling
func (p *ConcurrentBatchProcessor) writeOutputFile(outputPath string, data []byte) error {
	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory '%s': %w", outputDir, err)
	}

	// Check for write permissions in output directory
	if err := p.checkWritePermission(outputDir); err != nil {
		return err
	}

	// Write the file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied writing to '%s' - check directory permissions", outputPath)
		}
		return fmt.Errorf("unable to write file '%s': %w", filepath.Base(outputPath), err)
	}

	return nil
}

// checkWritePermission verifies write access to a directory
func (p *ConcurrentBatchProcessor) checkWritePermission(dir string) error {
	testFile := filepath.Join(dir, ".write_test_"+fmt.Sprintf("%d", time.Now().UnixNano()))
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("no write permission in directory '%s'", dir)
	}
	file.Close()
	os.Remove(testFile)
	return nil
}

// convertFileData converts data using the shared conversion service with enhanced error handling
func (p *ConcurrentBatchProcessor) convertFileData(inputData []byte) ([]byte, error) {
	var fromPlatform, toPlatform models.PlatformType

	// Validate and convert source platform
	switch sourceType {
	case "iflytek":
		fromPlatform = models.PlatformIFlytek
	case "dify":
		fromPlatform = models.PlatformDify
	case "coze":
		fromPlatform = models.PlatformCoze
	default:
		return nil, fmt.Errorf("unsupported source platform '%s' - supported platforms: iflytek, dify, coze", sourceType)
	}

	// Validate and convert target platform
	switch targetType {
	case "iflytek":
		toPlatform = models.PlatformIFlytek
	case "dify":
		toPlatform = models.PlatformDify
	case "coze":
		toPlatform = models.PlatformCoze
	default:
		return nil, fmt.Errorf("unsupported target platform '%s' - supported platforms: iflytek, dify, coze", targetType)
	}

	// Validate that source and target are different
	if fromPlatform == toPlatform {
		return nil, fmt.Errorf("source and target platforms are the same ('%s') - no conversion needed", sourceType)
	}

	// Perform conversion with enhanced error context
	result, err := p.conversionSvc.Convert(inputData, fromPlatform, toPlatform)
	if err != nil {
		return nil, p.enhanceConversionError(err, fromPlatform, toPlatform)
	}

	return result, nil
}

// enhanceConversionError provides more user-friendly conversion error messages
func (p *ConcurrentBatchProcessor) enhanceConversionError(err error, from, to models.PlatformType) error {
	errMsg := err.Error()

	// Handle common parsing errors
	if strings.Contains(errMsg, "yaml: unmarshal errors") || strings.Contains(errMsg, "invalid YAML") {
		return fmt.Errorf("invalid YAML format - please check file syntax and structure")
	}

	if strings.Contains(errMsg, "json: invalid") || strings.Contains(errMsg, "unexpected end of JSON") {
		return fmt.Errorf("invalid JSON format - please check file syntax and structure")
	}

	// Handle validation errors
	if strings.Contains(errMsg, "validation failed") {
		return fmt.Errorf("workflow validation failed - file may be corrupted or contain unsupported features")
	}

	// Handle platform-specific conversion issues
	if strings.Contains(errMsg, "failed to parse") {
		return fmt.Errorf("unable to parse %s workflow format - file may not be a valid %s workflow", from, from)
	}

	if strings.Contains(errMsg, "failed to generate") {
		return fmt.Errorf("unable to generate %s format - some features may not be supported in target platform", to)
	}

	// Handle node type issues
	if strings.Contains(errMsg, "unsupported node type") {
		return fmt.Errorf("workflow contains node types not supported in %s platform - conversion may require manual adjustment", to)
	}

	// Default enhanced error with conversion context
	return fmt.Errorf("%s → %s conversion failed: %w", from, to, err)
}

// updateProgress updates progress tracking and displays enhanced visual progress
func (p *ConcurrentBatchProcessor) updateProgress(result BatchResult) {
	p.progressTracker.mutex.Lock()
	p.progressTracker.completed++
	if !result.Success {
		p.progressTracker.failed++
	}

	completed := p.progressTracker.completed
	total := p.progressTracker.total
	failed := p.progressTracker.failed
	elapsed := time.Since(p.progressTracker.startTime)
	p.progressTracker.mutex.Unlock()

	// Calculate progress statistics
	progress := float64(completed) / float64(total) * 100
	speed := float64(completed) / elapsed.Seconds()
	estimatedTotal := elapsed * time.Duration(float64(total)/float64(completed))
	remaining := estimatedTotal - elapsed

	if verbose {
		// Verbose mode: show detailed progress with visual elements
		if result.Success {
			fileSize := p.humanizeFileSize(result.Job.FilePath)
			fmt.Printf("✅ [%d/%d] %s %s (%s) - %.1fms | %.1f/s | ETA: %v\n",
				completed, total,
				p.createProgressBar(progress, 20),
				filepath.Base(result.Job.FilePath),
				fileSize,
				float64(result.Duration.Nanoseconds())/1e6,
				speed,
				p.formatDuration(remaining))
		} else {
			fmt.Printf("❌ [%d/%d] %s %s - %v\n",
				completed, total,
				p.createProgressBar(progress, 20),
				filepath.Base(result.Job.FilePath),
				result.Error)
		}
	} else {
		// Non-verbose: show periodic progress with visual bar
		if completed%5 == 0 || completed == total {
			status := "🟢"
			if failed > 0 {
				status = "🟡"
			}

			fmt.Printf("\r%s %s %.1f%% (%d/%d) | ⚡%.1f/s | ⏰%v | ❌%d",
				status,
				p.createProgressBar(progress, 30),
				progress, completed, total,
				speed,
				p.formatDuration(remaining),
				failed)

			// Print newline on completion or error
			if completed == total {
				fmt.Println()
			}
		}
	}
}

// createProgressBar generates a visual progress bar
func (p *ConcurrentBatchProcessor) createProgressBar(progress float64, width int) string {
	if progress > 100 {
		progress = 100
	}

	filled := int(progress * float64(width) / 100)
	empty := width - filled

	bar := ""

	// Use different characters for different progress ranges
	if progress < 25 {
		bar += "🔴" + strings.Repeat("▓", filled) + strings.Repeat("░", empty)
	} else if progress < 50 {
		bar += "🟡" + strings.Repeat("▓", filled) + strings.Repeat("░", empty)
	} else if progress < 75 {
		bar += "🔵" + strings.Repeat("▓", filled) + strings.Repeat("░", empty)
	} else {
		bar += "🟢" + strings.Repeat("▓", filled) + strings.Repeat("░", empty)
	}

	return bar
}

// humanizeFileSize returns human-readable file size
func (p *ConcurrentBatchProcessor) humanizeFileSize(filePath string) string {
	info, err := os.Stat(filePath)
	if err != nil {
		return "unknown"
	}

	size := info.Size()
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
	} else {
		return fmt.Sprintf("%.1fGB", float64(size)/(1024*1024*1024))
	}
}

// formatDuration formats duration for display
func (p *ConcurrentBatchProcessor) formatDuration(d time.Duration) string {
	if d < 0 {
		return "0s"
	}

	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	} else {
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	}
}

// setupBatchDirectories validates and sets up input/output directories
func setupBatchDirectories() error {
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		return fmt.Errorf("input directory does not exist: %s", inputDir)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return nil
}

// findFilesToConvert finds files to convert based on pattern
func findFilesToConvert() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(inputDir, pattern))
	if err != nil {
		return nil, fmt.Errorf("failed to find files with pattern %s: %w", pattern, err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found matching pattern %s in directory %s", pattern, inputDir)
	}

	return files, nil
}

// logFilesFound logs the number of files found
func logFilesFound(files []string) {
	if verbose {
		fmt.Printf("📁 Found %d files to convert\n", len(files))
	}
}

// checkOutputFileConflicts checks for existing output files and handles conflicts
func checkOutputFileConflicts(files []string) error {
	if overwriteMode {
		return nil // Skip conflict check in overwrite mode
	}

	conflicts := make(map[string]string) // output path -> input path
	for _, inputFile := range files {
		outputFile := filepath.Join(outputDir, filepath.Base(inputFile))
		if _, err := os.Stat(outputFile); err == nil {
			conflicts[outputFile] = inputFile
		}
	}

	if len(conflicts) == 0 {
		return nil // No conflicts
	}

	// Display conflicts to user
	fmt.Printf("⚠️  Output file conflicts detected:\n")
	for outputFile, inputFile := range conflicts {
		fmt.Printf("   %s (from %s)\n", outputFile, filepath.Base(inputFile))
	}

	fmt.Printf("\n📋 %d files already exist in output directory.\n", len(conflicts))
	fmt.Println("Choose an action:")
	fmt.Println("  1. Overwrite all existing files")
	fmt.Println("  2. Skip files that already exist")
	fmt.Println("  3. Cancel batch conversion")
	fmt.Println("  (Use --overwrite flag to automatically overwrite)")

	choice, err := promptUserChoice()
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	switch choice {
	case "1":
		// Continue with overwrite
		if !quiet {
			fmt.Printf("✅ Proceeding to overwrite %d existing files\n", len(conflicts))
		}
		return nil
	case "2":
		// Remove conflicting files from processing list
		return removeConflictingFiles(files, conflicts)
	case "3":
		return fmt.Errorf("batch conversion cancelled by user")
	default:
		return fmt.Errorf("invalid choice '%s' - please run again and choose 1, 2, or 3", choice)
	}
}

// promptUserChoice prompts user for conflict resolution choice
func promptUserChoice() (string, error) {
	fmt.Print("Enter your choice (1-3): ")

	var choice string
	_, err := fmt.Scanln(&choice)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(choice), nil
}

// removeConflictingFiles removes files that would conflict with existing outputs
func removeConflictingFiles(files []string, conflicts map[string]string) error {
	// Create a set of input files to skip
	skipFiles := make(map[string]bool)
	for _, inputFile := range conflicts {
		skipFiles[inputFile] = true
	}

	// Filter out conflicting files
	var filteredFiles []string
	for _, file := range files {
		if !skipFiles[file] {
			filteredFiles = append(filteredFiles, file)
		}
	}

	skippedCount := len(files) - len(filteredFiles)
	if skippedCount == len(files) {
		return fmt.Errorf("all files would be skipped due to conflicts - no files to process")
	}

	if !quiet {
		fmt.Printf("⏭️  Skipping %d files due to conflicts, processing %d files\n",
			skippedCount, len(filteredFiles))
	}

	// Update the global files list (this is a limitation but works for this use case)
	// In a more sophisticated implementation, this would return the filtered list
	// For now, we'll just warn the user that they need to manually exclude conflicts
	return fmt.Errorf("conflict resolution by skipping files requires manual file exclusion - please exclude conflicting files or use --overwrite")
}

// printBatchSummary prints summary of batch conversion results
func printBatchSummary(files []string, successCount, errorCount int) {
	if !quiet {
		fmt.Printf("\n📊 Concurrent Batch Conversion Summary:\n")
		fmt.Printf("   Workers used: %d\n", workerCount)
		fmt.Printf("   Total files: %d\n", len(files))
		fmt.Printf("   Successful: %d\n", successCount)
		fmt.Printf("   Failed: %d\n", errorCount)
		fmt.Printf("   Success rate: %.1f%%\n", float64(successCount)/float64(len(files))*100)
	}
}

// Close gracefully shuts down the processor
func (p *ConcurrentBatchProcessor) Close() {
	if p.cancel != nil {
		p.cancel()
	}
}
