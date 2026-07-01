package autostart

import "log"

// Enable enables auto-start on Linux (currently unsupported)
func Enable() error {
	log.Println("Auto-start is currently only supported on Windows")
	return nil
}

// Disable disables auto-start on Linux (currently unsupported)
func Disable() error {
	log.Println("Auto-start is currently only supported on Windows")
	return nil
}
