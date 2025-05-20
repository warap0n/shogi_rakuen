// usecase/game_usecase_test.go
package usecase_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"shogi-rakuen/model"
	"shogi-rakuen/usecase"
	"shogi-rakuen/usecase/input"
)

// --- フェイクリポジトリ ---
type FakeGameRepo struct {
	saved *model.Game
	store map[string]*model.Game
}

func NewFakeGameRepo() *FakeGameRepo {
	return &FakeGameRepo{store: make(map[string]*model.Game)}
}

func (f *FakeGameRepo) Save(g *model.Game) (*model.Game, error) {
	f.saved = g
	f.store[g.ID] = g
	return g, nil
}

func (f *FakeGameRepo) FindByID(id string) (*model.Game, error) {
	if g, ok := f.store[id]; ok {
		return g, nil
	}
	return nil, errors.New("not found")
}

// --- StartGame ---
func TestStartGame_Success(t *testing.T) {
	repo := NewFakeGameRepo()
	uc := usecase.NewGameUsecase(repo)

	in := input.StartGameInput{BlackID: "alice", WhiteID: "bob"}
	g, err := uc.StartGame(in)
	assert.NoError(t, err)

	saved := repo.saved
	assert.NotNil(t, saved)
	assert.Equal(t, "alice", saved.PlayerBlackID)
	assert.Equal(t, "bob", saved.PlayerWhiteID)
	assert.NotEmpty(t, saved.ID)
	assert.Same(t, saved, g)
}

func TestStartGame_InvalidPlayers(t *testing.T) {
	repo := NewFakeGameRepo()
	uc := usecase.NewGameUsecase(repo)

	t.Run("same player", func(t *testing.T) {
		_, err := uc.StartGame(input.StartGameInput{BlackID: "x", WhiteID: "x"})
		assert.ErrorIs(t, err, usecase.ErrSamePlayer)
	})

	t.Run("empty black ID", func(t *testing.T) {
		_, err := uc.StartGame(input.StartGameInput{BlackID: "", WhiteID: "y"})
		assert.ErrorIs(t, err, usecase.ErrInvalidPlayerID)
	})

	t.Run("empty white ID", func(t *testing.T) {
		_, err := uc.StartGame(input.StartGameInput{BlackID: "x", WhiteID: ""})
		assert.ErrorIs(t, err, usecase.ErrInvalidPlayerID)
	})
}

// --- GetGameByID ---
func TestGetGameByID_Success(t *testing.T) {
	repo := NewFakeGameRepo()
	uc := usecase.NewGameUsecase(repo)

	id := uuid.NewString()
	// あらかじめリポジトリに格納しておく
	game := model.NewGame()
	game.ID = id
	repo.store[id] = game

	got, err := uc.GetGameByID(id)
	assert.NoError(t, err)
	assert.Same(t, game, got)
}

func TestGetGameByID_NotFound(t *testing.T) {
	repo := NewFakeGameRepo()
	uc := usecase.NewGameUsecase(repo)

	_, err := uc.GetGameByID("no-such")
	assert.ErrorIs(t, err, usecase.ErrGameNotFound)
}

// --- ApplyMove ---
func TestApplyMove_Success(t *testing.T) {
	repo := NewFakeGameRepo()
	uc := usecase.NewGameUsecase(repo)

	id := uuid.NewString()
	g0 := model.NewGame()
	g0.ID = id
	repo.store[id] = g0

	in := input.ApplyMoveInput{
		GameID:    id,
		From:      model.Position{Rank: 6, File: 0},
		To:        model.Position{Rank: 5, File: 0},
		Promote:   false,
		Drop:      false,
		DropPiece: model.Pawn,
	}

	g1, err := uc.ApplyMove(in)
	assert.NoError(t, err)
	assert.Len(t, g1.Moves, 1)
	assert.Equal(t, model.White, g1.Turn)
	assert.Same(t, repo.saved, g1)
}

func TestApplyMove_GameNotFound(t *testing.T) {
	repo := NewFakeGameRepo()
	uc := usecase.NewGameUsecase(repo)

	_, err := uc.ApplyMove(input.ApplyMoveInput{GameID: "missing"})
	assert.ErrorIs(t, err, usecase.ErrGameNotFound)
}

func TestApplyMove_DomainError(t *testing.T) {
	repo := NewFakeGameRepo()
	uc := usecase.NewGameUsecase(repo)

	id := uuid.NewString()
	g0 := model.NewGame()
	g0.ID = id
	repo.store[id] = g0

	in := input.ApplyMoveInput{
		GameID: id,
		From:   model.Position{Rank: 5, File: 5},
		To:     model.Position{Rank: 6, File: 5},
	}

	_, err := uc.ApplyMove(in)
	assert.ErrorIs(t, err, usecase.ErrInvalidMove)
}

func TestApplyMove_AlreadyFinished(t *testing.T) {
	repo := NewFakeGameRepo()
	uc := usecase.NewGameUsecase(repo)

	id := uuid.NewString()
	g0 := model.NewGame()
	g0.ID = id
	g0.Finished = true
	repo.store[id] = g0

	in := input.ApplyMoveInput{
		GameID:    id,
		From:      model.Position{Rank: 6, File: 0},
		To:        model.Position{Rank: 5, File: 0},
		Promote:   false,
		Drop:      false,
		DropPiece: model.Pawn,
	}

	_, err := uc.ApplyMove(in)
	assert.ErrorIs(t, err, usecase.ErrGameAlreadyFinished)
}

// --- ListMoves ---
func TestListMoves_Success(t *testing.T) {
	repo := NewFakeGameRepo()
	uc := usecase.NewGameUsecase(repo)

	id := uuid.NewString()
	g0 := model.NewGame()
	g0.ID = id
	// 事前に何手か指しておく
	g0.Moves = []model.Move{
		{From: model.Position{Rank: 6, File: 0}, To: model.Position{Rank: 5, File: 0}},
		{From: model.Position{Rank: 6, File: 1}, To: model.Position{Rank: 5, File: 1}},
	}
	repo.store[id] = g0

	moves, err := uc.ListMoves(id)
	assert.NoError(t, err)
	assert.Len(t, moves, 2)
}

func TestListMoves_NotFound(t *testing.T) {
	repo := NewFakeGameRepo()
	uc := usecase.NewGameUsecase(repo)

	_, err := uc.ListMoves("absent")
	assert.ErrorIs(t, err, usecase.ErrGameNotFound)
}
