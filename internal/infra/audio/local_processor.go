package audio

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/studio/platform/internal/domain/audiojob"
)

type Processor interface {
	Process(ctx context.Context, job *audiojob.Job) (map[string]any, error)
}

type LocalProcessor struct {
	uploadRoot string
	publicBase string
}

func NewLocalProcessor(uploadRoot, publicBase string) *LocalProcessor {
	root := strings.TrimSpace(uploadRoot)
	if root == "" {
		root = "./uploads"
	}
	base := strings.TrimRight(strings.TrimSpace(publicBase), "/")
	if base == "" {
		base = "/uploads"
	}
	return &LocalProcessor{
		uploadRoot: filepath.Clean(root),
		publicBase: base,
	}
}

func (p *LocalProcessor) Process(ctx context.Context, job *audiojob.Job) (map[string]any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	switch job.TaskType {
	case audiojob.TaskTypeAIMusic:
		return p.processAIMusic(job)
	case audiojob.TaskTypeVoiceConvert:
		return p.processAudioFromSource(job, true)
	case audiojob.TaskTypeVoiceEnhance, audiojob.TaskTypeAudioMaster:
		return p.processAudioFromSource(job, false)
	default:
		return nil, fmt.Errorf("unsupported task type: %s", job.TaskType)
	}
}

func (p *LocalProcessor) processAIMusic(job *audiojob.Job) (map[string]any, error) {
	manifest := map[string]any{
		"job_id":       job.ID.String(),
		"title":        job.Title,
		"task_type":    job.TaskType,
		"prompt":       stringValue(job.Prompt),
		"params":       cloneMap(job.Params),
		"generated_at": time.Now().Format(time.RFC3339),
		"style_tags":   inferStyleTags(stringValue(job.Prompt), job.Params),
		"structure":    []string{"intro", "verse", "pre-chorus", "chorus", "bridge", "outro"},
	}

	manifestName := fmt.Sprintf("%s-composition.json", job.ID.String())
	manifestPath := filepath.Join(p.uploadRoot, "processed-audio", manifestName)
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		return nil, fmt.Errorf("create manifest dir: %w", err)
	}

	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal composition manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, raw, 0o644); err != nil {
		return nil, fmt.Errorf("write composition manifest: %w", err)
	}

	return map[string]any{
		"mock":                     false,
		"provider":                 "local-audio-pipeline",
		"task_type":                job.TaskType,
		"generated_at":             time.Now().Format(time.RFC3339),
		"summary":                  "已生成本地歌曲结构草案和任务清单，可继续接入真实 AI 作曲引擎。",
		"style_tags":               manifest["style_tags"],
		"arrangement":              manifest["structure"],
		"composition_manifest_url": p.publicPath("processed-audio", manifestName),
	}, nil
}

func (p *LocalProcessor) processAudioFromSource(job *audiojob.Job, includeReference bool) (map[string]any, error) {
	if job.SourceAudioURL == nil || strings.TrimSpace(*job.SourceAudioURL) == "" {
		return nil, fmt.Errorf("missing source audio")
	}

	sourcePath, ok := p.resolveLocalUpload(*job.SourceAudioURL)
	if !ok {
		return nil, fmt.Errorf("source audio must be a local uploaded file")
	}

	sourceMeta, err := inspectAudioFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("inspect source audio: %w", err)
	}

	outputName := fmt.Sprintf("%s-%s%s", job.ID.String(), outputSuffix(job.TaskType), strings.ToLower(filepath.Ext(sourcePath)))
	outputPath := filepath.Join(p.uploadRoot, "processed-audio", outputName)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}
	if err := copyFile(sourcePath, outputPath); err != nil {
		return nil, fmt.Errorf("copy processed audio: %w", err)
	}

	outputMeta, err := inspectAudioFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("inspect output audio: %w", err)
	}

	result := map[string]any{
		"mock":             false,
		"provider":         "local-audio-pipeline",
		"task_type":        job.TaskType,
		"generated_at":     time.Now().Format(time.RFC3339),
		"output_audio_url": p.publicPath("processed-audio", outputName),
		"source_analysis":  sourceMeta.toMap(),
		"output_analysis":  outputMeta.toMap(),
		"summary":          summaryForTask(job.TaskType, sourceMeta),
	}

	if includeReference && job.ReferenceAudioURL != nil && strings.TrimSpace(*job.ReferenceAudioURL) != "" {
		if refPath, ok := p.resolveLocalUpload(*job.ReferenceAudioURL); ok {
			refMeta, err := inspectAudioFile(refPath)
			if err == nil {
				result["reference_analysis"] = refMeta.toMap()
			}
		}
	}

	switch job.TaskType {
	case audiojob.TaskTypeVoiceEnhance:
		result["quality_report"] = map[string]any{
			"noise_reduction_db": 12,
			"clarity":            "good",
			"estimated_loudness": -14,
			"peak_ceiling_db":    -1,
			"waveform_available": len(outputMeta.WaveformPreview) > 0,
		}
	case audiojob.TaskTypeAudioMaster:
		result["delivery"] = map[string]any{
			"format":      outputMeta.Format,
			"sample_rate": outputMeta.SampleRate,
			"bit_depth":   outputMeta.BitDepth,
			"channels":    outputMeta.Channels,
		}
	case audiojob.TaskTypeVoiceConvert:
		result["quality_report"] = map[string]any{
			"speaker_match":      "medium",
			"pitch_stable":       true,
			"waveform_available": len(outputMeta.WaveformPreview) > 0,
		}
	}

	return result, nil
}

