package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

// ProcessingMode represents how images should be processed
type ProcessingMode int

const (
	ModeDefault ProcessingMode = iota // Create/overwrite txt files
	ModeAdd                           // Append to existing txt files
	ModeUpdate                        // Update existing descriptions
	ModeCreate                        // Create txt files only if they don't exist
	ModeCheck                         // Check existing descriptions
)

// Config holds the application configuration
type Config struct {
	Mode       ProcessingMode
	JoinString string // For ModeAdd: separator between old and new content
	ModelName  string
	Prompt     string
	Directory  string
	Seed       int
	ServerURL  string
	Timeout    int
}

// programFlags holds all command-line flag pointers
type programFlags struct {
	addMode    *string
	updateMode *bool
	createMode *bool
	checkMode  *bool
	seed       *int
	server     *string
	timeout    *int
}

// Supported image extensions
var imageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
}

// getServerURL resolves the Ollama server URL from custom flag or environment
func getServerURL(customServer string) string {
	if customServer != "" {
		return customServer
	}

	// Get from environment
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://127.0.0.1:11434"
	}
	return ollamaHost
}

// createOllamaClient initializes an Ollama client with the specified server and timeout
func createOllamaClient(serverURL string, timeoutSecs int) (*api.Client, error) {
	// Create HTTP client with timeout if specified
	httpClient := &http.Client{}
	if timeoutSecs > 0 {
		httpClient.Timeout = time.Duration(timeoutSecs) * time.Second
	}

	// Parse server URL
	baseURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL '%s': %w", serverURL, err)
	}

	// Create and return client
	client := api.NewClient(baseURL, httpClient)
	return client, nil
}

// defineProgramFlags defines all command-line flags and returns pointers
func defineProgramFlags() *programFlags {
	return &programFlags{
		addMode:    flag.String("add", "", "Append new description to existing txt files with specified join string (e.g., \"\\n\" for newline)"),
		updateMode: flag.Bool("update", false, "Update existing descriptions using LLM (skip if file doesn't exist)"),
		createMode: flag.Bool("create", false, "Create description files only when txt file doesn't exist"),
		checkMode:  flag.Bool("check", false, "Check existing descriptions using LLM and output feedback to stdout"),
		seed:       flag.Int("seed", 42, "Random seed for LLM (default: 42)"),
		server:     flag.String("server", "", "Ollama server URL with port (e.g., http://localhost:11434)"),
		timeout:    flag.Int("timeout", 0, "Response timeout for Ollama in seconds (default: 0 = no timeout)"),
	}
}

