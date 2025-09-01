package GoRapidOCR

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

type OcrArgs struct {
	EnsureAscii    string `ocrArg:"--ensureAscii"`    // Enable (1)/Disable (0) ASCII escape output	Default: 0
	Models         string `ocrArg:"--models"`         // Model directory path, absolute or relative	Default: "models"
	Det            string `ocrArg:"--det"`            // Detection model name	Default: "ch_PP-OCRv3_det_infer.onnx"
	Cls            string `ocrArg:"--cls"`            // Classification model name	Default: "ch_ppocr_mobile_v2.0_cls_infer.onnx"
	Rec            string `ocrArg:"--rec"`            // Recognition model name	Default: "ch_PP-OCRv3_rec_infer.onnx"
	Keys           string `ocrArg:"--keys"`           // Recognition dictionary name	Default: "ppocr_keys_v1.txt"
	DoAngle        string `ocrArg:"--doAngle"`        // Enable (1)/Disable (0) text direction detection	Default: 1
	MostAngle      string `ocrArg:"--mostAngle"`      // Enable (1)/Disable (0) angle voting	Default: 1
	NumThread      string `ocrArg:"--numThread"`      // Number of threads	Default: 4
	Padding        string `ocrArg:"--padding"`        // Preprocessing border width, optimizes narrow image recognition	Default: 50
	MaxSideLen     string `ocrArg:"--maxSideLen"`     // Image long side shrink value, improves large image speed	Default: 1024
	BoxScoreThresh string `ocrArg:"--boxScoreThresh"` // Text box confidence threshold	Default: 0.5
	BoxThresh      string `ocrArg:"--boxThresh"`      // Default: 0.3
	UnClipRatio    string `ocrArg:"--unClipRatio"`    // Single text box size multiplier	Default: 1.6
	ImagePath      string `ocrArg:"--image_path"`     // Initial image path	Default: ""
}

const clipboardImagePath = `clipboard`

func (o OcrArgs) CmdString() string {
	var s string
	v := reflect.ValueOf(o)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsZero() {
			continue
		}
		f := v.Type().Field(i)
		if f.Tag.Get(ocrArgTag) == "" {
			continue
		}
		// value := v.Field(i).Elem().Interface()
		value := v.Field(i).Interface()

		switch valueType := value.(type) {
		case *bool:
			if *valueType {
				s += fmt.Sprintf("%s=1 ", f.Tag.Get(ocrArgTag))
			} else {
				s += fmt.Sprintf("%s=0 ", f.Tag.Get(ocrArgTag))
			}
		default:
			if v.Field(i).Kind() == reflect.Ptr {
				s += fmt.Sprintf("%s=%v ", f.Tag.Get(ocrArgTag), v.Field(i).Elem().Interface())
			} else {
				s += fmt.Sprintf("%s=%v ", f.Tag.Get(ocrArgTag), value)
			}
		}
	}
	s = strings.TrimSpace(s)
	return s
}

