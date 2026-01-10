package specifications

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"propertyProject/internal"
)

// HTTPGreeterAdapter wraps HTTP handlers to satisfy the GreeterContract
type HTTPGreeterAdapter struct {
	handler *internal.Handler
}

func NewHTTPGreeterAdapter(handler *internal.Handler) *HTTPGreeterAdapter {
	return &HTTPGreeterAdapter{handler: handler}
}

func (a *HTTPGreeterAdapter) Greet(location string) string {
	var handlerFunc http.HandlerFunc
	var path string

	switch location {
	case internal.LocationWorld:
		handlerFunc = a.handler.HelloWorldHandler
		path = "/hello-world"
	case internal.LocationUK:
		handlerFunc = a.handler.HelloUKHandler
		path = "/hello-uk"
	default:
		handlerFunc = a.handler.HelloWorldHandler
		path = "/hello-world"
	}

	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	handlerFunc(rec, req)

	body, _ := io.ReadAll(rec.Body)
	// Extract message from HTML response - look for content between <h2> tags
	content := string(body)
	start := strings.Index(content, "<h2>")
	end := strings.Index(content, "</h2>")
	if start != -1 && end != -1 {
		return content[start+4 : end]
	}
	return content
}

// TestGreeter_HTTP runs specs against the HTTP adapter (end-to-end)
func TestGreeter_HTTP(t *testing.T) {
	// Change to project root so templates can be found
	projectRoot := findProjectRoot()
	originalDir, _ := os.Getwd()
	os.Chdir(projectRoot)
	defer os.Chdir(originalDir)

	greeter := internal.NewGreeter()
	handler := internal.NewHandler(greeter)
	adapter := NewHTTPGreeterAdapter(handler)

	GreeterSpec(t, adapter)
}

func findProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}
