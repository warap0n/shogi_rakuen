package model

// Move は盤上の１手を表します
type Move struct {
	From    Position // 打った駒は From=nil, Drop=true 扱いにする
	To      Position
	Promote bool // 成り動作をしたか
	Drop    bool // 打（持ち駒を打つ）かどうか
}
