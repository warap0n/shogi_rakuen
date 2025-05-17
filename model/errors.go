package model

import "errors"

// Board／Game に関連するエラーを一箇所にまとめる

// 盤外参照
var ErrOutOfBounds = errors.New("position out of board")

// すでにゲーム終了済み
var ErrGameAlreadyFinished = errors.New("game already finished")

// 持ち駒がないのに打とうとした
var ErrNoPieceToDrop = errors.New("no piece to drop")

// 移動元に駒がない
var ErrNoPieceAtSource = errors.New("no piece at source")
