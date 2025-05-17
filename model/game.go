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

// NewGame は初期配置から始まる Game を返します
func NewGame() *Game {
	return &Game{
		Board:    NewBoard(),
		Turn:     Black,
		Captured: map[Color][]*Piece{Black: {}, White: {}},
	}
}

func (g *Game) ApplyMove(m Move) error {
	if g.Finished {
		return errors.New("game already finished")
	}

	var p *Piece
	var err error

	// 1) Drop（打ち）か移動か
	if m.Drop {
		pcs := g.Captured[g.Turn]
		if len(pcs) == 0 {
			return errors.New("no piece to drop")
		}
		p = pcs[len(pcs)-1]
		g.Captured[g.Turn] = pcs[:len(pcs)-1]
	} else {
		// 移動元の駒を取得
		p, err = g.Board.PieceAt(m.From)
		if errors.Is(err, ErrOutOfBounds) {
			return fmt.Errorf("source %v: %w", m.From, ErrOutOfBounds)
		}
		if err != nil {
			return err
		}
		if p == nil {
			return errors.New("no piece at source")
		}

		// 移動元を空にする
		if err := g.Board.SetPiece(m.From, nil); err != nil {
			return err
		}
	}

	// 2) 成り処理
	if m.Promote {
		p.Promoted = true
	}

	// 3) 移動先にいる駒を取得
	captured, err := g.Board.PieceAt(m.To)
	if errors.Is(err, ErrOutOfBounds) {
		return fmt.Errorf("destination %v: %w", m.To, ErrOutOfBounds)
	}
	if err != nil {
		return err
	}

	// 取った駒は持ち駒に追加（成り戻し）
	if captured != nil {
		captured.Promoted = false
		g.Captured[g.Turn] = append(g.Captured[g.Turn], captured)
	}

	// 駒を置く
	if err := g.Board.SetPiece(m.To, p); err != nil {
		return err
	}

	// 4) ターン切り替え・履歴追加
	g.Turn = 1 - g.Turn
	g.Moves = append(g.Moves, m)

	return nil
}
