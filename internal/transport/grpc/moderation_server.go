package grpc

import (
	"context"

	"github.com/studio/platform/internal/infra/moderation"
	moderationv1 "github.com/studio/platform/proto/moderation/v1"
)

// ModerationServer implements moderationv1.ModerationServiceServer.
type ModerationServer struct {
	moderationv1.UnimplementedModerationServiceServer
	moderator moderation.Moderator
}

// NewModerationServer creates a new ModerationServer.
func NewModerationServer(m moderation.Moderator) *ModerationServer {
	return &ModerationServer{moderator: m}
}

// ReviewText reviews text content.
func (s *ModerationServer) ReviewText(ctx context.Context, req *moderationv1.ReviewTextRequest) (*moderationv1.ReviewResult, error) {
	decision, reason, err := s.moderator.ReviewText(ctx, req.Text)
	if err != nil {
		return nil, err
	}
	var labels []string
	if reason != "" {
		labels = []string{reason}
	}
	return &moderationv1.ReviewResult{
		Decision: string(decision),
		Labels:   labels,
	}, nil
}

// ReviewImage reviews image content.
func (s *ModerationServer) ReviewImage(ctx context.Context, req *moderationv1.ReviewImageRequest) (*moderationv1.ReviewResult, error) {
	decision, reason, err := s.moderator.ReviewImage(ctx, req.Url)
	if err != nil {
		return nil, err
	}
	var labels []string
	if reason != "" {
		labels = []string{reason}
	}
	return &moderationv1.ReviewResult{
		Decision: string(decision),
		Labels:   labels,
	}, nil
}
