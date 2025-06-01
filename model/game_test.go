// model/game_test.go

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- 正常系 ---

func TestNewGameWithPlayers_Success(t *testing.T) {
	blackID := "alice"
	whiteID := "bob"

	g, err := NewGameWithPlayers(blackID, whiteID)
	assert.NoError(t, err)
	assert.NotNil(t, g)

	// プレイヤーID が設定されていること
	assert.Equal(t, blackID, g.PlayerBlackID)
	assert.Equal(t, whiteID, g.PlayerWhiteID)
}

func TestApplyMove_NormalMove(t *testing.T) {
	g, err := NewGameWithPlayers("black", "white")
	require.NoError(t, err)

	// Black pawn を (6,0) に置いておく
	from := Position{Rank: 6, File: 0}
	to := Position{Rank: 5, File: 0}
	require.NoError(t, g.Board.SetPiece(from, &Piece{Type: Pawn, Color: Black}))

	m := Move{From: from, To: to, Promote: false, Drop: false}

	err = g.ApplyMove(m)
	assert.NoError(t, err)

	// 移動先に Pawn が置かれている
	p, err := g.Board.PieceAt(to)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, Pawn, p.Type)
	assert.Equal(t, Black, p.Color)

	// 元のマスは空
	p, err = g.Board.PieceAt(from)
	assert.NoError(t, err)
	assert.Nil(t, p)

	// ターンが切り替わり、履歴に残る
	assert.Equal(t, White, g.Turn)
	assert.Len(t, g.Moves, 1)
	assert.Equal(t, m, g.Moves[0])
}

func TestApplyMove_Promotion_Success(t *testing.T) {
	g, err := NewGameWithPlayers("black", "white")
	require.NoError(t, err)

	// プロモーション成功例: (2,0)→(1,0) に動いて成る
	from := Position{Rank: 2, File: 0}
	to := Position{Rank: 1, File: 0}
	// (2,0) に黒の Pawn を置いておく
	require.NoError(t, g.Board.SetPiece(from, &Piece{Type: Pawn, Color: Black}))

	m := Move{From: from, To: to, Promote: true, Drop: false}

	err = g.ApplyMove(m)
	assert.NoError(t, err)

	// (1,0) に Promoted Pawn（Tokin）がいる
	p, err := g.Board.PieceAt(to)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, Pawn, p.Type)
	assert.True(t, p.Promoted, "プロモートフラグが立っていること")
}

func TestApplyMove_Capture(t *testing.T) {
	g, err := NewGameWithPlayers("black", "white")
	require.NoError(t, err)

	// Black lance を (6,3) に置き、White silver を (3,3) に置く
	from := Position{Rank: 6, File: 3}
	target := Position{Rank: 3, File: 3}
	require.NoError(t, g.Board.SetPiece(from, &Piece{Type: Lance, Color: Black}))
	require.NoError(t, g.Board.SetPiece(target, &Piece{Type: Silver, Color: White, Promoted: false}))

	m := Move{From: from, To: target, Promote: false, Drop: false}

	err = g.ApplyMove(m)
	assert.NoError(t, err)

	// (3,3) には黒の Lance が置かれている
	pAtDest, err := g.Board.PieceAt(target)
	assert.NoError(t, err)
	assert.NotNil(t, pAtDest)
	assert.Equal(t, Lance, pAtDest.Type)
	assert.Equal(t, Black, pAtDest.Color)
	assert.False(t, pAtDest.Promoted)

	// (6,3) は空
	pAtFrom, err := g.Board.PieceAt(from)
	assert.NoError(t, err)
	assert.Nil(t, pAtFrom)

	// Black のキャプチャリストに Silver が入っている
	caps := g.Captured[Black]
	require.Len(t, caps, 1)
	capturedPiece := caps[0]
	assert.Equal(t, Silver, capturedPiece.Type)
	assert.Equal(t, Black, capturedPiece.Color)
	assert.False(t, capturedPiece.Promoted)
}

