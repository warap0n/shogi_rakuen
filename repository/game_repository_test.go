// repository/game_repository_test.go

package repository_test

import (
	"testing"

	"shogi-rakuen/model"
	"shogi-rakuen/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB_GameRepository はインメモリ SQLite DB を初期化し、
// Game テーブルおよび MoveEntity テーブルをマイグレーションします。
func setupTestDB_GameRepository(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })

	require.NoError(t, db.AutoMigrate(&model.Game{}, &model.MoveEntity{}))
	return db
}

// newEmptyGame は model.NewGameWithPlayers を使って初期状態の Game を作り、ID を固定値にします。
func newEmptyGame() *model.Game {
	g, _ := model.NewGameWithPlayers("black", "white")
	g.ID = "test-game-1"
	return g
}

// newSampleMove はサンプルの Move を返します。
// ここでは「初期盤面の (6,6) にある歩を (6,5) に動かす」例です。
func newSampleMove() model.Move {
	return model.Move{
		From:    model.Position{Rank: 6, File: 6},
		To:      model.Position{Rank: 5, File: 6},
		Promote: false,
		Drop:    false,
	}
}

func TestGameRepository_CreateGame(t *testing.T) {
	db := setupTestDB_GameRepository(t)
	repo := repository.NewGameRepository(db)

	g := newEmptyGame()
	created, err := repo.CreateGame(g)
	require.NoError(t, err)

	assert.Equal(t, g.ID, created.ID)

	var dbGame model.Game
	err = db.First(&dbGame, "id = ?", g.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "black", dbGame.PlayerBlackID)
	assert.Equal(t, "white", dbGame.PlayerWhiteID)
}

func TestGameRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB_GameRepository(t)
	repo := repository.NewGameRepository(db)

	_, err := repo.FindByID("does-not-exist")
	assert.ErrorIs(t, err, repository.ErrGameNotFound)
}

func TestGameRepository_FindByID_WithValidMove(t *testing.T) {
	db := setupTestDB_GameRepository(t)
	repo := repository.NewGameRepository(db)

	g := newEmptyGame()
	_, err := repo.CreateGame(g)
	require.NoError(t, err)

	move1 := newSampleMove()
	_, err = repo.AppendMove(g, move1)
	require.NoError(t, err)

	found, err := repo.FindByID(g.ID)
	require.NoError(t, err)

	// 盤面再構築済み → Moves スライスに 1 手入っている
	assert.Len(t, found.Moves, 1)
	assert.Equal(t, move1, found.Moves[0])

	// Board 上にも (6,5) に黒の歩がいるはず
	p, err := found.Board.PieceAt(move1.To)
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, model.Pawn, p.Type)
	assert.Equal(t, model.Black, p.Color)
}

func TestGameRepository_FindByID_ErrorOnLoadMoves(t *testing.T) {
	db := setupTestDB_GameRepository(t)
	repo := repository.NewGameRepository(db)

	g := newEmptyGame()
	_, err := repo.CreateGame(g)
	require.NoError(t, err)

	// DB をクローズして強制的に Err を再現
	sqlDB, _ := db.DB()
	sqlDB.Close()

	_, err = repo.FindByID(g.ID)
	assert.Error(t, err)
}
