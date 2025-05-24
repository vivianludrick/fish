package utils

import (
	"fmt"
	"os"
	"strings"
)

func GetCacheFileName(cacheDir string, fileName string) string {
	fileName = fileName[strings.LastIndex(fileName, "/")+1:]
	return fmt.Sprintf("%s/%s.gob", cacheDir, fileName)
}

func EnsureDir(dirName string) error {
	return os.MkdirAll(dirName, 0755)
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}
