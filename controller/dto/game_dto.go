package dto

import (
	"shogi-rakuen/model"
)

// StartGameRequest はゲーム開始リクエストの DTO
type StartGameRequest struct {
	BlackID string `json:"black_id" validate:"required"`
	WhiteID string `json:"white_id" validate:"required"`
}

// PieceDTO は盤上・持ち駒の駒情報を JSON 用に変換した構造体
type PieceDTO struct {
	Type     model.PieceType `json:"type"`
	Color    int             `json:"color"`    // 0=先手,1=後手
	Promoted bool            `json:"promoted"` // 成駒か
}

// GameResponse はゲームの現在状態を返却する DTO
type GameResponse struct {
	ID            string        `json:"id"`
	PlayerBlackID string        `json:"black_id"`
	PlayerWhiteID string        `json:"white_id"`
	Turn          int           `json:"turn"` // 次の手番
	Finished      bool          `json:"finished"`
	Winner        int           `json:"winner,omitempty"` // 終了後の勝者 0=先手,1=後手
	Board         [][]*PieceDTO `json:"board"`            // 9x9 盤面
	Captured      [][]*PieceDTO `json:"captured"`         // [2][]持ち駒
	Moves         []*MoveDTO    `json:"moves"`            // 指し手履歴
}

// NewGameResponse は model.Game から GameResponse を生成します
func NewGameResponse(g *model.Game) *GameResponse {
	resp := &GameResponse{
		ID:            g.ID,
		PlayerBlackID: g.PlayerBlackID,
		PlayerWhiteID: g.PlayerWhiteID,
		Turn:          int(g.Turn),
		Finished:      g.Finished,
		Board:         toPieceGrid(g.Board),
		Captured:      toCapturedGrid(g.Captured),
		Moves:         NewMovesResponse(g.Moves),
	}
	if g.Finished {
		resp.Winner = int(g.Winner)
	}
	return resp
}

// ApplyMoveRequest は指し手 API の入力 DTO
type ApplyMoveRequest struct {
	From      model.Position  `json:"from" validate:"required"`
	To        model.Position  `json:"to" validate:"required"`
	Promote   bool            `json:"promote"`
	Drop      bool            `json:"drop"`
	DropPiece model.PieceType `json:"drop_piece,omitempty"`
}

// MoveDTO は一手分の情報を返す DTO
type MoveDTO struct {
	From      model.Position  `json:"from"`
	To        model.Position  `json:"to"`
	Promote   bool            `json:"promote"`
	Drop      bool            `json:"drop"`
	DropPiece model.PieceType `json:"drop_piece,omitempty"`
}

// NewMovesResponse は model.Move の配列を []*MoveDTO に変換します
func NewMovesResponse(ms []model.Move) []*MoveDTO {
	dtos := make([]*MoveDTO, len(ms))
	for i, m := range ms {
		dtos[i] = &MoveDTO{
			From:      m.From,
			To:        m.To,
			Promote:   m.Promote,
			Drop:      m.Drop,
			DropPiece: m.DropPiece,
		}
	}
	return dtos
}

// toPieceGrid は Board オブジェクトを 9x9 の PieceDTO グリッドに変換します
func toPieceGrid(board *model.Board) [][]*PieceDTO {
	const size = 9
	grid := make([][]*PieceDTO, size)
	for r := 0; r < size; r++ {
		row := make([]*PieceDTO, size)
		for f := 0; f < size; f++ {
			p, _ := board.PieceAt(model.Position{Rank: r, File: f})
			if p != nil {
				row[f] = &PieceDTO{Type: p.Type, Color: int(p.Color), Promoted: p.Promoted}
			} else {
				row[f] = nil
			}
		}
		grid[r] = row
	}
	return grid
}

// toCapturedGrid は Captured マップを [2][]*PieceDTO に変換します
func toCapturedGrid(captured map[model.Color][]*model.Piece) [][]*PieceDTO {
	grid := make([][]*PieceDTO, 2)
	for c := model.Black; c <= model.White; c++ {
		pcs := captured[c]
		dtos := make([]*PieceDTO, len(pcs))
		for i, p := range pcs {
			dtos[i] = &PieceDTO{Type: p.Type, Color: int(p.Color), Promoted: p.Promoted}
		}
		grid[int(c)] = dtos
	}
	return grid
}
