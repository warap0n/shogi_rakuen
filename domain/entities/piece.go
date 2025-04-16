package entities

import "errors"

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

func (p *Piece) CanPromote() bool {
	return p.Type != Kin && p.Type != Ou
}

func (p *Piece) Promote() error {
	if !p.CanPromote() {
		return errors.New("this piece cannot be promoted")
	}
	p.IsPromoted = true
	return nil
}

func (p *Piece) Demote() error {
	if !p.IsPromoted {
		return errors.New("this piece is not promoted")
	}
	p.IsPromoted = false
	return nil
}
