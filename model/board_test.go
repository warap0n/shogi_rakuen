// model/board_test.go
package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBoard_InitialPlacement(t *testing.T) {
	b := NewBoard()

	// White King at (0,4)
	p, err := b.PieceAt(Position{Rank: 0, File: 4})
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, King, p.Type)
	assert.Equal(t, White, p.Color)

	// White Rook at (1,1)
	p, err = b.PieceAt(Position{Rank: 1, File: 1})
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, Rook, p.Type)
	assert.Equal(t, White, p.Color)

	// Black Rook at (7,7)
	p, err = b.PieceAt(Position{Rank: 7, File: 7})
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, Rook, p.Type)
	assert.Equal(t, Black, p.Color)

	// Empty square at (4,4)
	p, err = b.PieceAt(Position{Rank: 4, File: 4})
	assert.NoError(t, err)
	assert.Nil(t, p)
}

func TestPieceAt_OutOfBounds(t *testing.T) {
	b := NewBoard()

	_, err := b.PieceAt(Position{Rank: -1, File: 0})
	assert.ErrorIs(t, err, ErrOutOfBounds)

	_, err = b.PieceAt(Position{Rank: 0, File: 9})
	assert.ErrorIs(t, err, ErrOutOfBounds)
}

func TestSetPiece_Normal(t *testing.T) {
	b := NewBoard()
	pos := Position{Rank: 4, File: 4}
	pawn := &Piece{Type: Pawn, Color: Black}

	// in-bounds ならエラーなし
	err := b.SetPiece(pos, pawn)
	assert.NoError(t, err)

	// 正しくセットされている
	got, err := b.PieceAt(pos)
	assert.NoError(t, err)
	assert.Equal(t, pawn, got)
}

func TestSetPiece_Overwrite(t *testing.T) {
	b := NewBoard()
	pos := Position{Rank: 4, File: 4}
	first := &Piece{Type: Pawn, Color: Black}
	second := &Piece{Type: Gold, Color: White}

	// まず Pawn を置く
	assert.NoError(t, b.SetPiece(pos, first))
	got, err := b.PieceAt(pos)
	assert.NoError(t, err)
	assert.Equal(t, first, got)

	// さらに Gold で上書き
	assert.NoError(t, b.SetPiece(pos, second))
	got, err = b.PieceAt(pos)
	assert.NoError(t, err)
	assert.Equal(t, second, got)
}

func TestSetPiece_OutOfBounds(t *testing.T) {
	b := NewBoard()

	err := b.SetPiece(Position{Rank: -1, File: 0}, &Piece{Type: Pawn})
	assert.ErrorIs(t, err, ErrOutOfBounds)
	assert.Contains(t, err.Error(), ErrOutOfBounds.Error())
}
