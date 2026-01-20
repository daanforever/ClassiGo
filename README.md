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

### Command Syntax

```bash
classigo <model-name> <prompt-file> [directory]
```

**Parameters:**
- `<model-name>` - Name of the Ollama vision model to use (e.g., `glm4-v-flash`)
- `<prompt-file>` - Path to a text file containing the prompt for image description
- `[directory]` - (Optional) Directory containing images. Defaults to current directory if not specified

### Basic Usage

1. Create a prompt file (e.g., `prompt.txt`):
   ```
   Напиши от 10 до 30 слов описывающих изображение
   ```

2. Process images in the current directory:
   ```bash
   go run main.go glm4-v-flash ./prompt.txt
   ```

   Or if you built the executable:
   ```bash
   ./classigo glm4-v-flash ./prompt.txt
   ```

### Specify a Directory

Process images in a specific directory:
```bash
go run main.go glm4-v-flash ./prompt.txt /path/to/images
```

Or:
```bash
./classigo glm4-v-flash ./prompt.txt /path/to/images
```

### Example Output

```
Using model: glm4-v-flash
Using prompt: Напиши от 10 до 30 слов описывающих изображение
Processing images in directory: ./photos

Found 3 image(s) to process.

[1/3] Processing: sunset.jpg...
  ✓ Saved: sunset.txt (2.34 sec)

[2/3] Processing: cat.png...
  ✓ Saved: cat.txt (1.87 sec)

[3/3] Processing: landscape.jpg...
  ✓ Saved: landscape.txt (2.15 sec)

==================================================
Processing complete!
Success: 3 | Errors: 0 | Total: 3
```

## How It Works

1. Reads the prompt from the specified prompt file
2. Scans the specified directory for image files
3. For each image:
   - Reads the image file
   - Sends it to the specified Ollama vision model with the custom prompt
   - Saves the model's response to a `.txt` file with the same name as the image
4. Displays progress, timing, and summary statistics

## Configuration

### Using Different Models

Simply specify a different model name when running the application:

```bash
./classigo llava ./prompt.txt ./images
```

Make sure the model is installed in Ollama:
```bash
ollama pull llava
```

### Customizing Prompts

Create different prompt files for different use cases:

**prompt_detailed.txt:**
```
Provide a detailed description of this image in 50-100 words, including colors, objects, and mood.
```

**prompt_short.txt:**
```
Describe this image in one sentence.
```

**prompt_russian.txt:**
```
Напиши от 10 до 30 слов описывающих изображение
```

Then use them as needed:
```bash
./classigo glm4-v-flash ./prompt_detailed.txt ./images
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
