package handlers

import (
	"bytes"
	"errors"
	"net/url"
	"strings"

	"eman-backend/services"

	"github.com/gofiber/fiber/v2"
)

var errNoFileUploaded = errors.New("no file uploaded")

func uploadFromRequest(c *fiber.Ctx, storage *services.StorageService) (string, error) {
	contentType := c.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		file, err := c.FormFile("file")
		if err != nil {
			return "", errNoFileUploaded
		}
		return storage.UploadFile(file)
	}

	filename := c.Get("X-Filename")
	if filename == "" {
		filename = c.Query("filename")
	}
	if filename == "" {
		return "", errors.New("missing X-Filename header")
	}
	if decoded, err := url.QueryUnescape(filename); err == nil {
		filename = decoded
	}

	contentLength := int64(c.Request().Header.ContentLength())
	body := c.Request().BodyStream()
	if body == nil {
		raw := c.Body()
		if len(raw) == 0 {
			return "", errors.New("request body is empty")
		}
		body = bytes.NewReader(raw)
		if contentLength <= 0 {
			contentLength = int64(len(raw))
		}
	}

	if contentLength <= 0 {
		return "", errors.New("content length required")
	}

	return storage.UploadStream(filename, contentType, contentLength, body)
}
