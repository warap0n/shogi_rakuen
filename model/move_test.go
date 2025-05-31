package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoveEntity_ToDomain(t *testing.T) {
	entity := MoveEntity{
		ID:        100,
		GameID:    "game-test",
		Index:     2,
		From:      "77", // 内部的には (6,6)
		To:        "76", // 内部的には (6,5)
		Drop:      false,
		DropPiece: "",
		Promote:   true,
		CreatedAt: time.Now(),
	}

	mv := entity.ToDomain()

	// Position フィールドの復元チェック
	expectFrom, err := ParsePosition("77")
	require.NoError(t, err)
	expectTo, err := ParsePosition("76")
	require.NoError(t, err)

	assert.Equal(t, expectFrom, mv.From)
	assert.Equal(t, expectTo, mv.To)

	// その他のプロパティも確認
	assert.False(t, mv.Drop)
	assert.Equal(t, PieceType(""), mv.DropPiece)
	assert.True(t, mv.Promote)
}

func TestMove_ToEntity(t *testing.T) {
	// "77" → (6,6), "81" → (7,0)
	fromPos, err := ParsePosition("77")
	require.NoError(t, err)
	toPos, err := ParsePosition("81")
	require.NoError(t, err)

	original := Move{
		From:      fromPos,
		To:        toPos,
		Drop:      true,
		DropPiece: PieceType("P"),
		Promote:   false,
	}

	entity := original.ToEntity("game-xyz", 3)

	assert.Equal(t, "game-xyz", entity.GameID)
	assert.Equal(t, 3, entity.Index)
	assert.Equal(t, "77", entity.From)
	assert.Equal(t, "81", entity.To)
	assert.True(t, entity.Drop)
	assert.Equal(t, "P", entity.DropPiece)
	assert.False(t, entity.Promote)
}

// 往復変換 (Move → MoveEntity → Move) が値を壊さないこと
func TestRoundTrip_MoveEntity(t *testing.T) {
	fromPos, err := ParsePosition("33")
	require.NoError(t, err)
	toPos, err := ParsePosition("59")
	require.NoError(t, err)

	original := Move{
		From:      fromPos,
		To:        toPos,
		Drop:      false,
		DropPiece: PieceType(""),
		Promote:   true,
	}

	entity := original.ToEntity("round-trip-id", 0)
	reconstructed := entity.ToDomain()

	assert.Equal(t, original.From, reconstructed.From)
	assert.Equal(t, original.To, reconstructed.To)
	assert.Equal(t, original.Drop, reconstructed.Drop)
	assert.Equal(t, original.DropPiece, reconstructed.DropPiece)
	assert.Equal(t, original.Promote, reconstructed.Promote)
}

func TestMoveEntity_TableName(t *testing.T) {
	var me MoveEntity
	assert.Equal(t, "moves", me.TableName())
}

func TestMoveEntity_ToDomain_WithDrop(t *testing.T) {
	// "DropPiece" に "P" をセットすると、ToDomain で PieceType("P") が返るはず
	entity := MoveEntity{
		ID:        5,
		GameID:    "game-drop",
		Index:     1,
		From:      "55", // (内部 (4,4))
		To:        "54", // (内部 (4,3))
		Drop:      true,
		DropPiece: "P", // 歩を打つケース
		Promote:   false,
		CreatedAt: time.Now(),
	}

	mv := entity.ToDomain()

	// Drop=true のケースなので、mv.DropPiece に "P" が入っている
	expectedPosFrom, err := ParsePosition("55")
	require.NoError(t, err)
	expectedPosTo, err := ParsePosition("54")
	require.NoError(t, err)

	assert.Equal(t, expectedPosFrom, mv.From)
	assert.Equal(t, expectedPosTo, mv.To)
	assert.True(t, mv.Drop)
	assert.Equal(t, PieceType("P"), mv.DropPiece)
	assert.False(t, mv.Promote)
}
