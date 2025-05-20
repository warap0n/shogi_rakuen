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

	// Save に渡された値が repo.saved に入っている
	saved := repo.saved
	assert.NotNil(t, saved)
	assert.Equal(t, "alice", saved.PlayerBlackID)
	assert.Equal(t, "bob", saved.PlayerWhiteID)
	assert.NotEmpty(t, saved.ID)

	// ユースケースの戻り値も同一ポインタ
	assert.Same(t, saved, g)
}

func TestStartGame_InvalidPlayers(t *testing.T) {
	repo := NewFakeGameRepo()
	uc := usecase.NewGameUsecase(repo)

	// 同じ ID
	_, err := uc.StartGame(input.StartGameInput{BlackID: "x", WhiteID: "x"})
	assert.Error(t, err)

	// 空文字
	_, err = uc.StartGame(input.StartGameInput{BlackID: "", WhiteID: "y"})
	assert.Error(t, err)
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
	// 初期ゲームを repo に登録
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

	// 1手指されていること
	assert.Len(t, g1.Moves, 1)
	// ターンが切り替わっていること
	assert.Equal(t, model.White, g1.Turn)
	// リポジトリに保存されたのは同一オブジェクト
	assert.Same(t, repo.saved, g1)
}

func TestApplyMove_GameNotFound(t *testing.T) {
	repo := NewFakeGameRepo()
	uc := usecase.NewGameUsecase(repo)

	in := input.ApplyMoveInput{GameID: "missing"}
	_, err := uc.ApplyMove(in)
	assert.ErrorIs(t, err, usecase.ErrGameNotFound)
}

func TestApplyMove_DomainError(t *testing.T) {
	repo := NewFakeGameRepo()
	uc := usecase.NewGameUsecase(repo)

	// まずゲームを作成してリポジトリに登録
	id := uuid.NewString()
	g0 := model.NewGame()
	g0.ID = id
	repo.store[id] = g0

	// 空マスから移動しようとして model.ErrNoPieceAtSource を返す
	in := input.ApplyMoveInput{
		GameID: id,
		From:   model.Position{Rank: 5, File: 5},
		To:     model.Position{Rank: 6, File: 5},
	}

	_, err := uc.ApplyMove(in)
	// ここではユースケース層で ErrInvalidMove にマッピングされているはず
	assert.ErrorIs(t, err, usecase.ErrInvalidMove)
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
