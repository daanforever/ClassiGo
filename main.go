package main

import (
	"context"
	"flag"
	"fmt"
	"log"
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
)

// Supported image extensions
var imageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
}

func main() {
	// Define flags
	addMode := flag.Bool("add", false, "Append new description to existing txt files (skip if file doesn't exist)")
	updateMode := flag.Bool("update", false, "Update existing descriptions using LLM (skip if file doesn't exist)")
	flag.Parse()

	// Validate flags are mutually exclusive
	if *addMode && *updateMode {
		log.Fatalf("Error: --add and --update flags cannot be used together")
	}

	// Determine processing mode
	mode := ModeDefault
	if *addMode {
		mode = ModeAdd
	} else if *updateMode {
		mode = ModeUpdate
	}

	// Parse positional arguments
	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--add | --update] <model-name> <prompt-file> [directory]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Modes:\n")
		fmt.Fprintf(os.Stderr, "  (default)  Create/overwrite description files\n")
		fmt.Fprintf(os.Stderr, "  --add      Append new description to existing txt files (skip if file doesn't exist)\n")
		fmt.Fprintf(os.Stderr, "  --update   Update existing descriptions using LLM (skip if file doesn't exist)\n\n")
		fmt.Fprintf(os.Stderr, "Example: %s glm4-v-flash ./prompt.txt ./images\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Example: %s --add glm4-v-flash ./prompt.txt ./images\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	modelName := args[0]
	promptFile := args[1]
	directory := "."
	if len(args) > 2 {
		directory = args[2]
	}

	// Read prompt from file
	promptData, err := os.ReadFile(promptFile)
	if err != nil {
		log.Fatalf("Error reading prompt file '%s': %v", promptFile, err)
	}
	prompt := strings.TrimSpace(string(promptData))
	if prompt == "" {
		log.Fatalf("Prompt file '%s' is empty", promptFile)
	}

	// Validate directory
	dirInfo, err := os.Stat(directory)
	if err != nil {
		log.Fatalf("Error accessing directory '%s': %v", directory, err)
	}
	if !dirInfo.IsDir() {
		log.Fatalf("Path '%s' is not a directory", directory)
	}

	fmt.Printf("Using model: %s\n", modelName)
	fmt.Printf("Using prompt: %s\n", prompt)
	fmt.Printf("Processing images in directory: %s\n", directory)

	// Display mode
	switch mode {
	case ModeAdd:
		fmt.Printf("Mode: Append to existing descriptions\n\n")
	case ModeUpdate:
		fmt.Printf("Mode: Update existing descriptions\n\n")
	default:
		fmt.Printf("Mode: Create/overwrite descriptions\n\n")
	}

	// Initialize Ollama client
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatalf("Failed to create Ollama client: %v\nMake sure Ollama is installed and running.", err)
	}

	// Scan directory for image files
	files, err := os.ReadDir(directory)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
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
	if mode == ModeAdd || mode == ModeUpdate {
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
	}

	if len(imageFiles) == 0 {
		if mode == ModeAdd || mode == ModeUpdate {
			fmt.Println("No image files with existing txt files found in the directory.")
		} else {
			fmt.Println("No image files found in the directory.")
		}
		return
	}

	fmt.Printf("Found %d image(s) to process.\n\n", len(imageFiles))

	// Process each image
	successCount := 0
	errorCount := 0

	for i, filename := range imageFiles {
		fmt.Printf("[%d/%d] Processing: %s...\n", i+1, len(imageFiles), filename)

		imagePath := filepath.Join(directory, filename)

		// Process the image
		if err := processImage(client, imagePath, modelName, prompt, mode); err != nil {
			fmt.Printf("  ❌ Error: %v\n", err)
			errorCount++
		} else {
			successCount++
		}
		fmt.Println()
	}

	// Print summary
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Processing complete!\n")
	fmt.Printf("Success: %d | Errors: %d | Total: %d\n", successCount, errorCount, len(imageFiles))
}

func processImage(client *api.Client, imagePath string, modelName string, prompt string, mode ProcessingMode) error {
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
		finalPrompt = fmt.Sprintf("%s\n\nExisting description:\n%s\n\nPlease update and improve the above description.", prompt, strings.TrimSpace(string(existingContent)))
	}

	// Prepare request
	req := &api.GenerateRequest{
		Model:  modelName,
		Prompt: finalPrompt,
		Images: []api.ImageData{imgData},
	}

	// Collect response
	var response strings.Builder

	respFunc := func(resp api.GenerateResponse) error {
		response.WriteString(resp.Response)
		return nil
	}

	// Call Ollama API
	ctx := context.Background()
	err = client.Generate(ctx, req, respFunc)
	if err != nil {
		return fmt.Errorf("failed to generate description: %w", err)
	}

	// Write response based on mode
	var outFile *os.File
	switch mode {
	case ModeAdd:
		// Open file in append mode
		outFile, err = os.OpenFile(txtPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open output file for appending: %w", err)
		}
		defer outFile.Close()

		// Write separator and new description
		_, err = outFile.WriteString("\n\n" + response.String())
		if err != nil {
			return fmt.Errorf("failed to append to output file: %w", err)
		}

		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("  ✓ Appended: %s (%.2f sec)\n", filepath.Base(txtPath), elapsed)

	case ModeUpdate:
		// Overwrite file with updated description
		outFile, err = os.Create(txtPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outFile.Close()

		_, err = outFile.WriteString(response.String())
		if err != nil {
			return fmt.Errorf("failed to write to output file: %w", err)
		}

		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("  ✓ Updated: %s (%.2f sec)\n", filepath.Base(txtPath), elapsed)

	default: // ModeDefault
		// Create or truncate output file
		outFile, err = os.Create(txtPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outFile.Close()

		_, err = outFile.WriteString(response.String())
		if err != nil {
			return fmt.Errorf("failed to write to output file: %w", err)
		}

		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("  ✓ Saved: %s (%.2f sec)\n", filepath.Base(txtPath), elapsed)
	}

	return nil
}
