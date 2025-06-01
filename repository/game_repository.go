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
	// Board/Moves/Captured は gorm:"-" なので保存されない
	if err := r.db.Create(g).Error; err != nil {
		return nil, err
	}
	return g, nil
}

// FindByID は「Board と Captured を含めた状態で」Game を返します。
// 1) games テーブルから ID, PlayerBlackID, PlayerWhiteID, Turn, Finished, Winner だけ取得
// 2) model.NewGameWithPlayers で初期盤面＋空の持ち駒を生成し、上記メタ情報をセット
// 3) moves テーブルを idx 昇順で取得し、一手ずつ ApplyMove する。
func (r *gameRepository) FindByID(id string) (*model.Game, error) {
	// 1) メタ情報だけ取得
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

	// 2) 初期盤面を生成
	latestGame, err := model.NewGameWithPlayers(meta.PlayerBlackID, meta.PlayerWhiteID)
	if err != nil {
		return nil, err
	}
	latestGame.ID = meta.ID
	latestGame.Turn = meta.Turn
	latestGame.Finished = meta.Finished
	latestGame.Winner = meta.Winner

	// 3) moves テーブルをすべて取得し、ApplyMove で再構築（無効な手はスキップする）
	var entities []model.MoveEntity
	if err := r.db.
		Where("game_id = ?", id).
		Order("idx").
		Find(&entities).Error; err != nil {
		return nil, err
	}
	for _, me := range entities {
		domainMove := me.ToDomain()
		if applyErr := latestGame.ApplyMove(domainMove); applyErr != nil {
			// Todo: 無効なMoveの処理
			// そもそもMoveに無効なものが含まれないような設計にするのが理想
			return nil, applyErr
		}
	}

	return latestGame, nil
}

// AppendMove は一手を moves テーブルに追加し、in-memory の Game.Moves にも追加します。
// 盤面チェックは行わず、あくまで「履歴を保存するだけ」。
func (r *gameRepository) AppendMove(g *model.Game, move model.Move) (*model.Game, error) {
	idx := len(g.Moves)
	me := move.ToEntity(g.ID, idx)
	if err := r.db.Create(&me).Error; err != nil {
		return nil, err
	}
	g.Moves = append(g.Moves, move)
	return g, nil
}
