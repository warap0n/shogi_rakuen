// model/piece_test.go
package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidMove_Pawn(t *testing.T) {
	b := NewBoard()

	// Black pawn: 前に1マスだけ
	bp := &Piece{Type: Pawn, Color: Black}
	assert.True(t, bp.ValidMove(
		Position{Rank: 6, File: 4},
		Position{Rank: 5, File: 4},
		b,
	))
	assert.False(t, bp.ValidMove(
		Position{Rank: 6, File: 4},
		Position{Rank: 6, File: 5},
		b,
	))

	// White pawn: 後ろ方向
	wp := &Piece{Type: Pawn, Color: White}
	assert.True(t, wp.ValidMove(
		Position{Rank: 2, File: 4},
		Position{Rank: 3, File: 4},
		b,
	))
	assert.False(t, wp.ValidMove(
		Position{Rank: 2, File: 4},
		Position{Rank: 1, File: 4},
		b,
	))
}

func TestValidMove_Lance(t *testing.T) {
	// 空盤なら一直線前方はOK
	b := &Board{}
	lp := &Piece{Type: Lance, Color: Black}
	assert.True(t, lp.ValidMove(
		Position{Rank: 6, File: 2},
		Position{Rank: 0, File: 2},
		b,
	))
	// 横移動・後ろ移動は不可
	assert.False(t, lp.ValidMove(
		Position{Rank: 6, File: 2},
		Position{Rank: 6, File: 3},
		b,
	))
	assert.False(t, lp.ValidMove(
		Position{Rank: 6, File: 2},
		Position{Rank: 7, File: 2},
		b,
	))
}

func TestValidMove_Knight(t *testing.T) {
	b := &Board{}
	kp := &Piece{Type: Knight, Color: Black}
	assert.True(t, kp.ValidMove(
		Position{Rank: 6, File: 4},
		Position{Rank: 4, File: 5},
		b,
	))
	assert.False(t, kp.ValidMove(
		Position{Rank: 6, File: 4},
		Position{Rank: 5, File: 6},
		b,
	))
}

func TestValidMove_Silver(t *testing.T) {
	b := &Board{}
	sp := &Piece{Type: Silver, Color: White}
	// 斜め前（前斜め左右）
	assert.True(t, sp.ValidMove(
		Position{Rank: 3, File: 3},
		Position{Rank: 4, File: 4},
		b,
	))
	assert.True(t, sp.ValidMove(
		Position{Rank: 3, File: 3},
		Position{Rank: 4, File: 2},
		b,
	))
	// 直進前のみ
	assert.True(t, sp.ValidMove(
		Position{Rank: 3, File: 3},
		Position{Rank: 4, File: 3},
		b,
	))
	// 後ろ斜めも合法
	assert.True(t, sp.ValidMove(
		Position{Rank: 3, File: 3},
		Position{Rank: 2, File: 2},
		b,
	))
	assert.True(t, sp.ValidMove(
		Position{Rank: 3, File: 3},
		Position{Rank: 2, File: 4},
		b,
	))
	// 横移動・後退直進・長距離は不可
	assert.False(t, sp.ValidMove(
		Position{Rank: 3, File: 3},
		Position{Rank: 3, File: 4},
		b,
	))
	assert.False(t, sp.ValidMove(
		Position{Rank: 3, File: 3},
		Position{Rank: 1, File: 3},
		b,
	))
}

func TestValidMove_Gold(t *testing.T) {
	b := &Board{}
	gp := &Piece{Type: Gold, Color: Black}
	// 前後左右・前斜めOK
	directions := []Position{
		{Rank: 3, File: 3}, {Rank: 5, File: 3},
		{Rank: 4, File: 2}, {Rank: 4, File: 4},
		{Rank: 3, File: 2}, {Rank: 3, File: 4},
	}
	for _, dst := range directions {
		assert.True(t, gp.ValidMove(
			Position{Rank: 4, File: 3},
			dst,
			b,
		), "Gold should move to", dst)
	}
	// 後ろ斜めは不可
	assert.False(t, gp.ValidMove(
		Position{Rank: 4, File: 3},
		Position{Rank: 5, File: 2},
		b,
	))
}

