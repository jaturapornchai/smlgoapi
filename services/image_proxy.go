package services

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"golang.org/x/image/draw"
)

type ImageProxy struct {
	cache          *cache.Cache
	cacheDir       string
	maxSize        int64 // Maximum file size in bytes
	allowedDomains []string
	mutex          sync.RWMutex
	inFlight       map[string]bool
}

func NewImageProxy() *ImageProxy {
	// Create cache directory
	cacheDir := "./image_cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Printf("❌ [imgproxy] Failed to create cache directory: %v", err)
	} else {
		log.Printf("📁 [imgproxy] Cache directory ready: %s", cacheDir)
	}
	return &ImageProxy{
		cache:          cache.New(24*time.Hour, 1*time.Hour),
		cacheDir:       cacheDir,
		maxSize:        10 * 1024 * 1024, // 10MB
		allowedDomains: []string{},       // Empty array = allow all domains
		inFlight:       make(map[string]bool),
	}
}

func (ip *ImageProxy) isAllowedDomain(imageURL string) bool {
	if len(ip.allowedDomains) == 0 {
		return true // Allow all if no restrictions
	}

	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return false
	}

	for _, domain := range ip.allowedDomains {
		if strings.Contains(parsedURL.Host, domain) {
			return true
		}
	}
	return false
}

func (ip *ImageProxy) getCacheKey(imageURL string, width, height int) string {
	// Include resize parameters in cache key
	cacheString := fmt.Sprintf("%s_w%d_h%d", imageURL, width, height)
	hash := md5.Sum([]byte(cacheString))
	return hex.EncodeToString(hash[:])
}

func (ip *ImageProxy) getCacheFilePath(cacheKey string) string {
	return filepath.Join(ip.cacheDir, cacheKey)
}

// HeadHandler handles HEAD requests for image proxy
func (ip *ImageProxy) HeadHandler(c *gin.Context) {
	imageURL := c.Query("url")
	log.Printf("🔍 [imgproxy-head] Request URL: %s", imageURL)

	if imageURL == "" {
		log.Printf("❌ [imgproxy-head] Empty URL - returning 400")
		c.Status(http.StatusBadRequest)
		return
	}

	// Check domain
	if !ip.isAllowedDomain(imageURL) {
		log.Printf("❌ [imgproxy-head] Domain not allowed for URL: %s - returning 403", imageURL)
		c.Status(http.StatusForbidden)
		return
	}

	log.Printf("✅ [imgproxy-head] Domain allowed for URL: %s", imageURL)

	// Get resize parameters
	widthStr := c.Query("w")
	heightStr := c.Query("h")

	var width, height int
	var err error

	if widthStr != "" {
		width, err = strconv.Atoi(widthStr)
		if err != nil || width <= 0 || width > 2000 {
			c.Status(http.StatusBadRequest)
			return
		}
	}

	if heightStr != "" {
		height, err = strconv.Atoi(heightStr)
		if err != nil || height <= 0 || height > 2000 {
			c.Status(http.StatusBadRequest)
			return
		}
	}
	// Generate cache key
	cacheKey := ip.getCacheKey(imageURL, width, height)
	cacheFilePath := ip.getCacheFilePath(cacheKey)
	log.Printf("🔍 [imgproxy-head] Cache key: %s, Cache path: %s", cacheKey, cacheFilePath)

	// Check if cached file exists
	if info, err := os.Stat(cacheFilePath); err == nil {
		// File exists, set appropriate headers
		log.Printf("✅ [imgproxy-head] Found in cache: %s (size: %d bytes)", cacheFilePath, info.Size())
		contentType := ip.getContentType(cacheFilePath)
		c.Header("Content-Type", contentType)
		c.Header("Content-Length", fmt.Sprintf("%d", info.Size()))
		c.Header("Cache-Control", "public, max-age=86400")
		c.Status(http.StatusOK)
		return
	}
	log.Printf("🔍 [imgproxy-head] Not in cache, checking remote URL: %s", imageURL)

	// Check if we can fetch the original image (HEAD request)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("HEAD", imageURL, nil)
	if err != nil {
		log.Printf("❌ [imgproxy-head] Failed to create HEAD request: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}

	req.Header.Set("User-Agent", "SMLGOAPI-ImageProxy/1.0")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ [imgproxy-head] Failed to fetch remote URL: %v", err)
		c.Status(http.StatusNotFound)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [imgproxy-head] Remote server returned status: %d", resp.StatusCode)
		c.Status(resp.StatusCode)
		return
	}

	log.Printf("✅ [imgproxy-head] Remote image exists, status: %d", resp.StatusCode)

	// Image exists, set headers and return 200
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	c.Header("Content-Type", contentType)
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		c.Header("Content-Length", contentLength)
	}
	c.Header("Cache-Control", "public, max-age=86400")
	c.Status(http.StatusOK)
}

