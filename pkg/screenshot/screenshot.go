package screenshot

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/kbinani/screenshot"
	"autoscreen/pkg/config"
)

func Capture(cfg *config.TaskConfig) error {
	if cfg.SaveDirectory == "" {
		return fmt.Errorf("save directory is not set")
	}

	err := os.MkdirAll(cfg.SaveDirectory, 0755)
	if err != nil {
		return fmt.Errorf("failed to create save directory: %v", err)
	}

	var img *image.RGBA

	if cfg.Mode == "region" {
		rect := image.Rect(cfg.RegionX, cfg.RegionY, cfg.RegionX+cfg.RegionW, cfg.RegionY+cfg.RegionH)
		img, err = screenshot.CaptureRect(rect)
	} else {
		bounds := screenshot.GetDisplayBounds(0)
		img, err = screenshot.CaptureRect(bounds)
	}

	if err != nil {
		return fmt.Errorf("failed to capture screenshot: %v", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	ext := cfg.Format
	if ext == "" {
		ext = "png"
	}
	filename := fmt.Sprintf("screenshot_%s.%s", timestamp, ext)
	filepath := filepath.Join(cfg.SaveDirectory, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	if ext == "jpg" || ext == "jpeg" {
		return jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	}
	return png.Encode(file, img)
}
