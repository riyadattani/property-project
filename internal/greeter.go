package internal

type GreeterService struct{}

func NewGreeter() *GreeterService {
	return &GreeterService{}
}

func (g *GreeterService) Greet(location string) string {
	switch location {
	case LocationUK:
		return "Hello, UK!"
	case LocationWorld:
		return "Hello, World!"
	default:
		return "Hello, World!"
	}
}
