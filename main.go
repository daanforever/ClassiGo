package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ollama/ollama/api"
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
	// Parse command-line arguments
	directory := "."
	if len(os.Args) > 1 {
		directory = os.Args[1]
	}

	// Validate directory
	dirInfo, err := os.Stat(directory)
	if err != nil {
		log.Fatalf("Error accessing directory '%s': %v", directory, err)
	}
	if !dirInfo.IsDir() {
		log.Fatalf("Path '%s' is not a directory", directory)
	}

	fmt.Printf("Processing images in directory: %s\n\n", directory)

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

	if len(imageFiles) == 0 {
		fmt.Println("No image files found in the directory.")
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
		if err := processImage(client, imagePath); err != nil {
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

func processImage(client *api.Client, imagePath string) error {
	// Read image file
	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	// Prepare output file path
	ext := filepath.Ext(imagePath)
	txtPath := strings.TrimSuffix(imagePath, ext) + ".txt"

	// Create or truncate output file
	outFile, err := os.Create(txtPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Prepare request
	req := &api.GenerateRequest{
		Model:  "glm4-v-flash",
		Prompt: "Напиши от 10 до 30 слов описывающих изображение",
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

	// Write response to file
	_, err = outFile.WriteString(response.String())
	if err != nil {
		return fmt.Errorf("failed to write to output file: %w", err)
	}

	fmt.Printf("  ✓ Saved: %s\n", filepath.Base(txtPath))
	return nil
}
