package filestorage

import (
	"bytes"
	"fmt"
	"os"
	"sync"
)

type FileInfo struct {
	Name string
	Type string
	Path string
}

type FileStorage struct {
	mutex      sync.RWMutex
	fileFolder string
	files      map[string]*FileInfo
}

func NewFileStorage(fileFolder string) *FileStorage {
	return &FileStorage{
		fileFolder: fileFolder,
		files:      make(map[string]*FileInfo),
	}
}

func (storage *FileStorage) Save(fileName, fileType string, data bytes.Buffer) (string, error) {
	filePath := fmt.Sprintf("%s/%s%s", storage.fileFolder, fileName, fileType)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("Can not create file: %w", err)
	}

	_, err = data.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("Error while writing data: %w", err)
	}

	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	storage.files[fileName] = &FileInfo{
		Name: fileName,
		Type: fileType,
		Path: filePath,
	}

	return fileName, nil
}
