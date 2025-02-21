中文|[English](README.md) 


# RapidOcr

[![Go Reference](https://pkg.go.dev/badge/github.com/doraemonkeys/paddleocr.svg)](https://pkg.go.dev/github.com/doraemonkeys/paddleocr)



Go语言实现的对RapidOcr-json的简单封装。

## 安装

1. RapidOcr 下载程序并解压。
2. 安装 GoRapidOcr

   ```go
   go get github.com/topascend/GoRapidOcr
   ```

## 快速开始
    
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