package uploader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/schollz/progressbar/v3"
)

const (
	ChunkSize       = 5 * 1024 * 1024 // 5MB chunks
	MaxConcurrent   = 3                // Max concurrent uploads
	UploadCacheDir  = ".studio-cli/uploads"
)

// ChunkedUploader handles chunked file uploads
type ChunkedUploader struct {
	apiURL      string
	accessToken string
	cacheDir    string
}

// NewChunkedUploader creates a new chunked uploader
func NewChunkedUploader(apiURL, accessToken string) (*ChunkedUploader, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, UploadCacheDir)
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &ChunkedUploader{
		apiURL:      apiURL,
		accessToken: accessToken,
		cacheDir:    cacheDir,
	}, nil
}

// UploadState tracks upload progress
type UploadState struct {
	FilePath       string   `json:"file_path"`
	TotalChunks    int      `json:"total_chunks"`
	UploadedChunks []int    `json:"uploaded_chunks"`
	UploadID       string   `json:"upload_id"`
}

// Upload uploads a file in chunks with progress tracking
func (u *ChunkedUploader) Upload(filePath string, gameSlug string, version string) (string, error) {
	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	fileSize := fileInfo.Size()
	totalChunks := int((fileSize + ChunkSize - 1) / ChunkSize)

	fmt.Printf("📦 File size: %.2f MB\n", float64(fileSize)/(1024*1024))
	fmt.Printf("📊 Total chunks: %d\n", totalChunks)

	// Initialize upload
	uploadID, err := u.initUpload(gameSlug, version, fileInfo.Name(), fileSize)
	if err != nil {
		return "", fmt.Errorf("failed to initialize upload: %w", err)
	}

	// Create progress bar
	bar := progressbar.NewOptions64(
		fileSize,
		progressbar.OptionSetDescription("Uploading"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(50),
		progressbar.OptionThrottle(100),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
	)

	// Upload chunks
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Use semaphore for concurrent uploads
	sem := make(chan struct{}, MaxConcurrent)
	var wg sync.WaitGroup
	errChan := make(chan error, totalChunks)

	for i := 0; i < totalChunks; i++ {
		wg.Add(1)
		go func(chunkIndex int) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			offset := int64(chunkIndex) * ChunkSize
			size := ChunkSize
			if offset+int64(size) > fileSize {
				size = int(fileSize - offset)
			}

			chunk := make([]byte, size)
			if _, err := file.ReadAt(chunk, offset); err != nil && err != io.EOF {
				errChan <- fmt.Errorf("failed to read chunk %d: %w", chunkIndex, err)
				return
			}

			if err := u.uploadChunk(uploadID, chunkIndex, chunk); err != nil {
				errChan <- fmt.Errorf("failed to upload chunk %d: %w", chunkIndex, err)
				return
			}

			_ = bar.Add(size)
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	if err := <-errChan; err != nil {
		return "", err
	}

	// Finalize upload
	downloadURL, err := u.finalizeUpload(uploadID)
	if err != nil {
		return "", fmt.Errorf("failed to finalize upload: %w", err)
	}

	return downloadURL, nil
}

// initUpload initializes a chunked upload
func (u *ChunkedUploader) initUpload(gameSlug, version, filename string, fileSize int64) (string, error) {
	reqBody := map[string]any{
		"game_slug": gameSlug,
		"version":   version,
		"filename":  filename,
		"file_size": fileSize,
	}

	data, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", u.apiURL+"/api/v1/admin/uploads/init", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+u.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("init upload failed with status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			UploadID string `json:"upload_id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Data.UploadID, nil
}

// uploadChunk uploads a single chunk
func (u *ChunkedUploader) uploadChunk(uploadID string, chunkIndex int, data []byte) error {
	url := fmt.Sprintf("%s/api/v1/admin/uploads/%s/chunks/%d", u.apiURL, uploadID, chunkIndex)

	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+u.accessToken)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("upload chunk failed with status %d", resp.StatusCode)
	}

	return nil
}

// finalizeUpload finalizes the upload and returns download URL
func (u *ChunkedUploader) finalizeUpload(uploadID string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/admin/uploads/%s/finalize", u.apiURL, uploadID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+u.accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("finalize upload failed with status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			DownloadURL string `json:"download_url"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Data.DownloadURL, nil
}
