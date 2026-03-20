package oss

import (
	"context"
	"testing"
	"time"

	"github.com/studio/platform/configs"
)

func TestLocalStorageGeneratePresignedURL(t *testing.T) {
	t.Parallel()

	storage := NewLocalStorage(configs.OSSConfig{Provider: "local"})

	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "absolute http url",
			key:  "https://cdn.example.com/audio/demo.mp3",
			want: "https://cdn.example.com/audio/demo.mp3",
		},
		{
			name: "leading slash path",
			key:  "/uploads/audio/demo.mp3",
			want: "/uploads/audio/demo.mp3",
		},
		{
			name: "uploads relative path",
			key:  "uploads/images/demo.webp",
			want: "/uploads/images/demo.webp",
		},
		{
			name: "plain relative path",
			key:  "processed-audio/demo.wav",
			want: "/uploads/processed-audio/demo.wav",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := storage.GeneratePresignedURL(context.Background(), tc.key, time.Hour)
			if err != nil {
				t.Fatalf("GeneratePresignedURL() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("GeneratePresignedURL() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLocalStorageGenerateUploadPolicyUnsupported(t *testing.T) {
	t.Parallel()

	storage := NewLocalStorage(configs.OSSConfig{Provider: "local"})
	_, err := storage.GenerateUploadPolicy(context.Background(), "posts/demo/", time.Hour)
	if err == nil {
		t.Fatal("expected GenerateUploadPolicy to fail in local mode")
	}
}
