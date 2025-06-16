# SMLGOAPI - ClickHouse REST API Backend with Vector Search

REST API backend สำหรับเชื่อมต่อกับ ClickHouse database โดยใช้ Go + Gin framework พร้อม TF-IDF Vector Search

## 🚀 Features

- ✅ RESTful API endpoints
- ✅ ClickHouse native protocol connection
- ✅ **TF-IDF Vector Search** - ค้นหาสินค้าด้วย semantic similarity
- ✅ **Thai/English Text Processing** - รองรับการประมวลผลข้อความทั้งไทยและอังกฤษ
- ✅ CORS support สำหรับ frontend integration
- ✅ Enhanced logging with detailed search analytics
- ✅ Graceful shutdown
- ✅ Health check endpoint
- ✅ JSON response format
- ✅ Error handling

## 🛠️ Configuration

แก้ไขไฟล์ `.env` เพื่อเปลี่ยน configuration:

```bash
# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8080

# ClickHouse Configuration
CLICKHOUSE_HOST=161.35.98.110
CLICKHOUSE_PORT=9000
CLICKHOUSE_USER=sml2
CLICKHOUSE_PASSWORD=Md5WyoEwHfR1q6
CLICKHOUSE_SECURE=false
CLICKHOUSE_DATABASE=sml2
```

## 📁 Project Structure

```
smlgoapi/
├── .env                       # Environment configuration
├── .vscode/
│   └── tasks.json            # VS Code tasks
├── config/
│   └── config.go            # Configuration management
├── handlers/
│   └── api.go               # API route handlers (with vector search)
├── models/
│   └── models.go            # Data models
├── services/
│   ├── clickhouse.go        # ClickHouse service layer
│   └── vector_db.go         # TF-IDF Vector Database service
├── main.go                  # Main application entry point
├── test_clickhouse_legacy.go # Legacy CLI test tool
├── go.mod                   # Go module dependencies
└── README.md               # Project documentation
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21 หรือสูงกว่า
- การเชื่อมต่อกับ ClickHouse server

### Installation

1. Clone หรือ download โปรเจค
2. ติดตั้ง dependencies:
   ```bash
   go mod tidy
   ```

### Running the API Server

#### วิธีที่ 1: Run จาก source code
```bash
go run main.go
```

#### วิธีที่ 2: Build แล้วรัน executable
```bash
# Build
go build -o smlgoapi.exe main.go

# Run
./smlgoapi.exe
```

#### วิธีที่ 3: ใช้ VS Code Tasks
- กด `Ctrl+Shift+P` แล้วเลือก "Tasks: Run Task"
- เลือก "Run SMLGOAPI Server"

Server จะรันที่: **http://localhost:8080**

## 🌐 API Endpoints

### Health Check
```
GET /health
```
ตรวจสอบสถานะของ API และการเชื่อมต่อ database

### Vector Search (⭐ NEW!)
```
POST /search
Content-Type: application/json

{
  "query": "motor",
  "limit": 10,
  "offset": 0
}
```
ค้นหาสินค้าด้วย TF-IDF Vector Similarity

**JSON Body Parameters:**
- `query` (required): คำค้นหา (รองรับไทย/อังกฤษ)
- `limit` (optional): จำนวนผลลัพธ์ (default: 10, max: 100)
- `offset` (optional): เริ่มต้นจากผลลัพธ์ที่ (default: 0)

**Response:**
```json
{
  "success": true,
  "message": "Search completed successfully",
  "data": {
    "data": [
      {
        "id": "000123",
        "name": "COIL OEM BENZ SPRINTER",
        "similarity_score": 0.856,
        "metadata": {
          "code": "000123",
          "unit": "ใบ",
          "balance_qty": 2.0,
          "supplier_code": "ซ034"
        }
      }
    ],
    "total_count": 150,
    "query": "coil",
    "duration_ms": 45.2
  }
}
```

### Database Information
```
GET /api/tables
```
ดึงรายชื่อตารางทั้งหมดในฐานข้อมูล

## 📋 API Response Format

### Standard Response
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {...}
}
```

### Error Response
```json
{
  "success": false,
  "error": "Error message description"
}
```

## 🔍 Vector Search Features

### Text Processing
- **Thai Text**: ใช้ GSE (Go Segmenter Engine) สำหรับการตัดคำภาษาไทย
- **English Text**: ใช้ Snowball Stemming Algorithm
- **Multi-language**: ตรวจจับภาษาอัตโนมัติ

### TF-IDF Algorithm
- **Term Frequency (TF)**: ความถี่ของคำในเอกสาร
- **Inverse Document Frequency (IDF)**: น้ำหนักความสำคัญของคำ
- **Cosine Similarity**: การคำนวณความคล้ายคลึงระหว่าง vectors

