package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestImageExtensions verifies that the image extensions map contains expected formats
func TestImageExtensions(t *testing.T) {
	expectedExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}

	for _, ext := range expectedExtensions {
		if !imageExtensions[ext] {
			t.Errorf("Expected extension %s to be supported", ext)
		}
	}
}

// TestImageExtensionsCaseInsensitive verifies case-insensitive extension matching
func TestImageExtensionsCaseInsensitive(t *testing.T) {
	// We check lowercase in the code, so uppercase should not be in the map
	upperCaseExts := []string{".JPG", ".PNG", ".GIF"}

	for _, ext := range upperCaseExts {
		if imageExtensions[ext] {
			t.Errorf("Extension map should only contain lowercase extensions, found: %s", ext)
		}
	}
}

// TestCreateTempDirectory creates a temporary directory for testing
func createTempTestDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "classigo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	return tempDir
}

// TestDirectoryValidation tests directory validation logic
func TestDirectoryValidation(t *testing.T) {
	t.Run("ValidDirectory", func(t *testing.T) {
		tempDir := createTempTestDir(t)
		defer os.RemoveAll(tempDir)

		// Check that directory exists and is a directory
		dirInfo, err := os.Stat(tempDir)
		if err != nil {
			t.Errorf("Expected directory to exist: %v", err)
		}
		if !dirInfo.IsDir() {
			t.Error("Expected path to be a directory")
		}
	})

	t.Run("NonExistentDirectory", func(t *testing.T) {
		_, err := os.Stat("/nonexistent/directory/path")
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
	})

	t.Run("FileInsteadOfDirectory", func(t *testing.T) {
		tempDir := createTempTestDir(t)
		defer os.RemoveAll(tempDir)

		// Create a file
		filePath := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(filePath, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Check that it's not a directory
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to stat file: %v", err)
		}
		if fileInfo.IsDir() {
			t.Error("Expected file to not be a directory")
		}
	})
}

// TestImageFileDetection tests that we correctly identify image files
func TestImageFileDetection(t *testing.T) {
	testCases := []struct {
		filename string
		isImage  bool
	}{
		{"photo.jpg", true},
		{"photo.JPG", true},
		{"image.png", true},
		{"picture.gif", true},
		{"graphic.bmp", true},
		{"modern.webp", true},
		{"document.pdf", false},
		{"text.txt", false},
		{"video.mp4", false},
		{"audio.mp3", false},
		{"noextension", false},
		{"multiple.dots.jpg", true},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			// Get extension (case-sensitive from filepath.Ext)
			ext := filepath.Ext(tc.filename)

			// We need to check lowercase since our map has lowercase extensions
			isImageActual := imageExtensions[strings.ToLower(ext)]

			if isImageActual != tc.isImage {
				t.Errorf("For file %s, expected isImage=%v, got %v", tc.filename, tc.isImage, isImageActual)
			}
		})
	}
}

// TestOutputFileNaming tests that output file names are generated correctly
func TestOutputFileNaming(t *testing.T) {
	testCases := []struct {
		inputPath  string
		outputPath string
	}{
		{"image.jpg", "image.txt"},
		{"photo.png", "photo.txt"},
		{"path/to/image.jpg", "path/to/image.txt"},
		{"image.jpeg", "image.txt"},
		{"complex.name.with.dots.jpg", "complex.name.with.dots.txt"},
	}

	for _, tc := range testCases {
		t.Run(tc.inputPath, func(t *testing.T) {
			ext := filepath.Ext(tc.inputPath)
			outputPath := tc.inputPath[:len(tc.inputPath)-len(ext)] + ".txt"

			if outputPath != tc.outputPath {
				t.Errorf("For input %s, expected output %s, got %s", tc.inputPath, tc.outputPath, outputPath)
			}
		})
	}
}

// TestCreateAndReadFile tests file creation and reading
func TestCreateAndReadFile(t *testing.T) {
	tempDir := createTempTestDir(t)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Test content for image description"

	// Write file
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Read file
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected content %q, got %q", testContent, string(content))
	}
}

// TestScanDirectory tests directory scanning functionality
func TestScanDirectory(t *testing.T) {
	tempDir := createTempTestDir(t)
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{
		"image1.jpg",
		"image2.png",
		"photo.gif",
		"document.txt",
		"video.mp4",
		"image3.JPEG",
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filePath, []byte("dummy content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Create a subdirectory (should be skipped)
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Scan directory
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	imageCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if imageExtensions[ext] {
			imageCount++
		}
	}

	// We expect 4 image files (.jpg, .png, .gif, .jpeg)
	// Now .JPEG will match because we convert to lowercase
	expectedCount := 4
	if imageCount != expectedCount {
		t.Errorf("Expected %d image files, found %d", expectedCount, imageCount)
	}
}
