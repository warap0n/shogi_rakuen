package model

// Position は盤面の升目を (File, Rank) で表します。
// どちらも 1〜9 の範囲を取る想定です
type Position struct {
	Rank int // 段 0..8
	File int // 筋 0..8
}

// OnBoard は盤内かどうかをチェックします
func (p Position) OnBoard() bool {
	return p.File >= 0 && p.File <= 8 && p.Rank >= 0 && p.Rank <= 8
}
