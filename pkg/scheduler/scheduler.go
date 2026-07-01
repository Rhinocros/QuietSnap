package scheduler

import (
	"log"
	"sync"
	"time"

	"autoscreen/pkg/config"
	"autoscreen/pkg/screenshot"
)

var (
	stopChan  chan struct{}
	isRunning bool
	mu        sync.Mutex
)

func Start(cfg *config.Config) error {
	mu.Lock()
	defer mu.Unlock()

	if isRunning {
		return nil
	}

	stopChan = make(chan struct{})
	isRunning = true

	go runSchedule(cfg, stopChan)
	return nil
}

func Stop() {
	mu.Lock()
	defer mu.Unlock()

	if !isRunning {
		return
	}

	close(stopChan)
	isRunning = false
}

func IsRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return isRunning
}

func isWithinBounds(cfg *config.Config) (bool, bool) {
	now := time.Now()
	loc, _ := time.LoadLocation("Local")

	// 1. Check Global Dates
	if cfg.StartDate != "" {
		startDate, err := time.ParseInLocation("2006-01-02", cfg.StartDate, loc)
		if err == nil {
			startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc)
			if now.Before(startDate) {
				return false, false
			}
		}
	}

	if cfg.EndDate != "" {
		endDate, err := time.ParseInLocation("2006-01-02", cfg.EndDate, loc)
		if err == nil {
			// end of the day
			endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, loc)
			if now.After(endDate) {
				return false, true // expired
			}
		}
	}

	// 2. Check Daily Time Window
	if cfg.DailyStartTime != "" && cfg.DailyEndTime != "" {
		startTime, err1 := time.ParseInLocation("15:04", cfg.DailyStartTime, loc)
		endTime, err2 := time.ParseInLocation("15:04", cfg.DailyEndTime, loc)
		if err1 == nil && err2 == nil {
			// Create today's start and end time
			todayStart := time.Date(now.Year(), now.Month(), now.Day(), startTime.Hour(), startTime.Minute(), 0, 0, loc)
			todayEnd := time.Date(now.Year(), now.Month(), now.Day(), endTime.Hour(), endTime.Minute(), 59, 999999999, loc)
			
			// If end time is smaller than start time (e.g. 22:00 to 02:00 next day)
			if todayEnd.Before(todayStart) {
				// We consider if now > todayStart OR now < todayEnd (cross midnight)
				if now.Before(todayStart) && now.After(todayEnd) {
					return false, false
				}
			} else {
				// Normal case
				if now.Before(todayStart) || now.After(todayEnd) {
					return false, false
				}
			}
		}
	}

	return true, false
}

func runSchedule(cfg *config.Config, stop <-chan struct{}) {
	interval := time.Duration(cfg.IntervalMinutes) * time.Minute
	if interval < time.Minute {
		interval = time.Minute
	}

	// Check immediately
	if within, expired := isWithinBounds(cfg); expired {
		log.Println("End date reached, stopping scheduler.")
		Stop()
		return
	} else if within {
		err := screenshot.Capture(cfg)
		if err != nil {
			log.Println("Screenshot error:", err)
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			within, expired := isWithinBounds(cfg)
			if expired {
				log.Println("End date reached, stopping scheduler.")
				Stop()
				return
			}
			if within {
				err := screenshot.Capture(cfg)
				if err != nil {
					log.Println("Screenshot error:", err)
				}
			}
		case <-stop:
			return
		}
	}
}
