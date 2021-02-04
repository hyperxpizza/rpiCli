package filestorage

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/sys/unix"
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

func (storage *FileStorage) CheckCapacity() uint64 {
	var stat unix.Statfs_t

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	unix.Statfs(wd, &stat)

	// Available blocks * size per block = available space in bytes
	return stat.Bavail * uint64(stat.Bsize)
}

func (storage *FileStorage) SearchFile(searchFile string) (string, error) {
	var files []string

	err := filepath.Walk(storage.fileFolder, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})

	if err != nil {
		return "", err
	}

	for _, file := range files {
		fmt.Println(file)

		if file == searchFile {
			return file, nil
		}
	}

	return "", fmt.Errorf("File not found")
}