func TestValidMove_King(t *testing.T) {
	b := &Board{}
	king := &Piece{Type: King, Color: White}
	// 周囲８マス
	for dr := -1; dr <= 1; dr++ {
		for df := -1; df <= 1; df++ {
			if dr == 0 && df == 0 {
				continue
			}
			assert.True(t, king.ValidMove(
				Position{Rank: 4, File: 4},
				Position{Rank: 4 + dr, File: 4 + df},
				b,
			))
		}
	}
	// 2マス以上は不可
	assert.False(t, king.ValidMove(
		Position{Rank: 4, File: 4},
		Position{Rank: 6, File: 4},
		b,
	))
}

func TestValidMove_Bishop(t *testing.T) {
	b := &Board{}
	bp := &Piece{Type: Bishop, Color: Black}
	// 斜めスライド
	assert.True(t, bp.ValidMove(
		Position{Rank: 4, File: 4},
		Position{Rank: 1, File: 1},
		b,
	))
	assert.True(t, bp.ValidMove(
		Position{Rank: 4, File: 4},
		Position{Rank: 7, File: 7},
		b,
	))
	// 中間に障害物を設置
	b.SetPiece(Position{Rank: 3, File: 3}, &Piece{Type: Pawn, Color: White})
	assert.False(t, bp.ValidMove(
		Position{Rank: 4, File: 4},
		Position{Rank: 1, File: 1},
		b,
	))
}

func TestValidMove_Rook(t *testing.T) {
	b := &Board{}
	rp := &Piece{Type: Rook, Color: White}
	// 縦横スライド
	assert.True(t, rp.ValidMove(
		Position{Rank: 2, File: 4},
		Position{Rank: 8, File: 4},
		b,
	))
	assert.True(t, rp.ValidMove(
		Position{Rank: 2, File: 4},
		Position{Rank: 2, File: 0},
		b,
	))
	// 中間に障害物を設置
	b.SetPiece(Position{Rank: 4, File: 4}, &Piece{Type: Pawn, Color: Black})
	assert.False(t, rp.ValidMove(
		Position{Rank: 2, File: 4},
		Position{Rank: 8, File: 4},
		b,
	))
}

func TestValidMove_Promoted(t *testing.T) {
	b := &Board{}
	// 成銀は金動作
	sp := &Piece{Type: Silver, Color: Black, Promoted: true}
	assert.True(t, sp.ValidMove(
		Position{Rank: 4, File: 4},
		Position{Rank: 3, File: 4},
		b,
	))
	assert.False(t, sp.ValidMove(
		Position{Rank: 4, File: 4},
		Position{Rank: 5, File: 5},
		b,
	))

	// 馬：角+王
	horse := &Piece{Type: Bishop, Color: White, Promoted: true}
	assert.True(t, horse.ValidMove(
		Position{Rank: 4, File: 4},
		Position{Rank: 1, File: 1},
		b,
	))
	assert.True(t, horse.ValidMove(
		Position{Rank: 4, File: 4},
		Position{Rank: 5, File: 4},
		b,
	))

	// 龍：飛車+王
	dragon := &Piece{Type: Rook, Color: Black, Promoted: true}
	assert.True(t, dragon.ValidMove(
		Position{Rank: 3, File: 4},
		Position{Rank: 0, File: 4},
		b,
	))
	assert.True(t, dragon.ValidMove(
		Position{Rank: 3, File: 4},
		Position{Rank: 4, File: 5},
		b,
	))
}

func TestValidMove_OutOfBoundsAndSameSquare(t *testing.T) {
	b := &Board{}
	p := &Piece{Type: King, Color: Black}
	// 盤外
	assert.False(t, p.ValidMove(
		Position{Rank: 4, File: 4},
		Position{Rank: 9, File: 4},
		b,
	))
	// 同一マス
	assert.False(t, p.ValidMove(
		Position{Rank: 4, File: 4},
		Position{Rank: 4, File: 4},
		b,
	))
}