func (ip *ImageProxy) ProxyHandler(c *gin.Context) {
	imageURL := c.Query("url")
	if imageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "URL parameter is required",
		})
		return
	}

	// Get resize parameters
	widthStr := c.Query("w")
	heightStr := c.Query("h")

	var width, height int
	var err error

	if widthStr != "" {
		width, err = strconv.Atoi(widthStr)
		if err != nil || width <= 0 || width > 2000 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid width parameter (1-2000)",
			})
			return
		}
	}

	if heightStr != "" {
		height, err = strconv.Atoi(heightStr)
		if err != nil || height <= 0 || height > 2000 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid height parameter (1-2000)",
			})
			return
		}
	}

	// Generate cache key
	cacheKey := ip.getCacheKey(imageURL, width, height)

	// Check if already processing this image
	ip.mutex.Lock()
	if ip.inFlight[cacheKey] {
		ip.mutex.Unlock()
		// Wait a bit and retry
		time.Sleep(100 * time.Millisecond)
		ip.mutex.RLock()
		processing := ip.inFlight[cacheKey]
		ip.mutex.RUnlock()

		if processing {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Image is being processed, please try again",
			})
			return
		}
	} else {
		ip.inFlight[cacheKey] = true
		ip.mutex.Unlock()

		// Ensure cleanup
		defer func() {
			ip.mutex.Lock()
			delete(ip.inFlight, cacheKey)
			ip.mutex.Unlock()
		}()
	}

	// Check domain permissions
	// Validate URL
	parsedURL, err := url.Parse(imageURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid URL - must be http or https",
		})
		return
	}

	// Check allowed domains
	if !ip.isAllowedDomain(imageURL) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Domain not allowed"})
		return
	}

	cacheFilePath := ip.getCacheFilePath(cacheKey)

	// Check if cached file exists and is recent
	if fileInfo, err := os.Stat(cacheFilePath); err == nil {
		if time.Since(fileInfo.ModTime()) < 24*time.Hour {
			log.Printf("📸 [imgproxy] Serving cached image: %s (size: %dx%d)", imageURL, width, height)
			ip.serveCachedFile(c, cacheFilePath)
			return
		}
	}

	// Fetch image from URL
	resizeInfo := ""
	if width > 0 || height > 0 {
		resizeInfo = fmt.Sprintf(" (resize: %dx%d)", width, height)
	}
	log.Printf("🔄 [imgproxy] Fetching new image: %s%s", imageURL, resizeInfo)
	ip.fetchAndCacheImage(c, imageURL, cacheFilePath, width, height)
}

func (ip *ImageProxy) serveCachedFile(c *gin.Context, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to open cached file",
		})
		return
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get file info",
		})
		return
	}

	// Set headers
	c.Header("Content-Type", ip.getContentType(filePath))
	c.Header("Cache-Control", "public, max-age=31536000")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Copy file to response
	io.Copy(c.Writer, file)
}

