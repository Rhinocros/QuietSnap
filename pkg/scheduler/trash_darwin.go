package scheduler

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

func MoveToTrash(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	
	script := fmt.Sprintf(`tell application "Finder" to move POSIX file "%s" to trash`, absPath)
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("osascript failed: %v", err)
	}
	return nil
}
