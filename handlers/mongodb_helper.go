package handlers

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FetchScheduleConfig retrieves the schedule configuration from MongoDB Atlas
func FetchScheduleConfig(collectionName string, filter map[string]interface{}) (map[string]interface{}, error) {
	if atlasDB == nil {
		return nil, fmt.Errorf("MongoDB Atlas is not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := atlasDB.Collection(collectionName)
	var result map[string]interface{}

	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to find schedule: %w", err)
	}

	return result, nil
}

// FetchAllActiveSchedules retrieves all active email schedules from MongoDB Atlas
func FetchAllActiveSchedules(collectionName string) ([]map[string]interface{}, error) {
	if atlasDB == nil {
		return nil, fmt.Errorf("MongoDB Atlas is not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := atlasDB.Collection(collectionName)

	// Filter for enabled schedules only
	filter := bson.M{"enabled": true}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schedules: %w", err)
	}
	defer cursor.Close(ctx)

	var schedules []map[string]interface{}
	if err := cursor.All(ctx, &schedules); err != nil {
		return nil, fmt.Errorf("failed to decode schedules: %w", err)
	}

	return schedules, nil
}

// LogSchedulerExecution logs a scheduler execution attempt to MongoDB
func LogSchedulerExecution(scheduleID, scheduleName, status string, details map[string]interface{}) error {
	if atlasDB == nil {
		return fmt.Errorf("MongoDB Atlas is not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := atlasDB.Collection("scheduler_executions")

	executionLog := bson.M{
		"schedule_id":   scheduleID,
		"schedule_name": scheduleName,
		"status":        status,
		"timestamp":     time.Now(),
		"details":       details,
	}

	if status == "started" {
		executionLog["started_at"] = time.Now()
	} else {
		executionLog["completed_at"] = time.Now()
	}

	_, err := collection.InsertOne(ctx, executionLog)
	if err != nil {
		return fmt.Errorf("failed to log execution: %w", err)
	}

	return nil
}

// RecordSchedulerFailure records a failed execution for retry tracking
func RecordSchedulerFailure(scheduleID, scheduleName, errorMsg string, retryCount int) error {
	if atlasDB == nil {
		return fmt.Errorf("MongoDB Atlas is not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := atlasDB.Collection("scheduler_failures")

	// Calculate next retry time with exponential backoff
	var nextRetryDelay time.Duration
	switch retryCount {
	case 0:
		nextRetryDelay = 0 // Immediate retry
	case 1:
		nextRetryDelay = 5 * time.Minute
	case 2:
		nextRetryDelay = 15 * time.Minute
	case 3:
		nextRetryDelay = 1 * time.Hour
	default:
		nextRetryDelay = 24 * time.Hour // Dead letter queue
	}

	failureRecord := bson.M{
		"schedule_id":   scheduleID,
		"schedule_name": scheduleName,
		"failed_at":     time.Now(),
		"error_message": errorMsg,
		"retry_count":   retryCount,
		"next_retry_at": time.Now().Add(nextRetryDelay),
		"status":        "pending_retry",
	}

	if retryCount >= 4 {
		failureRecord["status"] = "dead_letter"
	}

	_, err := collection.InsertOne(ctx, failureRecord)
	if err != nil {
		return fmt.Errorf("failed to record failure: %w", err)
	}

	return nil
}

// GetPendingRetries retrieves failed executions that should be retried
func GetPendingRetries() ([]map[string]interface{}, error) {
	if atlasDB == nil {
		return nil, fmt.Errorf("MongoDB Atlas is not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := atlasDB.Collection("scheduler_failures")

	// Find failures that are due for retry
	filter := bson.M{
		"status":        "pending_retry",
		"next_retry_at": bson.M{"$lte": time.Now()},
		"retry_count":   bson.M{"$lt": 4}, // Max 4 retries
	}

	opts := options.Find().SetSort(bson.M{"failed_at": 1}).SetLimit(50)

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending retries: %w", err)
	}
	defer cursor.Close(ctx)

	var retries []map[string]interface{}
	if err := cursor.All(ctx, &retries); err != nil {
		return nil, fmt.Errorf("failed to decode retries: %w", err)
	}

	return retries, nil
}

// UpdateRetryStatus updates the status of a retry attempt
func UpdateRetryStatus(failureID interface{}, newStatus string) error {
	if atlasDB == nil {
		return fmt.Errorf("MongoDB Atlas is not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := atlasDB.Collection("scheduler_failures")

	filter := bson.M{"_id": failureID}
	update := bson.M{
		"$set": bson.M{
			"status":        newStatus,
			"last_retry_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update retry status: %w", err)
	}

	return nil
}

// GetDueSchedules retrieves schedules that are due to run now
func GetDueSchedules(collectionName string) ([]map[string]interface{}, error) {
	if atlasDB == nil {
		return nil, fmt.Errorf("MongoDB Atlas is not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := atlasDB.Collection(collectionName)

	// Find schedules that are:
	// 1. Enabled
	// 2. next_run_time is in the past or now
	filter := bson.M{
		"enabled":       true,
		"next_run_time": bson.M{"$lte": time.Now()},
	}

	opts := options.Find().SetSort(bson.M{"next_run_time": 1}).SetLimit(100)

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch due schedules: %w", err)
	}
	defer cursor.Close(ctx)

	var schedules []map[string]interface{}
	if err := cursor.All(ctx, &schedules); err != nil {
		return nil, fmt.Errorf("failed to decode schedules: %w", err)
	}

	return schedules, nil
}

// UpdateScheduleNextRun updates the next run time and last run time for a schedule
func UpdateScheduleNextRun(scheduleID string, nextRunTime time.Time, lastRunTime time.Time) error {
	if atlasDB == nil {
		return fmt.Errorf("MongoDB Atlas is not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := atlasDB.Collection("email_schedules")

	filter := bson.M{"schedule_id": scheduleID}
	update := bson.M{
		"$set": bson.M{
			"next_run_time": nextRunTime,
			"last_run_time": lastRunTime,
			"updated_at":    time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	return nil
}