func (ip *ImageProxy) fetchAndCacheImage(c *gin.Context, imageURL, cacheFilePath string, width, height int) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create request",
		})
		return
	}

	req.Header.Set("User-Agent", "SMLGOAPI-ImageProxy/1.0")
	req.Header.Set("Accept", "image/*")

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch image: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Failed to fetch image: HTTP %d", resp.StatusCode),
		})
		return
	}

	// Check content length
	if resp.ContentLength > ip.maxSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"success": false,
			"error":   "Image too large (max 10MB)",
		})
		return
	}

	// Ensure cache directory exists
	cacheDir := filepath.Dir(cacheFilePath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create cache directory: " + err.Error(),
		})
		return
	}

	// Read image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to read image data",
		})
		return
	}

	// Process image (resize if needed)
	finalImageData := imageData
	if width > 0 || height > 0 {
		resizedData, err := ip.resizeImage(imageData, width, height)
		if err != nil {
			log.Printf("⚠️ [imgproxy] Resize failed, serving original: %v", err)
		} else {
			finalImageData = resizedData
			log.Printf("🔧 [imgproxy] Image resized to %dx%d", width, height)
		}
	}

	// Save to cache
	err = os.WriteFile(cacheFilePath, finalImageData, 0644)
	if err != nil {
		log.Printf("❌ [imgproxy] Failed to save to cache: %v", err)
	}

	// Set response headers
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = ip.getContentType(imageURL)
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=31536000")
	c.Header("Content-Length", fmt.Sprintf("%d", len(finalImageData)))

	// Send response
	c.Writer.Write(finalImageData)

	resizeInfo := ""
	if width > 0 || height > 0 {
		resizeInfo = fmt.Sprintf(" (resized to %dx%d)", width, height)
	}
	log.Printf("✅ [imgproxy] Image cached successfully: %s%s", imageURL, resizeInfo)
}

func (ip *ImageProxy) getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	default:
		return "image/jpeg"
	}
}

// GetStats returns cache statistics
func (ip *ImageProxy) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"cache_items":     ip.cache.ItemCount(),
		"cache_dir":       ip.cacheDir,
		"max_size_mb":     ip.maxSize / (1024 * 1024),
		"allowed_domains": ip.allowedDomains,
	}
}

func (ip *ImageProxy) resizeImage(imageData []byte, targetWidth, targetHeight int) ([]byte, error) {
	// Decode image
	img, format, err := image.Decode(strings.NewReader(string(imageData)))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	// Get original dimensions
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// Calculate new dimensions maintaining aspect ratio
	newWidth, newHeight := ip.calculateDimensions(originalWidth, originalHeight, targetWidth, targetHeight)

	// Create new image
	resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Resize using high-quality algorithm
	draw.CatmullRom.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil) // Encode to bytes
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(&buf, resized)
	case "gif":
		err = gif.Encode(&buf, resized, nil)
	case "webp":
		// WebP encoding (note: webp package doesn't export Encode in some versions)
		// Fallback to JPEG for WebP
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 90})
	default:
		// Default to JPEG for unknown formats
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 90})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode resized image: %v", err)
	}

	return buf.Bytes(), nil
}

func (ip *ImageProxy) calculateDimensions(originalWidth, originalHeight, targetWidth, targetHeight int) (int, int) {
	// If both dimensions are specified, use them directly
	if targetWidth > 0 && targetHeight > 0 {
		return targetWidth, targetHeight
	}

	// If only width is specified, maintain aspect ratio
	if targetWidth > 0 && targetHeight == 0 {
		ratio := float64(targetWidth) / float64(originalWidth)
		return targetWidth, int(float64(originalHeight) * ratio)
	}

	// If only height is specified, maintain aspect ratio
	if targetHeight > 0 && targetWidth == 0 {
		ratio := float64(targetHeight) / float64(originalHeight)
		return int(float64(originalWidth) * ratio), targetHeight
	}

	// If no dimensions specified, return original
	return originalWidth, originalHeight
}
