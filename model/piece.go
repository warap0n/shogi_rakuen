package model

// Color は先手／後手を表します
type Color int

const (
	Black Color = iota
	White
)

// PieceType は駒の種類を表します
type PieceType int

const (
	King PieceType = iota
	Rook
	Bishop
	Gold
	Silver
	Knight
	Lance
	Pawn
)

// Piece は１枚の駒を表します
type Piece struct {
	Type     PieceType // 種類
	Color    Color     // 先手/後手
	Promoted bool      // 成り駒かどうか
}
