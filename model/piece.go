package model

// Color は先手／後手を表します
type Color int

const (
	Black Color = iota
	White
)

// PieceType は駒の種類を表します
type PieceType int

const (
	King PieceType = iota
	Rook
	Bishop
	Gold
	Silver
	Knight
	Lance
	Pawn
)

// Piece は１枚の駒を表します
type Piece struct {
	Type     PieceType // 種類
	Color    Color     // 先手/後手
	Promoted bool      // 成り駒かどうか
}

func (pt PieceType) Promotable() bool {
	switch pt {
	case Pawn, Lance, Knight, Silver, Bishop, Rook:
		return true
	default:
		return false
	}
}

// ValidMove は「駒を from から to に移動する手」が合法かどうかを返します。
//   - Drop（打ち）や成りは別処理。
//   - 手番は p.Color、盤面は b を参照。
//   - p.Promoted が true の場合、歩・香・桂・銀 は金将と同じ動きに。
//   - 中継マスのチェックが必要な駒（香・飛・角）は盤面上を走査。
//   - toに味方駒がいるかどうかはGAME側でチェックする。
func (p *Piece) ValidMove(from, to Position, b *Board) bool {
	// 1) 盤外チェック
	if !to.OnBoard() {
		return false
	}
	// 2) 同じマスには動けない
	if from == to {
		return false
	}
	// 3) 成り駒の動きを種別ごとにハンドリング
	if p.Promoted {
		switch p.Type {
		case Pawn, Lance, Knight, Silver:
			// 成歩・成香・成桂・成銀 は金将と同じ
			return validGold(p.Color, from, to)
		case Bishop:
			// 馬：角の動き + 周囲１マス（王の動き）
			return validBishop(from, to, b) ||
				validKing(from, to)
		case Rook:
			// 龍：飛車の動き + 周囲１マス（王の動き）
			return validRook(from, to, b) ||
				validKing(from, to)
		default:
			// Gold／King はそもそも成らないはずなので false
			return false
		}
	}
	// 4) 非プロモート駒の動き
	switch p.Type {
	case Pawn:
		return validPawn(p.Color, from, to)
	case Lance:
		return validLance(p.Color, from, to, b)
	case Knight:
		return validKnight(p.Color, from, to)
	case Silver:
		return validSilver(p.Color, from, to)
	case Gold:
		return validGold(p.Color, from, to)
	case King:
		return validKing(from, to)
	case Bishop:
		return validBishop(from, to, b)
	case Rook:
		return validRook(from, to, b)
	default:
		return false
	}
}

// ===== 以下、駒種ごとのヘルパー =====

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ----------------------
//
//	歩
//
// ----------------------
func validPawn(color Color, f, t Position) bool {
	dr := t.Rank - f.Rank
	df := t.File - f.File
	if df != 0 {
		return false
	}
	if color == Black {
		return dr == -1
	}
	return dr == 1
}

// ----------------------
//
//	香車：一直線前方（盤の端まで）、途中に駒がない
//
// ----------------------
func validLance(color Color, f, t Position, b *Board) bool {
	df := t.File - f.File
	dr := t.Rank - f.Rank
	if df != 0 {
		return false
	}
	// 進行方向チェック
	step := -1
	if color == White {
		step = 1
	}
	if dr == 0 || (dr/step) <= 0 {
		return false
	}
	// 中間マスに駒がないか
	for r := f.Rank + step; r != t.Rank; r += step {
		if piece, _ := b.PieceAt(Position{Rank: r, File: f.File}); piece != nil {
			return false
		}
	}
	return true
}

// ----------------------
//
//	桂馬：2前方＋1横
//
// ----------------------
func validKnight(color Color, f, t Position) bool {
	dr := t.Rank - f.Rank
	df := t.File - f.File
	if abs(df) != 1 {
		return false
	}
	if color == Black {
		return dr == -2
	}
	return dr == 2
}

// ----------------------
//
//	銀：斜め４方向＋前方直進
//
// ----------------------
func validSilver(color Color, f, t Position) bool {
	dr := t.Rank - f.Rank
	df := t.File - f.File
	// 斜めは常にOK
	if abs(dr) == 1 && abs(df) == 1 {
		return true
	}
	// 直進は「前方のみ」
	if df == 0 {
		if color == Black && dr == -1 {
			return true
		}
		if color == White && dr == 1 {
			return true
		}
	}
	return false
}

// ----------------------
//
//	金：前後左右＋前斜め＋後直進（後ろ斜めは不可）
//
// ----------------------
func validGold(color Color, f, t Position) bool {
	dr := t.Rank - f.Rank
	df := t.File - f.File
	// １マス移動
	if abs(dr) > 1 || abs(df) > 1 {
		return false
	}
	// 後ろ斜め禁止
	if color == Black && dr == 1 && abs(df) == 1 {
		return false
	}
	if color == White && dr == -1 && abs(df) == 1 {
		return false
	}
	// それ以外の１マスはOK
	return true
}

// ----------------------
//
//	王：周囲８方向すべて１マス
//
// ----------------------
func validKing(f, t Position) bool {
	dr := t.Rank - f.Rank
	df := t.File - f.File
	return abs(dr) <= 1 && abs(df) <= 1
}

// ----------------------
//
//	角：斜め全方向スライド
//
// ----------------------
func validBishop(f, t Position, b *Board) bool {
	dr := t.Rank - f.Rank
	df := t.File - f.File
	if abs(dr) != abs(df) {
		return false
	}
	stepR := 1
	if dr < 0 {
		stepR = -1
	}
	stepF := 1
	if df < 0 {
		stepF = -1
	}
	// 中間経路チェック
	for r, c := f.Rank+stepR, f.File+stepF; r != t.Rank; r, c = r+stepR, c+stepF {
		if piece, _ := b.PieceAt(Position{Rank: r, File: c}); piece != nil {
			return false
		}
	}
	return true
}

// ----------------------
//
//	飛車：縦横全方向スライド
//
// ----------------------
func validRook(f, t Position, b *Board) bool {
	dr := t.Rank - f.Rank
	df := t.File - f.File
	// 直線かどうか
	if dr != 0 && df != 0 {
		return false
	}
	// 進行方向とステップ
	stepR, stepF := 0, 0
	if dr != 0 {
		if dr > 0 {
			stepR = 1
		} else {
			stepR = -1
		}
	} else {
		if df > 0 {
			stepF = 1
		} else {
			stepF = -1
		}
	}
	// 中間経路チェック
	for r, c := f.Rank+stepR, f.File+stepF; r != t.Rank || c != t.File; r, c = r+stepR, c+stepF {
		if piece, _ := b.PieceAt(Position{Rank: r, File: c}); piece != nil {
			return false
		}
	}
	return true
}
