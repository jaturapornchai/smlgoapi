package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"smlgoapi/handlers"
)

// SimpleScheduler runs in a loop checking for due email tasks
type SimpleScheduler struct {
	checkInterval time.Duration
	apiURL        string
	stopChan      chan struct{}
}

// NewSimpleScheduler creates a new scheduler instance
func NewSimpleScheduler(checkIntervalMinutes int, apiURL string) *SimpleScheduler {
	if checkIntervalMinutes <= 0 {
		checkIntervalMinutes = 1 // Default to 1 minute
	}

	return &SimpleScheduler{
		checkInterval: time.Duration(checkIntervalMinutes) * time.Minute,
		apiURL:        apiURL,
		stopChan:      make(chan struct{}),
	}
}

// Start begins the scheduler loop
func (s *SimpleScheduler) Start() {
	log.Printf("📧 Starting Simple Email Scheduler (checking every %v)", s.checkInterval)

	// Run immediately on start
	go s.checkAndSendDueEmails()

	// Then run on interval
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go s.checkAndSendDueEmails()
		case <-s.stopChan:
			log.Println("🛑 Stopping Simple Email Scheduler")
			return
		}
	}
}

// Stop gracefully stops the scheduler
func (s *SimpleScheduler) Stop() {
	close(s.stopChan)
}

// checkAndSendDueEmails checks for schedules that are due and sends emails
func (s *SimpleScheduler) checkAndSendDueEmails() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ Panic in scheduler: %v", r)
		}
	}()

	// Get due schedules from MongoDB
	schedules, err := handlers.GetDueSchedules("email_schedules")
	if err != nil {
		log.Printf("⚠️  Failed to get due schedules: %v", err)
		return
	}

	if len(schedules) == 0 {
		log.Println("✓ No due schedules at", time.Now().Format("2006-01-02 15:04:05"))
		return
	}

	log.Printf("📬 Found %d due schedule(s)", len(schedules))

	for _, schedule := range schedules {
		scheduleID, _ := schedule["schedule_id"].(string)
		scheduleName, _ := schedule["schedule_name"].(string)
		shopID, _ := schedule["shop_id"].(string)

		if scheduleID == "" || shopID == "" {
			log.Printf("⚠️  Skipping invalid schedule (missing ID or shopID)")
			continue
		}

		log.Printf("⏰ Processing: %s (%s)", scheduleName, scheduleID)

		// Send the email using the API endpoint
		err := s.sendReportEmail(shopID, scheduleID)

		if err != nil {
			log.Printf("❌ Failed to send email for %s: %v", scheduleID, err)

			// Log failure to MongoDB
			_ = handlers.LogSchedulerExecution(scheduleID, scheduleName, "failed", map[string]interface{}{
				"error": err.Error(),
			})

			// Record for retry
			_ = handlers.RecordSchedulerFailure(scheduleID, scheduleName, err.Error(), 0)
		} else {
			log.Printf("✅ Successfully sent email for: %s", scheduleID)

			// Log success
			_ = handlers.LogSchedulerExecution(scheduleID, scheduleName, "success", map[string]interface{}{
				"sent_at": time.Now(),
			})

			// Calculate and update next run time
			nextRunTime := s.calculateNextRunTime(schedule)
			err = handlers.UpdateScheduleNextRun(scheduleID, nextRunTime, time.Now())
			if err != nil {
				log.Printf("⚠️  Failed to update next run time for %s: %v", scheduleID, err)
			} else {
				log.Printf("📅 Next run for %s: %s", scheduleID, nextRunTime.Format("2006-01-02 15:04:05"))
			}
		}
	}
}

// sendReportEmail calls the existing API endpoint to send email
func (s *SimpleScheduler) sendReportEmail(shopID, scheduleID string) error {
	payload := map[string]string{
		"shopid":      shopID,
		"schedule_id": scheduleID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/v1/sendreportemail", s.apiURL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Minute} // 5 min timeout for PDF generation
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// calculateNextRunTime calculates when the schedule should run next
func (s *SimpleScheduler) calculateNextRunTime(schedule map[string]interface{}) time.Time {
	now := time.Now()

	// Get schedule pattern
	pattern, _ := schedule["schedule_pattern"].(string)

	// Get interval in minutes (if specified)
	intervalMinutes, _ := schedule["interval_minutes"].(float64)
	if intervalMinutes > 0 {
		return now.Add(time.Duration(intervalMinutes) * time.Minute)
	}

	// Parse based on pattern
	switch pattern {
	case "hourly":
		return now.Add(1 * time.Hour)
	case "daily":
		// Run at the same time tomorrow
		return now.Add(24 * time.Hour)
	case "weekly":
		return now.Add(7 * 24 * time.Hour)
	case "monthly":
		// Add one month
		return now.AddDate(0, 1, 0)
	default:
		// Default to daily if pattern not recognized
		log.Printf("⚠️  Unknown schedule pattern '%s', defaulting to daily", pattern)
		return now.Add(24 * time.Hour)
	}
}
