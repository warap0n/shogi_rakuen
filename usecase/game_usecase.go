// usecase/game_usecase.go
package usecase

import (
	"errors"
	"shogi-rakuen/model"
	"shogi-rakuen/repository"
	"shogi-rakuen/usecase/input"

	"github.com/google/uuid"
)

// IGameUsecase は将棋対局に関するユースケースを定義します
type IGameUsecase interface {
	StartGame(in input.StartGameInput) (*model.Game, error)
	GetGameByID(id string) (*model.Game, error)
	ApplyMove(in input.ApplyMoveInput) (*model.Game, error)
	ListMoves(id string) ([]model.Move, error)
}

// GameUsecase は IGameUsecase の実装です
type GameUsecase struct {
	gr repository.IGameRepository
}

// NewGameUsecase は GameUsecase のコンストラクタ
func NewGameUsecase(gr repository.IGameRepository) IGameUsecase {
	return &GameUsecase{gr: gr}
}

func (u *GameUsecase) StartGame(in input.StartGameInput) (*model.Game, error) {
	// ドメインモデルでプレイヤー検証
	g, err := model.NewGameWithPlayers(in.BlackID, in.WhiteID)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidPlayerID):
			return nil, ErrInvalidPlayerID
		case errors.Is(err, model.ErrSamePlayer):
			return nil, ErrSamePlayer
		default:
			return nil, err
		}
	}
	// ユースケース層で一意の ID を付与
	g.ID = uuid.NewString()

	// 永続化
	return u.gr.Save(g)
}

func (u *GameUsecase) GetGameByID(id string) (*model.Game, error) {
	g, err := u.gr.FindByID(id)
	if err != nil {
		return nil, ErrGameNotFound
	}
	return g, nil
}

func (u *GameUsecase) ApplyMove(in input.ApplyMoveInput) (*model.Game, error) {
	// 1) 既存対局を取得
	g, err := u.gr.FindByID(in.GameID)
	if err != nil {
		return nil, ErrGameNotFound
	}
	// 2) ドメインロジックで一手適用
	m := model.Move{
		From:      in.From,
		To:        in.To,
		Promote:   in.Promote,
		Drop:      in.Drop,
		DropPiece: in.DropPiece,
	}
	if err := g.ApplyMove(m); err != nil {
		switch {
		case errors.Is(err, model.ErrGameAlreadyFinished):
			return nil, ErrGameAlreadyFinished
		case errors.Is(err, model.ErrNoPieceAtSource),
			errors.Is(err, model.ErrOutOfBounds),
			errors.Is(err, model.ErrNoPieceToDrop),
			errors.Is(err, model.ErrInvalidPromotionPiece),
			errors.Is(err, model.ErrInvalidPromotionZone):
			return nil, ErrInvalidMove
		default:
			// 想定外のエラーはそのまま返す
			return nil, err
		}
	}
	// 3) 更新を保存
	if _, err := u.gr.Save(g); err != nil {
		return nil, err
	}
	return g, nil
}

func (u *GameUsecase) ListMoves(id string) ([]model.Move, error) {
	g, err := u.gr.FindByID(id)
	if err != nil {
		return nil, ErrGameNotFound
	}
	return g.Moves, nil
}
