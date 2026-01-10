package internal

const (
	LocationWorld = "world"
	LocationUK    = "uk"
)

type Greeter interface {
	Greet(location string) string
}
