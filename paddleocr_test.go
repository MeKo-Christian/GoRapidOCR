package GoRapidOCR

import (
	"testing"
)

func TestOcrArgs_CmdString(t *testing.T) {
	tests := []struct {
		name string
		o    OcrArgs
		want string
	}{
		{"empty", OcrArgs{}, ""},
		{"ensureAscii", OcrArgs{EnsureAscii: "1"}, "--ensureAscii=1"},
		{"models", OcrArgs{Models: "custom_models"}, "--models=custom_models"},
		{"det", OcrArgs{Det: "custom_det.onnx"}, "--det=custom_det.onnx"},
		{"cls", OcrArgs{Cls: "custom_cls.onnx"}, "--cls=custom_cls.onnx"},
		{"rec", OcrArgs{Rec: "custom_rec.onnx"}, "--rec=custom_rec.onnx"},
		{"keys", OcrArgs{Keys: "custom_keys.txt"}, "--keys=custom_keys.txt"},
		{"doAngle", OcrArgs{DoAngle: "0"}, "--doAngle=0"},
		{"mostAngle", OcrArgs{MostAngle: "0"}, "--mostAngle=0"},
		{"numThread", OcrArgs{NumThread: "8"}, "--numThread=8"},
		{"padding", OcrArgs{Padding: "100"}, "--padding=100"},
		{"maxSideLen", OcrArgs{MaxSideLen: "2048"}, "--maxSideLen=2048"},
		{"boxScoreThresh", OcrArgs{BoxScoreThresh: "0.6"}, "--boxScoreThresh=0.6"},
		{"boxThresh", OcrArgs{BoxThresh: "0.4"}, "--boxThresh=0.4"},
		{"unClipRatio", OcrArgs{UnClipRatio: "2.0"}, "--unClipRatio=2.0"},
		{"imagePath", OcrArgs{ImagePath: "test.png"}, "--image_path=test.png"},
		{"multiple", OcrArgs{EnsureAscii: "1", Models: "models", Det: "det.onnx"}, "--ensureAscii=1 --models=models --det=det.onnx"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.CmdString(); got != tt.want {
				t.Errorf("OcrArgs.CmdString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPpocr(t *testing.T) {
	type args struct {
		exePath string
		args    OcrArgs
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"1", args{"",
			OcrArgs{}}, true},
		{"2", args{`E:\path\to\your\PaddleOCR-json.exe`,
			OcrArgs{}}, false},
		{"3", args{`.\PaddleOCR-json_v.1.3.1\PaddleOCR-json.exe`,
			OcrArgs{}}, false},
		{"3", args{`PaddleOCR-json_v.1.3.1\PaddleOCR-json.exe`,
			OcrArgs{}}, false},
		{"4", args{`.\PaddleOCR-json.exe`,
			OcrArgs{}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPpocr(tt.args.exePath, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPpocr() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
			if closeErr := p.Close(); closeErr != nil {
				t.Errorf("Close() error = %v", closeErr)
			}
		}
		})
	}
}
