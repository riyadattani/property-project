package specifications

import (
	"testing"

	"propertyProject/internal"
)

// GreeterContract defines the interface that all Greeter implementations must satisfy
type GreeterContract interface {
	Greet(location string) string
}

// GreeterSpec runs the specification tests against any Greeter implementation
func GreeterSpec(t *testing.T, greeter GreeterContract) {
	t.Run("ReturnsHelloWorld", func(t *testing.T) {
		result := greeter.Greet(internal.LocationWorld)
		expected := "Hello, World!"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("ReturnsHelloUK", func(t *testing.T) {
		result := greeter.Greet(internal.LocationUK)
		expected := "Hello, UK!"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

// TestGreeter_Domain runs specs against the pure domain implementation
func TestGreeter_Domain(t *testing.T) {
	greeter := internal.NewGreeter()
	GreeterSpec(t, greeter)
}
