# ClassiGo - Image Classifier using Ollama

ClassiGo is a Go application that automatically generates descriptive text for images using Ollama's vision models. It processes all images in a directory and saves AI-generated descriptions to corresponding text files.

## Features

- 🖼️ Processes multiple image formats (JPG, JPEG, PNG, GIF, BMP, WEBP)
- 🤖 Uses Ollama's `glm4-v-flash` vision model for image description
- 📝 Saves descriptions to `.txt` files alongside images
- ⚡ Batch processing with progress tracking
- 🛡️ Robust error handling and informative logging
- 🌍 Supports Russian language prompts and responses

## Prerequisites

Before using ClassiGo, ensure you have:

1. **Go** (version 1.21 or higher)
   - Download from: https://golang.org/dl/

2. **Ollama** installed and running
   - Download from: https://ollama.ai/
   - Install the required model:
     ```bash
     ollama pull glm4-v-flash
     ```

## Installation

1. Clone or download this repository:
   ```bash
   git clone <repository-url>
   cd ClassiGo
   ```

2. Download dependencies:
   ```bash
   go mod tidy
   ```

3. Build the application (optional):
   ```bash
   go build -o classigo
   ```

## Usage

### Basic Usage

Process images in the current directory:
```bash
go run main.go
```

Or if you built the executable:
```bash
./classigo
```

### Specify a Directory

Process images in a specific directory:
```bash
go run main.go /path/to/images
```

Or:
```bash
./classigo /path/to/images
```

### Example Output

```
Processing images in directory: ./photos

Found 3 image(s) to process.

[1/3] Processing: sunset.jpg...
  ✓ Saved: sunset.txt

[2/3] Processing: cat.png...
  ✓ Saved: cat.txt

[3/3] Processing: landscape.jpg...
  ✓ Saved: landscape.txt

==================================================
Processing complete!
Success: 3 | Errors: 0 | Total: 3
```

## How It Works

1. Scans the specified directory for image files
2. For each image:
   - Reads the image file
   - Sends it to Ollama's `glm4-v-flash` model
   - Requests a description in Russian (10-30 words)
   - Saves the description to a `.txt` file with the same name as the image
3. Displays progress and summary statistics

## Configuration

### Changing the Model

To use a different Ollama model, edit `main.go` and change the model name:

```go
req := &api.GenerateRequest{
    Model:  "your-model-name",  // Change this
    Prompt: "Напиши от 10 до 30 слов описывающих изображение",
    Images: []api.ImageData{imgData},
}
```

### Changing the Prompt

To modify the description prompt, edit the `Prompt` field in the request:

```go
Prompt: "Your custom prompt here",
```

### Adding More Image Formats

To support additional image formats, add them to the `imageExtensions` map in `main.go`:

```go
var imageExtensions = map[string]bool{
    ".jpg":  true,
    ".jpeg": true,
    ".png":  true,
    ".tiff": true,  // Add new formats here
    // ... etc
}
```

## Troubleshooting

### "go: command not found"
Install Go from https://golang.org/dl/

### "Failed to create Ollama client"
- Make sure Ollama is installed and running
- Check that the Ollama service is accessible (default: http://localhost:11434)

### "failed to generate description"
- Ensure the `glm4-v-flash` model is installed: `ollama pull glm4-v-flash`
- Check that Ollama has enough resources to run the model
- Verify the image file is not corrupted

## Project Structure

```
ClassiGo/
├── go.mod           # Go module definition
├── go.sum           # Dependency checksums
├── main.go          # Main application code
└── README.md        # This file
```

## License

This project is provided as-is for educational and personal use.

## Contributing

Feel free to submit issues and enhancement requests!
