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

// StartGame は新しい対局を開始し、永続化します
func (gu *GameUsecase) StartGame(in input.StartGameInput) (*model.Game, error) {
	g, err := model.NewGameWithPlayers(in.BlackID, in.WhiteID)
	if err != nil {
		return nil, mapStartError(err)
	}
	g.ID = uuid.NewString()
	return gu.gr.CreateGame(g)
}

// GetGameByID は対局情報を取得します
func (gu *GameUsecase) GetGameByID(id string) (*model.Game, error) {
	return gu.loadGame(id)
}

// ApplyMove は一手を適用し、永続化します
func (gu *GameUsecase) ApplyMove(in input.ApplyMoveInput) (*model.Game, error) {
	g, err := gu.loadGame(in.GameID)
	if err != nil {
		return nil, err
	}

	m := gu.convertToMove(in)
	if err := gu.applyMoveToGame(g, m); err != nil {
		return nil, err
	}

	if err := gu.persistMove(g, m); err != nil {
		return nil, err
	}
	return g, nil
}

// ListMoves は対局の全手を取得します
func (gu *GameUsecase) ListMoves(id string) ([]model.Move, error) {
	g, err := gu.loadGame(id)
	if err != nil {
		return nil, err
	}
	return g.Moves, nil
}

// loadGame は対局の取得とエラー変換を行います
func (gu *GameUsecase) loadGame(id string) (*model.Game, error) {
	g, err := gu.gr.FindByID(id)
	if err != nil {
		return nil, ErrGameNotFound
	}
	return g, nil
}

// convertToMove は入力 DTO からドメイン Move を組み立てます
func (gu *GameUsecase) convertToMove(in input.ApplyMoveInput) model.Move {
	return model.Move{
		From:      in.From,
		To:        in.To,
		Promote:   in.Promote,
		Drop:      in.Drop,
		DropPiece: in.DropPiece,
	}
}

// applyMoveToGame はモデルのビジネスロジックを呼び出し、エラーをマッピングします
func (gu *GameUsecase) applyMoveToGame(g *model.Game, m model.Move) error {
	if err := g.ApplyMove(m); err != nil {
		switch {
		case errors.Is(err, model.ErrGameAlreadyFinished):
			return ErrGameAlreadyFinished
		case errors.Is(err, model.ErrNoPieceAtSource),
			errors.Is(err, model.ErrOutOfBounds),
			errors.Is(err, model.ErrNoPieceToDrop),
			errors.Is(err, model.ErrInvalidPromotionPiece),
			errors.Is(err, model.ErrInvalidPromotionZone):
			return ErrInvalidMove
		default:
			return err
		}
	}
	return nil
}

// persistMove は一手を永続化します
func (gu *GameUsecase) persistMove(g *model.Game, m model.Move) error {
	_, err := gu.gr.AppendMove(g, m)
	return err
}

// mapStartError は StartGame のエラーをユースケースエラーにマッピング
func mapStartError(err error) error {
	switch {
	case errors.Is(err, model.ErrInvalidPlayerID):
		return ErrInvalidPlayerID
	case errors.Is(err, model.ErrSamePlayer):
		return ErrSamePlayer
	default:
		return err
	}
}
