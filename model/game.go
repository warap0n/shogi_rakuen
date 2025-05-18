// model/game.go
package model

import (
	"errors"
	"fmt"
)

// Game は１局の対局を管理します
type Game struct {
	Board    *Board
	Turn     Color
	Moves    []Move
	Captured map[Color][]*Piece // 先手／後手の持ち駒
	Finished bool
	Winner   Color
}

func NewGame() *Game {
	return &Game{
		Board:    NewBoard(),
		Turn:     Black,
		Captured: map[Color][]*Piece{Black: {}, White: {}},
	}
}

func (g *Game) ApplyMove(m Move) error {
	// 0) ゲーム終了後はエラー
	if g.Finished {
		return ErrGameAlreadyFinished
	}

	var p *Piece
	var err error

	// 1) Drop（打ち）か移動か
	if m.Drop {
		pcs := g.Captured[g.Turn]
		// 1-1) DropPiece と一致する駒を手持ちから探す
		idx := -1
		for i, cp := range pcs {
			if cp.Type == m.DropPiece {
				idx = i
				break
			}
		}
		if idx < 0 {
			return ErrNoPieceToDrop
		}
		// 1-2) その駒を取り出し、手持ちから除去
		p = pcs[idx]
		g.Captured[g.Turn] = append(pcs[:idx], pcs[idx+1:]...)

	} else {
		// 1') 移動元の駒を取得
		p, err = g.Board.PieceAt(m.From)
		if errors.Is(err, ErrOutOfBounds) {
			return fmt.Errorf("source %v: %w", m.From, ErrOutOfBounds)
		}
		if err != nil {
			return err
		}
		if p == nil {
			return ErrNoPieceAtSource
		}
		// 1'') 移動元を空に
		if err := g.Board.SetPiece(m.From, nil); err != nil {
			return err
		}
	}

	// 2) 成り処理

	if m.Promote {
		// 2-1) そもそもプロモート可能な駒か
		if !p.Type.Promotable() {
			return ErrInvalidPromotionPiece
		}
		// 2-2) 移動元 or 移動先がプロモーションゾーンか
		if !(m.From.InPromotionZone(g.Turn) || m.To.InPromotionZone(g.Turn)) {
			return ErrInvalidPromotionZone
		}
		p.Promoted = true
	}

	// 3) 移動先の駒を取得
	captured, err := g.Board.PieceAt(m.To)
	if errors.Is(err, ErrOutOfBounds) {
		return fmt.Errorf("destination %v: %w", m.To, ErrOutOfBounds)
	}
	if err != nil {
		return err
	}
	// 3') 取った駒は持ち駒に追加（成り戻し）
	if captured != nil {
		captured.Promoted = false
		g.Captured[g.Turn] = append(g.Captured[g.Turn], captured)
	}

	// 4) 駒を置く
	if err := g.Board.SetPiece(m.To, p); err != nil {
		return err
	}

	// 5) ターン切り替え・履歴追加
	g.Turn = 1 - g.Turn
	g.Moves = append(g.Moves, m)

	return nil
}
