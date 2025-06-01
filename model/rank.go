package model

// RankType はユーザーの級・段を表す列挙型です
// 'kyu' = 級、'dan' = 段
// DBには文字列(varchar)として保存します
type RankType string

const (
	RankTypeKyu RankType = "kyu" // 級
	RankTypeDan RankType = "dan" // 段
)

// AllRankTypes はすべての RankType を列挙したスライスです
var AllRankTypes = []RankType{
	RankTypeKyu,
	RankTypeDan,
}

// String は RankType を文字列表現に変換します
func (rt RankType) String() string {
	return string(rt)
}
