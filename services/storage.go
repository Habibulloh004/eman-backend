package services

import (
	"eman-backend/config"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type StorageService struct {
	cfg *config.Config
}

func NewStorageService(cfg *config.Config) *StorageService {
	return &StorageService{cfg: cfg}
}

// AllowedExtensions for file uploads
var AllowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".mp4":  true,
	".webm": true,
	".mov":  true,
}

// UploadFile saves an uploaded file to storage
func (s *StorageService) UploadFile(file *multipart.FileHeader) (string, error) {
	// Validate file size
	maxFileSize := int64(s.cfg.MaxUploadSizeMB) * 1024 * 1024
	if file.Size > maxFileSize {
		return "", fmt.Errorf("file too large, max size is %dMB", s.cfg.MaxUploadSizeMB)
	}

	// Get file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !AllowedExtensions[ext] {
		return "", fmt.Errorf("file type not allowed: %s", ext)
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s_%s%s", time.Now().Format("20060102"), uuid.New().String()[:8], ext)

	// Create upload directory if not exists
	uploadPath := filepath.Join(s.cfg.UploadDir, time.Now().Format("2006/01"))
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Full file path
	fullPath := filepath.Join(uploadPath, filename)

	// Open source file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy content
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Return relative path for URL
	relPath := filepath.Join(time.Now().Format("2006/01"), filename)
	return relPath, nil
}

// DeleteFile removes a file from storage
func (s *StorageService) DeleteFile(relativePath string) error {
	fullPath := filepath.Join(s.cfg.UploadDir, relativePath)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}

	return os.Remove(fullPath)
}

// GetFilePath returns the full filesystem path for a relative path
func (s *StorageService) GetFilePath(relativePath string) string {
	return filepath.Join(s.cfg.UploadDir, relativePath)
}
