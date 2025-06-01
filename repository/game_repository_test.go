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
	// テストが終わったらクローズ
	t.Cleanup(func() { sqlDB.Close() })

	// model.Game と model.MoveEntity のマイグレーション
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
// ここでは「7g→7f」（内部座標 (6,6) → (6,5)）を例とします。
func newSampleMove() model.Move {
	return model.Move{
		From:    model.Position{Rank: 7, File: 7},
		To:      model.Position{Rank: 7, File: 6},
		Promote: false,
		Drop:    false,
	}
}

func TestGameRepository_CreateGame(t *testing.T) {
	db := setupTestDB_GameRepository(t)
	repo := repository.NewGameRepository(db)

	// 新規対局オブジェクトを作成
	g := newEmptyGame()

	// CreateGame を呼び出し
	created, err := repo.CreateGame(g)
	require.NoError(t, err)

	// 戻り値のチェック：ID がそのまま返ってくることと、DB上にレコードが存在する
	assert.Equal(t, g.ID, created.ID)

	var dbGame model.Game
	err = db.First(&dbGame, "id = ?", g.ID).Error
	require.NoError(t, err)
	assert.Equal(t, g.ID, dbGame.ID)
	assert.Equal(t, "black", dbGame.PlayerBlackID)
	assert.Equal(t, "white", dbGame.PlayerWhiteID)
}

func TestGameRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB_GameRepository(t)
	repo := repository.NewGameRepository(db)

	// 存在しない ID で検索 → ErrGameNotFound
	_, err := repo.FindByID("non-existent-id")
	assert.ErrorIs(t, err, repository.ErrGameNotFound)
}

func TestGameRepository_FindByID_WithMoves(t *testing.T) {
	db := setupTestDB_GameRepository(t)
	repo := repository.NewGameRepository(db)

	// 1) CreateGame で対局を作成
	g := newEmptyGame()
	_, err := repo.CreateGame(g)
	require.NoError(t, err)

	// 2) AppendMove を呼んで 1 手だけ登録
	move1 := newSampleMove()
	_, err = repo.AppendMove(g, move1)
	require.NoError(t, err)

	// 3) FindByID で取得して Moves に反映されているか確認
	found, err := repo.FindByID(g.ID)
	require.NoError(t, err)

	assert.Len(t, found.Moves, 1)
	assert.Equal(t, move1, found.Moves[0])
}

func TestGameRepository_AppendMove_MultipleMoves(t *testing.T) {
	db := setupTestDB_GameRepository(t)
	repo := repository.NewGameRepository(db)

	// 新しい対局を作成
	g := newEmptyGame()
	_, err := repo.CreateGame(g)
	require.NoError(t, err)

	// 0手目: 初期状態。Moves は空
	found0, err := repo.FindByID(g.ID)
	require.NoError(t, err)
	assert.Len(t, found0.Moves, 0)

	// 1手目
	move1 := newSampleMove()
	_, err = repo.AppendMove(g, move1)
	require.NoError(t, err)

	// 2手目：ここでは「7c→7d」を例とします（実際の盤面は無視して良い）
	move2 := model.Move{
		From:    model.Position{Rank: 3, File: 7},
		To:      model.Position{Rank: 3, File: 6},
		Promote: false,
		Drop:    false,
	}
	_, err = repo.AppendMove(g, move2)
	require.NoError(t, err)

	// FindByID で取得し、Moves が 2 件あること
	found2, err := repo.FindByID(g.ID)
	require.NoError(t, err)
	assert.Len(t, found2.Moves, 2)
	assert.Equal(t, move1, found2.Moves[0])
	assert.Equal(t, move2, found2.Moves[1])
}

func TestGameRepository_FindByID_ErrorOnLoadMoves(t *testing.T) {
	db := setupTestDB_GameRepository(t)
	repo := repository.NewGameRepository(db)

	// 1) CreateGame で対局を作成
	g := newEmptyGame()
	_, err := repo.CreateGame(g)
	require.NoError(t, err)

	// 2) DB クローズして強制エラー再現
	sqlDB, _ := db.DB()
	sqlDB.Close()

	_, err = repo.FindByID(g.ID)
	assert.Error(t, err)
}

func TestGameRepository_AppendMove_InvalidMove(t *testing.T) {
	db := setupTestDB_GameRepository(t)
	repo := repository.NewGameRepository(db)

	// 新規対局を作成
	g := newEmptyGame()
	_, err := repo.CreateGame(g)
	require.NoError(t, err)

	// “無効な一手” を作成（駒が存在しない座標からの移動）
	invalid := model.Move{
		From:    model.Position{Rank: 5, File: 5},
		To:      model.Position{Rank: 5, File: 4},
		Promote: false,
		Drop:    false,
	}

	// 無効な一手でも、AppendMove 自体はエラーを返さず履歴に追加される
	_, err = repo.AppendMove(g, invalid)
	require.NoError(t, err)

	// FindByID して invalid が単に履歴に含まれていることを確認
	found, err := repo.FindByID(g.ID)
	require.NoError(t, err)
	assert.Len(t, found.Moves, 1)
	assert.Equal(t, invalid, found.Moves[0])
}
