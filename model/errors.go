// model/errors.go
package model

import "errors"

var (
	// 盤外参照
	ErrOutOfBounds = errors.New("position out of board")
	// ゲーム終了後に手を指そうとした
	ErrGameAlreadyFinished = errors.New("game already finished")
	// 打ち駒が手持ちにない
	ErrNoPieceToDrop = errors.New("no piece to drop")
	// 移動元に駒がない
	ErrNoPieceAtSource = errors.New("no piece at source")
	// この駒ではその移動ができない
	ErrInvalidMove = errors.New("invalid move for this piece")
	// そもそもその駒種はプロモート不可
	ErrInvalidPromotionPiece = errors.New("piece cannot promote")
	// プロモーションゾーン外でプロモートしようとした
	ErrInvalidPromotionZone = errors.New("promotion not allowed outside promotion zone")
	// 味方の駒を取ろうとした
	ErrCannotCaptureOwnPiece = errors.New("cannot capture own piece")
	// PlayerID が空
	ErrInvalidPlayerID = errors.New("player ID must be non-empty")
	// PlayerID が同じ
	ErrSamePlayer = errors.New("black and white player must differ")
)
