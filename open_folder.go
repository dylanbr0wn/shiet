package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

// openFolderFn opens a directory in the OS file manager. Swapped in tests.
var openFolderFn = openFolder

// openFolder opens dir in the platform file manager (Finder / Explorer / xdg-open).
// Uses Start so the process is detached and does not block the app.
func openFolder(dir string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", dir)
	case "windows":
		cmd = exec.Command("explorer", dir)
	default:
		cmd = exec.Command("xdg-open", dir)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open folder %q: %w", dir, err)
	}
	return nil
}
