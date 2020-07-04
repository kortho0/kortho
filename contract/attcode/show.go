package typeclass

type Show interface {
	Show() ([]byte, error)
}
