package moderation

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"  //nolint:gosec // Aliyun Green API requires MD5
	"crypto/sha1" //nolint:gosec // Aliyun Green API requires SHA1
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Decision represents the result of a moderation check.
type Decision string

const (
	DecisionPass   Decision = "pass"
	DecisionBlock  Decision = "block"
	DecisionReview Decision = "review"
)

// Moderator is the interface for content moderation.
type Moderator interface {
	// ReviewText checks text content. Returns Decision and a reason.
	ReviewText(ctx context.Context, text string) (Decision, string, error)
	// ReviewImage checks an image URL. Returns Decision and a reason.
	ReviewImage(ctx context.Context, imageURL string) (Decision, string, error)
}

// AliyunGreen implements Moderator using Aliyun Content Safety (Green) API.
type AliyunGreen struct {
	accessKeyID     string
	accessKeySecret string
	// region endpoint, e.g. "green-cip.cn-shanghai.aliyuncs.com"
	endpoint string
	client   *http.Client
}

// NewAliyunGreen creates a new AliyunGreen moderator.
func NewAliyunGreen(accessKeyID, accessKeySecret, endpoint string) *AliyunGreen {
	if endpoint == "" {
		endpoint = "green-cip.cn-shanghai.aliyuncs.com"
	}
	return &AliyunGreen{
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		endpoint:        endpoint,
		client:          &http.Client{Timeout: 10 * time.Second},
	}
}

// ReviewText submits text to Aliyun Green for safety review.
func (g *AliyunGreen) ReviewText(ctx context.Context, text string) (Decision, string, error) {
	reqBody := map[string]any{
		"tasks": []map[string]any{
			{
				"dataId":  uuid.New().String(),
				"content": text,
			},
		},
		"scenes": []string{"antispam"},
	}
	result, err := g.callAPI(ctx, "/green/text/scan", reqBody)
	if err != nil {
		return DecisionReview, "", fmt.Errorf("AliyunGreen.ReviewText: %w", err)
	}
	return parseDecision(result), parseReason(result), nil
}

// ReviewImage submits an image URL to Aliyun Green for safety review.
func (g *AliyunGreen) ReviewImage(ctx context.Context, imageURL string) (Decision, string, error) {
	reqBody := map[string]any{
		"tasks": []map[string]any{
			{
				"dataId": uuid.New().String(),
				"url":    imageURL,
			},
		},
		"scenes": []string{"porn", "terrorism", "ad"},
	}
	result, err := g.callAPI(ctx, "/green/image/scan", reqBody)
	if err != nil {
		return DecisionReview, "", fmt.Errorf("AliyunGreen.ReviewImage: %w", err)
	}
	return parseDecision(result), parseReason(result), nil
}

// callAPI sends a request to the Aliyun Green API with HMAC-SHA1 authentication.
func (g *AliyunGreen) callAPI(ctx context.Context, path string, body any) (map[string]any, error) {
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	now := time.Now().UTC()
	date := now.Format("Mon, 02 Jan 2006 15:04:05 GMT")
	nonce := uuid.New().String()

	// MD5 of body
	h := md5.New() //nolint:gosec
	h.Write(bodyJSON)
	contentMD5 := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// String to sign
	stringToSign := fmt.Sprintf("POST\n%s\napplication/json; charset=UTF-8\n%s\nx-acs-signature-nonce:%s\nx-acs-version:2018-05-09\n%s",
		contentMD5, date, nonce, path)

	// HMAC-SHA1 signature
	mac := hmac.New(sha1.New, []byte(g.accessKeySecret)) //nolint:gosec
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	authHeader := fmt.Sprintf("acs %s:%s", g.accessKeyID, signature)
	url := fmt.Sprintf("https://%s%s", g.endpoint, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Date", date)
	req.Header.Set("Content-MD5", contentMD5)
	req.Header.Set("x-acs-signature-nonce", nonce)
	req.Header.Set("x-acs-version", "2018-05-09")
	req.Header.Set("Authorization", authHeader)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return result, nil
}

// parseDecision extracts the first task's suggestion from the API response.
func parseDecision(result map[string]any) Decision {
	data, _ := result["data"].([]any)
	if len(data) == 0 {
		return DecisionReview
	}
	task, _ := data[0].(map[string]any)
	results, _ := task["results"].([]any)
	if len(results) == 0 {
		return DecisionPass
	}
	for _, r := range results {
		rm, _ := r.(map[string]any)
		if rm["suggestion"] == "block" {
			return DecisionBlock
		}
	}
	return DecisionPass
}

func parseReason(result map[string]any) string {
	data, _ := result["data"].([]any)
	if len(data) == 0 {
		return ""
	}
	task, _ := data[0].(map[string]any)
	results, _ := task["results"].([]any)
	for _, r := range results {
		rm, _ := r.(map[string]any)
		if rm["suggestion"] == "block" {
			label, _ := rm["label"].(string)
			return label
		}
	}
	return ""
}
