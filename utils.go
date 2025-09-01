package GoRapidOCR

import "os"

// Check if file exists
func fileIsExist(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	if f.IsDir() {
		return false
	}
	return true
}
