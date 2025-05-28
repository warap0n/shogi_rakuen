// model/game.go
package model

import (
	"errors"
	"fmt"
)

// Game は１局の対局を管理します
type Game struct {
	ID            string
	Board         *Board
	Turn          Color
	Moves         []Move
	Captured      map[Color][]*Piece // 先手／後手の持ち駒
	Finished      bool
	Winner        Color
	PlayerBlackID string
	PlayerWhiteID string
}

// 盤面のみの初期化
func NewGame() *Game {
	return &Game{
		Board:    NewBoard(),
		Turn:     Black,
		Captured: map[Color][]*Piece{Black: {}, White: {}},
	}
}

// 先手・後手の PlayerID を指定して盤面を初期化
func NewGameWithPlayers(blackID, whiteID string) (*Game, error) {
	if blackID == "" || whiteID == "" {
		return nil, ErrInvalidPlayerID
	}
	if blackID == whiteID {
		return nil, ErrSamePlayer
	}
	g := NewGame()
	g.PlayerBlackID = blackID
	g.PlayerWhiteID = whiteID
	return g, nil
}

// ApplyMove は１手を適用します（Drop or Move → Promotion → Capture → Place → NextTurn）
func (g *Game) ApplyMove(m Move) error {
	if g.Finished {
		return ErrGameAlreadyFinished
	}

	// 1) 駒を取得（持駒 or 盤上）
	p, err := g.takePiece(m)
	if err != nil {
		return err
	}

	// 2) 成り処理
	if m.Promote {
		if err := g.handlePromotion(p, m); err != nil {
			return err
		}
	}

	// 3) 相手駒を取る
	if err := g.handleCapture(m); err != nil {
		return err
	}

	// 4) 駒を配置
	if err := g.handlePlace(p, m); err != nil {
		return err
	}

	// 5) ターン更新と履歴
	g.nextTurn(m)
	return nil
}

// takePiece は持駒から取るか盤上から取るかを判断して駒を取得し元の場所をクリアします
func (g *Game) takePiece(m Move) (*Piece, error) {
	if m.Drop {
		return g.extractCapturedPiece(m)
	}
	return g.extractBoardPiece(m)
}

// extractCapturedPiece は持駒から指定駒を取り出し、スライスから除去します
func (g *Game) extractCapturedPiece(m Move) (*Piece, error) {
	piece, rest, err := findAndRemovePiece(g.Captured[g.Turn], m.DropPiece)
	if err != nil {
		return nil, err
	}
	g.Captured[g.Turn] = rest
	return piece, nil
}

// extractBoardPiece は盤上の指定座標から駒を取り出し、そのマスを空にします
func (g *Game) extractBoardPiece(m Move) (*Piece, error) {
	p, err := g.Board.PieceAt(m.From)
	if errors.Is(err, ErrOutOfBounds) {
		return nil, fmt.Errorf("source %v: %w", m.From, ErrOutOfBounds)
	}
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrNoPieceAtSource
	}
	if err := g.Board.SetPiece(m.From, nil); err != nil {
		return nil, err
	}
	return p, nil
}

// findAndRemovePiece は pieces から最初に見つかった駒種を取り出し、残りのスライスを返します
func findAndRemovePiece(pieces []*Piece, pt PieceType) (*Piece, []*Piece, error) {
	for i, cp := range pieces {
		if cp.Type == pt {
			return cp, append(pieces[:i], pieces[i+1:]...), nil
		}
	}
	return nil, nil, ErrNoPieceToDrop
}

// handlePromotion は成り条件をチェックし、駒に成りフラグを立てます
func (g *Game) handlePromotion(p *Piece, m Move) error {
	if !p.Type.Promotable() {
		return ErrInvalidPromotionPiece
	}
	if !(m.From.InPromotionZone(g.Turn) || m.To.InPromotionZone(g.Turn)) {
		return ErrInvalidPromotionZone
	}
	p.Promoted = true
	return nil
}

// handleCapture は移動先に敵駒があれば持駒に追加します
func (g *Game) handleCapture(m Move) error {
	cap, err := g.Board.PieceAt(m.To)
	if errors.Is(err, ErrOutOfBounds) {
		return fmt.Errorf("destination %v: %w", m.To, ErrOutOfBounds)
	}
	if err != nil {
		return err
	}
	if cap != nil {
		cap.Promoted = false
		g.Captured[g.Turn] = append(g.Captured[g.Turn], cap)
	}
	return nil
}

// handlePlace は駒 p を指定座標へ配置します
func (g *Game) handlePlace(p *Piece, m Move) error {
	return g.Board.SetPiece(m.To, p)
}

// nextTurn はターンを切替え、手番履歴に追加します
func (g *Game) nextTurn(m Move) {
	g.Turn = 1 - g.Turn
	g.Moves = append(g.Moves, m)
}
