package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/studio/platform/internal/domain/post"
	"github.com/studio/platform/internal/infra/embedding"
	"github.com/studio/platform/internal/pkg/apperr"
)

const (
	userInterestVecTTL = time.Hour
	interestVecPrefix  = "user:interest_vec:"
)

// similarFinder is an optional interface a post.Repository can satisfy to
// support pgvector cosine-similarity queries.
type similarFinder interface {
	FindSimilar(ctx context.Context, vec []float64, limit int) ([]*post.Post, error)
}

// likedLister is an optional interface for fetching a user's liked posts.
type likedLister interface {
	ListLikedByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*post.Post, int64, error)
}

// RecommendationService computes personalised post recommendations.
type RecommendationService struct {
	postRepo post.Repository
	embedder embedding.Embedder
	redis    *redis.Client // may be nil
}

// NewRecommendationService creates a RecommendationService.
func NewRecommendationService(postRepo post.Repository, embedder embedding.Embedder, rdb *redis.Client) *RecommendationService {
	return &RecommendationService{
		postRepo: postRepo,
		embedder: embedder,
		redis:    rdb,
	}
}

// GetRecommended returns up to limit posts recommended for userID.
// Falls back to explore (latest approved) when no interest vector is available.
func (s *RecommendationService) GetRecommended(ctx context.Context, userID uuid.UUID, limit int) ([]*post.Post, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	finder, hasFinder := s.postRepo.(similarFinder)
	if !hasFinder {
		return s.fallback(ctx, limit)
	}

	vec, err := s.getUserInterestVec(ctx, userID)
	if err != nil || len(vec) == 0 {
		return s.fallback(ctx, limit)
	}

	posts, err := finder.FindSimilar(ctx, vec, limit)
	if err != nil || len(posts) == 0 {
		return s.fallback(ctx, limit)
	}
	return posts, nil
}

// UpdateUserInterestVec recomputes the user's interest vector from recent liked posts.
func (s *RecommendationService) UpdateUserInterestVec(ctx context.Context, userID uuid.UUID) error {
	lister, ok := s.postRepo.(likedLister)
	if !ok {
		return nil
	}

	recentLiked, _, err := lister.ListLikedByUser(ctx, userID, 1, 10)
	if err != nil || len(recentLiked) == 0 {
		return nil
	}

	mean, count := s.computeMeanVec(recentLiked)
	if count == 0 {
		return nil
	}

	if s.redis != nil {
		key := interestVecPrefix + userID.String()
		data, _ := json.Marshal(mean)
		_ = s.redis.Set(ctx, key, data, userInterestVecTTL).Err()
	}
	return nil
}

func (s *RecommendationService) getUserInterestVec(ctx context.Context, userID uuid.UUID) ([]float64, error) {
	if s.redis != nil {
		key := interestVecPrefix + userID.String()
		data, err := s.redis.Get(ctx, key).Bytes()
		if err == nil {
			var vec []float64
			if json.Unmarshal(data, &vec) == nil {
				return vec, nil
			}
		}
	}

	lister, ok := s.postRepo.(likedLister)
	if !ok {
		return nil, fmt.Errorf("repo does not support likedLister")
	}

	recentLiked, _, err := lister.ListLikedByUser(ctx, userID, 1, 10)
	if err != nil || len(recentLiked) == 0 {
		return nil, fmt.Errorf("no liked posts")
	}

	mean, count := s.computeMeanVec(recentLiked)
	if count == 0 {
		return nil, fmt.Errorf("no embeddings computed")
	}

	if s.redis != nil {
		key := interestVecPrefix + userID.String()
		data, _ := json.Marshal(mean)
		_ = s.redis.Set(ctx, key, data, userInterestVecTTL).Err()
	}
	return mean, nil
}

func (s *RecommendationService) computeMeanVec(posts []*post.Post) ([]float64, int) {
	dims := s.embedder.Dims()
	mean := make([]float64, dims)
	count := 0

	for _, p := range posts {
		text := p.Title + " " + p.Content
		vec, err := s.embedder.Embed(text)
		if err != nil || len(vec) != dims {
			continue
		}
		for i, v := range vec {
			mean[i] += v
		}
		count++
	}

	if count > 0 {
		for i := range mean {
			mean[i] /= float64(count)
		}
	}
	return mean, count
}

func (s *RecommendationService) fallback(ctx context.Context, limit int) ([]*post.Post, error) {
	vis := post.VisibilityPublic
	mod := post.ModerationApproved
	posts, _, err := s.postRepo.List(ctx, post.ListFilter{
		Visibility:       &vis,
		ModerationStatus: &mod,
		Page:             1,
		PageSize:         limit,
	})
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询推荐内容失败", err)
	}
	return posts, nil
}
