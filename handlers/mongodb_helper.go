package handlers

import (
	"context"
	"fmt"
	"time"
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
