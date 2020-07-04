package typeclass

type Ofg interface {
	LessThan(Ofg) bool
	MoreThan(Ofg) bool
}