// OcrFile processes the OCR for a given image file path using the specified OCR arguments.
// It returns the raw OCR result as bytes and any error encountered.
func OcrFile(exePath, imagePath string, argsCnf OcrArgs) ([]byte, error) {
	p, err := NewPpocr(exePath, argsCnf)
	if err != nil {
		return nil, err
	}
	defer p.Close()
	b, err := p.OcrFile(imagePath)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func OcrFileAndParse(exePath, imagePath string, argsCnf OcrArgs) (Result, error) {
	data, err := OcrFile(exePath, imagePath, argsCnf)
	if err != nil {
		return Result{}, err
	}
	return ParseResult(data)
}

type Ppocr struct {
	exePath         string
	args            OcrArgs
	ppLock          *sync.Mutex
	restartExitChan chan struct{}
	internalErr     error

	cmdStdout io.ReadCloser
	cmdStdin  io.WriteCloser
	cmd       *exec.Cmd
	// 无缓冲同步信号通道，close()中接收，Run()中发送。
	// Run()退出必须有对应close方法的调用
	runGoroutineExitedChan chan struct{}
	// startTime time.Time
}

// NewPpocr creates a new instance of the Ppocr struct with the provided executable path
// and OCR arguments.
// It initializes the OCR process and returns a pointer to the Ppocr instance
// and any error encountered.
//
// It is the caller's responsibility to close the Ppocr instance when finished.
func NewPpocr(exePath string, args OcrArgs) (*Ppocr, error) {
	if !fileIsExist(exePath) {
		return nil, fmt.Errorf("executable file %s not found", exePath)
	}
	p := &Ppocr{
		exePath:                exePath,
		args:                   args,
		ppLock:                 new(sync.Mutex),
		restartExitChan:        make(chan struct{}),
		runGoroutineExitedChan: make(chan struct{}),
	}

	p.ppLock.Lock()
	defer p.ppLock.Unlock()
	err := p.initPpocr(exePath, args)
	if err == nil {
		go p.restartTimer()
	} else {
		p.close()
	}
	return p, err
}

// Locked call, need to close on error
func (p *Ppocr) initPpocr(exePath string, args OcrArgs) error {
	var cmdSlash string
	if runtime.GOOS == "windows" {
		cmdSlash = "\\"
	} else {
		cmdSlash = "/"
	}
	p.cmd = exec.Command("."+cmdSlash+filepath.Base(exePath), strings.Fields(args.CmdString())...)
	cmdDir := filepath.Dir(exePath)
	if cmdDir == "." {
		cmdDir = ""
	}
	p.cmd.Dir = cmdDir
	wc, err := p.cmd.StdinPipe()
	if err != nil {
		return err
	}
	rc, err := p.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	p.cmdStdin = wc
	p.cmdStdout = rc

	var stderrBuffer bytes.Buffer
	p.cmd.Stderr = &stderrBuffer

	err = p.cmd.Start()
	if err != nil {
		return fmt.Errorf("OCR process start failed: %v", err)
	}

	go func() {
		p.internalErr = nil
		err := p.cmd.Wait()
		// fmt.Println("Run() OCR process exited, error:", err)
		if err != nil {
			p.internalErr = err
		}
		p.runGoroutineExitedChan <- struct{}{}
	}()

	buf := make([]byte, 4096)
	start := 0
	for {
		n, err := rc.Read(buf[start:])
		if err != nil {
			if p.internalErr != nil {
				return fmt.Errorf("OCR init failed: %v,run error: %v", err, p.internalErr)
			}
			return fmt.Errorf("OCR init failed, error: %v, output: %s %s", err, buf[:start], stderrBuffer.String())
		}
		start += n
		if start >= len(buf) {
			return fmt.Errorf("OCR init failed: output too long")
		}
		if bytes.Contains(buf[:start], []byte("OCR init completed.")) {
			break
		}
	}
	return p.internalErr
}

// Close cleanly shuts down the OCR process associated with the Ppocr instance.
// It releases any resources and terminates the OCR process.
//
// Warning: This method should only be called once.
func (p *Ppocr) Close() error {
	p.ppLock.Lock()
	defer p.ppLock.Unlock()
	// close(p.restartExitChan) // 只能关闭一次
	select {
	case <-p.restartExitChan:
		return fmt.Errorf("OCR process has been closed")
	default:
		close(p.restartExitChan)
	}
	p.internalErr = fmt.Errorf("OCR process has been closed")
	return p.close()
}

func (p *Ppocr) close() (err error) {
	select {
	case <-p.runGoroutineExitedChan:
		return nil
	default:
	}
	defer func() {
		// 可能的情况：Run刚退出，p.exited还没设置为true
		if r := recover(); r != nil {
			err = fmt.Errorf("close panic: %v", r)
		}
		// fmt.Println("wait OCR runGoroutineExitedChan")
		<-p.runGoroutineExitedChan
		// fmt.Println("wait OCR runGoroutineExitedChan done")
	}()
	if p.cmd == nil {
		return nil
	}
	if p.cmd.ProcessState != nil && p.cmd.ProcessState.Exited() {
		return nil
	}
	if err := p.cmdStdin.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "close cmdIn error: %v\n", err)
	}
	if err := p.cmd.Process.Kill(); err != nil {
		return err
	}
	// fmt.Println("kill OCR process success")
	return nil
}

// Timed restart process to reduce memory usage (OCR program has memory leak)
func (p *Ppocr) restartTimer() {
	// ticker := time.NewTicker(10 * time.Second)
	ticker := time.NewTicker(20 * time.Minute)
	for {
		select {
		case <-ticker.C:
			// fmt.Println("restart OCR process")
			p.ppLock.Lock()
			_ = p.close()
			p.internalErr = p.initPpocr(p.exePath, p.args)
			p.ppLock.Unlock()
			// fmt.Println("restart OCR process done")
		case <-p.restartExitChan:
			// fmt.Println("exit OCR process")
			return
		}
	}
}

type imageData struct {
	Path       string `json:"image_path,omitempty"`
	ContentB64 []byte `json:"image_base64,omitempty"`
}

