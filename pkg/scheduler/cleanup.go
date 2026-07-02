package scheduler

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"autoscreen/pkg/config"
)

func executeCleanup(cfg *config.TaskConfig, isStartup bool) {
	if cfg.DeletePolicy == "" || cfg.DeletePolicy == "never" {
		return
	}

	// For previous_run, it should only execute once at startup.
	if cfg.DeletePolicy == "previous_run" && !isStartup {
		return
	}

	files, err := os.ReadDir(cfg.SaveDirectory)
	if err != nil {
		log.Printf("Cleanup error: failed to read directory %s: %v", cfg.SaveDirectory, err)
		return
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if !strings.HasPrefix(name, "screenshot_") {
			continue
		}
		if !strings.HasSuffix(name, ".png") && !strings.HasSuffix(name, ".jpg") && !strings.HasSuffix(name, ".jpeg") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		modTime := info.ModTime()
		shouldDelete := false

		switch cfg.DeletePolicy {
		case "previous_run":
			// If it's startup, just wipe existing screenshot files
			shouldDelete = true
		case "yesterday":
			// Delete files from yesterday or older
			if modTime.Before(todayStart) {
				shouldDelete = true
			}
		case "keep_n_days":
			if cfg.DeleteDays > 0 {
				cutoff := todayStart.AddDate(0, 0, -cfg.DeleteDays+1)
				if modTime.Before(cutoff) {
					shouldDelete = true
				}
			}
		}

		if shouldDelete {
			fullPath := filepath.Join(cfg.SaveDirectory, name)
			err := MoveToTrash(fullPath)
			if err != nil {
				log.Printf("Failed to move %s to trash: %v", name, err)
			} else {
				log.Printf("Moved %s to trash (Policy: %s)", name, cfg.DeletePolicy)
			}
		}
	}
}
