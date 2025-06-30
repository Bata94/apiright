package core

import (
	"os"
)

func WriteDummyFile(filePath, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}

func RemoveDummyFile(filePath string) error {
	return os.Remove(filePath)
}

func CreateDummyDir(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

func RemoveDummyDir(dirPath string) error {
	return os.RemoveAll(dirPath)
}
