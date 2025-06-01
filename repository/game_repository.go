// repository/game_repository.go

package repository

import (
	"errors"

	"shogi-rakuen/model"

	"gorm.io/gorm"
)

var (
	// Game が見つからなかった時のエラー
	ErrGameNotFound = errors.New("game not found")
)

type IGameRepository interface {
	CreateGame(g *model.Game) (*model.Game, error)
	FindByID(id string) (*model.Game, error)
	AppendMove(g *model.Game, move model.Move) (*model.Game, error)
}

type gameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) IGameRepository {
	return &gameRepository{db: db}
}

// CreateGame は新規ゲームをデータベースに作成します
func (r *gameRepository) CreateGame(g *model.Game) (*model.Game, error) {
	// Game テーブルに INSERT（Board/Moves/Captured は gorm:"-" なので保存されない）
	if err := r.db.Create(g).Error; err != nil {
		return nil, err
	}
	return g, nil
}

// FindByID は games テーブルのメタ情報を取得し、
// MoveEntity をすべて読み込んで Game.Moves に詰めるだけの実装に変更。
// ──────────────────────────────────────────────────────────────
// これにより、どんな手でも履歴として返却され、無効な手が混ざっていてもエラーとはなりません。
// Board/Captured は常に nil のままなので、盤面再構築が必要な場合は
// 呼び出し側（ユースケース層）で model.NewGameWithPlayers + for loop(ApplyMove) を行ってください。
func (r *gameRepository) FindByID(id string) (*model.Game, error) {
	// 1) games テーブルからメタ情報だけ取得
	var meta model.Game
	err := r.db.
		Model(&model.Game{}).
		Select("id", "player_black_id", "player_white_id", "turn", "finished", "winner").
		Where("id = ?", id).
		Take(&meta).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrGameNotFound
	}
	if err != nil {
		return nil, err
	}

	// 2) 取得した meta をベースに、Board/Captured は初期化せず Moves だけ使える Game を作成
	game := &model.Game{
		ID:            meta.ID,
		PlayerBlackID: meta.PlayerBlackID,
		PlayerWhiteID: meta.PlayerWhiteID,
		Turn:          meta.Turn,
		Finished:      meta.Finished,
		Winner:        meta.Winner,
		Board:         nil,                   // あえて nil のまま
		Moves:         make([]model.Move, 0), // 履歴をここに詰める
		Captured:      nil,                   // あえて nil のまま
	}

	// 3) moves テーブルを idx 昇順で取得し、Game.Moves に ToDomain() したものを append
	var entities []model.MoveEntity
	if err := r.db.
		Where("game_id = ?", id).
		Order("idx").
		Find(&entities).Error; err != nil {
		return nil, err
	}
	for _, me := range entities {
		game.Moves = append(game.Moves, me.ToDomain())
	}

	// 4) Game.Moves にすべての履歴が入り、Board/Captured は引き続き nil の状態で返す
	return game, nil
}

// AppendMove は一手を moves テーブルに追加し、in-memory の Game.Moves も更新して返します。
// Domain の盤面ロジックは呼ばず、純粋に「履歴として保存するだけ」です。
func (r *gameRepository) AppendMove(g *model.Game, move model.Move) (*model.Game, error) {
	idx := len(g.Moves)
	me := move.ToEntity(g.ID, idx)
	if err := r.db.Create(&me).Error; err != nil {
		return nil, err
	}
	// in-memory でも Moves に追加
	g.Moves = append(g.Moves, move)
	return g, nil
}
