package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/game"
	"github.com/studio/platform/internal/pkg/apperr"
)

// GameService handles game-related business logic
type GameService struct {
	gameRepo game.Repository
}

// NewGameService creates a new GameService
func NewGameService(gameRepo game.Repository) *GameService {
	return &GameService{
		gameRepo: gameRepo,
	}
}

// CreateGameInput represents input for creating a game
type CreateGameInput struct {
	Slug        string     `json:"slug" binding:"required"`
	Title       string     `json:"title" binding:"required"`
	Subtitle    *string    `json:"subtitle,omitempty"`
	Description *string    `json:"description,omitempty"`
	CoverKey    *string    `json:"cover_key,omitempty"`
	BannerKey   *string    `json:"banner_key,omitempty"`
	TrailerURL  *string    `json:"trailer_url,omitempty"`
	Genre       []string   `json:"genre"`
	Tags        []string   `json:"tags"`
	Engine      string     `json:"engine"`
	ReleaseDate *time.Time `json:"release_date,omitempty"`
}

// CreateGame creates a new game
func (s *GameService) CreateGame(ctx context.Context, developerID uuid.UUID, input CreateGameInput) (*game.Game, error) {
	// Check if slug exists
	exists, err := s.gameRepo.ExistsBySlug(ctx, input.Slug)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "检查游戏标识失败", err)
	}
	if exists {
		return nil, apperr.New(apperr.CodeInvalidParam, "游戏标识已存在")
	}

	// Create game entity
	g := &game.Game{
		ID:          uuid.New(),
		Slug:        input.Slug,
		Title:       input.Title,
		Subtitle:    input.Subtitle,
		Description: input.Description,
		CoverKey:    input.CoverKey,
		BannerKey:   input.BannerKey,
		TrailerURL:  input.TrailerURL,
		Genre:       input.Genre,
		Tags:        input.Tags,
		Engine:      input.Engine,
		Status:      game.StatusActive,
		ReleaseDate: input.ReleaseDate,
		DeveloperID: developerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database
	if err := s.gameRepo.Create(ctx, g); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "创建游戏失败", err)
	}

	return g, nil
}

// UpdateGameInput represents input for updating a game
type UpdateGameInput struct {
	Slug        *string    `json:"slug,omitempty"`
	Title       *string    `json:"title,omitempty"`
	Subtitle    *string    `json:"subtitle,omitempty"`
	Description *string    `json:"description,omitempty"`
	CoverKey    *string    `json:"cover_key,omitempty"`
	BannerKey   *string    `json:"banner_key,omitempty"`
	TrailerURL  *string    `json:"trailer_url,omitempty"`
	Genre       []string   `json:"genre,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	Engine      *string    `json:"engine,omitempty"`
	Status      *string    `json:"status,omitempty"`
	ReleaseDate *time.Time `json:"release_date,omitempty"`
}

// UpdateGame updates a game
func (s *GameService) UpdateGame(ctx context.Context, gameID uuid.UUID, input UpdateGameInput) (*game.Game, error) {
	// Get game
	g, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询游戏失败", err)
	}

	// Update fields if provided
	if input.Slug != nil && *input.Slug != g.Slug {
		exists, err := s.gameRepo.ExistsBySlug(ctx, *input.Slug)
		if err != nil {
			return nil, apperr.Wrap(apperr.CodeInternalError, "检查游戏标识失败", err)
		}
		if exists {
			return nil, apperr.New(apperr.CodeInvalidParam, "游戏标识已存在")
		}
		g.Slug = *input.Slug
	}

	if input.Title != nil {
		g.Title = *input.Title
	}

	if input.Subtitle != nil {
		g.Subtitle = input.Subtitle
	}

	if input.Description != nil {
		g.Description = input.Description
	}

	if input.CoverKey != nil {
		g.CoverKey = input.CoverKey
	}

	if input.BannerKey != nil {
		g.BannerKey = input.BannerKey
	}

	if input.TrailerURL != nil {
		g.TrailerURL = input.TrailerURL
	}

	if input.Genre != nil {
		g.Genre = input.Genre
	}

	if input.Tags != nil {
		g.Tags = input.Tags
	}

	if input.Engine != nil {
		g.Engine = *input.Engine
	}

	if input.Status != nil {
		g.Status = game.Status(*input.Status)
	}

	if input.ReleaseDate != nil {
		g.ReleaseDate = input.ReleaseDate
	}

	g.UpdatedAt = time.Now()

	// Save to database
	if err := s.gameRepo.Update(ctx, g); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "更新游戏失败", err)
	}

	return g, nil
}

// GetGameByID retrieves a game by ID
func (s *GameService) GetGameByID(ctx context.Context, gameID uuid.UUID) (*game.Game, error) {
	g, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询游戏失败", err)
	}
	return g, nil
}

// GetGameBySlug retrieves a game by slug
func (s *GameService) GetGameBySlug(ctx context.Context, slug string) (*game.Game, error) {
	g, err := s.gameRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询游戏失败", err)
	}
	return g, nil
}

// ListGamesInput represents input for listing games
type ListGamesInput struct {
	Page     int
	PageSize int
	Status   *game.Status
	Genre    string
	Tag      string
	Search   string
}

// ListGamesOutput represents output for listing games
type ListGamesOutput struct {
	Games []*game.Game `json:"games"`
	Total int64        `json:"total"`
	Page  int          `json:"page"`
	Size  int          `json:"size"`
}

// ListGames retrieves games with pagination and filters
func (s *GameService) ListGames(ctx context.Context, input ListGamesInput) (*ListGamesOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}

	filter := game.ListFilter{
		Page:     input.Page,
		PageSize: input.PageSize,
		Status:   input.Status,
		Genre:    input.Genre,
		Tag:      input.Tag,
		Search:   input.Search,
	}

	games, total, err := s.gameRepo.List(ctx, filter)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询游戏列表失败", err)
	}

	return &ListGamesOutput{
		Games: games,
		Total: total,
		Page:  input.Page,
		Size:  len(games),
	}, nil
}

// DeleteGame deletes a game
func (s *GameService) DeleteGame(ctx context.Context, gameID uuid.UUID) error {
	_, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "查询游戏失败", err)
	}

	if err := s.gameRepo.Delete(ctx, gameID); err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.CodeInternalError, "删除游戏失败", err)
	}

	return nil
}