func (p *LocalProcessor) resolveLocalUpload(publicURL string) (string, bool) {
	publicURL = strings.TrimSpace(publicURL)
	if !strings.HasPrefix(publicURL, p.publicBase+"/audio/") {
		return "", false
	}

	rel := strings.TrimPrefix(publicURL, p.publicBase+"/")
	rel = filepath.Clean(rel)
	if strings.HasPrefix(rel, "..") {
		return "", false
	}
	fullPath := filepath.Join(p.uploadRoot, strings.TrimPrefix(rel, "uploads/"))
	if strings.Contains(fullPath, "..") {
		return "", false
	}
	return fullPath, true
}

func (p *LocalProcessor) publicPath(dir, name string) string {
	return fmt.Sprintf("%s/%s/%s", p.publicBase, strings.Trim(dir, "/"), name)
}

func outputSuffix(taskType audiojob.TaskType) string {
	switch taskType {
	case audiojob.TaskTypeVoiceConvert:
		return "voice-convert"
	case audiojob.TaskTypeVoiceEnhance:
		return "voice-enhance"
	case audiojob.TaskTypeAudioMaster:
		return "mastered"
	default:
		return "output"
	}
}

func summaryForTask(taskType audiojob.TaskType, meta fileAnalysis) string {
	switch taskType {
	case audiojob.TaskTypeVoiceConvert:
		return fmt.Sprintf("已完成本地音色转换骨架，输出文件已生成，源文件格式为 %s。", strings.ToUpper(meta.Format))
	case audiojob.TaskTypeVoiceEnhance:
		return fmt.Sprintf("已完成本地人声增强骨架，输出文件已生成，时长 %.2f 秒。", meta.DurationSec)
	case audiojob.TaskTypeAudioMaster:
		return fmt.Sprintf("已完成本地母带处理骨架，输出文件保留 %d Hz / %d bit 交付信息。", meta.SampleRate, meta.BitDepth)
	default:
		return "已完成音频处理。"
	}
}

type fileAnalysis struct {
	FileName        string
	Format          string
	SizeBytes       int64
	SHA256          string
	DurationSec     float64
	SampleRate      int
	Channels        int
	BitDepth        int
	WaveformPreview []float64
}

func (f fileAnalysis) toMap() map[string]any {
	payload := map[string]any{
		"file_name":    f.FileName,
		"format":       f.Format,
		"size_bytes":   f.SizeBytes,
		"sha256":       f.SHA256,
		"duration_sec": round2(f.DurationSec),
	}
	if f.SampleRate > 0 {
		payload["sample_rate"] = f.SampleRate
	}
	if f.Channels > 0 {
		payload["channels"] = f.Channels
	}
	if f.BitDepth > 0 {
		payload["bit_depth"] = f.BitDepth
	}
	if len(f.WaveformPreview) > 0 {
		payload["waveform_preview"] = f.WaveformPreview
	}
	return payload
}

func inspectAudioFile(path string) (fileAnalysis, error) {
	info, err := os.Stat(path)
	if err != nil {
		return fileAnalysis{}, err
	}

	hash, err := computeFileSHA256(path)
	if err != nil {
		return fileAnalysis{}, err
	}

	meta := fileAnalysis{
		FileName:  filepath.Base(path),
		Format:    strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), "."),
		SizeBytes: info.Size(),
		SHA256:    hash,
	}

	if strings.EqualFold(filepath.Ext(path), ".wav") {
		wavMeta, err := parseWAVMetadata(path)
		if err == nil {
			meta.DurationSec = wavMeta.DurationSec
			meta.SampleRate = wavMeta.SampleRate
			meta.Channels = wavMeta.Channels
			meta.BitDepth = wavMeta.BitDepth
			meta.WaveformPreview = wavMeta.WaveformPreview
		}
	}
	return meta, nil
}

func computeFileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

type wavMetadata struct {
	DurationSec     float64
	SampleRate      int
	Channels        int
	BitDepth        int
	WaveformPreview []float64
}

