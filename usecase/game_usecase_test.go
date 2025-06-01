package usecase_test

import (
	"testing"

	"shogi-rakuen/model"
	"shogi-rakuen/repository"
	"shogi-rakuen/usecase"
	"shogi-rakuen/usecase/input"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGameRepository struct {
	mock.Mock
}

func (m *MockGameRepository) CreateGame(g *model.Game) (*model.Game, error) {
	args := m.Called(g)
	return args.Get(0).(*model.Game), args.Error(1)
}

func (m *MockGameRepository) FindByID(id string) (*model.Game, error) {
	args := m.Called(id)
	return args.Get(0).(*model.Game), args.Error(1)
}

func (m *MockGameRepository) AppendMove(g *model.Game, move model.Move) (*model.Game, error) {
	args := m.Called(g, move)
	return args.Get(0).(*model.Game), args.Error(1)
}

func setup() (*MockGameRepository, usecase.IGameUsecase) {
	mockRepo := new(MockGameRepository)
	uc := usecase.NewGameUsecase(mockRepo)
	return mockRepo, uc
}

func newTestGame() *model.Game {
	g, _ := model.NewGameWithPlayers("black", "white")
	g.ID = "game123"
	return g
}

// --- Normal Cases ---

func TestStartGame(t *testing.T) {
	mockRepo, uc := setup()

	input := input.StartGameInput{BlackID: "black", WhiteID: "white"}
	expected, _ := model.NewGameWithPlayers("black", "white")
	expected.ID = uuid.NewString()

	mockRepo.On("CreateGame", mock.AnythingOfType("*model.Game")).Return(expected, nil)

	game, err := uc.StartGame(input)

	assert.NoError(t, err)
	assert.Equal(t, expected.ID, game.ID)
	mockRepo.AssertExpectations(t)
}

func TestGetGameByID(t *testing.T) {
	mockRepo, uc := setup()
	game := newTestGame()

	mockRepo.On("FindByID", "game123").Return(game, nil)

	result, err := uc.GetGameByID("game123")

	assert.NoError(t, err)
	assert.Equal(t, "game123", result.ID)
}

func TestApplyMove(t *testing.T) {
	mockRepo, uc := setup()
	game := newTestGame()

	input := input.ApplyMoveInput{
		GameID:  "game123",
		From:    model.Position{Rank: 7, File: 7},
		To:      model.Position{Rank: 7, File: 6},
		Promote: false,
	}

	expectedMove := model.Move{
		From:    input.From,
		To:      input.To,
		Promote: input.Promote,
	}

	mockRepo.On("FindByID", "game123").Return(game, nil)
	mockRepo.On("AppendMove", game, expectedMove).Return(game, nil)

	result, err := uc.ApplyMove(input)

	assert.NoError(t, err)
	assert.Equal(t, "game123", result.ID)
	mockRepo.AssertExpectations(t)
}

func TestListMoves(t *testing.T) {
	mockRepo, uc := setup()
	game := newTestGame()
	game.Moves = []model.Move{
		{From: model.Position{Rank: 7, File: 7}, To: model.Position{Rank: 7, File: 6}},
		{From: model.Position{Rank: 3, File: 3}, To: model.Position{Rank: 3, File: 4}},
	}

	mockRepo.On("FindByID", "game123").Return(game, nil)

	moves, err := uc.ListMoves("game123")

	assert.NoError(t, err)
	assert.Len(t, moves, 2)
	assert.Equal(t, model.Position{Rank: 7, File: 7}, moves[0].From)
	assert.Equal(t, model.Position{Rank: 3, File: 3}, moves[1].From)
}

// --- Error Cases ---

func TestStartGame_SamePlayerError(t *testing.T) {
	_, uc := setup()

	input := input.StartGameInput{BlackID: "same", WhiteID: "same"}

	game, err := uc.StartGame(input)

	assert.Nil(t, game)
	assert.ErrorIs(t, err, usecase.ErrSamePlayer)
}

func TestGetGameByID_NotFound(t *testing.T) {
	mockRepo, uc := setup()

	// 型を明示して nil を返す
	mockRepo.On("FindByID", "not-found").Return((*model.Game)(nil), repository.ErrGameNotFound)

	result, err := uc.GetGameByID("not-found")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, usecase.ErrGameNotFound)
}

func TestApplyMove_AlreadyFinished(t *testing.T) {
	mockRepo, uc := setup()
	game := newTestGame()
	game.Finished = true

	mockRepo.On("FindByID", "game123").Return(game, nil)

	input := input.ApplyMoveInput{
		GameID: "game123",
		From:   model.Position{Rank: 7, File: 7},
		To:     model.Position{Rank: 7, File: 6},
	}

	result, err := uc.ApplyMove(input)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, usecase.ErrGameAlreadyFinished)
}

func TestApplyMove_InvalidMove(t *testing.T) {
	mockRepo, uc := setup()

	game, _ := model.NewGameWithPlayers("black", "white")
	game.ID = "game123"

	// 盤上に駒が存在しないマスから動かす
	input := input.ApplyMoveInput{
		GameID: "game123",
		From:   model.Position{Rank: 5, File: 5},
		To:     model.Position{Rank: 5, File: 4},
	}

	mockRepo.On("FindByID", "game123").Return(game, nil)

	result, err := uc.ApplyMove(input)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, usecase.ErrInvalidMove)
}

func TestListMoves_GameNotFound(t *testing.T) {
	mockRepo, uc := setup()

	// 型を明示して nil を返す
	mockRepo.On("FindByID", "notfound").Return((*model.Game)(nil), repository.ErrGameNotFound)

	result, err := uc.ListMoves("notfound")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, usecase.ErrGameNotFound)
}
