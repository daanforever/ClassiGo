---
name: Add Command Line Flags
overview: Add `--add` and `--update` command line flags to ClassiGo for appending descriptions to existing txt files and updating existing descriptions using the LLM.
todos:
  - id: parse_flags
    content: Add flag parsing with --add and --update, validate mutual exclusivity
    status: completed
  - id: create_mode_type
    content: Create ProcessingMode type with constants for Default, Add, Update
    status: completed
  - id: filter_images
    content: Add filtering logic to skip images without txt files for --add and --update modes
    status: completed
  - id: refactor_process
    content: Modify processImage() to handle all three processing modes
    status: completed
  - id: update_usage
    content: Update usage message to document new flags
    status: completed
  - id: update_docs
    content: Update README.md with new flags documentation and examples
    status: completed
isProject: false
---

# Add Command Line Flags for Append and Update Modes

## Implementation Approach

We'll use Go's `flag` package to parse command line flags and modify the existing processing logic to support three modes: default (overwrite), add (append), and update (refresh).

## Key Changes

### 1. Update Argument Parsing

In `main()` function (currently lines 25-36):

```go
// Current implementation uses os.Args directly
modelName := os.Args[1]
promptFile := os.Args[2]
directory := "."
if len(os.Args) > 3 {
    directory = os.Args[3]
}
```

**Change to:** Use `flag` package to parse flags and positional arguments:

- Add `--add` boolean flag
- Add `--update` boolean flag  
- Validate that flags are mutually exclusive
- Parse remaining positional arguments (model-name, prompt-file, directory)

### 2. Create Processing Mode Type

Add an enum/constant to represent the three processing modes:

- `ModeDefault` - Create/overwrite txt files (current behavior)
- `ModeAdd` - Append to existing txt files (skip if no txt file exists)
- `ModeUpdate` - Read existing description, send to LLM with context, update file

### 3. Modify File Filtering Logic

In `main()` function where images are collected (lines 73-82):

**Current:** Collects all image files

**Add logic for `--add` mode:**

- After collecting image files, filter to only keep images that have corresponding txt files
- Example: If `photo.jpg` exists but `photo.txt` doesn't, skip `photo.jpg`

**Add logic for `--update` mode:**

- After collecting image files, filter to only keep images that have corresponding txt files
- Similar to `--add` but will read the txt content instead of skipping

### 4. Refactor `processImage()` Function

Current `processImage()` function (lines 116-169) always creates/overwrites files.

**Modify signature:**

```go
func processImage(client *api.Client, imagePath string, modelName string, prompt string, mode ProcessingMode) error
```

**Add mode-specific logic:**

**For ModeDefault (current behavior):**

- Keep existing logic: `os.Create()` to overwrite

**For ModeAdd:**

- Check if txt file exists (it should, based on filtering)
- Open file in append mode: `os.OpenFile(txtPath, os.O_APPEND|os.O_WRONLY, 0644)`
- Generate description using LLM with same prompt
- Write separator (e.g., `\n\n`) before appending new description
- Append the new description to the file

**For ModeUpdate:**

- Read existing txt file content
- Modify the prompt to include existing description as context
  - Example: `[Original Prompt]\n\nExisting description:\n[Current Content]\n\nPlease update and improve the above description.`
- Send to LLM with the modified prompt
- Overwrite the txt file with the updated description

### 5. Update Usage Message

Modify the usage message (line 28) to reflect new flags:

```
Usage: classigo [--add | --update] <model-name> <prompt-file> [directory]

Modes:
  (default)  Create/overwrite description files
  --add      Append new description to existing txt files (skip if file doesn't exist)
  --update   Update existing descriptions using LLM (skip if file doesn't exist)
```

### 6. Update Tests

No test changes needed based on the AGENTS.md rules - these are internal logic changes that don't affect existing function signatures or behavior when flags aren't used.

### 7. Update Documentation

Update [README.md](README.md) to document the new flags:

- Add section explaining the three processing modes
- Add examples showing `--add` and `--update` usage
- Explain when to use each mode

## Implementation Order

1. Add flag parsing and mode validation in `main()`
2. Create `ProcessingMode` type constants
3. Update image file filtering logic for each mode
4. Modify `processImage()` signature and add mode-specific logic
5. Update usage message
6. Update README.md documentation

## Edge Cases to Handle

- Ensure mutual exclusivity of `--add` and `--update` flags
- Handle case where txt file exists but is empty (for update mode)
- Handle file permission errors when appending or updating
- Ensure clear user feedback about which files are being skipped