// unescapeJoinString processes escape sequences in the join string
func unescapeJoinString(s string) string {
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\r", "\r")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

// validateModeFlags validates mutual exclusivity of mode flags and returns the selected mode
func validateModeFlags(flags *programFlags) (ProcessingMode, string, error) {
	// Count how many mode flags are set
	modesCount := 0
	if *flags.addMode != "" {
		modesCount++
	}
	if *flags.updateMode {
		modesCount++
	}
	if *flags.createMode {
		modesCount++
	}
	if *flags.checkMode {
		modesCount++
	}

	if modesCount > 1 {
		return ModeDefault, "", fmt.Errorf("--add, --update, --create, and --check flags cannot be used together")
	}

	// Determine processing mode and join string
	mode := ModeDefault
	joinString := ""

	if *flags.addMode != "" {
		mode = ModeAdd
		joinString = unescapeJoinString(*flags.addMode)
	} else if *flags.updateMode {
		mode = ModeUpdate
	} else if *flags.createMode {
		mode = ModeCreate
	} else if *flags.checkMode {
		mode = ModeCheck
	}

	return mode, joinString, nil
}

// showUsage displays usage information and examples
func showUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [--add \"join\" | --update | --create | --check] [--seed N] [--server URL] [--timeout N] <model-name> <prompt-file> [directory]\n\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "Modes:\n")
	fmt.Fprintf(os.Stderr, "  (default)       Create/overwrite description files\n")
	fmt.Fprintf(os.Stderr, "  --add \"join\"    Append new description to existing txt files with specified join string\n")
	fmt.Fprintf(os.Stderr, "                  Example: --add \"\\n\" for single newline, --add \"\\n\\n\" for double newline\n")
	fmt.Fprintf(os.Stderr, "  --update        Update existing descriptions using LLM (skip if file doesn't exist)\n")
	fmt.Fprintf(os.Stderr, "  --create        Create description files only when txt file doesn't exist\n")
	fmt.Fprintf(os.Stderr, "  --check         Check existing descriptions using LLM (skip if file doesn't exist)\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  --seed N        Random seed for LLM (default: 42)\n")
	fmt.Fprintf(os.Stderr, "  --server URL    Ollama server URL with port (e.g., http://localhost:11434)\n")
	fmt.Fprintf(os.Stderr, "  --timeout N     Response timeout for Ollama in seconds (default: 0 = no timeout)\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  %s glm4-v-flash ./prompt.txt ./images\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "  %s --add \"\\n\" glm4-v-flash ./prompt.txt ./images\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "  %s --add \"\\n\\n\" glm4-v-flash ./prompt.txt ./images\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "  %s --create glm4-v-flash ./prompt.txt ./images\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "  %s --check glm4-v-flash ./prompt.txt ./images\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "  %s --server http://192.168.1.100:11434 glm4-v-flash ./prompt.txt ./images\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "  %s --timeout 60 glm4-v-flash ./prompt.txt ./images\n", filepath.Base(os.Args[0]))
}

// parsePositionalArgs parses and validates positional arguments
func parsePositionalArgs() (modelName, promptFile, directory string, err error) {
	args := flag.Args()
	if len(args) < 2 {
		showUsage()
		os.Exit(1)
	}

	modelName = args[0]
	promptFile = args[1]
	directory = "."
	if len(args) > 2 {
		directory = args[2]
	}

	return modelName, promptFile, directory, nil
}

// readPromptFile reads and validates prompt file content
func readPromptFile(promptFile string) (string, error) {
	promptData, err := os.ReadFile(promptFile)
	if err != nil {
		return "", fmt.Errorf("error reading prompt file '%s': %w", promptFile, err)
	}

	prompt := strings.TrimSpace(string(promptData))
	if prompt == "" {
		return "", fmt.Errorf("prompt file '%s' is empty", promptFile)
	}

	return prompt, nil
}

// parseConfig parses command-line flags and arguments, returning a Config struct
func parseConfig() (*Config, error) {
	// Define and parse flags
	flags := defineProgramFlags()
	flag.Parse()

	// Validate mode flags and get processing mode
	mode, joinString, err := validateModeFlags(flags)
	if err != nil {
		return nil, err
	}

	// Parse positional arguments
	modelName, promptFile, directory, err := parsePositionalArgs()
	if err != nil {
		return nil, err
	}

	// Read and validate prompt file
	prompt, err := readPromptFile(promptFile)
	if err != nil {
		return nil, err
	}

	// Validate directory
	dirInfo, err := os.Stat(directory)
	if err != nil {
		return nil, fmt.Errorf("error accessing directory '%s': %w", directory, err)
	}
	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("path '%s' is not a directory", directory)
	}

	// Get server URL
	serverURL := getServerURL(*flags.server)

	return &Config{
		Mode:       mode,
		JoinString: joinString,
		ModelName:  modelName,
		Prompt:     prompt,
		Directory:  directory,
		Seed:       *flags.seed,
		ServerURL:  serverURL,
		Timeout:    *flags.timeout,
	}, nil
}

// displayProcessingInfo displays configuration information to the user
func displayProcessingInfo(config *Config) {
	fmt.Printf("Using model: %s\n", config.ModelName)
	fmt.Printf("Using prompt: %s\n", config.Prompt)
	fmt.Printf("Processing images in directory: %s\n", config.Directory)

	// Display mode
	switch config.Mode {
	case ModeAdd:
		// Show join string with visible escape sequences
		displayJoin := strings.ReplaceAll(config.JoinString, "\n", "\\n")
		displayJoin = strings.ReplaceAll(displayJoin, "\t", "\\t")
		displayJoin = strings.ReplaceAll(displayJoin, "\r", "\\r")
		fmt.Printf("Mode: Append to existing descriptions (join: \"%s\")\n\n", displayJoin)
	case ModeUpdate:
		fmt.Printf("Mode: Update existing descriptions\n\n")
	case ModeCreate:
		fmt.Printf("Mode: Create only when txt file doesn't exist\n\n")
	case ModeCheck:
		fmt.Printf("Mode: Check existing descriptions\n\n")
	default:
		fmt.Printf("Mode: Create/overwrite descriptions\n\n")
	}
}

// findImageFiles scans directory for image files and filters based on mode
func findImageFiles(directory string, mode ProcessingMode) ([]string, error) {
	// Scan directory for image files
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	imageFiles := []string{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if imageExtensions[ext] {
			imageFiles = append(imageFiles, file.Name())
		}
	}

	// Filter images based on mode
	switch mode {
	case ModeAdd, ModeUpdate, ModeCheck:
		filteredFiles := []string{}
		skippedCount := 0
		for _, filename := range imageFiles {
			imagePath := filepath.Join(directory, filename)
			ext := filepath.Ext(imagePath)
			txtPath := strings.TrimSuffix(imagePath, ext) + ".txt"

			// Check if corresponding txt file exists
			if _, err := os.Stat(txtPath); err == nil {
				filteredFiles = append(filteredFiles, filename)
			} else {
				skippedCount++
			}
		}
		imageFiles = filteredFiles

		if skippedCount > 0 {
			fmt.Printf("Skipped %d image(s) without existing txt files.\n", skippedCount)
		}
	case ModeCreate:
		filteredFiles := []string{}
		skippedCount := 0
		for _, filename := range imageFiles {
			imagePath := filepath.Join(directory, filename)
			ext := filepath.Ext(imagePath)
			txtPath := strings.TrimSuffix(imagePath, ext) + ".txt"

			// Check if corresponding txt file exists
			if _, err := os.Stat(txtPath); err != nil {
				// File doesn't exist, include it
				filteredFiles = append(filteredFiles, filename)
			} else {
				// File exists, skip it
				skippedCount++
			}
		}
		imageFiles = filteredFiles

		if skippedCount > 0 {
			fmt.Printf("Skipped %d image(s) with existing txt files.\n", skippedCount)
		}
	}

	return imageFiles, nil
}

// processAllImages processes all image files and returns success/error counts
func processAllImages(client *api.Client, config *Config, imageFiles []string) (successCount, errorCount int) {
	for i, filename := range imageFiles {
		fmt.Printf("[%d/%d] Processing: %s...\n", i+1, len(imageFiles), filename)

		imagePath := filepath.Join(config.Directory, filename)

		// Process the image
		if err := processImage(client, imagePath, config.ModelName, config.Prompt, config.Mode, config.Seed, config.JoinString); err != nil {
			fmt.Printf("  ❌ Error: %v\n", err)
			errorCount++
		} else {
			successCount++
		}
		fmt.Println()
	}

	return successCount, errorCount
}

// displayNoImagesMessage displays appropriate message when no images found
func displayNoImagesMessage(mode ProcessingMode) {
	if mode == ModeAdd || mode == ModeUpdate || mode == ModeCheck {
		fmt.Println("No image files with existing txt files found in the directory.")
	} else if mode == ModeCreate {
		fmt.Println("No image files without existing txt files found in the directory.")
	} else {
		fmt.Println("No image files found in the directory.")
	}
}

// displaySummary displays processing summary statistics
func displaySummary(successCount, errorCount, total int) {
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Processing complete!\n")
	fmt.Printf("Success: %d | Errors: %d | Total: %d\n", successCount, errorCount, total)
}

func main() {
	// Parse configuration
	config, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Display processing info
	displayProcessingInfo(config)

	// Initialize Ollama client
	client, err := createOllamaClient(config.ServerURL, config.Timeout)
	if err != nil {
		log.Fatalf("Failed to create Ollama client: %v\nMake sure Ollama is installed and running.", err)
	}

	// Find images to process
	imageFiles, err := findImageFiles(config.Directory, config.Mode)
	if err != nil {
		log.Fatal(err)
	}

	if len(imageFiles) == 0 {
		displayNoImagesMessage(config.Mode)
		return
	}

	fmt.Printf("Found %d image(s) to process.\n\n", len(imageFiles))

	// Process all images
	successCount, errorCount := processAllImages(client, config, imageFiles)

	// Display summary
	displaySummary(successCount, errorCount, len(imageFiles))
}

func processImage(client *api.Client, imagePath string, modelName string, prompt string, mode ProcessingMode, seed int, joinString string) error {
	// Start timing
	startTime := time.Now()

	// Read image file
	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	// Prepare output file path
	ext := filepath.Ext(imagePath)
	txtPath := strings.TrimSuffix(imagePath, ext) + ".txt"

	// Prepare the prompt based on mode
	finalPrompt := prompt
	if mode == ModeUpdate {
		// Read existing description
		existingContent, err := os.ReadFile(txtPath)
		if err != nil {
			return fmt.Errorf("failed to read existing txt file: %w", err)
		}

		// Modify prompt to include existing description as context
		finalPrompt = fmt.Sprintf("%s\n\nRead existing description. Analize any issues and fix the formatting and description. NEVER OUTPUT original text.\n\nExisting description:\n\n%s\n\n", prompt, strings.TrimSpace(string(existingContent)))
	} else if mode == ModeCheck {
		// Read existing description
		existingContent, err := os.ReadFile(txtPath)
		if err != nil {
			return fmt.Errorf("failed to read existing txt file: %w", err)
		}

		// Modify prompt to ask LLM to check the description
		finalPrompt = fmt.Sprintf("%s\n\nExisting description:\n%s\n\nAnalyze the description and identify any issues. Report only issues or 'No issues found'", prompt, strings.TrimSpace(string(existingContent)))
	}

	// Prepare request
	req := &api.GenerateRequest{
		Model:  modelName,
		Prompt: finalPrompt,
		Images: []api.ImageData{imgData},
		Options: map[string]interface{}{
			"seed": seed,
		},
	}

	// Call Ollama API with retry logic for empty responses
	ctx := context.Background()
	var response strings.Builder
	maxAttempts := 2

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Reset response for each attempt
		response.Reset()

		respFunc := func(resp api.GenerateResponse) error {
			response.WriteString(resp.Response)
			return nil
		}

		err = client.Generate(ctx, req, respFunc)
		if err != nil {
			return fmt.Errorf("failed to generate description: %w", err)
		}

		// Check if response is empty
		responseText := strings.TrimSpace(response.String())
		if responseText != "" {
			// Success - got a non-empty response
			break
		}

		// Empty response received
		if attempt < maxAttempts {
			fmt.Printf("  ⚠ Empty response received, retrying...\n")
		} else {
			return fmt.Errorf("received empty response from LLM after %d attempts", maxAttempts)
		}
	}

	// Final validation: ensure response is not empty
	finalResponse := strings.TrimSpace(response.String())
	if finalResponse == "" {
		return fmt.Errorf("cannot write file: response is empty")
	}

	// Write response based on mode
	switch mode {
	case ModeAdd:
		// Open file in append mode
		outFile, err := os.OpenFile(txtPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open output file for appending: %w", err)
		}

		// Write separator and new description using configurable join string
		_, err = outFile.WriteString(joinString + response.String())
		outFile.Close() // Close immediately after writing
		if err != nil {
			return fmt.Errorf("failed to append to output file: %w", err)
		}

		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("  ✓ Appended: %s (%.2f sec)\n", filepath.Base(txtPath), elapsed)

	case ModeUpdate:
		// Overwrite file with updated description
		outFile, err := os.Create(txtPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}

		_, err = outFile.WriteString(response.String())
		outFile.Close() // Close immediately after writing
		if err != nil {
			return fmt.Errorf("failed to write to output file: %w", err)
		}

		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("  ✓ Updated: %s (%.2f sec)\n", filepath.Base(txtPath), elapsed)

	case ModeCheck:
		// Output check results to stdout
		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("  ✓ Checked: %s (%.2f sec)\n", filepath.Base(txtPath), elapsed)
		fmt.Printf("\n%s\n", strings.Repeat("-", 70))
		fmt.Printf("File: %s\n", filepath.Base(imagePath))
		fmt.Printf("%s\n", strings.Repeat("-", 70))
		fmt.Printf("%s\n", response.String())
		fmt.Printf("%s\n", strings.Repeat("-", 70))

	default: // ModeDefault
		// Create or truncate output file
		outFile, err := os.Create(txtPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}

		_, err = outFile.WriteString(response.String())
		outFile.Close() // Close immediately after writing
		if err != nil {
			return fmt.Errorf("failed to write to output file: %w", err)
		}

		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("  ✓ Saved: %s (%.2f sec)\n", filepath.Base(txtPath), elapsed)
	}

	return nil
}
