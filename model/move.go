package model

// Move は盤上の１手を表します
type Move struct {
	From      Position  // 移動元（Drop のとき無視）
	To        Position  // 移動先
	Drop      bool      // true=打ち
	DropPiece PieceType // Drop=true のとき、打つ駒の種類
	Promote   bool      // 成り
}