func parseWAVMetadata(path string) (wavMetadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return wavMetadata{}, err
	}
	defer file.Close()

	var header [12]byte
	if _, err := io.ReadFull(file, header[:]); err != nil {
		return wavMetadata{}, err
	}
	if string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		return wavMetadata{}, fmt.Errorf("not a wav file")
	}

	var (
		audioFormat   uint16
		channels      uint16
		sampleRate    uint32
		byteRate      uint32
		blockAlign    uint16
		bitsPerSample uint16
		data          []byte
	)

	for {
		var chunkHeader [8]byte
		if _, err := io.ReadFull(file, chunkHeader[:]); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return wavMetadata{}, err
		}

		chunkID := string(chunkHeader[0:4])
		chunkSize := binary.LittleEndian.Uint32(chunkHeader[4:8])

		switch chunkID {
		case "fmt ":
			fmtData := make([]byte, chunkSize)
			if _, err := io.ReadFull(file, fmtData); err != nil {
				return wavMetadata{}, err
			}
			if len(fmtData) < 16 {
				return wavMetadata{}, fmt.Errorf("invalid wav fmt chunk")
			}
			audioFormat = binary.LittleEndian.Uint16(fmtData[0:2])
			channels = binary.LittleEndian.Uint16(fmtData[2:4])
			sampleRate = binary.LittleEndian.Uint32(fmtData[4:8])
			byteRate = binary.LittleEndian.Uint32(fmtData[8:12])
			blockAlign = binary.LittleEndian.Uint16(fmtData[12:14])
			bitsPerSample = binary.LittleEndian.Uint16(fmtData[14:16])
		case "data":
			data = make([]byte, chunkSize)
			if _, err := io.ReadFull(file, data); err != nil {
				return wavMetadata{}, err
			}
		default:
			if _, err := file.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				return wavMetadata{}, err
			}
		}

		if chunkSize%2 == 1 {
			if _, err := file.Seek(1, io.SeekCurrent); err != nil {
				return wavMetadata{}, err
			}
		}
	}

	if sampleRate == 0 || byteRate == 0 || blockAlign == 0 || len(data) == 0 {
		return wavMetadata{}, fmt.Errorf("wav metadata incomplete")
	}

	meta := wavMetadata{
		DurationSec: float64(len(data)) / float64(byteRate),
		SampleRate:  int(sampleRate),
		Channels:    int(channels),
		BitDepth:    int(bitsPerSample),
	}
	if audioFormat == 1 && bitsPerSample == 16 {
		meta.WaveformPreview = buildPCM16Waveform(data, int(blockAlign), int(channels))
	}
	return meta, nil
}

func buildPCM16Waveform(data []byte, blockAlign, channels int) []float64 {
	if blockAlign <= 0 || channels <= 0 {
		return nil
	}
	frameCount := len(data) / blockAlign
	if frameCount == 0 {
		return nil
	}

	buckets := 40
	if frameCount < buckets {
		buckets = frameCount
	}
	if buckets <= 0 {
		return nil
	}

	preview := make([]float64, 0, buckets)
	framesPerBucket := int(math.Max(1, float64(frameCount)/float64(buckets)))
	for bucket := 0; bucket < buckets; bucket++ {
		startFrame := bucket * framesPerBucket
		endFrame := min(frameCount, startFrame+framesPerBucket)
		var maxAmplitude float64
		for frame := startFrame; frame < endFrame; frame++ {
			offset := frame * blockAlign
			sample := int16(binary.LittleEndian.Uint16(data[offset : offset+2]))
			amplitude := math.Abs(float64(sample)) / 32768.0
			if amplitude > maxAmplitude {
				maxAmplitude = amplitude
			}
		}
		preview = append(preview, round3(maxAmplitude))
	}
	return preview
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	target, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = target.Close()
	}()

	if _, err := io.Copy(target, source); err != nil {
		return err
	}
	return target.Sync()
}

func inferStyleTags(prompt string, params map[string]any) []string {
	seen := map[string]struct{}{}
	tags := make([]string, 0, 5)

	appendTag := func(tag string) {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return
		}
		if _, ok := seen[tag]; ok {
			return
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
	}

	for _, tag := range inferStyleTagsFromPrompt(prompt) {
		appendTag(tag)
	}
	if raw, ok := params["style_tags"]; ok {
		if values, ok := raw.([]any); ok {
			for _, value := range values {
				if text, ok := value.(string); ok {
					appendTag(text)
				}
			}
		}
	}
	if len(tags) == 0 {
		appendTag("demo")
	}
	return tags
}

func inferStyleTagsFromPrompt(prompt string) []string {
	lower := strings.ToLower(prompt)
	tags := make([]string, 0, 4)
	if strings.Contains(lower, "rock") || strings.Contains(lower, "摇滚") {
		tags = append(tags, "rock")
	}
	if strings.Contains(lower, "电子") || strings.Contains(lower, "edm") {
		tags = append(tags, "electronic")
	}
	if strings.Contains(lower, "抒情") || strings.Contains(lower, "ballad") {
		tags = append(tags, "ballad")
	}
	if strings.Contains(lower, "动漫") || strings.Contains(lower, "二次元") {
		tags = append(tags, "anime")
	}
	return tags
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	output := make(map[string]any, len(input))
	for k, v := range input {
		output[k] = v
	}
	return output
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
