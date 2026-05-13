package writer

import (
	"reflect"
	"testing"

	"github.com/gofurry/fiberx/internal/renderer"
)

func TestWriteRefusesOverwrite(t *testing.T) {
	targetDir := t.TempDir()
	rendered := renderer.Result{
		Files: []renderer.File{
			{Path: "main.go", Content: []byte("package main\n")},
		},
	}

	result, err := New(targetDir).Write(rendered)
	if err != nil {
		t.Fatalf("first Write() returned error: %v", err)
	}
	if result.WrittenFiles != 1 {
		t.Fatalf("expected one written file, got %d", result.WrittenFiles)
	}
	if !reflect.DeepEqual(result.WrittenPaths, []string{"main.go"}) {
		t.Fatalf("expected written paths to be tracked, got %#v", result.WrittenPaths)
	}

	if _, err := New(targetDir).Write(rendered); err == nil {
		t.Fatal("expected second Write() to fail on overwrite")
	}
}
