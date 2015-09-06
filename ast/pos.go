package ast

//Position is position of source code
type Position struct {
	Begin int
	End   int
}

//Pos interface
type Pos interface {
	SetPosition(Position)
	GetPosition() Position
}

//PosImpl provide implementation for Pos
type PosImpl struct {
	pos Position
}

//SetPosition function
func (pi *PosImpl) SetPosition(p Position) {
	pi.pos = p
}

//GetPosition function
func (pi *PosImpl) GetPosition() Position {
	return pi.pos
}
