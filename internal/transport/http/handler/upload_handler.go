package handler

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/pkg/response"
)

// UploadHandler handles file uploads
type UploadHandler struct {
	uploadDir string
	maxSize   int64 // in bytes
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(uploadDir string, maxSize int64) *UploadHandler {
	return &UploadHandler{
		uploadDir: uploadDir,
		maxSize:   maxSize,
	}
}

// UploadImage uploads an image file
func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, fmt.Errorf("获取文件失败: %w", err))
		return
	}

	// Check file size
	if file.Size > h.maxSize {
		response.Error(c, fmt.Errorf("文件大小超过限制 (%dMB)", h.maxSize/1024/1024))
		return
	}

	// Check file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	if !allowedExts[ext] {
		response.Error(c, fmt.Errorf("不支持的文件格式: %s", ext))
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filepath := filepath.Join(h.uploadDir, "images", filename)

	// Save file
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		response.Error(c, fmt.Errorf("保存文件失败: %w", err))
		return
	}

	// Return URL
	url := fmt.Sprintf("/uploads/images/%s", filename)
	response.Success(c, gin.H{
		"url":      url,
		"filename": filename,
		"size":     file.Size,
	})
}

// UploadAudioFile uploads an audio file
func (h *UploadHandler) UploadAudioFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, fmt.Errorf("获取文件失败: %w", err))
		return
	}

	// Check file size
	maxAudioSize := int64(100 * 1024 * 1024) // 100MB
	if file.Size > maxAudioSize {
		response.Error(c, fmt.Errorf("文件大小超过限制 (100MB)"))
		return
	}

	// Check file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".mp3":  true,
		".wav":  true,
		".flac": true,
		".ogg":  true,
		".m4a":  true,
	}

	if !allowedExts[ext] {
		response.Error(c, fmt.Errorf("不支持的文件格式: %s", ext))
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filepath := filepath.Join(h.uploadDir, "audio", filename)

	// Save file
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		response.Error(c, fmt.Errorf("保存文件失败: %w", err))
		return
	}

	// Return URL
	url := fmt.Sprintf("/uploads/audio/%s", filename)
	response.Success(c, gin.H{
		"url":      url,
		"filename": filename,
		"size":     file.Size,
	})
}
