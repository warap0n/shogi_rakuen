package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- 正常系 ---

func TestApplyMove_NormalMove(t *testing.T) {
	g := NewGame()

	// Black pawn from Rank=6, File=0 to Rank=5, File=0
	from := Position{Rank: 6, File: 0}
	to := Position{Rank: 5, File: 0}
	m := Move{From: from, To: to, Promote: false, Drop: false}

	err := g.ApplyMove(m)
	assert.NoError(t, err)

	p, err := g.Board.PieceAt(to)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, Pawn, p.Type)
	assert.Equal(t, Black, p.Color)

	p, err = g.Board.PieceAt(from)
	assert.NoError(t, err)
	assert.Nil(t, p)

	assert.Equal(t, White, g.Turn)
	assert.Len(t, g.Moves, 1)
}

func TestApplyMove_Promotion_Success(t *testing.T) {
	g := NewGame()
	// 先手のプロモート成功例: Black のプロモゾーンは Rank<=2
	// 先手 Pawn を敵陣(ランク2)から一つ進めてプロモート
	from := Position{Rank: 2, File: 0} // プロモゾーン内
	to := Position{Rank: 1, File: 0}   // プロモゾーン内
	m := Move{From: from, To: to, Promote: true, Drop: false}

	err := g.ApplyMove(m)
	assert.NoError(t, err)

	p, err := g.Board.PieceAt(to)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.True(t, p.Promoted, "プロモートフラグが立っていること")
}

func TestApplyMove_Capture(t *testing.T) {
	g := NewGame()
	// Place White silver at Rank=3, File=3
	target := Position{Rank: 3, File: 3}
	_ = g.Board.SetPiece(target, &Piece{Type: Silver, Color: White})

	// Black pawn captures
	m := Move{From: Position{Rank: 6, File: 3}, To: target, Promote: false, Drop: false}

	err := g.ApplyMove(m)
	assert.NoError(t, err)

	caps := g.Captured[Black]
	assert.Len(t, caps, 1)
	assert.Equal(t, Silver, caps[0].Type)
	assert.False(t, caps[0].Promoted)
}

func TestApplyMove_Drop(t *testing.T) {
	g := NewGame()
	// Give Black a pawn in hand
	g.Captured[Black] = append(g.Captured[Black], &Piece{Type: Pawn, Color: Black})

	m := Move{To: Position{Rank: 4, File: 4}, Drop: true, DropPiece: Pawn}

	err := g.ApplyMove(m)
	assert.NoError(t, err)

	p, err := g.Board.PieceAt(m.To)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, Pawn, p.Type)

	assert.Empty(t, g.Captured[Black])
}

// --- 異常系 ---

func TestApplyMove_InvalidPromotionPiece(t *testing.T) {
	g := NewGame()
	// 金将(Gold)はそもそもプロモート不可
	// まず盤上に金を適当な場所に置く
	pos := Position{Rank: 2, File: 0}
	g.Board.SetPiece(pos, &Piece{Type: Gold, Color: Black})

	m := Move{From: pos, To: Position{Rank: 1, File: 0}, Promote: true, Drop: false}

	err := g.ApplyMove(m)
	assert.ErrorIs(t, err, ErrInvalidPromotionPiece)
}

func TestApplyMove_InvalidPromotionZone(t *testing.T) {
	g := NewGame()
	// Pawn を自陣 (プロモゾーン外) でプロモート宣言
	from := Position{Rank: 4, File: 0}
	to := Position{Rank: 3, File: 0}
	// ここに Pawn をセット
	g.Board.SetPiece(from, &Piece{Type: Pawn, Color: Black})

	m := Move{From: from, To: to, Promote: true, Drop: false}

	err := g.ApplyMove(m)
	assert.ErrorIs(t, err, ErrInvalidPromotionZone)
}

func TestApplyMove_GameAlreadyFinished(t *testing.T) {
	g := NewGame()
	g.Finished = true

	m := Move{From: Position{Rank: 6, File: 0}, To: Position{Rank: 5, File: 0}}
	err := g.ApplyMove(m)
	assert.ErrorIs(t, err, ErrGameAlreadyFinished)
}

func TestApplyMove_NoPieceToDrop(t *testing.T) {
	g := NewGame()
	m := Move{To: Position{Rank: 4, File: 4}, Drop: true, DropPiece: Pawn}

	err := g.ApplyMove(m)
	assert.ErrorIs(t, err, ErrNoPieceToDrop)
}

func TestApplyMove_NoPieceAtSource(t *testing.T) {
	g := NewGame()
	src := Position{Rank: 6, File: 0}
	_ = g.Board.SetPiece(src, nil)

	m := Move{From: src, To: Position{Rank: 5, File: 0}, Drop: false}
	err := g.ApplyMove(m)
	assert.ErrorIs(t, err, ErrNoPieceAtSource)
}

func TestApplyMove_OutOfBounds_Source(t *testing.T) {
	g := NewGame()
	m := Move{From: Position{Rank: 0, File: -1}, To: Position{Rank: 0, File: 0}}
	err := g.ApplyMove(m)
	// wrapped ErrOutOfBounds
	assert.ErrorIs(t, err, ErrOutOfBounds)
	assert.Contains(t, err.Error(), "source")
}

func TestApplyMove_OutOfBounds_Destination(t *testing.T) {
	g := NewGame()
	m := Move{From: Position{Rank: 6, File: 0}, To: Position{Rank: 0, File: 9}}
	err := g.ApplyMove(m)
	// wrapped ErrOutOfBounds
	assert.ErrorIs(t, err, ErrOutOfBounds)
	assert.Contains(t, err.Error(), "destination")
}
