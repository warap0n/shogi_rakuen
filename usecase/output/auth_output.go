package output

import "shogi-rakuen/model"

type SignupOutput struct {
	ID        uint
	Email     string
	Username  string
	RankType  model.RankType
	RankLevel int
}

type LoginOutput struct {
	Token string
}

type GetUserByIdOutput struct {
	ID        uint
	Email     string
	Username  string
	RankType  model.RankType
	RankLevel int
}