func TestApplyMove_MultiCaptureDemotion(t *testing.T) {
	g, err := NewGameWithPlayers("black", "white")
	require.NoError(t, err)

	// Black bishop を (6,6) に置き、White silver を (3,3)、White knight を (2,2) に置く
	from := Position{Rank: 6, File: 6}
	pos1 := Position{Rank: 3, File: 3}
	pos2 := Position{Rank: 2, File: 2}
	require.NoError(t, g.Board.SetPiece(from, &Piece{Type: Bishop, Color: Black, Promoted: false}))
	require.NoError(t, g.Board.SetPiece(pos1, &Piece{Type: Silver, Color: White, Promoted: true}))
	require.NoError(t, g.Board.SetPiece(pos2, &Piece{Type: Knight, Color: White, Promoted: true}))

	// 1 手目: (6,6)→(3,3) でまず Silver をキャプチャ
	move1 := Move{From: from, To: pos1, Promote: false, Drop: false}
	require.NoError(t, g.ApplyMove(move1))

	// applyMove後は g.Turn=White になるので、意図的に黒番に戻す
	g.Turn = Black

	// 2 手目: (3,3)→(2,2) で Knight をキャプチャ
	move2 := Move{From: pos1, To: pos2, Promote: false, Drop: false}
	require.NoError(t, g.ApplyMove(move2))

	// (2,2) には黒の Bishop がいる
	pAtPos2, err := g.Board.PieceAt(pos2)
	require.NoError(t, err)
	require.NotNil(t, pAtPos2)
	assert.Equal(t, Bishop, pAtPos2.Type)
	assert.Equal(t, Black, pAtPos2.Color)
	assert.False(t, pAtPos2.Promoted)

	// Black のキャプチャリストには、最初に Silver、次に Knight が順に格納されている
	caps := g.Captured[Black]
	require.Len(t, caps, 2)

	assert.Equal(t, Silver, caps[0].Type)
	assert.Equal(t, Black, caps[0].Color)
	assert.False(t, caps[0].Promoted)

	assert.Equal(t, Knight, caps[1].Type)
	assert.Equal(t, Black, caps[1].Color)
	assert.False(t, caps[1].Promoted)
}

func TestApplyMove_CaptureThenDrop(t *testing.T) {
	g, err := NewGameWithPlayers("black", "white")
	require.NoError(t, err)

	// Black rook を (6,3) に置き、White silver を (3,3) に置く
	from := Position{Rank: 6, File: 3}
	target := Position{Rank: 3, File: 3}
	require.NoError(t, g.Board.SetPiece(from, &Piece{Type: Rook, Color: Black}))
	require.NoError(t, g.Board.SetPiece(target, &Piece{Type: Silver, Color: White, Promoted: false}))

	// キャプチャ
	moveCapture := Move{From: from, To: target, Promote: false, Drop: false}
	require.NoError(t, g.ApplyMove(moveCapture))

	// g.Turn=White になるので、続けて Black に戻す
	g.Turn = Black

	// 持ち駒に Silver が入っていることを確認
	caps := g.Captured[Black]
	require.Len(t, caps, 1)

	// ドロップを実施
	dropPos := Position{Rank: 4, File: 4}
	moveDrop := Move{To: dropPos, Drop: true, DropPiece: Silver}
	require.NoError(t, g.ApplyMove(moveDrop))

	// dropPos に Silver が置かれ、所有者は Black
	pAtDrop, err := g.Board.PieceAt(dropPos)
	require.NoError(t, err)
	require.NotNil(t, pAtDrop)
	assert.Equal(t, Silver, pAtDrop.Type)
	assert.Equal(t, Black, pAtDrop.Color)
	assert.False(t, pAtDrop.Promoted)

	// Black のキャプチャリストから Silver が削除されていること
	assert.Len(t, g.Captured[Black], 0)
}

func TestApplyMove_DropDemotedPiece(t *testing.T) {
	g, err := NewGameWithPlayers("black", "white")
	require.NoError(t, err)

	// White promoted pawn (Tokin) を (2,2) に置き、Black bishop を (6,6) に置く
	pos := Position{Rank: 2, File: 2}
	require.NoError(t, g.Board.SetPiece(pos, &Piece{Type: Pawn, Color: White, Promoted: true}))
	from := Position{Rank: 6, File: 6}
	require.NoError(t, g.Board.SetPiece(from, &Piece{Type: Bishop, Color: Black}))

	// キャプチャ
	moveCapture := Move{From: from, To: pos, Promote: false, Drop: false}
	require.NoError(t, g.ApplyMove(moveCapture))

	// g.Turn=White になるので、Black に戻す
	g.Turn = Black

	// 持ち駒に Pawn (非成) が入っている
	caps := g.Captured[Black]
	require.Len(t, caps, 1)
	assert.Equal(t, Pawn, caps[0].Type)
	assert.False(t, caps[0].Promoted)

	// ドロップ
	dropPos := Position{Rank: 5, File: 5}
	moveDrop := Move{To: dropPos, Drop: true, DropPiece: Pawn}
	require.NoError(t, g.ApplyMove(moveDrop))

	// dropPos に配置された Pawn が非成のまま
	pAtDrop, err := g.Board.PieceAt(dropPos)
	require.NoError(t, err)
	require.NotNil(t, pAtDrop)
	assert.Equal(t, Pawn, pAtDrop.Type)
	assert.False(t, pAtDrop.Promoted)
}

