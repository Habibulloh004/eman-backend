package services

import (
	"bufio"
	"eman-backend/config"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chai2010/webp"
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
	".bmp":  true,
	".tif":  true,
	".tiff": true,
	".mp4":  true,
	".webm": true,
	".mov":  true,
	".mp3":  true,
	".wav":  true,
	".ogg":  true,
	".m4a":  true,
	".aac":  true,
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".ppt":  true,
	".pptx": true,
	".txt":  true,
	".rtf":  true,
	".csv":  true,
	".odt":  true,
	".ods":  true,
	".odp":  true,
	".zip":  true,
	".rar":  true,
	".7z":   true,
}

var AllowedMimeTypes = map[string]bool{
	"image/jpeg":         true,
	"image/jpg":          true,
	"image/png":          true,
	"image/gif":          true,
	"image/webp":         true,
	"image/bmp":          true,
	"image/tiff":         true,
	"video/mp4":          true,
	"video/webm":         true,
	"video/quicktime":    true,
	"application/mp4":    true,
	"application/webm":   true,
	"audio/mpeg":         true,
	"audio/mp3":          true,
	"audio/wav":          true,
	"audio/x-wav":        true,
	"audio/wave":         true,
	"audio/ogg":          true,
	"audio/mp4":          true,
	"audio/x-m4a":        true,
	"audio/aac":          true,
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
	"application/vnd.ms-powerpoint":                                             true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"text/plain":      true,
	"text/csv":        true,
	"text/rtf":        true,
	"application/rtf": true,
	"application/vnd.oasis.opendocument.text":         true,
	"application/vnd.oasis.opendocument.spreadsheet":  true,
	"application/vnd.oasis.opendocument.presentation": true,
	"application/zip":              true,
	"application/x-zip-compressed": true,
	"application/vnd.rar":          true,
	"application/x-rar-compressed": true,
	"application/x-7z-compressed":  true,
}

var ConvertibleToWebP = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

// UploadFile saves an uploaded file to storage
func (s *StorageService) UploadFile(file *multipart.FileHeader) (string, error) {
	// Validate file size
	maxFileSize := int64(s.cfg.MaxUploadSizeMB) * 1024 * 1024
	if file.Size > maxFileSize {
		return "", fmt.Errorf("file too large, max size is %dMB", s.cfg.MaxUploadSizeMB)
	}

	// Open source file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	contentType := ""
	if file.Header != nil {
		contentType = file.Header.Get("Content-Type")
	}

	return s.saveFromReader(src, file.Filename, contentType)
}

// UploadStream saves a streamed upload to storage.
func (s *StorageService) UploadStream(filename string, contentType string, size int64, body io.Reader) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("missing filename")
	}

	maxFileSize := int64(s.cfg.MaxUploadSizeMB) * 1024 * 1024
	if size <= 0 {
		return "", fmt.Errorf("content length required")
	}
	if size > maxFileSize {
		return "", fmt.Errorf("file too large, max size is %dMB", s.cfg.MaxUploadSizeMB)
	}

	return s.saveFromReader(body, filename, contentType)
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

