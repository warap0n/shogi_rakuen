package model

import "time"

// Move は盤上の１手を表します
type Move struct {
	From      Position  `json:"from" gorm:"-"`
	To        Position  `json:"to" gorm:"-"`
	Drop      bool      `json:"drop" gorm:"-"`
	DropPiece PieceType `json:"drop_piece" gorm:"-"`
	Promote   bool      `json:"promote" gorm:"-"`
}

type MoveEntity struct {
	ID        uint      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	GameID    string    `json:"game_id" gorm:"column:game_id;not null;index"`
	Index     int       `json:"idx" gorm:"column:idx;not null"`
	From      string    `json:"from" gorm:"column:from;not null"`
	To        string    `json:"to" gorm:"column:to;not null"`
	Drop      bool      `json:"drop" gorm:"column:drop;not null"`
	DropPiece string    `json:"drop_piece" gorm:"column:drop_piece;type:char(1)"`
	Promote   bool      `json:"promote" gorm:"column:promote;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime;not null"`
}

// TableName を定義して GORM に moves テーブルとマッピングさせる
func (MoveEntity) TableName() string {
	return "moves"
}

// ToDomain は MoveEntity → ドメイン Move への変換
func (e MoveEntity) ToDomain() Move {
	fromPos, _ := ParsePosition(e.From) // 文字列から Position を復元するヘルパー
	toPos, _ := ParsePosition(e.To)
	var dropPiece PieceType
	if e.Drop {
		dropPiece = PieceType(e.DropPiece)
	}
	return Move{
		From:      fromPos,
		To:        toPos,
		Drop:      e.Drop,
		DropPiece: dropPiece,
		Promote:   e.Promote,
	}
}

// ToEntity はドメイン Move → MoveEntity への変換
func (m Move) ToEntity(gameID string, idx int) MoveEntity {
	return MoveEntity{
		GameID:    gameID,
		Index:     idx, // 何手目
		From:      m.From.String(),
		To:        m.To.String(),
		Drop:      m.Drop,
		DropPiece: string(m.DropPiece),
		Promote:   m.Promote,
		CreatedAt: time.Now(),
	}
}