### Search Analytics
- Real-time logging ของการค้นหา
- ระยะเวลาการประมวลผล
- จำนวนผลลัพธ์และ similarity scores
- การตรวจจับภาษา

## 📊 Sample API Calls

### Using PowerShell

**Note**: Search endpoint uses GET method with base64 encoded JSON parameters. All parameters (query, limit, offset) are packed into a JSON object and encoded as base64.

```powershell
# Health check
Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET

# Get tables
Invoke-RestMethod -Uri "http://localhost:8008/api/tables" -Method GET

# Vector search (English) - POST with JSON body
$searchBody = @{
    query = "coil"
    limit = 5
    offset = 0
} | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8008/search" -Method POST -Body $searchBody -ContentType "application/json"

# Vector search (Thai) - POST with JSON body
$searchBody = @{
    query = "คอยล์"
    limit = 3
    offset = 0
} | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8008/search" -Method POST -Body $searchBody -ContentType "application/json"

# Search with pagination
$searchBody = @{
    query = "compressor"
    limit = 10
    offset = 20
} | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8008/search" -Method POST -Body $searchBody -ContentType "application/json"
```

### Using curl

**Note**: Uses POST method with JSON body.

```bash
# Vector search (English)
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{"query":"coil","limit":5,"offset":0}'

# Vector search (Thai)
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{"query":"คอยล์","limit":3,"offset":0}'

# Search with pagination
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \```

### Using JavaScript (Frontend)

**Note**: Uses POST method with JSON body.

```javascript
// Helper function to search
async function searchProducts(query, limit = 10, offset = 0) {
  const searchBody = {
    query: query,
    limit: limit,
    offset: offset
  };
  
  const response = await fetch('http://localhost:8008/search', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(searchBody)
  });
  
  return await response.json();
}

// Vector search (English)
searchProducts('coil', 5, 0)
  .then(data => {
    console.log('Search results:', data.data.data);
    console.log('Total found:', data.data.total_count);
    console.log('Duration:', data.data.duration_ms + 'ms');
  });

// Advanced search with Thai
searchProducts('คอมเพรสเซอร์', 10, 0)
  .then(data => console.log(data));

// Search with special characters and symbols
searchProducts('AC/DC Motor 220V 50Hz', 5, 0)
  .then(data => console.log(data));

// Advanced usage with error handling
async function advancedSearch(query, limit = 10, offset = 0) {
  try {
    const data = await searchProducts(query, limit, offset);
    if (data.success) {
      console.log(`Found ${data.data.total_count} results in ${data.data.duration_ms}ms`);
      return data.data.data;
    } else {
      console.error('Search failed:', data.error);
      return [];
    }
  } catch (error) {
    console.error('Search error:', error);
    return [];
  }
}

// Usage
advancedSearch('bearing', 5).then(results => console.log(results));
```

## 🔧 Dependencies

- [gin-gonic/gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [gin-contrib/cors](https://github.com/gin-contrib/cors) - CORS middleware
- [clickhouse-go/v2](https://github.com/ClickHouse/clickhouse-go) - ClickHouse driver
- [joho/godotenv](https://github.com/joho/godotenv) - Environment variables loader
- [go-ego/gse](https://github.com/go-ego/gse) - Go Segmenter Engine (Thai text processing)
- [kljensen/snowball](https://github.com/kljensen/snowball) - Snowball stemming algorithm

## 🔍 Development Tools

### Legacy CLI Test Tool
```bash
go run test_clickhouse_legacy.go
```
รันเครื่องมือทดสอบแบบ CLI (สำหรับ debug)

## 📝 Notes

- API รองรับ CORS สำหรับการเชื่อมต่อจาก frontend
- ใช้ native ClickHouse protocol (port 9000)
- Response เป็น JSON format
- Vector search ใช้ TF-IDF algorithm พร้อม cosine similarity
- รองรับการประมวลผลข้อความไทยและอังกฤษ
- มี enhanced logging สำหรับ search analytics
- มี graceful shutdown เมื่อปิด server
- Error handling ครอบคลุม

## 🚀 Ready for Frontend Integration!

API พร้อมให้ frontend frameworks เชื่อมต่อ เช่น:
- React.js
- Vue.js  
- Angular
- Next.js
- หรือ vanilla JavaScript

**Base URL:** `http://localhost:8080`

### ⭐ Vector Search Highlights:
- **Intelligent Search**: ค้นหาด้วย semantic similarity แทนการค้นหาแบบ exact match
- **Multi-language Support**: รองรับภาษาไทยและอังกฤษ
- **Real-time Analytics**: logging รายละเอียดการค้นหาแบบ real-time
- **High Performance**: TF-IDF algorithm ที่ได้รับการปรับปรุงสำหรับ performance
