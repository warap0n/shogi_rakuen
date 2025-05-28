package model

import "time"

// Move は盤上の１手を表します
type Move struct {
	From      Position  // 移動元（Drop のとき無視）
	To        Position  // 移動先
	Drop      bool      // true=打ち
	DropPiece PieceType // Drop=true のとき、打つ駒の種類
	Promote   bool      // 成り

}

// MoveEntity は DB 上の moves テーブルに対応するエンティティです
type MoveEntity struct {
	ID        uint      `gorm:"primaryKey"`     // 自動付番
	GameID    string    `gorm:"not null;index"` // どの対局か
	Index     int       `gorm:"not null"`       // 何手目か
	From      string    `gorm:"not null"`       // 文字列化した Position (例: "7g")
	To        string    `gorm:"not null"`
	Drop      bool      `gorm:"not null"`
	DropPiece string    `gorm:"size:1"` // 駒種を文字で
	Promote   bool      `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"` // 打った時刻
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
		Index:     idx,
		From:      m.From.String(), // Position.String() を使って "7g" の形式に
		To:        m.To.String(),
		Drop:      m.Drop,
		DropPiece: string(m.DropPiece),
		Promote:   m.Promote,
		CreatedAt: time.Now(), // 現在時刻を自動設定
	}
}
