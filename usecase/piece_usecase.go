package usecase

import "shogi-rakuen/model"

type IPiece interface {
	Promote(piece *model.Piece) error
}

type pieceUsecase struct {
}
