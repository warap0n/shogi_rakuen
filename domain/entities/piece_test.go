package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanPromote(t *testing.T) {
	tests := []struct {
		name        string
		pieceType   PieceType
		wantPromote bool
	}{
		{"歩は成れる", Fu, true},
		{"香は成れる", Kyo, true},
		{"桂は成れる", Kei, true},
		{"銀は成れる", Gin, true},
		{"角は成れる", Kaku, true},
		{"飛は成れる", Hisha, true},
		{"金は成れない", Kin, false},
		{"王は成れない", Ou, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Piece{Type: tt.pieceType}
			assert.Equal(t, tt.wantPromote, p.CanPromote())
		})
	}
}

func TestPromote(t *testing.T) {
	t.Run("成れる駒はPromoteでIsPromotedがtrueになる", func(t *testing.T) {
		p := &Piece{Type: Fu}
		err := p.Promote()
		assert.NoError(t, err)
		assert.True(t, p.IsPromoted)
	})

	t.Run("成れない駒はPromoteでエラーが返る", func(t *testing.T) {
		p := &Piece{Type: Kin}
		err := p.Promote()
		assert.Error(t, err)
		assert.False(t, p.IsPromoted)
	})
}

func TestDemote(t *testing.T) {
	t.Run("DemoteでIsPromotedがfalseになる", func(t *testing.T) {
		p := &Piece{Type: Kaku, IsPromoted: true}
		err := p.Demote()
		assert.NoError(t, err)
		assert.False(t, p.IsPromoted)
	})

	t.Run("DemoteでIsPromotedがfalseの時はエラーが返る", func(t *testing.T) {
		p := &Piece{Type: Kaku, IsPromoted: false}
		err := p.Demote()
		assert.Error(t, err)
		assert.False(t, p.IsPromoted)
	})
}