func (s *StorageService) saveFromReader(reader io.Reader, originalName string, contentType string) (string, error) {
	normalizedContentType := normalizeContentType(contentType)

	buffered := bufio.NewReader(reader)
	if normalizedContentType == "" {
		if sniff, err := buffered.Peek(512); err == nil || len(sniff) > 0 {
			normalizedContentType = normalizeContentType(http.DetectContentType(sniff))
		}
	}

	ext := strings.ToLower(filepath.Ext(originalName))
	if ext == "" {
		ext = extensionFromContentType(normalizedContentType)
	}

	if !isAllowedType(ext, normalizedContentType) {
		if ext == "" {
			return "", fmt.Errorf("file type not allowed")
		}
		return "", fmt.Errorf("file type not allowed: %s", ext)
	}

	now := time.Now()
	uploadPath := filepath.Join(s.cfg.UploadDir, now.Format("2006/01"))
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	baseName := fmt.Sprintf("%s_%s", now.Format("20060102"), uuid.New().String()[:8])

	if shouldConvertToWebP(ext, normalizedContentType) {
		filename := baseName + ".webp"
		fullPath := filepath.Join(uploadPath, filename)

		if err := s.writeWebP(fullPath, buffered); err != nil {
			return "", err
		}

		return filepath.Join(now.Format("2006/01"), filename), nil
	}

	if ext == "" {
		return "", fmt.Errorf("missing file extension")
	}

	filename := baseName + ext
	fullPath := filepath.Join(uploadPath, filename)
	if err := writeStream(fullPath, buffered); err != nil {
		return "", err
	}

	return filepath.Join(now.Format("2006/01"), filename), nil
}

func (s *StorageService) writeWebP(fullPath string, reader io.Reader) error {
	img, _, err := image.Decode(reader)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	quality := s.cfg.WebPQuality
	if quality <= 0 {
		quality = 85
	}
	if quality > 100 {
		quality = 100
	}

	options := &webp.Options{
		Quality: float32(quality),
	}

	if s.cfg.WebPLossless {
		options.Lossless = true
	}
	if s.cfg.WebPExact {
		options.Exact = true
	}

	if err := webp.Encode(dst, img, options); err != nil {
		return fmt.Errorf("failed to encode webp: %w", err)
	}

	return nil
}

func writeStream(fullPath string, reader io.Reader) error {
	dst, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

func shouldConvertToWebP(ext string, contentType string) bool {
	contentType = normalizeContentType(contentType)
	switch contentType {
	case "image/jpeg", "image/jpg", "image/png":
		return true
	case "image/gif", "image/webp", "image/bmp", "image/tiff":
		return false
	}
	if ext == ".gif" || ext == ".webp" {
		return false
	}
	return ConvertibleToWebP[ext]
}

func normalizeContentType(contentType string) string {
	if contentType == "" {
		return ""
	}
	parts := strings.Split(contentType, ";")
	return strings.TrimSpace(strings.ToLower(parts[0]))
}

func extensionFromContentType(contentType string) string {
	switch contentType {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/bmp":
		return ".bmp"
	case "image/tiff":
		return ".tiff"
	case "video/mp4", "application/mp4":
		return ".mp4"
	case "video/webm", "application/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	case "audio/mpeg", "audio/mp3":
		return ".mp3"
	case "audio/wav", "audio/x-wav", "audio/wave":
		return ".wav"
	case "audio/ogg":
		return ".ogg"
	case "audio/mp4", "audio/x-m4a":
		return ".m4a"
	case "audio/aac":
		return ".aac"
	case "application/pdf":
		return ".pdf"
	case "application/msword":
		return ".doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return ".docx"
	case "application/vnd.ms-excel":
		return ".xls"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return ".xlsx"
	case "application/vnd.ms-powerpoint":
		return ".ppt"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return ".pptx"
	case "text/plain":
		return ".txt"
	case "text/csv":
		return ".csv"
	case "text/rtf", "application/rtf":
		return ".rtf"
	case "application/vnd.oasis.opendocument.text":
		return ".odt"
	case "application/vnd.oasis.opendocument.spreadsheet":
		return ".ods"
	case "application/vnd.oasis.opendocument.presentation":
		return ".odp"
	case "application/zip", "application/x-zip-compressed":
		return ".zip"
	case "application/vnd.rar", "application/x-rar-compressed":
		return ".rar"
	case "application/x-7z-compressed":
		return ".7z"
	default:
		return ""
	}
}

func isAllowedType(ext string, contentType string) bool {
	if ext != "" && AllowedExtensions[ext] {
		return true
	}
	if contentType != "" && AllowedMimeTypes[contentType] {
		return true
	}
	return false
}
