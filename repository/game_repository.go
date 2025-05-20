// Todo: 雛形のみ作成

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

// IGameRepository は対局永続化用のインターフェース
type IGameRepository interface {
	Save(g *model.Game) (*model.Game, error)
	FindByID(id string) (*model.Game, error)
}

type gameRepository struct {
	db *gorm.DB
}

// NewGameRepository は GORM ベースの IGameRepository を返します
func NewGameRepository(db *gorm.DB) IGameRepository {
	return &gameRepository{db: db}
}

// Save は対局オブジェクトを保存または更新します。
// GORM の Save は primary key が空なら INSERT、埋まっていれば UPDATE 相当です。
func (r *gameRepository) Save(g *model.Game) (*model.Game, error) {
	if err := r.db.Save(g).Error; err != nil {
		return nil, err
	}
	return g, nil
}

// FindByID は指定された ID の対局を返します。
// 見つからない場合は ErrGameNotFound を返します。
func (r *gameRepository) FindByID(id string) (*model.Game, error) {
	var g model.Game
	err := r.db.
		Where("id = ?", id).
		Preload("Board").    // 必要に応じて関連もロード
		Preload("Captured"). // map[string][]*Piece の扱いは別途カスタムが要るかも
		First(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrGameNotFound
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}
