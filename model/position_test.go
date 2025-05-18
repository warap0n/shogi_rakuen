// model/position_test.go
package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPosition_InPromotionZone_Black(t *testing.T) {
	// Black のプロモーションゾーンは Rank 0,1,2
	for _, r := range []int{-1, 0, 1, 2, 3} {
		pos := Position{Rank: r, File: 4}
		want := (r >= 0 && r <= 2)
		assert.Equalf(t, want, pos.InPromotionZone(Black),
			"Black, Rank=%d → InPromotionZone は %v", r, want)
	}
}

func TestPosition_InPromotionZone_White(t *testing.T) {
	// White のプロモーションゾーンは Rank 6,7,8
	for _, r := range []int{5, 6, 7, 8, 9} {
		pos := Position{Rank: r, File: 4}
		want := (r >= 6 && r <= 8)
		assert.Equalf(t, want, pos.InPromotionZone(White),
			"White, Rank=%d → InPromotionZone は %v", r, want)
	}
}

func TestPosition_InPromotionZone_InvalidColor(t *testing.T) {
	// 定義外の Color を渡したら false が返る
	pos := Position{Rank: 1, File: 1}
	const UnknownColor Color = 42
	assert.False(t, pos.InPromotionZone(UnknownColor))
}
