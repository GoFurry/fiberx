package writer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofurry/fiberx/internal/renderer"
)

type Result struct {
	DryRun       bool
	WrittenFiles int
	WrittenPaths []string
	TargetDir    string
}

type Writer struct {
	TargetDir string
}

func New(targetDir string) Writer {
	return Writer{TargetDir: targetDir}
}

func (w Writer) Write(rendered renderer.Result) (Result, error) {
	if w.TargetDir == "" {
		return Result{}, fmt.Errorf("target directory cannot be empty")
	}

	if err := os.MkdirAll(w.TargetDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create target directory %q: %w", w.TargetDir, err)
	}

	writtenPaths := make([]string, 0, len(rendered.Files))
	for _, file := range rendered.Files {
		target := filepath.Join(w.TargetDir, filepath.FromSlash(file.Path))
		if _, err := os.Stat(target); err == nil {
			return Result{}, fmt.Errorf("refusing to overwrite existing file %q", target)
		} else if !os.IsNotExist(err) {
			return Result{}, fmt.Errorf("stat target file %q: %w", target, err)
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return Result{}, fmt.Errorf("create parent directory for %q: %w", target, err)
		}
		if err := os.WriteFile(target, file.Content, 0o644); err != nil {
			return Result{}, fmt.Errorf("write target file %q: %w", target, err)
		}
		writtenPaths = append(writtenPaths, filepath.ToSlash(file.Path))
	}

	return Result{
		DryRun:       false,
		WrittenFiles: len(rendered.Files),
		WrittenPaths: writtenPaths,
		TargetDir:    w.TargetDir,
	}, nil
}
