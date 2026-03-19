package audio

import (
	"context"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/studio/platform/internal/domain/audiojob"
)

func TestLocalProcessorProcessVoiceEnhance(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sourceDir := filepath.Join(root, "audio")
	require.NoError(t, os.MkdirAll(sourceDir, 0o755))

	sourcePath := filepath.Join(sourceDir, "sample.wav")
	require.NoError(t, writeTestWAV(sourcePath, 16000, time.Second))

	processor := NewLocalProcessor(root, "/uploads")
	job := &audiojob.Job{
		ID:             uuid.New(),
		Title:          "voice enhance",
		TaskType:       audiojob.TaskTypeVoiceEnhance,
		SourceAudioURL: stringPtr("/uploads/audio/sample.wav"),
	}

	result, err := processor.Process(context.Background(), job)
	require.NoError(t, err)

	outputURL, ok := result["output_audio_url"].(string)
	require.True(t, ok)
	assert.Contains(t, outputURL, "/uploads/processed-audio/")

	sourceAnalysis, ok := result["source_analysis"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "wav", sourceAnalysis["format"])
	assert.Equal(t, 16000, sourceAnalysis["sample_rate"])
	assert.Equal(t, 1, sourceAnalysis["channels"])

	outputAnalysis, ok := result["output_analysis"].(map[string]any)
	require.True(t, ok)
	preview, ok := outputAnalysis["waveform_preview"].([]float64)
	require.True(t, ok)
	assert.NotEmpty(t, preview)

	outputPath := filepath.Join(root, "processed-audio", filepath.Base(outputURL))
	_, err = os.Stat(outputPath)
	require.NoError(t, err)
}

func TestLocalProcessorProcessAIMusic(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	processor := NewLocalProcessor(root, "/uploads")
	prompt := "做一首偏动漫感的电子流行歌，副歌抓耳"
	job := &audiojob.Job{
		ID:       uuid.New(),
		Title:    "ai music",
		TaskType: audiojob.TaskTypeAIMusic,
		Prompt:   &prompt,
		Params: map[string]any{
			"style_tags": []any{"bright", "female-vocal"},
		},
	}

	result, err := processor.Process(context.Background(), job)
	require.NoError(t, err)

	manifestURL, ok := result["composition_manifest_url"].(string)
	require.True(t, ok)
	assert.Contains(t, manifestURL, "/uploads/processed-audio/")
}

func writeTestWAV(path string, sampleRate int, duration time.Duration) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	const (
		channels      = 1
		bitsPerSample = 16
	)

	totalSamples := int(float64(sampleRate) * duration.Seconds())
	dataSize := totalSamples * channels * (bitsPerSample / 8)
	byteRate := sampleRate * channels * (bitsPerSample / 8)
	blockAlign := channels * (bitsPerSample / 8)
	riffSize := 36 + dataSize

	if _, err := file.WriteString("RIFF"); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint32(riffSize)); err != nil {
		return err
	}
	if _, err := file.WriteString("WAVE"); err != nil {
		return err
	}
	if _, err := file.WriteString("fmt "); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint32(16)); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint16(1)); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint16(channels)); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint32(sampleRate)); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint32(byteRate)); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint16(blockAlign)); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint16(bitsPerSample)); err != nil {
		return err
	}
	if _, err := file.WriteString("data"); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint32(dataSize)); err != nil {
		return err
	}

	for i := 0; i < totalSamples; i++ {
		sample := int16(math.Sin(2*math.Pi*440*float64(i)/float64(sampleRate)) * 28000)
		if err := binary.Write(file, binary.LittleEndian, sample); err != nil {
			return err
		}
	}
	return nil
}

func stringPtr(value string) *string {
	return &value
}
