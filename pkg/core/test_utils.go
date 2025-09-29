package core

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
)

// DummyFile represents a dummy file created for testing purposes.
type DummyFile struct {
	Body        *bytes.Buffer
	ContentType string
}

func CreateDummyMultipartFile(fileName, fileContent string) (*DummyFile, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, err
	}

	_, err = io.WriteString(part, fileContent)
	if err != nil {
		return nil, err
	}

	writer.Close()

	return &DummyFile{
		Body:        body,
		ContentType: writer.FormDataContentType(),
	}, nil
}

func ReadDummyFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

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
