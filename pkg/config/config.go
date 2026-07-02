package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type TaskConfig struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Enabled         bool   `json:"enabled"`
	SaveDirectory   string `json:"saveDirectory"`
	StartDate       string `json:"startDate"`
	EndDate         string `json:"endDate"`
	DailyStartTime  string `json:"dailyStartTime"`
	DailyEndTime    string `json:"dailyEndTime"`
	IntervalMinutes int    `json:"intervalMinutes"`
	Mode            string `json:"mode"`   // "fullscreen" or "region"
	Format          string `json:"format"` // "png" or "jpg"
	RegionX         int    `json:"regionX"`
	RegionY         int    `json:"regionY"`
	RegionW         int    `json:"regionW"`
	RegionH         int    `json:"regionH"`
	DeletePolicy    string `json:"deletePolicy"`
	DeleteDays      int    `json:"deleteDays"`
}

type Config struct {
	Tasks             []TaskConfig `json:"tasks"`
	AutoStart         bool         `json:"autoStart"`
	ShowUIOnAutoStart bool         `json:"showUIOnAutoStart"`
}

func GetConfigPath() string {
	execPath, err := os.Executable()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(filepath.Dir(execPath), "config.json")
}

func LoadConfig() *Config {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	defaultCfg := &Config{
		Tasks:             []TaskConfig{},
		AutoStart:         false,
		ShowUIOnAutoStart: true,
	}
	if err != nil {
		return defaultCfg
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultCfg
	}
	
	if cfg.Tasks == nil {
		cfg.Tasks = []TaskConfig{}
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
