package entities

type PieceType string

const (
	Fu    PieceType = "FU"
	Kyo   PieceType = "KYO"
	Kei   PieceType = "KEI"
	Gin   PieceType = "GIN"
	Kin   PieceType = "KIN"
	Kaku  PieceType = "KAKU"
	Hisha PieceType = "HISHA"
	Ou    PieceType = "OU"
)

type Piece struct {
	Type       PieceType
	IsPromoted bool
	PlayerId   string
}
