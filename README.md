# ClassiGo - Image Classifier using Vision LLMs

ClassiGo is a Go application that automatically generates descriptive text for images using vision models. By default it talks to Ollama; with `--api openai` it can use any OpenAI-compatible chat completions server (e.g. LM Studio). It processes all images in a directory and saves AI-generated descriptions to corresponding text files.

## Features

- 🖼️ Processes multiple image formats (JPG, JPEG, PNG, GIF, BMP, WEBP)
- 🤖 Supports Ollama (default) and OpenAI-compatible `/v1/chat/completions` APIs
- 📝 Saves descriptions to `.txt` files alongside images
- 🔄 Three processing modes: create, append, and update
- ⚡ Batch processing with progress tracking
- 🛡️ Robust error handling and informative logging
- 🌍 Supports Russian language prompts and responses

## Prerequisites

Before using ClassiGo, ensure you have:

1. **Go** (version 1.21 or higher)
   - Download from: https://golang.org/dl/

2. An LLM backend:
   - **Ollama** (default) installed and running — https://ollama.ai/
     ```bash
     ollama pull glm4-v-flash
     ```
   - Or an **OpenAI-compatible** server (e.g. LM Studio) with a vision model loaded

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
classigo [--add | --update | --create | --check] [--api ollama|openai] [--seed N] [--server URL] [--timeout N] [--text-only] <model-name> <prompt-file> [directory]
```

**Flags:**
- `--add` - Append new description to existing txt files (skip images without txt files)
- `--update` - Update existing descriptions using LLM (skip images without txt files)
- `--create` - Create description files only when txt file doesn't exist
- `--check` - Check existing descriptions using LLM (skip images without txt files)
- (no flag) - Create/overwrite description files (default behavior)

**Options:**
- `--api NAME` - LLM API backend: `ollama` (default) or `openai` (OpenAI-compatible chat completions)
- `--seed N` - Random seed for LLM (default: 42)
- `--server URL` - LLM server origin with port (e.g., `http://localhost:11434`). Do **not** append `/v1/...`; for OpenAI the client adds `/v1/chat/completions`. If omitted, uses `OLLAMA_HOST` or `http://127.0.0.1:11434`
- `--timeout N` - Response timeout for LLM in seconds (default: 0 = no timeout)
- `--text-only` - Do not send image data to the model (text-only request). Useful with `--update` or `--check`; in default/create/add modes the model cannot see the image pixels

**Parameters:**
- `<model-name>` - Name of the vision model to use (e.g., `glm4-v-flash`)
- `<prompt-file>` - Path to a text file containing the prompt for image description
- `[directory]` - (Optional) Directory containing images. Defaults to current directory if not specified

### Processing Modes

ClassiGo supports three different processing modes:

1. **Default Mode** (no flags)
   - Creates new description files or overwrites existing ones
   - Processes all image files in the directory
   - Best for: Initial description generation or complete regeneration

2. **Add Mode** (`--add` flag)
   - Appends new descriptions to existing txt files
   - Skips images that don't have corresponding txt files
   - Adds a separator before the new description
   - Best for: Adding alternative descriptions or multiple perspectives

3. **Update Mode** (`--update` flag)
   - Reads existing descriptions and asks the LLM to improve them
   - Skips images that don't have corresponding txt files
   - Includes the existing description as context in the prompt
   - Best for: Refining or improving existing descriptions

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

### Using Add Mode

To append new descriptions to existing txt files:
```bash
./classigo --add glm4-v-flash ./prompt.txt ./images
```

This will:
- Only process images that already have corresponding txt files
- Generate a new description using the prompt
- Append the new description to the existing txt file with a separator
- Skip any images without txt files

### Using Update Mode

To update and improve existing descriptions:
```bash
./classigo --update glm4-v-flash ./prompt.txt ./images
```

This will:
- Only process images that already have corresponding txt files
- Read the existing description from each txt file
- Send both the image and existing description to the LLM
- Ask the LLM to update and improve the description
- Overwrite the txt file with the improved description
- Skip any images without txt files

### Example Output

**Default Mode:**
```
Using model: glm4-v-flash
Using prompt: Напиши от 10 до 30 слов описывающих изображение
Processing images in directory: ./photos
Mode: Create/overwrite descriptions

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

**Add Mode:**
```
Using model: glm4-v-flash
Using prompt: Напиши от 10 до 30 слов описывающих изображение
Processing images in directory: ./photos
Mode: Append to existing descriptions

Found 2 image(s) to process.

