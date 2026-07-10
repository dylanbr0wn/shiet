package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	applog "github.com/dylanbr0wn/shiet/internal/log"
)

func TestLogPath(t *testing.T) {
	app := &App{logPath: "/tmp/shiet/shiet.log"}
	if got := app.LogPath(); got != "/tmp/shiet/shiet.log" {
		t.Fatalf("LogPath: got %q", got)
	}
	if got := (*App)(nil).LogPath(); got != "" {
		t.Fatalf("nil LogPath: got %q", got)
	}
}

func TestRevealLogFolder_createsDirAndOpens(t *testing.T) {
	root := t.TempDir()
	logFile := filepath.Join(root, "nested", "missing", "shiet.log")
	// File must not exist; only the parent dir should be created on reveal.
	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		t.Fatalf("expected missing log file, got err=%v", err)
	}

	var opened string
	prev := openFolderFn
	openFolderFn = func(dir string) error {
		opened = dir
		return nil
	}
	t.Cleanup(func() { openFolderFn = prev })

	var buf bytes.Buffer
	app := &App{log: applog.New(&buf), logPath: logFile}
	if err := app.RevealLogFolder(); err != nil {
		t.Fatalf("RevealLogFolder: %v", err)
	}

	wantDir := filepath.Dir(logFile)
	if opened != wantDir {
		t.Fatalf("opened dir: got %q want %q", opened, wantDir)
	}
	if st, err := os.Stat(wantDir); err != nil || !st.IsDir() {
		t.Fatalf("expected dir created at %q: err=%v", wantDir, err)
	}
	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		t.Fatalf("reveal must not create the log file itself, err=%v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"op":"log.reveal_folder"`) {
		t.Fatalf("expected reveal log, got %q", out)
	}
}

func TestRevealLogFolder_emptyPath(t *testing.T) {
	var buf bytes.Buffer
	app := &App{log: applog.New(&buf)}
	if err := app.RevealLogFolder(); err == nil {
		t.Fatal("expected error for empty log path")
	}
}