// --- 異常系 ---

func TestNewGameWithPlayers_InvalidPlayerID(t *testing.T) {
	_, err := NewGameWithPlayers("", "bob")
	assert.ErrorIs(t, err, ErrInvalidPlayerID)

	_, err = NewGameWithPlayers("alice", "")
	assert.ErrorIs(t, err, ErrInvalidPlayerID)

	_, err = NewGameWithPlayers("", "")
	assert.ErrorIs(t, err, ErrInvalidPlayerID)
}

func TestNewGameWithPlayers_SamePlayer(t *testing.T) {
	_, err := NewGameWithPlayers("alice", "alice")
	assert.ErrorIs(t, err, ErrSamePlayer)
}

func TestApplyMove_InvalidPromotionPiece(t *testing.T) {
	g, err := NewGameWithPlayers("black", "white")
	require.NoError(t, err)

	// Gold はそもそもプロモート不可なので (2,0)→(1,0) で ErrInvalidPromotionPiece
	from := Position{Rank: 2, File: 0}
	require.NoError(t, g.Board.SetPiece(from, &Piece{Type: Gold, Color: Black}))

	m := Move{From: from, To: Position{Rank: 1, File: 0}, Promote: true, Drop: false}

	err = g.ApplyMove(m)
	assert.ErrorIs(t, err, ErrInvalidPromotionPiece)
}

func TestApplyMove_InvalidPromotionZone(t *testing.T) {
	g, err := NewGameWithPlayers("black", "white")
	require.NoError(t, err)

	// 自陣の (4,0)→(3,0) でプロモート宣言すれば ErrInvalidPromotionZone
	from := Position{Rank: 4, File: 0}
	require.NoError(t, g.Board.SetPiece(from, &Piece{Type: Pawn, Color: Black}))

	m := Move{From: from, To: Position{Rank: 3, File: 0}, Promote: true, Drop: false}

	err = g.ApplyMove(m)
	assert.ErrorIs(t, err, ErrInvalidPromotionZone)
}

func TestApplyMove_GameAlreadyFinished(t *testing.T) {
	g, err := NewGameWithPlayers("black", "white")
	require.NoError(t, err)

	g.Finished = true
	m := Move{From: Position{Rank: 6, File: 0}, To: Position{Rank: 5, File: 0}}

	err = g.ApplyMove(m)
	assert.ErrorIs(t, err, ErrGameAlreadyFinished)
}

func TestApplyMove_NoPieceToDrop(t *testing.T) {
	g, err := NewGameWithPlayers("black", "white")
	require.NoError(t, err)

	m := Move{To: Position{Rank: 4, File: 4}, Drop: true, DropPiece: Pawn}
	err = g.ApplyMove(m)
	assert.ErrorIs(t, err, ErrNoPieceToDrop)
}

func TestApplyMove_NoPieceAtSource(t *testing.T) {
	g, err := NewGameWithPlayers("black", "white")
	require.NoError(t, err)

	src := Position{Rank: 6, File: 0}
	require.NoError(t, g.Board.SetPiece(src, nil)) // 明示的に空にする

	m := Move{From: src, To: Position{Rank: 5, File: 0}, Drop: false}
	err = g.ApplyMove(m)
	assert.ErrorIs(t, err, ErrNoPieceAtSource)
}

func TestApplyMove_OutOfBounds_Source(t *testing.T) {
	g, err := NewGameWithPlayers("black", "white")
	require.NoError(t, err)

	m := Move{From: Position{Rank: 0, File: -1}, To: Position{Rank: 0, File: 0}}
	err = g.ApplyMove(m)
	assert.ErrorIs(t, err, ErrOutOfBounds)
	assert.Contains(t, err.Error(), "source")
}

func TestApplyMove_OutOfBounds_Destination(t *testing.T) {
	g, err := NewGameWithPlayers("black", "white")
	require.NoError(t, err)

	// (9) は 0～8 の範囲外なので ErrInvalidMove に置き換え
	m := Move{From: Position{Rank: 6, File: 0}, To: Position{Rank: 0, File: 9}}
	from := Position{Rank: 6, File: 0}
	require.NoError(t, g.Board.SetPiece(from, &Piece{Type: Pawn, Color: Black}))

	err = g.ApplyMove(m)
	assert.ErrorIs(t, err, ErrInvalidMove)
}
