// usecase/input/start_game_input.go
package input

import "shogi-rakuen/model"

// StartGameInput は新規対局開始の入力
type StartGameInput struct {
	BlackID string
	WhiteID string
}

// ApplyMoveInput は一手指す API の入力
type ApplyMoveInput struct {
	GameID    string
	From      model.Position
	To        model.Position
	Promote   bool
	Drop      bool
	DropPiece model.PieceType
}
