package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	SaveDirectory     string `json:"saveDirectory"`
	StartDate         string `json:"startDate"`
	EndDate           string `json:"endDate"`
	DailyStartTime    string `json:"dailyStartTime"`
	DailyEndTime      string `json:"dailyEndTime"`
	IntervalMinutes   int    `json:"intervalMinutes"`
	Mode              string `json:"mode"`   // "fullscreen" or "region"
	Format            string `json:"format"` // "png" or "jpg"
	RegionX           int    `json:"regionX"`
	RegionY           int    `json:"regionY"`
	RegionW           int    `json:"regionW"`
	RegionH           int    `json:"regionH"`
	AutoStart         bool   `json:"autoStart"`
	ShowUIOnAutoStart bool   `json:"showUIOnAutoStart"`
}

func GetConfigPath() string {
	execPath, err := os.Executable()
	if err != nil {
		return "autoscreen_config.json"
	}
	return filepath.Join(filepath.Dir(execPath), "autoscreen_config.json")
}

func LoadConfig() *Config {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		// Default config
		return &Config{
			Mode:              "fullscreen",
			Format:            "png",
			IntervalMinutes:   5,
			AutoStart:         false,
			ShowUIOnAutoStart: true,
		}
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &Config{Mode: "fullscreen", Format: "png", IntervalMinutes: 5, ShowUIOnAutoStart: true}
	}
	return &cfg
}

func SaveConfig(cfg *Config) error {
	path := GetConfigPath()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
