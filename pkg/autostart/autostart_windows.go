package autostart

import (
	"log"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const appName = "QuietSnap"

func getExecPath() string {
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Clean(execPath)
}

func Enable() error {
	execPath := getExecPath()
	if execPath == "" {
		return nil
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		log.Println("Error opening registry:", err)
		return err
	}
	defer key.Close()

	cmd := `"` + execPath + `" --autostart`
	log.Println("Enabling autostart:", cmd)
	return key.SetStringValue(appName, cmd)
}

func Disable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	err = key.DeleteValue(appName)
	if err != nil && err != registry.ErrNotExist {
		return err
	}
	return nil
}
