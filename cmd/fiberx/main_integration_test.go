//go:build integration

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLIBuildP0(t *testing.T) {
	t.Setenv("FIBERX_MANIFEST_ROOT", manifestRootForCLI(t))

	workdir := t.TempDir()
	withWorkingDir(t, workdir, func() {
		_ = captureStdout(t, func() error {
			return run([]string{"new", "demo", "--preset", "light"})
		})
	})

	projectDir := filepath.Join(workdir, "demo")
	initGitRepoForCLI(t, projectDir)

	withWorkingDir(t, projectDir, func() {
		output := captureStdout(t, func() error {
			return run([]string{"build", "server", "--target", runtime.GOOS + "/" + runtime.GOARCH})
		})
		if !strings.Contains(output, "built project=demo") || !strings.Contains(output, "artifacts: 1") || !strings.Contains(output, "dry-run: false") {
			t.Fatalf("expected build summary output, got:\n%s", output)
		}
	})

	binaryPath := filepath.Join(projectDir, "dist", "server", runtime.GOOS+"_"+runtime.GOARCH, "demo")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}
	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("expected built artifact at %q: %v", binaryPath, err)
	}

	stalePath := filepath.Join(projectDir, "dist", "stale.txt")
	if err := os.WriteFile(stalePath, []byte("stale"), 0o644); err != nil {
		t.Fatalf("write stale artifact: %v", err)
	}

	withWorkingDir(t, projectDir, func() {
		_ = captureStdout(t, func() error {
			return run([]string{"build", "--clean", "--target", runtime.GOOS + "/" + runtime.GOARCH})
		})
	})
	if _, err := os.Stat(stalePath); !os.IsNotExist(err) {
		t.Fatalf("expected stale artifact to be removed by clean build, got %v", err)
	}
}

func TestCLIBuildP2(t *testing.T) {
	t.Setenv("FIBERX_MANIFEST_ROOT", manifestRootForCLI(t))

	workdir := t.TempDir()
	withWorkingDir(t, workdir, func() {
		_ = captureStdout(t, func() error {
			return run([]string{"new", "demo", "--preset", "light"})
		})
	})

	projectDir := filepath.Join(workdir, "demo")
	initGitRepoForCLI(t, projectDir)

	configPath := filepath.Join(projectDir, "fiberx.yaml")
	configBody := `project:
  name: demo
  module: github.com/example/demo
build:
  out_dir: dist
  clean: true
  parallel: true
  version:
    source: git
    package: github.com/example/demo/internal/version
  defaults:
    cgo: false
    trimpath: true
    ldflags:
      - "-s -w"
      - "-X github.com/example/demo/internal/version.Version={{.Version}}"
      - "-X github.com/example/demo/internal/version.Commit={{.Commit}}"
      - "-X github.com/example/demo/internal/version.BuildTime={{.BuildTime}}"
  checksum:
    enabled: true
    algorithm: sha256
  compress:
    upx:
      enabled: false
      level: 5
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - ` + runtime.GOOS + `/` + runtime.GOARCH + `
      archive:
        enabled: true
        format: auto
        files:
          - README.md
          - config
      pre_hooks: []
      post_hooks: []
`
	if err := os.WriteFile(configPath, []byte(configBody), 0o644); err != nil {
		t.Fatalf("write generated build config: %v", err)
	}

	withWorkingDir(t, projectDir, func() {
		output := captureStdout(t, func() error {
			return run([]string{"build", "server", "--target", runtime.GOOS + "/" + runtime.GOARCH, "--dry-run"})
		})
		if !strings.Contains(output, "build plan project=demo") || !strings.Contains(output, "dry-run: true") || !strings.Contains(output, "archive=") || !strings.Contains(output, "checksum:") || !strings.Contains(output, "build metadata:") || !strings.Contains(output, "release manifest:") {
			t.Fatalf("expected dry-run build plan output, got:\n%s", output)
		}
	})

	if _, err := os.Stat(filepath.Join(projectDir, "dist")); !os.IsNotExist(err) {
		t.Fatalf("expected dry-run not to create dist dir, got %v", err)
	}

	withWorkingDir(t, projectDir, func() {
		output := captureStdout(t, func() error {
			return run([]string{"build", "server", "--target", runtime.GOOS + "/" + runtime.GOARCH})
		})
		if !strings.Contains(output, "built project=demo") || !strings.Contains(output, "dry-run: false") || !strings.Contains(output, "archive=") || !strings.Contains(output, "checksum:") || !strings.Contains(output, "build metadata:") || !strings.Contains(output, "release manifest:") {
			t.Fatalf("expected p2 build output, got:\n%s", output)
		}
	})

	distDir := filepath.Join(projectDir, "dist", "server", runtime.GOOS+"_"+runtime.GOARCH)
	checksumPath := filepath.Join(projectDir, "dist", "SHA256SUMS")
	if _, err := os.Stat(checksumPath); err != nil {
		t.Fatalf("expected checksum output: %v", err)
	}

	var archivePath string
	if runtime.GOOS == "windows" {
		archivePath = filepath.Join(distDir, "demo_"+runtime.GOOS+"_"+runtime.GOARCH+".zip")
	} else {
		archivePath = filepath.Join(distDir, "demo_"+runtime.GOOS+"_"+runtime.GOARCH+".tar.gz")
	}
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("expected archive output at %q: %v", archivePath, err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, "dist", "build-metadata.json")); err != nil {
		t.Fatalf("expected build metadata output: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, "dist", "release-manifest.json")); err != nil {
		t.Fatalf("expected release manifest output: %v", err)
	}
}

