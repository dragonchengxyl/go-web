package handler

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

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

// UploadGameFile uploads a game file
func (h *UploadHandler) UploadGameFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, fmt.Errorf("获取文件失败: %w", err))
		return
	}

	// Check file size (allow larger files for games)
	maxGameSize := int64(2 * 1024 * 1024 * 1024) // 2GB
	if file.Size > maxGameSize {
		response.Error(c, fmt.Errorf("文件大小超过限制 (2GB)"))
		return
	}

	// Check file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".zip": true,
		".rar": true,
		".7z":  true,
		".exe": true,
		".dmg": true,
		".pkg": true,
	}

	if !allowedExts[ext] {
		response.Error(c, fmt.Errorf("不支持的文件格式: %s", ext))
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filepath := filepath.Join(h.uploadDir, "games", filename)

	// Save file
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		response.Error(c, fmt.Errorf("保存文件失败: %w", err))
		return
	}

	// Return URL
	url := fmt.Sprintf("/uploads/games/%s", filename)
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

// UploadChunk handles chunked file upload for large files
func (h *UploadHandler) UploadChunk(c *gin.Context) {
	uploadID := c.PostForm("upload_id")
	chunkIndex := c.PostForm("chunk_index")
	totalChunks := c.PostForm("total_chunks")

	if uploadID == "" {
		uploadID = uuid.New().String()
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, fmt.Errorf("获取文件失败: %w", err))
		return
	}

	// Save chunk
	chunkPath := filepath.Join(h.uploadDir, "chunks", uploadID, chunkIndex)
	if err := c.SaveUploadedFile(file, chunkPath); err != nil {
		response.Error(c, fmt.Errorf("保存分片失败: %w", err))
		return
	}

	response.Success(c, gin.H{
		"upload_id":    uploadID,
		"chunk_index":  chunkIndex,
		"total_chunks": totalChunks,
	})
}

// MergeChunks merges uploaded chunks into a single file
func (h *UploadHandler) MergeChunks(c *gin.Context) {
	uploadID := c.PostForm("upload_id")
	totalChunks := c.PostForm("total_chunks")
	filename := c.PostForm("filename")

	if uploadID == "" || totalChunks == "" || filename == "" {
		response.Error(c, fmt.Errorf("缺少必要参数"))
		return
	}

	// Create output file
	ext := filepath.Ext(filename)
	outputFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	outputPath := filepath.Join(h.uploadDir, "games", outputFilename)

	output, err := c.Writer.(*gin.ResponseWriter).File(outputPath)
	if err != nil {
		response.Error(c, fmt.Errorf("创建输出文件失败: %w", err))
		return
	}
	defer output.Close()

	// Merge chunks
	chunksDir := filepath.Join(h.uploadDir, "chunks", uploadID)
	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(chunksDir, fmt.Sprintf("%d", i))
		chunk, err := c.Writer.(*gin.ResponseWriter).File(chunkPath)
		if err != nil {
			response.Error(c, fmt.Errorf("读取分片失败: %w", err))
			return
		}

		if _, err := io.Copy(output, chunk); err != nil {
			chunk.Close()
			response.Error(c, fmt.Errorf("合并分片失败: %w", err))
			return
		}
		chunk.Close()
	}

	// Clean up chunks
	// os.RemoveAll(chunksDir)

	url := fmt.Sprintf("/uploads/games/%s", outputFilename)
	response.Success(c, gin.H{
		"url":      url,
		"filename": outputFilename,
	})
}
