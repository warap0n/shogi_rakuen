package model

import (
	"fmt"
)

// Position は盤面の升目を (File, Rank) で表します。
// どちらも 0〜8 の範囲を取る想定です
// String と ParsePosition で 1〜9 の文字列表記に変換できます。
type Position struct {
	Rank int // 段 0..8
	File int // 筋 0..8
}

// OnBoard は盤内かどうかをチェックします
func (p Position) OnBoard() bool {
	return p.File >= 0 && p.File <= 8 && p.Rank >= 0 && p.Rank <= 8
}

func (p Position) InPromotionZone(color Color) bool {
	switch color {
	case Black:
		return p.Rank >= 0 && p.Rank <= 2
	case White:
		return p.Rank >= 6 && p.Rank <= 8
	}
	return false
}

// String は Position を "12" のような 1〜9 の文字列に変換します
// File+1 を先、Rank+1 を後ろに連結（例: File=0,Rank=0 -> "11"）
func (p Position) String() string {
	return fmt.Sprintf("%d%d", p.File+1, p.Rank+1)
}

// ParsePosition は "12" のような文字列を解析して Position を返します
// 長さが 2 で、各文字が '1'〜'9' の範囲であることを期待します
func ParsePosition(s string) (Position, error) {
	if len(s) != 2 {
		return Position{}, fmt.Errorf("invalid position string: %s", s)
	}
	file := int(s[0] - '0')
	rank := int(s[1] - '0')
	if file < 1 || file > 9 || rank < 1 || rank > 9 {
		return Position{}, fmt.Errorf("position out of range: %s", s)
	}
	return Position{File: file - 1, Rank: rank - 1}, nil
}