[1/2] Processing: sunset.jpg...
  ✓ Appended: sunset.txt (2.10 sec)

[2/2] Processing: cat.png...
  ✓ Appended: cat.txt (1.95 sec)

==================================================
Processing complete!
Success: 2 | Errors: 0 | Total: 2
```

**Update Mode:**
```
Using model: glm4-v-flash
Using prompt: Напиши от 10 до 30 слов описывающих изображение
Processing images in directory: ./photos
Mode: Update existing descriptions

Found 2 image(s) to process.

[1/2] Processing: sunset.jpg...
  ✓ Updated: sunset.txt (2.45 sec)

[2/2] Processing: cat.png...
  ✓ Updated: cat.txt (2.20 sec)

==================================================
Processing complete!
Success: 2 | Errors: 0 | Total: 2
```

## How It Works

**Default Mode:**
1. Reads the prompt from the specified prompt file
2. Scans the specified directory for image files
3. For each image:
   - Reads the image file
   - Sends it to the specified Ollama vision model with the custom prompt
   - Saves the model's response to a `.txt` file with the same name as the image
4. Displays progress, timing, and summary statistics

**Add Mode (`--add`):**
1. Reads the prompt from the specified prompt file
2. Scans the specified directory for image files
3. Filters images to only those with existing txt files
4. For each filtered image:
   - Reads the image file
   - Generates a new description using the prompt
   - Appends the new description to the existing txt file (with separator)
5. Displays progress, timing, and summary statistics

**Update Mode (`--update`):**
1. Reads the prompt from the specified prompt file
2. Scans the specified directory for image files
3. Filters images to only those with existing txt files
4. For each filtered image:
   - Reads the image file and existing txt file
   - Sends both to the LLM with modified prompt asking to improve the description
   - Overwrites the txt file with the updated description
5. Displays progress, timing, and summary statistics

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

### Connecting to a Remote Server

By default, ClassiGo connects to `localhost:11434` (Ollama). To connect to a different server (e.g., a remote machine or custom port), use the `--server` flag:

```bash
./classigo --server http://192.168.1.100:11434 glm4-v-flash ./prompt.txt ./images
```

You can also combine this with other flags:

```bash
./classigo --server http://192.168.1.100:11434 --add glm4-v-flash ./prompt.txt ./images
```

**Note:** The server URL must include the protocol (`http://` or `https://`) and port number. Pass the **origin only** (no `/v1/chat/completions` or other path suffix).

### Using an OpenAI-Compatible API

With `--api openai`, ClassiGo sends `POST {server}/v1/chat/completions` (no API key). Use this with LM Studio, vLLM, LocalAI, or similar:

```bash
./classigo --api openai --server http://192.168.1.45:1234 my-vision-model ./prompt.txt ./images
```

If `--server` is omitted, the same default / `OLLAMA_HOST` resolution applies (usually you want an explicit `--server` for OpenAI backends).

### Setting a Timeout

By default, ClassiGo waits indefinitely for the LLM to respond. To set a timeout (useful for preventing hanging on slow models or network issues), use the `--timeout` flag:

```bash
./classigo --timeout 60 glm4-v-flash ./prompt.txt ./images
```

This will timeout after 60 seconds if the model doesn't respond. You can combine this with other flags:

```bash
./classigo --timeout 120 --server http://192.168.1.100:11434 glm4-v-flash ./prompt.txt ./images
```

**Note:** If a timeout occurs, the error will be reported and ClassiGo will continue processing the next image.

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

### "Failed to create LLM client"
- Make sure the backend is running (Ollama, or your OpenAI-compatible server)
- Check that `--server` points at the correct origin (default: http://127.0.0.1:11434)

### "failed to generate description"
- Ensure the model is available on the server (e.g. `ollama pull glm4-v-flash`)
- For `--api openai`, confirm the server exposes `/v1/chat/completions` and the model supports vision if you send images
- Verify the image file is not corrupted

## Project Structure

```
ClassiGo/
├── go.mod              # Go module definition
├── go.sum              # Dependency checksums
├── main.go             # CLI, modes, file processing
├── generator.go        # Shared LLM interface and helpers
├── ollama.go           # Ollama backend
├── openai.go           # OpenAI-compatible backend
├── main_test.go        # File/directory tests
├── generator_test.go   # API/MIME/URL/request-shape tests
├── AGENTS.md           # Instructions for AI agents
└── README.md           # This file
```

## License

This project is provided as-is for educational and personal use.

## Contributing

Feel free to submit issues and enhancement requests!