// OcrFile sends an image file path to the OCR process and retrieves the OCR result.
// It returns the OCR result as bytes and any error encountered.
func (p *Ppocr) OcrFile(imagePath string) ([]byte, error) {
	var data = imageData{Path: imagePath}
	dataJson, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	p.ppLock.Lock()
	defer p.ppLock.Unlock()
	if p.internalErr != nil {
		return nil, p.internalErr
	}
	return p.ocr(dataJson)
}

func (p *Ppocr) ocr(dataJson []byte) ([]byte, error) {
	_, err := p.cmdStdin.Write(dataJson)
	if err != nil {
		return nil, err
	}
	_, err = p.cmdStdin.Write([]byte("\n"))
	if err != nil {
		return nil, err
	}
	content := make([]byte, 1024*10)
	start := 0
	for {
		n, err := p.cmdStdout.Read(content[start:])
		if err != nil {
			return nil, err
		}
		start += n
		if start >= len(content) {
			content = append(content, make([]byte, 1024*10)...)
		}
		if content[start-1] == '\n' {
			break
		}
	}
	return content[:start], nil
}

// Ocr processes the OCR for a given image represented as a byte slice.
// It returns the OCR result as bytes and any error encountered.
func (p *Ppocr) Ocr(image []byte) ([]byte, error) {
	if p.internalErr != nil {
		return nil, p.internalErr
	}
	var data = imageData{ContentB64: image}
	dataJson, err := json.Marshal(data) //auto base64
	if err != nil {
		return nil, err
	}

	p.ppLock.Lock()
	defer p.ppLock.Unlock()
	return p.ocr(dataJson)
}

type Data struct {
	Rect  [][]int `json:"box"`
	Score float32 `json:"score"`
	Text  string  `json:"text"`
}

type Result struct {
	Code int
	Msg  string
	Data []Data
}

const (
	// CodeSuccess indicates that the OCR process was successful.
	CodeSuccess = 100
	// CodeNoText indicates that no text was recognized.
	CodeNoText = 101
)

// ParseResult parses the raw OCR result bytes into a slice of Result structs.
// It returns the parsed results and any error encountered during parsing.
func ParseResult(rawData []byte) (Result, error) {
	var resp map[string]any
	err := json.Unmarshal(rawData, &resp)
	if err != nil {
		return Result{}, err
	}
	var result = Result{}
	var resData = make([]Data, 0)
	if resp["code"] == nil {
		return Result{}, fmt.Errorf("no code in response")
	}
	if resp["code"].(float64) != 100 {
		result.Code = int(resp["code"].(float64))
		result.Msg = fmt.Sprintf("%v", resp["data"])
		return result, nil
	}
	if resp["data"] == nil {
		return Result{}, fmt.Errorf("no data in response")
	}
	dataSlice, ok := resp["data"]
	if !ok {
		return result, fmt.Errorf("data is not array")
	}
	result.Code = CodeSuccess
	result.Msg = "parse success"

	var data []any
	data, ok = dataSlice.([]any)
	if !ok {
		return result, fmt.Errorf("data is not array")
	}
	for _, v := range data {
		str, err := json.Marshal(v)
		if err != nil {
			return result, err
		}
		var r Data
		err = json.Unmarshal(str, &r)
		if err != nil {
			return result, err
		}
		resData = append(resData, r)
	}
	result.Data = resData
	return result, nil
}

// OcrFileAndParse processes the OCR for a given image file path and parses the result.
// It returns the parsed OCR results as a slice of Result structs and any error encountered.
func (p *Ppocr) OcrFileAndParse(imagePath string) (Result, error) {
	b, err := p.OcrFile(imagePath)
	if err != nil {
		return Result{}, err
	}
	return ParseResult(b)
}

// OcrAndParse processes and parses the OCR for a given image represented as a byte slice.
// It returns the parsed OCR results as a slice of Result structs and any error encountered.
func (p *Ppocr) OcrAndParse(image []byte) (Result, error) {
	b, err := p.Ocr(image)
	if err != nil {
		return Result{}, err
	}
	return ParseResult(b)
}

const ocrArgTag = "ocrArg"

// Deprecated: Only PaddleOCR-json v1.3.1 is supported.
//
// OcrClipboard processes the OCR for an image stored in the clipboard.
// It returns the raw OCR result as bytes and any error encountered.
func (p *Ppocr) OcrClipboard() ([]byte, error) {
	return p.OcrFile(clipboardImagePath)
}

// Deprecated: Only PaddleOCR-json v1.3.1 is supported.
//
// OcrClipboardAndParse processes the OCR for an image stored in the clipboard and parses the result.
// It returns the parsed OCR results as a slice of Result structs and any error encountered.
func (p *Ppocr) OcrClipboardAndParse() (Result, error) {
	return p.OcrFileAndParse(clipboardImagePath)
}
