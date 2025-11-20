package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	atlasClient *mongo.Client
	atlasDB     *mongo.Database
)

// InitMongoAtlas เชื่อมต่อ MongoDB Atlas
func InitMongoAtlas() error {
	uri := os.Getenv("MONGODB_ATLAS_URI")
	if uri == "" {
		log.Println("⚠️  MONGODB_ATLAS_URI not set, skipping Atlas connection")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB Atlas: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB Atlas: %w", err)
	}

	atlasClient = client

	// Get database name from env or use default
	dbName := os.Getenv("MONGODB_ATLAS_DBNAME")
	if dbName == "" {
		dbName = "atlas_db"
	}
	atlasDB = client.Database(dbName)

	log.Printf("✅ Connected to MongoDB Atlas successfully (DB: %s)", dbName)
	return nil
}

// DisconnectMongoAtlas ปิดการเชื่อมต่อ MongoDB Atlas
func DisconnectMongoAtlas() {
	if atlasClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := atlasClient.Disconnect(ctx); err != nil {
			log.Printf("❌ Error disconnecting from MongoDB Atlas: %v", err)
		} else {
			log.Println("✅ Disconnected from MongoDB Atlas")
		}
	}
}

// MongoAtlasUpdateHandler - Update/Insert (Upsert) document
// Frontend กำหนดโครงสร้างของ filter และ data เอง
func MongoAtlasUpdateHandler(c *gin.Context) {
	if atlasDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"code":    503,
			"message": "MongoDB Atlas is not connected",
		})
		return
	}

	var reqBody struct {
		Collection string                 `json:"collection"`
		Filter     map[string]interface{} `json:"filter"`     // Frontend กำหนด filter เอง
		Data       map[string]interface{} `json:"data"`       // Frontend กำหนด data เอง
		Upsert     bool                   `json:"upsert"`     // default: false
		ReplaceOne bool                   `json:"replaceone"` // true = replaceOne, false = updateOne with $set
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    400,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Validate required fields
	if reqBody.Collection == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    400,
			"message": "Collection name is required",
		})
		return
	}

	if reqBody.Filter == nil {
		reqBody.Filter = make(map[string]interface{})
	}

	if reqBody.Data == nil {
		reqBody.Data = make(map[string]interface{})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := atlasDB.Collection(reqBody.Collection)

	var result *mongo.UpdateResult
	var err error

	if reqBody.ReplaceOne {
		// ReplaceOne - แทนที่ document ทั้งหมด
		replaceOpts := options.Replace().SetUpsert(reqBody.Upsert)
		result, err = collection.ReplaceOne(ctx, reqBody.Filter, reqBody.Data, replaceOpts)
	} else {
		// UpdateOne with $set - update เฉพาะ field ที่ระบุ
		updateOpts := options.Update().SetUpsert(reqBody.Upsert)
		update := bson.M{"$set": reqBody.Data}
		result, err = collection.UpdateOne(ctx, reqBody.Filter, update, updateOpts)
	}

	if err != nil {
		log.Printf("❌ MongoDB Atlas update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    500,
			"message": "Failed to update document",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"code":           200,
		"matched_count":  result.MatchedCount,
		"modified_count": result.ModifiedCount,
		"upserted_id":    result.UpsertedID,
	})
}

// MongoAtlasDeleteHandler - Delete document(s)
// Frontend กำหนด filter เอง
func MongoAtlasDeleteHandler(c *gin.Context) {
	if atlasDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"code":    503,
			"message": "MongoDB Atlas is not connected",
		})
		return
	}

	var reqBody struct {
		Collection string                 `json:"collection"`
		Filter     map[string]interface{} `json:"filter"`     // Frontend กำหนด filter เอง
		DeleteMany bool                   `json:"delete_many"` // true = deleteMany, false = deleteOne
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    400,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Validate required fields
	if reqBody.Collection == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    400,
			"message": "Collection name is required",
		})
		return
	}

	if reqBody.Filter == nil {
		reqBody.Filter = make(map[string]interface{})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := atlasDB.Collection(reqBody.Collection)
	var deletedCount int64
	var deleteErr error

	if reqBody.DeleteMany {
		result, err := collection.DeleteMany(ctx, reqBody.Filter)
		if err != nil {
			deleteErr = err
		} else {
			deletedCount = result.DeletedCount
		}
	} else {
		result, err := collection.DeleteOne(ctx, reqBody.Filter)
		if err != nil {
			deleteErr = err
		} else {
			deletedCount = result.DeletedCount
		}
	}

	if deleteErr != nil {
		log.Printf("❌ MongoDB Atlas delete error: %v", deleteErr)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    500,
			"message": "Failed to delete document(s)",
			"error":   deleteErr.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"code":          200,
		"deleted_count": deletedCount,
	})
}

// MongoAtlasGetHandler - Get document(s)
// Frontend กำหนด filter, sort, projection เอง
func MongoAtlasGetHandler(c *gin.Context) {
	if atlasDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"code":    503,
			"message": "MongoDB Atlas is not connected",
		})
		return
	}

	var reqBody struct {
		Collection string                 `json:"collection"`
		Filter     map[string]interface{} `json:"filter"`     // Frontend กำหนด filter เอง
		Projection map[string]interface{} `json:"projection"` // กำหนด field ที่ต้องการ (optional)
		Sort       map[string]interface{} `json:"sort"`       // กำหนดการเรียงลำดับ (optional)
		Limit      int64                  `json:"limit"`      // จำนวนสูงสุด
		Skip       int64                  `json:"skip"`       // ข้าม (สำหรับ pagination)
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    400,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Validate required fields
	if reqBody.Collection == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"code":    400,
			"message": "Collection name is required",
		})
		return
	}

	if reqBody.Filter == nil {
		reqBody.Filter = make(map[string]interface{})
	}

	// Query options
	opts := options.Find()
	if reqBody.Limit > 0 {
		opts.SetLimit(reqBody.Limit)
	}
	if reqBody.Skip > 0 {
		opts.SetSkip(reqBody.Skip)
	}
	if reqBody.Projection != nil && len(reqBody.Projection) > 0 {
		opts.SetProjection(reqBody.Projection)
	}
	if reqBody.Sort != nil && len(reqBody.Sort) > 0 {
		opts.SetSort(reqBody.Sort)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := atlasDB.Collection(reqBody.Collection)
	cursor, err := collection.Find(ctx, reqBody.Filter, opts)
	if err != nil {
		log.Printf("❌ MongoDB Atlas find error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    500,
			"message": "Failed to query documents",
			"error":   err.Error(),
		})
		return
	}
	defer cursor.Close(ctx)

	// Decode results
	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		log.Printf("❌ MongoDB Atlas decode error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"code":    500,
			"message": "Failed to decode documents",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"code":   200,
		"count":  len(results),
		"data":   results,
	})
}