func TestCLIBuildP3Profile(t *testing.T) {
	t.Setenv("FIBERX_MANIFEST_ROOT", manifestRootForCLI(t))

	workdir := t.TempDir()
	withWorkingDir(t, workdir, func() {
		_ = captureStdout(t, func() error {
			return run([]string{"new", "demo", "--preset", "light"})
		})
	})

	projectDir := filepath.Join(workdir, "demo")
	initGitRepoForCLI(t, projectDir)

	withWorkingDir(t, projectDir, func() {
		output := captureStdout(t, func() error {
			return run([]string{"build", "server", "--profile", "prod", "--target", runtime.GOOS + "/" + runtime.GOARCH, "--dry-run"})
		})
		if !strings.Contains(output, "profile: prod") || !strings.Contains(output, "build metadata:") || !strings.Contains(output, "release manifest:") {
			t.Fatalf("expected profile dry-run output, got:\n%s", output)
		}
	})

	if _, err := os.Stat(filepath.Join(projectDir, "dist")); !os.IsNotExist(err) {
		t.Fatalf("expected dry-run not to create dist dir, got %v", err)
	}

	withWorkingDir(t, projectDir, func() {
		output := captureStdout(t, func() error {
			return run([]string{"build", "server", "--profile", "prod", "--target", runtime.GOOS + "/" + runtime.GOARCH})
		})
		if !strings.Contains(output, "profile: prod") {
			t.Fatalf("expected prod profile in build output, got:\n%s", output)
		}
	})

	if _, err := os.Stat(filepath.Join(projectDir, "dist", "prod", "build-metadata.json")); err != nil {
		t.Fatalf("expected profiled build metadata output: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, "dist", "prod", "release-manifest.json")); err != nil {
		t.Fatalf("expected profiled release manifest output: %v", err)
	}
}

func TestCLIBuildHookSafetyFlags(t *testing.T) {
	t.Setenv("FIBERX_MANIFEST_ROOT", manifestRootForCLI(t))

	workdir := t.TempDir()
	withWorkingDir(t, workdir, func() {
		_ = captureStdout(t, func() error {
			return run([]string{"new", "demo", "--preset", "light"})
		})
	})

	projectDir := filepath.Join(workdir, "demo")
	initGitRepoForCLI(t, projectDir)
	configPath := filepath.Join(projectDir, "fiberx.yaml")
	configBody := `project:
  name: demo
  module: github.com/example/demo
build:
  out_dir: dist
  clean: true
  parallel: false
  version:
    source: git
    package: github.com/example/demo/internal/version
  defaults:
    cgo: false
    trimpath: true
    ldflags:
      - "-s -w"
      - "-X github.com/example/demo/internal/version.Version={{.Version}}"
      - "-X github.com/example/demo/internal/version.Commit={{.Commit}}"
      - "-X github.com/example/demo/internal/version.BuildTime={{.BuildTime}}"
  checksum:
    enabled: false
    algorithm: sha256
  compress:
    upx:
      enabled: false
      level: 5
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - ` + runtime.GOOS + `/` + runtime.GOARCH + `
      archive:
        enabled: false
        format: auto
        files: []
      pre_hooks:
        - name: inspect
          command: ["go", "version"]
      post_hooks:
        - name: inspect-post
          command: ["go", "version"]
`
	if err := os.WriteFile(configPath, []byte(configBody), 0o644); err != nil {
		t.Fatalf("write generated build config: %v", err)
	}

	withWorkingDir(t, projectDir, func() {
		err := run([]string{"build", "server", "--target", runtime.GOOS + "/" + runtime.GOARCH})
		if err == nil || !strings.Contains(err.Error(), "rerun with --yes to approve them or --no-hooks to skip them") {
			t.Fatalf("expected non-interactive hook confirmation error, got %v", err)
		}
	})

	withWorkingDir(t, projectDir, func() {
		output := captureStdout(t, func() error {
			return run([]string{"build", "server", "--target", runtime.GOOS + "/" + runtime.GOARCH, "--no-hooks"})
		})
		if !strings.Contains(output, "hooks: skipped by --no-hooks") {
			t.Fatalf("expected --no-hooks output, got:\n%s", output)
		}
	})

	withWorkingDir(t, projectDir, func() {
		output := captureStdout(t, func() error {
			return run([]string{"build", "server", "--target", runtime.GOOS + "/" + runtime.GOARCH, "--yes"})
		})
		if !strings.Contains(output, "hooks: approved by --yes") {
			t.Fatalf("expected --yes output, got:\n%s", output)
		}
	})
}
