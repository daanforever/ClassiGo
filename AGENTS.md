# ClassiGo - Instructions for AI Agents

## Project Overview
ClassiGo is a command-line tool written in Go for automatic generation of text descriptions of images using vision models through the Ollama API.

## Running Tests

### Command to run all tests
```powershell
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User"); go test -v
```

### Command to run tests with coverage
```powershell
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User"); go test -v -cover
```

### Existing tests
- `TestImageExtensions` - validation of supported image extensions
- `TestImageExtensionsCaseInsensitive` - validation of case-insensitive extensions
- `TestDirectoryValidation` - directory validation check
- `TestImageFileDetection` - image file detection check
- `TestOutputFileNaming` - output file naming generation check
- `TestCreateAndReadFile` - file creation and reading check
- `TestScanDirectory` - directory scanning check

## Building the Application

### Build command
```powershell
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User"); go build -o classigo.exe
```

### Verifying successful build
```powershell
Test-Path classigo.exe
Get-Item classigo.exe | Select-Object Name, Length, LastWriteTime
```

## Updating Dependencies

### Command to update go.mod
```powershell
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User"); go mod tidy
```

## Code Formatting

### Command to format all files
```powershell
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User"); go fmt ./...
```

### Command to check code (vet)
```powershell
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User"); go vet ./...
```

## Running the Application

### Basic usage
```powershell
.\classigo.exe <model-name> [directory]
```

### Examples
```powershell
# Process images in current directory
.\classigo.exe glm4-v-flash .

# Process images in specified directory
.\classigo.exe glm4-v-flash ./images
```

## Project Structure

```
ClassiGo/
├── main.go           # Main application code
├── main_test.go      # Tests
├── go.mod            # Go dependencies
├── go.sum            # Dependency checksums
├── README.md         # User documentation
├── AGENTS.md         # This file - instructions for AI agents
└── classigo.exe      # Compiled executable file
```

## Requirements

### System requirements
- Go 1.25.6 or newer
- Windows (PowerShell)
- Ollama installed and running

### Go dependencies
- `github.com/ollama/ollama/api` - Ollama API client

### Supported image formats
- `.jpg`, `.jpeg`
- `.png`
- `.gif`
- `.bmp`
- `.webp`

## When to Update Tests

### Tests DO NOT require updating when:
- Adding internal logic (e.g., timers, logging)
- Changing output format
- Performance optimizations without API changes

### Tests REQUIRE updating when:
- Adding new supported file formats (update `TestImageExtensions`)
- Changing output file naming logic (update `TestOutputFileNaming`)
- Changing directory processing logic (update `TestDirectoryValidation`, `TestScanDirectory`)
- Adding new public functions (add new tests)
- Changing behavior of existing functions (update corresponding tests)

## Typical Workflow

1. **Make code changes**
   ```powershell
   # Edit main.go or other files
   ```

2. **Format code**
   ```powershell
   $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User"); go fmt ./...
   ```

3. **Check code**
   ```powershell
   $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User"); go vet ./...
   ```

4. **Run tests**
   ```powershell
   $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User"); go test -v
   ```

5. **Build application**
   ```powershell
   $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User"); go build -o classigo.exe
   ```

6. **Test executable**
   ```powershell
   .\classigo.exe glm4-v-flash ./test-images
   ```

## Debugging

### Checking Go version
```powershell
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User"); go version
```

### Checking dependencies
```powershell
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User"); go list -m all
```

### Checking Ollama status
```powershell
# Make sure Ollama is running and accessible
ollama list
```

