// model/board.go
package model

import (
	"fmt"
)

// Board は 9x9 の升目を持つ盤です。nil は空きマス。
type Board struct {
	Squares [9][9]*Piece
}

// NewBoard は全駒の初期配置をセットする
func NewBoard() *Board {
	b := &Board{}

	// ┏━━━━━━━━━━━━━━━━━━━━━━┓
	// ┃ White 側（上段 0-2）  ┃
	// ┗━━━━━━━━━━━━━━━━━━━━━━┛
	// Back rank (row 0)
	b.Squares[0][0] = &Piece{Type: Lance, Color: White}
	b.Squares[0][1] = &Piece{Type: Knight, Color: White}
	b.Squares[0][2] = &Piece{Type: Silver, Color: White}
	b.Squares[0][3] = &Piece{Type: Gold, Color: White}
	b.Squares[0][4] = &Piece{Type: King, Color: White}
	b.Squares[0][5] = &Piece{Type: Gold, Color: White}
	b.Squares[0][6] = &Piece{Type: Silver, Color: White}
	b.Squares[0][7] = &Piece{Type: Knight, Color: White}
	b.Squares[0][8] = &Piece{Type: Lance, Color: White}
	// Second rank (row 1)
	b.Squares[1][1] = &Piece{Type: Rook, Color: White}
	b.Squares[1][7] = &Piece{Type: Bishop, Color: White}
	// Pawn rank (row 2)
	for f := 0; f < 9; f++ {
		b.Squares[2][f] = &Piece{Type: Pawn, Color: White}
	}

	// ┏━━━━━━━━━━━━━━━━━━━━━━┓
	// ┃ Black 側（下段 6-8）  ┃
	// ┗━━━━━━━━━━━━━━━━━━━━━━┛
	// Pawn rank (row 6)
	for f := 0; f < 9; f++ {
		b.Squares[6][f] = &Piece{Type: Pawn, Color: Black}
	}
	// Second rank from bottom (row 7)
	b.Squares[7][1] = &Piece{Type: Bishop, Color: Black}
	b.Squares[7][7] = &Piece{Type: Rook, Color: Black}
	// Back rank (row 8)
	b.Squares[8][0] = &Piece{Type: Lance, Color: Black}
	b.Squares[8][1] = &Piece{Type: Knight, Color: Black}
	b.Squares[8][2] = &Piece{Type: Silver, Color: Black}
	b.Squares[8][3] = &Piece{Type: Gold, Color: Black}
	b.Squares[8][4] = &Piece{Type: King, Color: Black}
	b.Squares[8][5] = &Piece{Type: Gold, Color: Black}
	b.Squares[8][6] = &Piece{Type: Silver, Color: Black}
	b.Squares[8][7] = &Piece{Type: Knight, Color: Black}
	b.Squares[8][8] = &Piece{Type: Lance, Color: Black}

	return b
}

func (b *Board) PieceAt(pos Position) (*Piece, error) {
	if !pos.OnBoard() {
		return nil, ErrOutOfBounds
	}
	return b.Squares[pos.Rank][pos.File], nil
}

// SetPiece は盤外なら ErrOutOfBounds、盤内なら駒を置いて nil を返す
func (b *Board) SetPiece(pos Position, p *Piece) error {
	if !pos.OnBoard() {
		return fmt.Errorf("%w: %+v", ErrOutOfBounds, pos)
	}
	b.Squares[pos.Rank][pos.File] = p
	return nil
}
