# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoRapidOCR is a Go wrapper library for the hiroi-sora/PaddleOCR-json OCR engine. It provides a convenient interface for performing Optical Character Recognition on images by communicating with the PaddleOCR-json executable through stdin/stdout pipes.

## Development Commands

- `make test` - Run all tests
- `make build` - Build the project 
- `make lint` - Run go vet for code analysis
- `make fmt` - Format code using go fmt
- `make clean` - Clean build artifacts

## Core Architecture

### Main Components

1. **paddleocr.go** - Primary library containing:
   - `Ppocr` struct: Manages OCR process lifecycle with automatic restarts
   - `OcrArgs` struct: Configuration parameters for OCR engine
   - Communication via JSON over stdin/stdout pipes with external OCR executable

2. **utils.go** - Utility functions (file existence checking)

3. **paddleocr_test.go** - Unit tests for core functionality

### Key Design Patterns

- **Process Management**: The `Ppocr` struct manages a long-running external OCR process, handling process lifecycle, automatic restarts (every 20 minutes to prevent memory leaks), and clean shutdown
- **Concurrent Safety**: Uses mutex locks (`ppLock`) to ensure thread-safe access to the OCR process
- **Resource Management**: Implements proper cleanup patterns with `Close()` method and goroutine synchronization via channels
- **Configuration via Struct Tags**: `OcrArgs` uses reflection and struct tags (`ocrArg`) to convert Go struct fields to command line arguments

### Process Communication Flow

1. Initialize OCR process with command line arguments
2. Wait for "OCR init completed." message 
3. Send JSON data via stdin containing image path or base64 data
4. Read JSON response from stdout
5. Parse response into structured `Result` type

### External Dependencies

- Requires PaddleOCR-json executable to be downloaded separately
- Communicates with external process rather than using CGO or native Go OCR
- Supports both Windows and Linux platforms with different executable handling

## Testing Notes

Tests are designed to work with mock paths since the actual PaddleOCR-json executable may not be present in development environments. The `NewPpocr` function tests include various path scenarios including Windows-style paths.