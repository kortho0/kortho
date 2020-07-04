package typeclass

type Read interface {
	Read([]byte) ([]byte, error) 
}
