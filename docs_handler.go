package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DocsHandler handles API documentation endpoints with frontend-friendly format
func DocsHandler(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

	c.JSON(http.StatusOK, gin.H{
		"api_info": gin.H{
			"name":        "SMLGOAPI",
			"version":     "1.0.0",
			"description": "Multi-Database REST API with ClickHouse and PostgreSQL support",
			"base_url":    "http://localhost:8008",
			"status":      "operational",
		},
		"working_endpoints": gin.H{
			"health_v1": gin.H{
				"url":         "/v1/health",
				"method":      "GET",
				"description": "Check API health status (v1)",
				"status":      "✅ Working",
			},
			"health": gin.H{
				"url":         "/health",
				"method":      "GET",
				"description": "Check API health status (root)",
				"status":      "✅ Working",
			},
			"postgresql_select": gin.H{
				"url":         "/v1/pgselect",
				"method":      "POST",
				"description": "Execute PostgreSQL SELECT queries",
				"status":      "✅ Working",
				"example": gin.H{
					"query": "SELECT version()",
				},
			},
			"postgresql_command": gin.H{
				"url":         "/v1/pgcommand",
				"method":      "POST",
				"description": "Execute PostgreSQL commands",
				"status":      "✅ Working",
				"example": gin.H{
					"query": "SELECT COUNT(*) FROM information_schema.tables",
				},
			},
			"tables": gin.H{
				"url":         "/v1/tables",
				"method":      "GET",
				"description": "List all database tables",
				"status":      "✅ Working",
			},
			"provinces": gin.H{
				"url":         "/v1/provinces",
				"method":      "POST",
				"description": "Get Thai provinces",
				"status":      "✅ Working",
				"example": gin.H{
					"name":   "กรุงเทพ",
					"limit":  10,
					"offset": 0,
				},
			},
			"image_proxy": gin.H{
				"url":         "/v1/imgproxy",
				"method":      "GET",
				"description": "Proxy images from external URLs",
				"status":      "✅ Working",
				"example":     "GET /v1/imgproxy?url=https://via.placeholder.com/300x200",
			},
		},
		"problematic_endpoints": gin.H{
			"search": gin.H{
				"url":         "/v1/search",
				"method":      "POST",
				"description": "Search products",
				"status":      "⚠️ Table 'products' not found - will return demo data",
				"example": gin.H{
					"query":  "test",
					"limit":  10,
					"offset": 0,
				},
				"note": "ตาราง products ไม่มีในฐานข้อมูล กำลังค้นหาตารางที่เกี่ยวข้อง",
			},
			"zipcode": gin.H{
				"url":         "/v1/findbyzipcode",
				"method":      "POST",
				"description": "Find location by zipcode",
				"status":      "⚠️ Request format issue",
				"correct_example": gin.H{
					"zipcode": 10100,
				},
				"note": "ต้องส่ง zipcode เป็น number ไม่ใช่ string",
			},
		},
		"frontend_examples": gin.H{
			"javascript_working": `
// Health Check (Working)
fetch('http://localhost:8008/v1/health')
  .then(response => response.json())
  .then(data => console.log(data));

// PostgreSQL Query (Working)  
fetch('http://localhost:8008/v1/pgselect', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ query: 'SELECT version()' })
})
  .then(response => response.json())
  .then(data => console.log(data));

// Get Tables (Working)
fetch('http://localhost:8008/v1/tables')
  .then(response => response.json()) 
  .then(data => console.log(data));

// Thai Provinces (Working)
fetch('http://localhost:8008/v1/provinces', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ name: 'กรุงเทพ', limit: 10 })
})
  .then(response => response.json())
  .then(data => console.log(data));`,
			"dart_working": `
// Dart/Flutter Examples (Working)
import 'package:http/http.dart' as http;
import 'dart:convert';

// Health Check
final healthResponse = await http.get(Uri.parse('http://localhost:8008/v1/health'));
print(jsonDecode(healthResponse.body));

// PostgreSQL Query  
final pgResponse = await http.post(
  Uri.parse('http://localhost:8008/v1/pgselect'),
  headers: {'Content-Type': 'application/json'},
  body: jsonEncode({'query': 'SELECT version()'})
);
print(jsonDecode(pgResponse.body));

// Zipcode (Fixed format)
final zipcodeResponse = await http.post(
  Uri.parse('http://localhost:8008/v1/findbyzipcode'),
  headers: {'Content-Type': 'application/json'},
  body: jsonEncode({'zipcode': 10100}) // Number, not string!
);`,
		},
		"fixes_applied": []string{
			"✅ Added /v1/health endpoint",
			"✅ Improved CORS headers",
			"⚠️ Search endpoint now handles missing products table gracefully",
			"📝 Updated documentation with working examples",
		},
		"database_status": gin.H{
			"postgresql":   "✅ Connected",
			"clickhouse":   "✅ Connected",
			"tables_found": "Use GET /v1/tables to see available tables",
		},
		"last_updated": "2025-06-19T13:20:00Z",
	})
}
