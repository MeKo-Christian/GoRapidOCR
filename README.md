# GoRapidOCR

[![Go Reference](https://pkg.go.dev/badge/github.com/MeKo-Christian/GoRapidOCR.svg)](https://pkg.go.dev/github.com/MeKo-Christian/GoRapidOCR) [![Go Report Card](https://goreportcard.com/badge/github.com/MeKo-Christian/GoRapidOCR)](https://goreportcard.com/report/github.com/MeKo-Christian/GoRapidOCR) [![CI](https://github.com/MeKo-Christian/GoRapidOCR/actions/workflows/ci.yml/badge.svg)](https://github.com/MeKo-Christian/GoRapidOCR/actions/workflows/ci.yml)


A simple wrapper for hiroi-sora/PaddleOCR-json implemented in Go language.


## Installation

1. Download the program from [PaddleOCR-json releases](https://github.com/hiroi-sora/PaddleOCR-json/releases) and decompress it.

2. Install GoRapidOCR

   ```go
   go get github.com/MeKo-Christian/GoRapidOCR
   ```

## Development

First, install [Just](https://github.com/casey/just) if you haven't already.

Available commands:

- `just test` - Run tests
- `just build` - Build the project
- `just lint` - Run go vet
- `just fmt` - Format code
- `just check` - Run all checks (test, lint, fmt)
- `just deps` - Install/update dependencies
- `just test-cover` - Run tests with coverage
- `just all` - Build, test, and check everything

## Quick Start

```go
package main

import (
   "fmt"

   "github.com/MeKo-Christian/GoRapidOCR"
)

func main() {
   p, err := GoRapidOCR.NewPpocr("C:\\Users\\mypc\\Downloads\\RapidOCR-json_v0.2.0\\RapidOCR-json.exe",
      GoRapidOCR.OcrArgs{
         Models: "models",
         Det:    "ch_PP-OCRv4_det_server_infer.onnx",
      })
   if err != nil {
      panic(err)
   }
   defer p.Close()
   result, err := p.OcrFileAndParse(`C:\Users\mypc\Downloads\RapidOCR-json_v0.2.0\1.png`)
   if err != nil {
      panic(err)
   }
   if result.Code != GoRapidOCR.CodeSuccess {
      fmt.Println("orc failed:", result.Msg)
      return
   }
   fmt.Println(result.Data)
}

```

