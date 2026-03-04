package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/studio/platform/internal/domain/game"
)

// MockGameRepository is a mock implementation of game.Repository
type MockGameRepository struct {
	mock.Mock
}

func (m *MockGameRepository) Create(ctx context.Context, g *game.Game) error {
	args := m.Called(ctx, g)
	return args.Error(0)
}

func (m *MockGameRepository) Update(ctx context.Context, g *game.Game) error {
	args := m.Called(ctx, g)
	return args.Error(0)
}

func (m *MockGameRepository) GetByID(ctx context.Context, id uuid.UUID) (*game.Game, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.Game), args.Error(1)
}

func (m *MockGameRepository) GetBySlug(ctx context.Context, slug string) (*game.Game, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.Game), args.Error(1)
}

func (m *MockGameRepository) List(ctx context.Context, filter game.ListFilter) ([]*game.Game, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*game.Game), args.Get(1).(int64), args.Error(2)
}

func (m *MockGameRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGameRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	args := m.Called(ctx, slug)
	return args.Bool(0), args.Error(1)
}

// TestCreateGame_Success tests successful game creation
func TestCreateGame_Success(t *testing.T) {
	mockRepo := new(MockGameRepository)
	service := NewGameService(mockRepo)

	ctx := context.Background()
	developerID := uuid.New()
	input := CreateGameInput{
		Slug:        "test-game",
		Title:       "Test Game",
		Description: stringPtr("A test game"),
		Genre:       []string{"Action"},
		Tags:        []string{"indie"},
		Engine:      "Unity",
	}

	mockRepo.On("ExistsBySlug", ctx, input.Slug).Return(false, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*game.Game")).Return(nil)

	result, err := service.CreateGame(ctx, developerID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, input.Slug, result.Slug)
	assert.Equal(t, input.Title, result.Title)
	assert.Equal(t, developerID, result.DeveloperID)
	mockRepo.AssertExpectations(t)
}

// TestCreateGame_DuplicateSlug tests game creation with duplicate slug
func TestCreateGame_DuplicateSlug(t *testing.T) {
	mockRepo := new(MockGameRepository)
	service := NewGameService(mockRepo)

	ctx := context.Background()
	developerID := uuid.New()
	input := CreateGameInput{
		Slug:   "test-game",
		Title:  "Test Game",
		Genre:  []string{"Action"},
		Tags:   []string{"indie"},
		Engine: "Unity",
	}

	mockRepo.On("ExistsBySlug", ctx, input.Slug).Return(true, nil)

	result, err := service.CreateGame(ctx, developerID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "游戏标识已存在")
	mockRepo.AssertExpectations(t)
}

// TestGetGameBySlug tests retrieving a game by slug
func TestGetGameBySlug(t *testing.T) {
	mockRepo := new(MockGameRepository)
	service := NewGameService(mockRepo)

	ctx := context.Background()
	slug := "test-game"
	expectedGame := &game.Game{
		ID:    uuid.New(),
		Slug:  slug,
		Title: "Test Game",
	}

	mockRepo.On("GetBySlug", ctx, slug).Return(expectedGame, nil)

	result, err := service.GetGameBySlug(ctx, slug)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedGame.ID, result.ID)
	assert.Equal(t, expectedGame.Slug, result.Slug)
	mockRepo.AssertExpectations(t)
}

// TestListGames_WithFilters tests listing games with filters
func TestListGames_WithFilters(t *testing.T) {
	mockRepo := new(MockGameRepository)
	service := NewGameService(mockRepo)

	ctx := context.Background()
	input := ListGamesInput{
		Page:     1,
		PageSize: 10,
		Genre:    "Action",
		Search:   "test",
	}

	expectedGames := []*game.Game{
		{ID: uuid.New(), Slug: "game1", Title: "Game 1"},
		{ID: uuid.New(), Slug: "game2", Title: "Game 2"},
	}

	mockRepo.On("List", ctx, mock.AnythingOfType("game.ListFilter")).Return(expectedGames, int64(2), nil)

	result, err := service.ListGames(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result.Games))
	assert.Equal(t, int64(2), result.Total)
	mockRepo.AssertExpectations(t)
}

// TestSearchGames tests searching games
func TestSearchGames(t *testing.T) {
	mockRepo := new(MockGameRepository)
	service := NewGameService(mockRepo)

	ctx := context.Background()
	query := "action"
	limit := 5

	expectedGames := []*game.Game{
		{ID: uuid.New(), Slug: "action-game", Title: "Action Game"},
	}

	mockRepo.On("List", ctx, mock.AnythingOfType("game.ListFilter")).Return(expectedGames, int64(1), nil)

	result, err := service.SearchGames(ctx, query, limit)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, len(result))
	mockRepo.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}

