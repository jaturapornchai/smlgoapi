package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func handleImageUpload(c *gin.Context) {
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

	// Create timeout context for image upload operation
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.RequestTimeout)
	defer cancel()

	// Start debug trace
	startDebugTrace(reqID, "POST", "/imgupload")

	// Step 1: Parse JSON request
	addDebugStep(reqID, "Parse JSON Request", nil)

	var request ImageUploadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		completeDebugStep(reqID, "Parse JSON Request", "ERROR", nil, err.Error(), nil)
		completeDebugTrace(reqID, "ERROR")

		duration := time.Since(start).Seconds() * 1000
		errorResponse := ImageUploadResponse{
			Status:           "error",
			Message:          fmt.Sprintf("Invalid request format: %v", err),
			ProcessingTimeMS: duration,
		}
		printResponseDetails(reqID, 400, errorResponse, duration)
		c.JSON(400, errorResponse)
		return
	}

	// Set default image number if not provided
	if request.ImageNumber == 0 {
		request.ImageNumber = 1
	}

	completeDebugStep(reqID, "Parse JSON Request", "SUCCESS", map[string]interface{}{
		"barcode":        request.Barcode,
		"image_number":   request.ImageNumber,
		"use_multi_view": request.UseMultiView,
		"image_data_len": len(request.ImageData),
	}, "", nil)

	// Print request details
	requestBody := map[string]interface{}{
		"barcode":           request.Barcode,
		"image_number":      request.ImageNumber,
		"use_multi_view":    request.UseMultiView,
		"image_data_length": len(request.ImageData),
		"image_format":      "base64",
	}
	printRequestDetails("POST", "/imgupload", reqID, nil, requestBody)

	fmt.Printf("\n📸 [handleImageUpload] STARTING IMAGE UPLOAD OPERATION:\n")
	fmt.Printf("   [handleImageUpload] Barcode: %s\n", request.Barcode)
	fmt.Printf("   [handleImageUpload] Image Number: %d\n", request.ImageNumber)
	fmt.Printf("   [handleImageUpload] Multi-View Enabled: %t\n", request.UseMultiView)
	fmt.Printf("   [handleImageUpload] Timeout: %v\n", config.RequestTimeout)

	// Step 2: Validate and decode image data
	addDebugStep(reqID, "Validate Image Data", map[string]interface{}{
		"raw_image_data_length": len(request.ImageData),
	})

	// Check timeout before processing
	if ctx.Err() != nil {
		completeDebugStep(reqID, "Validate Image Data", "ERROR", nil, "Request timeout during validation", nil)
		completeDebugTrace(reqID, "ERROR")

		duration := time.Since(start).Seconds() * 1000
		errorResponse := ImageUploadResponse{
			Status:           "error",
			Message:          "Request timeout during validation",
			ProcessingTimeMS: duration,
		}
		c.JSON(408, errorResponse)
		return
	}

	// Remove data URL prefix if present
	var imageData string
	if strings.Contains(request.ImageData, ",") {
		parts := strings.Split(request.ImageData, ",")
		if len(parts) > 1 {
			imageData = parts[1]
		}
	} else {
		imageData = request.ImageData
	}

	// Decode base64 image
	imageBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		completeDebugStep(reqID, "Validate Image Data", "ERROR", nil, fmt.Sprintf("Invalid base64 image data: %v", err), nil)
		completeDebugTrace(reqID, "ERROR")

		duration := time.Since(start).Seconds() * 1000
		errorResponse := ImageUploadResponse{
			Status:           "error",
			Message:          fmt.Sprintf("Invalid image data: %v", err),
			Barcode:          request.Barcode,
			ImageNumber:      request.ImageNumber,
			ProcessingTimeMS: duration,
		}
		printResponseDetails(reqID, 400, errorResponse, duration)
		c.JSON(400, errorResponse)
		return
	}

	completeDebugStep(reqID, "Validate Image Data", "SUCCESS", map[string]interface{}{
		"decoded_image_size": len(imageBytes),
		"image_type":         "binary",
	}, "", nil)

	fmt.Printf("   [handleImageUpload] Decoded image size: %d bytes\n", len(imageBytes))

	// Step 3: Process image upload with timeout
	addDebugStep(reqID, "Process Image Upload", map[string]interface{}{
		"processing_mode": map[string]interface{}{
			"multi_view": request.UseMultiView,
			"barcode":    request.Barcode,
			"img_number": request.ImageNumber,
		},
	})

	// Process upload in a goroutine with timeout
	resultChan := make(chan ImageUploadResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		var totalViewsGenerated int
		var totalVectorsStored int
		var vectorSize int

		if request.UseMultiView {
			fmt.Printf("   [handleImageUpload] 🎭 MULTI-VIEW PROCESSING MODE\n")

			// Check timeout during processing
			if ctx.Err() != nil {
				errorChan <- ctx.Err()
				return
			}

			// Generate multiple views (mock implementation)
			views := []string{"front", "side", "top", "rotated_15", "rotated_30"}
			totalViewsGenerated = len(views)

			for i, view := range views {
				// Check timeout for each view
				if ctx.Err() != nil {
					errorChan <- ctx.Err()
					return
				}

				// Generate vector for each view (mock)
				vector, err := generateColorHistogram(imageBytes)
				if err != nil {
					errorChan <- fmt.Errorf("failed to generate vector for view %s: %v", view, err)
					return
				}

				vectorSize = len(vector)
				totalVectorsStored++

				fmt.Printf("     [handleImageUpload] Generated vector for %s view (%d/%d)\n", view, i+1, len(views))
			}

			fmt.Printf("   [handleImageUpload] Multi-view processing completed: %d views, %d vectors\n", totalViewsGenerated, totalVectorsStored)
		} else {
			fmt.Printf("   [handleImageUpload] 📸 SINGLE-VIEW PROCESSING MODE (fallback)\n")

			// Check timeout during processing
			if ctx.Err() != nil {
				errorChan <- ctx.Err()
				return
			}

			// Generate single vector
			vector, err := generateColorHistogram(imageBytes)
			if err != nil {
				errorChan <- fmt.Errorf("failed to generate vector: %v", err)
				return
			}

			vectorSize = len(vector)
			totalViewsGenerated = 1
			totalVectorsStored = 1

			fmt.Printf("   [handleImageUpload] Single-view processing completed: 1 vector generated\n")
		}

		// Mock database storage (in a real implementation, you would store in ClickHouse)
		fmt.Printf("   [handleImageUpload] 💾 Storing vectors in database...\n")

		// Check timeout before finalizing
		if ctx.Err() != nil {
			errorChan <- ctx.Err()
			return
		}

		response := ImageUploadResponse{
			Status:              "success",
			Message:             fmt.Sprintf("Image uploaded and processed successfully for barcode %s", request.Barcode),
			Barcode:             request.Barcode,
			ImageNumber:         request.ImageNumber,
			TotalViewsGenerated: totalViewsGenerated,
			TotalVectorsStored:  totalVectorsStored,
			VectorSize:          vectorSize,
			ProcessingTimeMS:    time.Since(start).Seconds() * 1000,
		}

		resultChan <- response
	}()

	// Wait for result or timeout
	select {
	case response := <-resultChan:
		completeDebugStep(reqID, "Process Image Upload", "SUCCESS", map[string]interface{}{
			"views_generated": response.TotalViewsGenerated,
			"vectors_stored":  response.TotalVectorsStored,
			"vector_size":     response.VectorSize,
		}, "", nil)
		completeDebugTrace(reqID, "SUCCESS")

		duration := time.Since(start).Seconds() * 1000
		response.ProcessingTimeMS = duration
		printResponseDetails(reqID, 200, response, duration)
		c.JSON(200, response)

	case err := <-errorChan:
		completeDebugStep(reqID, "Process Image Upload", "ERROR", nil, err.Error(), nil)
		completeDebugTrace(reqID, "ERROR")

		duration := time.Since(start).Seconds() * 1000
		errorResponse := ImageUploadResponse{
			Status:           "error",
			Message:          fmt.Sprintf("Image processing error: %v", err),
			Barcode:          request.Barcode,
			ImageNumber:      request.ImageNumber,
			ProcessingTimeMS: duration,
		}
		printResponseDetails(reqID, 500, errorResponse, duration)
		c.JSON(500, errorResponse)

	case <-ctx.Done():
		completeDebugStep(reqID, "Process Image Upload", "ERROR", nil, "Image upload operation timeout", nil)
		completeDebugTrace(reqID, "ERROR")

		duration := time.Since(start).Seconds() * 1000
		errorResponse := ImageUploadResponse{
			Status:           "error",
			Message:          "Image upload operation timeout",
			Barcode:          request.Barcode,
			ImageNumber:      request.ImageNumber,
			ProcessingTimeMS: duration,
		}
		printResponseDetails(reqID, 408, errorResponse, duration)
		c.JSON(408, errorResponse)
	}
}
