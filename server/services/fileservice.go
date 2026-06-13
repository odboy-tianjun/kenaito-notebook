package services

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const MaxFileSize = 6 * 1024 * 1024 // 6MB

type FileService struct {
	dataDir string
	baseURL string
}

func NewFileService(baseURL string) (*FileService, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dataDir := filepath.Join(homeDir, ".kenaito-notebook", "assets")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	log.Printf("[FileService] Assets directory: %s", dataDir)
	log.Printf("[FileService] Base URL: %s", baseURL)

	return &FileService{
		dataDir: dataDir,
		baseURL: strings.TrimRight(baseURL, "/"),
	}, nil
}

func (fs *FileService) GetBaseURL() string {
	return fs.baseURL
}

func (fs *FileService) SaveImage(base64Data string) (string, error) {
	parts := strings.Split(base64Data, ",")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid base64 format")
	}

	header := parts[0]
	ext := ".png"
	if strings.Contains(header, "jpeg") || strings.Contains(header, "jpg") {
		ext = ".jpg"
	} else if strings.Contains(header, "gif") {
		ext = ".gif"
	} else if strings.Contains(header, "webp") {
		ext = ".webp"
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	if int64(len(data)) > MaxFileSize {
		return "", fmt.Errorf("文件大小超过 6MB 限制")
	}

	filename := fmt.Sprintf("%s%s", uuid.New().String()[:8], ext)

	dateDir := time.Now().Format("2006/01")
	dirPath := filepath.Join(fs.dataDir, dateDir)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", err
	}

	filePath := filepath.Join(dirPath, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/api/files/assets/%s/%s", fs.baseURL, dateDir, filename), nil
}

func (fs *FileService) SaveFile(base64Data, filename string) (string, error) {
	parts := strings.Split(base64Data, ",")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid base64 format")
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	if int64(len(data)) > MaxFileSize {
		return "", fmt.Errorf("文件大小超过 6MB 限制")
	}

	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)
	uniqueFilename := fmt.Sprintf("%s_%s%s", nameWithoutExt, uuid.New().String()[:8], ext)

	dateDir := time.Now().Format("2006/01")
	dirPath := filepath.Join(fs.dataDir, dateDir)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", err
	}

	filePath := filepath.Join(dirPath, uniqueFilename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/api/files/assets/%s/%s", fs.baseURL, dateDir, uniqueFilename), nil
}

func (fs *FileService) GetAssetsPath() string {
	return fs.dataDir
}
