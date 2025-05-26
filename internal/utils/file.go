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

func DoesFileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

func IsFileModifiedAfterCaching(cacheFileName string, fileName string) bool {
	cacheFileInfo, err := os.Stat(cacheFileName)
	if err != nil {
		return true
	}
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return true
	}
	return cacheFileInfo.ModTime().Before(fileInfo.ModTime())
}
