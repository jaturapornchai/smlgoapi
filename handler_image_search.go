package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func handleImageSearch(c *gin.Context) {
	start := time.Now()
	reqID := getNextRequestID()

	// Check if request context is already cancelled
	if c.Request.Context().Err() != nil {
		c.JSON(408, map[string]interface{}{
			"error":           "Request timeout",
			"message":         "Request was cancelled before processing",
			"timeout_seconds": config.RequestTimeout.Seconds(),
		})
		return
	}

	// Create timeout context for image search operation
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.RequestTimeout)
	defer cancel()

	var request ImageSearchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		duration := time.Since(start).Seconds() * 1000
		errorResponse := map[string]string{"error": "Invalid request format"}
		printResponseDetails(reqID, 400, errorResponse, duration)
		c.JSON(400, errorResponse)
		return
	}

	if request.SimilarityThreshold == 0 {
		request.SimilarityThreshold = config.SimilarityThreshold
	}
	if request.Limit == 0 {
		request.Limit = 10
	}

	requestBody := map[string]interface{}{
		"similarity_threshold": request.SimilarityThreshold,
		"limit":                request.Limit,
		"use_multi_view":       request.UseMultiView,
		"image_data_length":    len(request.ImageData),
		"image_format":         "base64",
	}

	printRequestDetails("POST", "/imgsearch", reqID, nil, requestBody)

	fmt.Printf("\n🔍 [handleImageSearch] STARTING IMAGE SEARCH OPERATION:\n")
	fmt.Printf("   [handleImageSearch] Multi-View Enabled: %t\n", request.UseMultiView)
	fmt.Printf("   [handleImageSearch] Similarity Threshold: %.3f\n", request.SimilarityThreshold)
	fmt.Printf("   [handleImageSearch] Limit: %d\n", request.Limit)
	fmt.Printf("   [handleImageSearch] Timeout: %v\n", config.RequestTimeout)

	// Check timeout before processing
	if ctx.Err() != nil {
		c.JSON(408, map[string]interface{}{
			"error":           "Request timeout during validation",
			"timeout_seconds": config.RequestTimeout.Seconds(),
		})
		return
	}

	// Decode base64 image
	var imageData string
	if strings.Contains(request.ImageData, ",") {
		parts := strings.Split(request.ImageData, ",")
		if len(parts) > 1 {
			imageData = parts[1]
		}
	} else {
		imageData = request.ImageData
	}

	imageBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		duration := time.Since(start).Seconds() * 1000
		errorResponse := map[string]string{"error": fmt.Sprintf("Invalid image data: %v", err)}
		printResponseDetails(reqID, 400, errorResponse, duration)
		c.JSON(400, errorResponse)
		return
	}

	fmt.Printf("   [handleImageSearch] Decoded image size: %d bytes\n", len(imageBytes))

	// Process image search with timeout
	resultChan := make(chan ImageSearchResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		// Generate mock search results
		var results []ImageSearchResult
		var queryVectorSize int

		if request.UseMultiView {
			fmt.Printf("   [handleImageSearch] 🎭 MULTI-VIEW SEARCH MODE\n")
			queryVectorSize = 5 // 5 different views

			// Check timeout during processing
			if ctx.Err() != nil {
				errorChan <- ctx.Err()
				return
			}

			// Mock multi-view search results
			results = []ImageSearchResult{
				{
					Barcode:         "123456789",
					ImageNumber:     1,
					SimilarityScore: 0.95,
					Name:            "MultiView_123456789_1",
					Description:     "Multi-view processed image for barcode 123456789",
				},
				{
					Barcode:         "987654321",
					ImageNumber:     2,
					SimilarityScore: 0.87,
					Name:            "MultiView_987654321_2",
					Description:     "Multi-view processed image for barcode 987654321",
				},
			}
		} else {
			fmt.Printf("   [handleImageSearch] 📸 SINGLE-VIEW SEARCH MODE (fallback)\n")

			// Check timeout during processing
			if ctx.Err() != nil {
				errorChan <- ctx.Err()
				return
			}

			vector, err := generateColorHistogram(imageBytes)
			if err != nil {
				errorChan <- err
				return
			}
			queryVectorSize = len(vector)

			// Mock single-view search results
			results = []ImageSearchResult{
				{
					Barcode:         "111222333",
					ImageNumber:     1,
					SimilarityScore: 0.89,
					Name:            "Image_111222333_1",
					Description:     "Single-view processed image",
				},
			}
		}

		// Check timeout before filtering
		if ctx.Err() != nil {
			errorChan <- ctx.Err()
			return
		}

		// Filter by similarity threshold
		var filteredResults []ImageSearchResult
		for _, result := range results {
			if result.SimilarityScore >= request.SimilarityThreshold {
				filteredResults = append(filteredResults, result)
			}
		}

		// Limit results
		if len(filteredResults) > request.Limit {
			filteredResults = filteredResults[:request.Limit]
		}

		fmt.Printf("   [handleImageSearch] Found %d similar images\n", len(filteredResults))
		if len(filteredResults) > 0 {
			fmt.Printf("   [handleImageSearch] Best match: %s (similarity: %.4f)\n",
				filteredResults[0].Barcode, filteredResults[0].SimilarityScore)
		}

		response := ImageSearchResponse{
			TotalFound:       len(filteredResults),
			Results:          filteredResults,
			QueryVectorSize:  queryVectorSize,
			ProcessingTimeMS: time.Since(start).Seconds() * 1000,
		}

		resultChan <- response
	}()

	// Wait for result or timeout
	select {
	case response := <-resultChan:
		duration := time.Since(start).Seconds() * 1000
		printResponseDetails(reqID, 200, response, duration)
		c.JSON(200, response)
	case err := <-errorChan:
		duration := time.Since(start).Seconds() * 1000
		errorResponse := map[string]interface{}{
			"error":       fmt.Sprintf("Image search error: %v", err),
			"duration_ms": duration,
		}
		printResponseDetails(reqID, 500, errorResponse, duration)
		c.JSON(500, errorResponse)
	case <-ctx.Done():
		duration := time.Since(start).Seconds() * 1000
		errorResponse := map[string]interface{}{
			"error":           "Image search operation timeout",
			"timeout_seconds": config.RequestTimeout.Seconds(),
			"duration_ms":     duration,
		}
		printResponseDetails(reqID, 408, errorResponse, duration)
		c.JSON(408, errorResponse)
	}
}
