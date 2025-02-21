English|[中文](/README-ZH.md)  



# RapidOcr

[![Go Reference](https://pkg.go.dev/badge/github.com/doraemonkeys/paddleocr.svg)](https://pkg.go.dev/github.com/doraemonkeys/paddleocr) [![Go Report Card](https://goreportcard.com/badge/github.com/doraemonkeys/paddleocr)](https://goreportcard.com/report/github.com/doraemonkeys/paddleocr)


A simple wrapper for hiroi-sora/PaddleOCR-json implemented in Go language.


## Installation

1. Download the program from [PaddleOCR-json releases](https://github.com/hiroi-sora/PaddleOCR-json/releases) and decompress it.

2. install GoRapidOcr

   ```go
   go get github.com/topascend/GoRapidOcr
   ```

## Quick Start

```go
package main

import (
   "fmt"

   "github.com/topascend/GoRapidOcr"
)

func main() {
   p, err := GoRapidOcr.NewPpocr("C:\\Users\\mypc\\Downloads\\RapidOCR-json_v0.2.0\\RapidOCR-json.exe",
      GoRapidOcr.OcrArgs{
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
   if result.Code != GoRapidOcr.CodeSuccess {
      fmt.Println("orc failed:", result.Msg)
      return
   }
   fmt.Println(result.Data)
}

```

