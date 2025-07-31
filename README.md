# SMLGOAPI - ClickHouse REST API Backend with Vector Search

REST API backend สำหรับเชื่อมต่อกับ ClickHouse database โดยใช้ Go + Gin framework พร้อม TF-IDF Vector Search และ Universal SQL Execution

## 🔍 Search Features

### Multi-Language Full Text Search
- **Thai Language Support**: Full text search for Thai products (e.g., "ทองแดง", "คอยาว")
- **English Language Support**: Full text search for English products (e.g., "coil", "brake")
- **OR Logic**: Multi-word queries use OR logic for broader results
- **Case-Insensitive**: Uses PostgreSQL ILIKE for case-insensitive matching
- **Priority Scoring**: Results ranked by exact matches, code matches, and name matches
- **🤖 AI Query Enhancement**: AI-powered query enhancement for better results

### Search Endpoints
- **GET Method**: `/v1/search?q=query&limit=10&offset=0&ai=0`
- **POST Method**: `/v1/search` with JSON body `{"query": "search term", "limit": 10, "offset": 0, "ai": 0}`

### AI Parameter (Powered by DeepSeek AI)
- **ai=0**: ไม่ใช้ AI (default) - ค้นหาด้วยคำค้นต้นฉบับ
- **ai=1**: ใช้ DeepSeek AI ปรับปรุงคำค้น - AI จะช่วยแปล ปรับปรุง และแก้ไขคำค้นให้ดีขึ้น

### 🧠 DeepSeek AI Features:
- **Thai ↔ English Translation**: แปลคำไทยเป็นอังกฤษอัตโนมัติ
- **Typo Correction**: แก้ไขคำผิด (เช่น "break" → "brake")
- **Plural to Singular**: แปลง (เช่น "coils" → "coil")
- **Query Optimization**: ปรับปรุงคำค้นให้เหมาะสมกับฐานข้อมูลอะไหล่รถยนต์
- **Fallback Support**: หาก DeepSeek API ไม่พร้อมใช้งาน จะใช้ระบบสำรอง

### Search Examples
```bash
# English search (no AI)
curl "http://localhost:8008/v1/search?q=coil&limit=5&ai=0"

# Thai search with AI enhancement
curl "http://localhost:8008/v1/search?q=เบรค&limit=5&ai=1"

# Multi-word search with AI
curl "http://localhost:8008/v1/search?q=brake pad&limit=5&ai=1"

# POST method with AI enabled
curl -X POST "http://localhost:8008/v1/search" \
  -H "Content-Type: application/json" \
  -d '{"query": "เบรค", "limit": 5, "ai": 1}'

# AI query enhancement examples:
# "เบรค" with ai=1 → enhanced to "brake"
# "break" with ai=1 → corrected to "brake"  
# "coils" with ai=1 → normalized to "coil"
```

## 🤖 AI Agent Integration

SMLGOAPI รองรับการใช้งานโดย AI agents โดยมี **`/guide` endpoint** ที่ให้ข้อมูลครบถ้วนเกี่ยวกับ API:

### 📖 Guide Endpoint
- **URL**: `GET /guide`
- **Purpose**: Complete API documentation for AI agents
- **Response**: Comprehensive JSON with all endpoint details, examples, and best practices

```bash
curl http://localhost:8008/guide
```

### 🧠 AI Agent Features
- **Self-Documenting API**: AI agents can discover all capabilities via `/guide`
- **Universal SQL Execution**: Execute any SQL command or query via JSON
- **Consistent Response Format**: All endpoints return standardized JSON
- **Error Handling**: Complete error information for robust integration
- **Performance Metrics**: Duration tracking for all operations

### 📚 AI Integration Examples
- `ai_agent_example.py` - Complete Python example for AI agents
- `COMMAND_SELECT_API_GUIDE.md` - Detailed usage guide
- `AI_GUIDE_ENDPOINT_DOCUMENTATION.md` - AI-specific documentation

## 🚀 Git Deployment Guide

### 📋 การ Deploy ด้วย Git

#### วิธีการ Deploy แบบ Traditional

1. **ตรวจสอบการเปลี่ยนแปลง**
   ```bash
   git status
   git diff
   ```

2. **เพิ่มไฟล์ที่เปลี่ยนแปลง**
   ```bash
   git add .
   ```

3. **Commit การเปลี่ยนแปลง**
   ```bash
   git commit -m "Deploying the latest changes"
   ```

4. **Push ไปยัง Repository**
   ```bash
   git push
   ```

5. **Deploy ไปยัง Production Server**
   ```bash
   # เชื่อมต่อไปยัง production server
   ssh root@103.13.30.32
   
   # เข้าไปยังโฟลเดอร์โปรเจค
   cd /data/smlgoapi/
   
   # ดึง Docker image ล่าสุดจาก registry
   docker pull ghcr.io/jaturapornchai/smlgoapi:master
   
   # รีสตาร์ท containers
   docker compose up -d
   ```

### 🛠️ การตั้งค่า Production Environment

สำหรับการ deploy ไปยัง production server:

1. **ตรวจสอบการเชื่อมต่อ SSH**
   ```bash
   ssh root@103.13.30.32 "docker --version"
   ```

2. **ตรวจสอบสถานะ containers**
   ```bash
   ssh root@103.13.30.32 "cd /data/vectorapi-dev/ && docker compose ps"
   ```

3. **ดู logs การทำงาน**
   ```bash
   ssh root@103.13.30.32 "cd /data/vectorapi-dev/ && docker compose logs -f"
   ```

4. **Restart services (หากจำเป็น)**
   ```bash
   ssh root@103.13.30.32 "cd /data/vectorapi-dev/ && docker compose restart"
   ```

### 📦 Deploy Command Summary

```bash
# Local: Push code changes
git add .
git commit -m "Update features"
git push

# Production: Deploy to server
ssh root@103.13.30.32
cd /data/vectorapi-dev/
docker pull ghcr.io/smlsoft/vectordbapi:main
docker compose up -d
exit
```

### 🔧 Production Deployment

#### Docker Deployment (สำหรับ Production Server)

หากต้องการใช้ Docker ใน production server:

1. **Build Docker Image**
   ```bash
   docker build -t smlgoapi:latest .
   ```

2. **รัน Container**
   ```bash
   docker run -d \
     --name smlgoapi \
     -p 8080:8080 \
     -e SERVER_HOST=0.0.0.0 \
     -e SERVER_PORT=8080 \
     -e CLICKHOUSE_HOST=your-clickhouse-host \
     -e CLICKHOUSE_PORT=9000 \
     -e CLICKHOUSE_USER=your-user \
     -e CLICKHOUSE_PASSWORD=your-password \
     -e CLICKHOUSE_DATABASE=your-database \
     -v $(pwd)/image_cache:/root/image_cache \
     smlgoapi:latest
   ```

### 🛠️ การตั้งค่า Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SERVER_HOST` | Server bind address | `0.0.0.0` | No |
| `SERVER_PORT` | Server port | `8080` | No |
| `CLICKHOUSE_HOST` | ClickHouse hostname | `localhost` | Yes |
| `CLICKHOUSE_PORT` | ClickHouse port | `9000` | No |
| `CLICKHOUSE_USER` | ClickHouse username | `default` | No |
| `CLICKHOUSE_PASSWORD` | ClickHouse password | `` | No |
| `CLICKHOUSE_DATABASE` | ClickHouse database | `default` | No |
| `CLICKHOUSE_SECURE` | Use SSL connection | `false` | No |
| `GIN_MODE` | Gin framework mode | `debug` | No |

## 📖 วิธีใช้แบบ Traditional (ไม่ใช้ Docker)

### 🔧 การติดตั้งและเริ่มต้น

1. **ตรวจสอบ Go version**
   ```bash
   go version  # ต้องเป็น Go 1.23 หรือสูงกว่า
   ```

2. **Clone project และติดตั้ง dependencies**
   ```bash
   cd c:\test\smlgoapi
   go mod tidy
   ```

3. **ตั้งค่าไฟล์ .env** (ถ้าจำเป็น)
   ```bash   # ค่าปัจจุบันทำงานอยู่แล้ว
   SERVER_HOST=0.0.0.0
   SERVER_PORT=8080
   CLICKHOUSE_HOST=161.35.98.110
   CLICKHOUSE_PORT=9000
   CLICKHOUSE_USER=wawa
   CLICKHOUSE_PASSWORD=TEGmUnjQuiqjvFMY
   CLICKHOUSE_DATABASE=datawawa
   ```

4. **รัน Server**
   ```bash
   # วิธีที่ 1: รันตรง ๆ
   go run .
   
   # วิธีที่ 2: Build แล้วรัน
   go build -o smlgoapi.exe main.go
   ./smlgoapi.exe
   
   # วิธีที่ 3: ใช้ VS Code Task (กด Ctrl+Shift+P -> Tasks: Run Task -> Run SMLGOAPI Server)
   ```

5. **ตรวจสอบสถานะ**
   - Server จะรันที่: `http://localhost:8080`
   - ดู log ใน console เพื่อตรวจสอบสถานะ

### 🌐 การใช้งาน API

#### 1. Health Check - ตรวจสอบสถานะ
```bash
# PowerShell
Invoke-RestMethod -Uri "http://localhost:8080/health"

# curl
curl http://localhost:8080/health
```

**ผลลัพธ์ที่คาดหวัง:**
```json
{
  "status": "healthy",
  "timestamp": "2025-06-16T10:22:58+07:00",
  "version": "25.5.1.2782",
  "database": "connected"
}
```

#### 2. Vector Search - ค้นหาสินค้า

##### 🔍 การค้นหาพื้นฐาน
```bash
# PowerShell
$body = @{
    query = "motor"
    limit = 5
} | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/search" -Method POST -Body $body -ContentType "application/json"

# curl
curl -X POST http://localhost:8080/search \
  -H "Content-Type: application/json" \
  -d '{"query":"motor","limit":5}'
```

##### 🔍 การค้นหาภาษาไทย
```bash
# PowerShell
$body = @{
    query = "คอมเพรสเซอร์"
    limit = 3
} | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/search" -Method POST -Body $body -ContentType "application/json"

# curl
curl -X POST http://localhost:8080/search \
  -H "Content-Type: application/json" \
  -d '{"query":"คอมเพรสเซอร์","limit":3}'
```

##### 🔍 การค้นหาแบบมี pagination
```bash
# PowerShell
$body = @{
    query = "bearing"
    limit = 10
    offset = 20
} | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/search" -Method POST -Body $body -ContentType "application/json"
```

#### 3. Image Proxy - รูปภาพ

##### 🎨 การดึงรูปภาพผ่าน proxy (พร้อม resize)
```bash
# ดึงรูปขนาดเดิม
curl "http://localhost:8080/imgproxy?url=https://picsum.photos/200/300" -o original_image.jpg

# Resize เป็นขนาด 300x375
curl "http://localhost:8080/imgproxy?url=https://picsum.photos/800/600&w=300&h=375" -o resized_image.jpg

# Resize แค่ width (รักษา aspect ratio)
curl "http://localhost:8080/imgproxy?url=https://picsum.photos/400/400&w=200" -o width_only.jpg

# Resize แค่ height (รักษา aspect ratio)  
curl "http://localhost:8080/imgproxy?url=https://picsum.photos/600/400&h=150" -o height_only.jpg

# ดึงรูปจากสินค้าใน database
curl "http://localhost:8080/imgproxy?url=https://f.ptcdn.info/468/065/000/pw5l8933TR0cL0CH7f-o.jpg&w=200&h=200" -o product_thumbnail.jpg
```

##### 📐 Image Resize Parameters
- `w` (width): ความกว้างที่ต้องการ (1-2000 pixels)
- `h` (height): ความสูงที่ต้องการ (1-2000 pixels)
- **Aspect Ratio**: ถ้าระบุแค่ width หรือ height จะรักษาสัดส่วนเดิม
- **Cache**: รูปที่ resize แล้วจะถูก cache แยกต่างหากตาม size
- **Quality**: ใช้ JPEG quality 90% สำหรับการ resize

#### 4. ดูข้อมูล Database
```bash
# ดูรายชื่อตารางทั้งหมด
curl http://localhost:8080/api/tables
```

### 💻 การใช้งานจาก Frontend

#### JavaScript/TypeScript Example
```javascript
// ฟังก์ชันสำหรับค้นหา
async function searchProducts(query, limit = 10, offset = 0) {
  try {
    const response = await fetch('http://localhost:8080/search', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ query, limit, offset })
    });
    
    const data = await response.json();
    if (data.success) {
      console.log(`พบ ${data.data.total_count} รายการ ใช้เวลา ${data.data.duration_ms}ms`);
      return data.data.data;
    } else {
      console.error('ค้นหาไม่สำเร็จ:', data.error);
      return [];
    }
  } catch (error) {
    console.error('เกิดข้อผิดพลาด:', error);
    return [];
  }
}

// ตัวอย่างการใช้งาน
searchProducts('motor', 5).then(results => {
  results.forEach(product => {
    console.log(`${product.name} (Score: ${product.similarity_score.toFixed(3)})`);
  });
});

// ฟังก์ชันสำหรับแสดงรูปภาพ
function getImageProxyUrl(originalUrl) {
  return `http://localhost:8080/imgproxy?url=${encodeURIComponent(originalUrl)}`;
}

// ใช้งานใน HTML
// <img src="getImageProxyUrl('https://f.ptcdn.info/468/065/000/pw5l8933TR0cL0CH7f-o.jpg')" />
```

#### React Example
```jsx
import React, { useState, useEffect } from 'react';

function ProductSearch() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [loading, setLoading] = useState(false);

  const handleSearch = async () => {
    if (!query.trim()) return;
    
    setLoading(true);
    try {
      const response = await fetch('http://localhost:8080/search', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query, limit: 10 })
      });
      
      const data = await response.json();
      if (data.success) {
        setResults(data.data.data);
      }
    } catch (error) {
      console.error('Search error:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <input 
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="ค้นหาสินค้า..."
        onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
      />
      <button onClick={handleSearch} disabled={loading}>
        {loading ? 'กำลังค้นหา...' : 'ค้นหา'}
      </button>
      
      <div>
        {results.map(product => (
          <div key={product.id}>
            <h3>{product.name}</h3>
            <p>Score: {product.similarity_score.toFixed(3)}</p>
            {product.img_url && (
              <img 
                src={`http://localhost:8080/imgproxy?url=${encodeURIComponent(product.img_url)}`}
                alt={product.name}
                style={{width: 100, height: 100}}
              />
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
```

### 📱 การใช้งานจาก Flutter

#### การตั้งค่าเบื้องต้น

1. **เพิ่ม dependencies ใน pubspec.yaml**
```yaml
dependencies:
  flutter:
    sdk: flutter
  http: ^1.1.0
  cached_network_image: ^3.3.0
  
dev_dependencies:
  flutter_test:
    sdk: flutter
```

2. **เพิ่ม internet permission (Android)**
ใน `android/app/src/main/AndroidManifest.xml`:
```xml
<uses-permission android:name="android.permission.INTERNET" />
```

3. **เพิ่ม network configuration (iOS)**
ใน `ios/Runner/Info.plist`:
```xml
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSAllowsArbitraryLoads</key>
    <true/>
</dict>
```

#### 🔧 Flutter Service Class

สร้างไฟล์ `lib/services/smlgo_api_service.dart`:

```dart
import 'dart:convert';
import 'package:http/http.dart' as http;

class SMLGOApiService {
  static const String baseUrl = 'http://103.13.30.32:8008'; // เปลี่ยนเป็น IP ของเซิร์ฟเวอร์
  
  // Health Check
  static Future<Map<String, dynamic>?> checkHealth() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/health'),
        headers: {'Content-Type': 'application/json'},
      );
      
      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Health check error: $e');
      return null;
    }
  }
  
  // Vector Search
  static Future<SearchResult?> searchProducts({
    required String query,
    int limit = 10,
    int offset = 0,
  }) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/search'),
        headers: {'Content-Type': 'application/json'},
        body: json.encode({
          'query': query,
          'limit': limit,
          'offset': offset,
        }),
      );
      
      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        if (data['success'] == true) {
          return SearchResult.fromJson(data['data']);
        }
      }
      return null;
    } catch (e) {
      print('Search error: $e');
      return null;
    }
  }
  
  // Get Image Proxy URL
  static String getImageProxyUrl(String originalUrl) {
    return '$baseUrl/imgproxy?url=${Uri.encodeComponent(originalUrl)}';
  }
  
  // Get Tables
  static Future<List<dynamic>?> getTables() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/api/tables'),
        headers: {'Content-Type': 'application/json'},
      );
      
      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        if (data['success'] == true) {
          return data['data'];
        }
      }
      return null;
    } catch (e) {
      print('Get tables error: $e');
      return null;
    }
  }
}

// Data Models
class SearchResult {
  final List<Product> data;
  final int totalCount;
  final String query;
  final double durationMs;
  
  SearchResult({
    required this.data,
    required this.totalCount,
    required this.query,
    required this.durationMs,
  });
  
  factory SearchResult.fromJson(Map<String, dynamic> json) {
    return SearchResult(
      data: (json['data'] as List).map((item) => Product.fromJson(item)).toList(),
      totalCount: json['total_count'] ?? 0,
      query: json['query'] ?? '',
      durationMs: (json['duration_ms'] ?? 0.0).toDouble(),
    );
  }
}

class Product {
  final String id;
  final String name;
  final double similarityScore;
  final ProductMetadata metadata;
  final String? imgUrl;
  
  Product({
    required this.id,
    required this.name,
    required this.similarityScore,
    required this.metadata,
    this.imgUrl,
  });
  
  factory Product.fromJson(Map<String, dynamic> json) {
    return Product(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      similarityScore: (json['similarity_score'] ?? 0.0).toDouble(),
      metadata: ProductMetadata.fromJson(json['metadata'] ?? {}),
      imgUrl: json['img_url'],
    );
  }
}

class ProductMetadata {
  final double balanceQty;
  final String code;
  final String? imgUrl;
  final double price;
  final String supplierCode;
  final String unit;
  
  ProductMetadata({
    required this.balanceQty,
    required this.code,
    this.imgUrl,
    required this.price,
    required this.supplierCode,
    required this.unit,
  });
  
  factory ProductMetadata.fromJson(Map<String, dynamic> json) {
    return ProductMetadata(
      balanceQty: (json['balance_qty'] ?? 0.0).toDouble(),
      code: json['code'] ?? '',
      imgUrl: json['img_url'],
      price: (json['price'] ?? 0.0).toDouble(),
      supplierCode: json['supplier_code'] ?? '',
      unit: json['unit'] ?? '',
    );
  }
}
```

#### 🎨 Flutter UI Example

สร้างไฟล์ `lib/screens/product_search_screen.dart`:

```dart
import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../services/smlgo_api_service.dart';

class ProductSearchScreen extends StatefulWidget {
  @override
  _ProductSearchScreenState createState() => _ProductSearchScreenState();
}

class _ProductSearchScreenState extends State<ProductSearchScreen> {
  final TextEditingController _searchController = TextEditingController();
  List<Product> _products = [];
  bool _isLoading = false;
  bool _isHealthy = false;
  String _lastQuery = '';
  double _lastDuration = 0.0;
  int _totalCount = 0;

  @override
  void initState() {
    super.initState();
    _checkHealth();
  }

  Future<void> _checkHealth() async {
    final health = await SMLGOApiService.checkHealth();
    setState(() {
      _isHealthy = health != null && health['status'] == 'healthy';
    });
  }

  Future<void> _searchProducts() async {
    final query = _searchController.text.trim();
    if (query.isEmpty) return;

    setState(() {
      _isLoading = true;
    });

    final result = await SMLGOApiService.searchProducts(
      query: query,
      limit: 20,
    );

    setState(() {
      _isLoading = false;
      if (result != null) {
        _products = result.data;
        _lastQuery = result.query;
        _lastDuration = result.durationMs;
        _totalCount = result.totalCount;
      } else {
        _products = [];
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('SMLGO Product Search'),
        backgroundColor: Colors.blue[600],
        actions: [
          Padding(
            padding: EdgeInsets.only(right: 16),
            child: Center(
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    _isHealthy ? Icons.check_circle : Icons.error,
                    color: _isHealthy ? Colors.green : Colors.red,
                    size: 20,
                  ),
                  SizedBox(width: 4),
                  Text(
                    _isHealthy ? 'Online' : 'Offline',
                    style: TextStyle(fontSize: 12),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
      body: Column(
        children: [
          // Search Section
          Container(
            padding: EdgeInsets.all(16),
            color: Colors.grey[100],
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _searchController,
                    decoration: InputDecoration(
                      hintText: 'ค้นหาสินค้า... (เช่น motor, คอมเพรสเซอร์)',
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                      prefixIcon: Icon(Icons.search),
                      filled: true,
                      fillColor: Colors.white,
                    ),
                    onSubmitted: (_) => _searchProducts(),
                  ),
                ),
                SizedBox(width: 8),
                ElevatedButton(
                  onPressed: _isLoading ? null : _searchProducts,
                  child: _isLoading
                      ? SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : Text('ค้นหา'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.blue[600],
                    foregroundColor: Colors.white,
                    padding: EdgeInsets.symmetric(horizontal: 20, vertical: 15),
                  ),
                ),
              ],
            ),
          ),

          // Results Info
          if (_products.isNotEmpty)
            Container(
              padding: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              color: Colors.blue[50],
              child: Row(
                children: [
                  Text(
                    'พบ $_totalCount รายการ สำหรับ "$_lastQuery"',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                  Spacer(),
                  Text(
                    '${_lastDuration.toStringAsFixed(1)}ms',
                    style: TextStyle(color: Colors.grey[600]),
                  ),
                ],
              ),
            ),

          // Products List
          Expanded(
            child: _isLoading
                ? Center(child: CircularProgressIndicator())
                : _products.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.search, size: 64, color: Colors.grey),
                            SizedBox(height: 16),
                            Text(
                              'ค้นหาสินค้าเพื่อแสดงผลลัพธ์',
                              style: TextStyle(
                                fontSize: 16,
                                color: Colors.grey[600],
                              ),
                            ),
                          ],
                        ),
                      )
                    : ListView.builder(
                        itemCount: _products.length,
                        itemBuilder: (context, index) {
                          final product = _products[index];
                          return ProductCard(product: product);
                        },
                      ),
          ),
        ],
      ),
    );
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }
}

class ProductCard extends StatelessWidget {
  final Product product;

  const ProductCard({Key? key, required this.product}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Padding(
        padding: EdgeInsets.all(16),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Product Image
            Container(
              width: 80,
              height: 80,
              child: product.imgUrl != null
                  ? CachedNetworkImage(
                      imageUrl: SMLGOApiService.getImageProxyUrl(product.imgUrl!),
                      placeholder: (context, url) => Container(
                        color: Colors.grey[200],
                        child: Icon(Icons.image, color: Colors.grey),
                      ),
                      errorWidget: (context, url, error) => Container(
                        color: Colors.grey[200],
                        child: Icon(Icons.broken_image, color: Colors.grey),
                      ),
                      fit: BoxFit.cover,
                    )
                  : Container(
                      color: Colors.grey[200],
                      child: Icon(Icons.image, color: Colors.grey),
                    ),
            ),
            SizedBox(width: 16),

            // Product Info
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    product.name,
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                  SizedBox(height: 8),
                  Row(
                    children: [
                      Text(
                        'รหัส: ${product.metadata.code}',
                        style: TextStyle(color: Colors.grey[600]),
                      ),
                      Spacer(),
                      Container(
                        padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                        decoration: BoxDecoration(
                          color: Colors.green[100],
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Text(
                          '${(product.similarityScore * 100).toStringAsFixed(1)}%',
                          style: TextStyle(
                            color: Colors.green[700],
                            fontWeight: FontWeight.bold,
                            fontSize: 12,
                          ),
                        ),
                      ),
                    ],
                  ),
                  SizedBox(height: 4),
                  Row(
                    children: [
                      Text(
                        'คงเหลือ: ${product.metadata.balanceQty.toStringAsFixed(0)} ${product.metadata.unit}',
                        style: TextStyle(color: Colors.grey[600]),
                      ),
                      Spacer(),
                      Text(
                        '฿${product.metadata.price.toStringAsFixed(0)}',
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                          color: Colors.blue[700],
                        ),
                      ),
                    ],
                  ),
                  if (product.metadata.supplierCode.isNotEmpty)
                    Padding(
                      padding: EdgeInsets.only(top: 4),
                      child: Text(
                        'ผู้จำหน่าย: ${product.metadata.supplierCode}',
                        style: TextStyle(
                          color: Colors.grey[600],
                          fontSize: 12,
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ],
        ),
      );
    }
  }
}
```

#### 🖼️ การใช้งาน Image Proxy ใน Flutter

```dart
// การใช้งานพื้นฐาน
CachedNetworkImage(
  imageUrl: SMLGOApiService.getImageProxyUrl(product.imgUrl!),
  width: 100, height: 100, fit: BoxFit.cover,
)

// การใช้งานแบบ resize
CachedNetworkImage(
  imageUrl: SMLGOApiService.getImageProxyUrl(
    product.imgUrl!,
    width: 300,
    height: 200
  ),
  fit: BoxFit.cover,
)

// ใช้ preset sizes
CachedNetworkImage(
  imageUrl: SMLGOApiService.getThumbnailUrl(product.imgUrl!), // 150x150
  width: 50, height: 50,
)

CachedNetworkImage(
  imageUrl: SMLGOApiService.getMediumImageUrl(product.imgUrl!), // 400x300
  fit: BoxFit.cover,
)

// Product List ด้วย thumbnail
ListView.builder(
  itemBuilder: (context, index) {
    final product = products[index];
    return ListTile(
      leading: product.imgUrl != null 
        ? CachedNetworkImage(
            imageUrl: SMLGOApiService.getThumbnailUrl(product.imgUrl!),
            width: 60, height: 60, fit: BoxFit.cover,
            placeholder: (context, url) => Container(
              color: Colors.grey[200],
              child: Icon(Icons.image),
            ),
          )
        : Container(
            width: 60, height: 60,
            color: Colors.grey[200],
            child: Icon(Icons.image),
          ),
      title: Text(product.name),
      subtitle: Text('Score: ${(product.similarityScore * 100).toStringAsFixed(1)}%'),
    );
  },
)
```

#### 🚀 การใช้งานใน main.dart

```dart
import 'package:flutter/material.dart';
import 'screens/product_search_screen.dart';

void main() {
  runApp(MyApp());
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'SMLGO Product Search',
      theme: ThemeData(
        primarySwatch: Colors.blue,
        visualDensity: VisualDensity.adaptivePlatformDensity,
      ),
      home: ProductSearchScreen(),
      debugShowCheckedModeBanner: false,
    );
  }
}
```

#### 📝 ตัวอย่างไฟล์ flutter_client_example.dart

```dart
// ตัวอย่างการใช้งาน API เบื้องต้น
import 'dart:convert';
import 'package:http/http.dart' as http;

void main() async {
  // Health Check
  final healthResponse = await http.get(Uri.parse('http://localhost:8080/health'));
  print('Health: ${healthResponse.body}');
  
  // Search Products
  final searchResponse = await http.post(
    Uri.parse('http://localhost:8080/search'),
    headers: {'Content-Type': 'application/json'},
    body: json.encode({
      'query': 'motor',
      'limit': 5,
    }),
  );
  
  if (searchResponse.statusCode == 200) {
    final data = json.decode(searchResponse.body);
    print('Search Results: ${data['data']['total_count']} items found');
    
    for (var product in data['data']['data']) {
      print('- ${product['name']} (Score: ${product['similarity_score']})');
    }
  }
}
```

#### 🔧 การแก้ปัญหาสำหรับ Flutter

1. **Network Permission Issues**:
   ```dart
   // เพิ่มใน android/app/src/main/AndroidManifest.xml
   <uses-permission android:name="android.permission.INTERNET" />
   <uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
   ```

2. **CORS Issues**:
   - API มี CORS enabled แล้ว
   - ถ้ายังมีปัญหา ให้เปลี่ยน localhost เป็น IP address ของเครื่อง

3. **Image Loading Issues**:
   ```dart
   // ใช้ cached_network_image สำหรับการจัดการรูปภาพ
   CachedNetworkImage(
     imageUrl: SMLGOApiService.getImageProxyUrl(imageUrl),
     httpHeaders: {'User-Agent': 'Flutter App'},
   )
   ```

4. **Connection Issues**:
   ```dart
   // ตรวจสอบการเชื่อมต่อ
   static const String baseUrl = 'http://192.168.1.100:8080'; // ใช้ IP แทน localhost
   ```

#### 🔧 การตั้งค่า Network Configuration

##### สำหรับ Android (android/app/src/main/AndroidManifest.xml):
```xml
<manifest xmlns:android="http://schemas.android.com/apk/res/android">
    <!-- เพิ่ม permissions เหล่านี้ -->
    <uses-permission android:name="android.permission.INTERNET" />
    <uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
    
    <!-- สำหรับ HTTP (ไม่ใช่ HTTPS) -->
    <application
        android:usesCleartextTraffic="true"
        ... >
        ...
    </application>
</manifest>
```

##### สำหรับ iOS (ios/Runner/Info.plist):
```xml
<dict>
    <!-- เพิ่มการตั้งค่าเหล่านี้ -->
    <key>NSAppTransportSecurity</key>
    <dict>
        <key>NSAllowsArbitraryLoads</key>
        <true/>
    </dict>
    
    <!-- สำหรับ Camera/Photo Library (ถ้าต้องการ) -->
    <key>NSCameraUsageDescription</key>
    <string>This app needs camera access to scan barcodes</string>
    <key>NSPhotoLibraryUsageDescription</key>
    <string>This app needs photo library access to select images</string>
</dict>
```

##### การหา IP Address ของเครื่อง:
```bash
# Windows
ipconfig | findstr "IPv4"

# macOS/Linux  
ifconfig | grep "inet "

# ตัวอย่าง: เปลี่ยนใน Flutter code
static const String baseUrl = 'http://192.168.1.100:8080'; // เปลี่ยนเป็น IP ของเครื่องที่รัน API
```

### 📊 การตรวจสอบ Performance

```bash
# ดู log ใน console จะแสดง:
# - เวลาที่ใช้ในการค้นหา (duration_ms)
# - จำนวนผลลัพธ์ที่พบ
# - Similarity scores
# - การใช้งาน image cache

# ตัวอย่าง log:
# 🔍 Search: 'motor' (limit: 5)
# 📋 TOP RESULTS: 1. Motor ABC (Score: 0.856)
# ✅ SEARCH COMPLETED (45.2ms)
```

## 🐳 คู่มือ Deploy ด้วย Docker

### 📋 ข้อกำหนดเบื้องต้น

ก่อนเริ่ม deploy ต้องติดตั้งสิ่งเหล่านี้ก่อน:

1. **Docker Desktop** (สำหรับ Windows/Mac)
   ```bash
   # ตรวจสอบว่าติดตั้งแล้ว
   docker --version
   docker-compose --version
   ```

2. **Make** (optional แต่แนะนำ)
   ```bash
   # Windows: ติดตั้ง Make ผ่าน Chocolatey
   choco install make
   
   # หรือใช้ PowerShell ใน Windows ก็ได้
   ```

### 🚀 วิธี Deploy แบบง่าย (Quick Start)

#### 1. Build และ Run ด้วย Docker Compose

```bash
# สร้าง Docker image และรัน
make run

# หรือใช้คำสั่ง docker-compose ตรง ๆ
docker-compose up -d
```

#### 2. ตรวจสอบสถานะ

```bash
# ดูสถานะ container
make status

# ดู logs
make logs

# ทดสอบ API
make test
```

#### 3. เข้าใช้งาน API

- **API Base**: http://localhost:8008
- **Health Check**: http://localhost:8008/health
- **API Documentation**: http://localhost:8008/

### 📖 คำสั่ง Make ที่ใช้บ่อย

```bash
# แสดงคำสั่งทั้งหมด
make help

# Build Docker image
make build

# รัน Development mode (พร้อม hot reload)
make dev

# หยุด containers
make stop

# รีสตาร์ท
make restart

# ลบ containers และ images
make clean

# ดู logs แบบ real-time
make logs
```

### 🔧 การตั้งค่า Environment Variables

#### ไฟล์ .env (Development)
```env
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8008

# ClickHouse Configuration
CLICKHOUSE_HOST=161.35.98.110
CLICKHOUSE_PORT=9000
CLICKHOUSE_USER=wawa
CLICKHOUSE_PASSWORD=TEGmUnjQuiqjvFMY
CLICKHOUSE_SECURE=false
CLICKHOUSE_DATABASE=datawawa

# Application Configuration
GIN_MODE=release
CACHE_DIR=/app/cache
```

#### การแก้ไข docker-compose.yml
```yaml
version: '3.8'
services:
  smlgoapi:
    build: .
    ports:
      - "8008:8008"
    environment:
      # แก้ไขค่าเหล่านี้ตามต้องการ
      - CLICKHOUSE_HOST=your-clickhouse-host
      - CLICKHOUSE_USER=your-username
      - CLICKHOUSE_PASSWORD=your-password
    volumes:
      - ./image_cache:/root/image_cache
```

### 🌐 Deploy สำหรับ Production

#### 1. ใช้ Production docker-compose

```bash
# Pull image จาก registry
docker-compose -f docker-compose.prod.yml pull

# Deploy production
docker-compose -f docker-compose.prod.yml up -d

# หรือใช้ Make
make deploy
```

#### 2. Deploy บน Server

```bash
# SSH เข้า server
ssh root@103.13.30.32

# ไปยัง directory ที่ต้องการ
cd /data/vectorapi-dev/

# Clone หรือ pull โค้ดล่าสุด
git clone https://github.com/your-repo/smlgoapi.git
cd smlgoapi

# Deploy
make deploy

# ตรวจสอบสถานะ
make status
```

### 🔍 Troubleshooting

#### 1. ปัญหา Port ติดขัด
```bash
# ตรวจสอบ port ที่ใช้อยู่
netstat -an | findstr :8008

# หยุด process ที่ใช้ port
# Windows
taskkill /f /im docker.exe

# เปลี่ยน port ใน docker-compose.yml
ports:
  - "8009:8008"  # เปลี่ยนจาก 8008 เป็น 8009
```

#### 2. ปัญหา Docker build ล้มเหลว
```bash
# ลบ cache และ build ใหม่
docker system prune -a
make build

# ตรวจสอบ logs
docker-compose logs smlgoapi
```

#### 3. ปัญหาเชื่อมต่อ ClickHouse
```bash
# ตรวจสอบ network ใน container
docker-compose exec smlgoapi ping 161.35.98.110

# ตรวจสอบ environment variables
docker-compose exec smlgoapi env | grep CLICKHOUSE
```

#### 4. ปัญหา Image Cache
```bash
# ลบ cache และสร้างใหม่
rm -rf ./image_cache
mkdir ./image_cache

# restart container
make restart
```

### 📊 Monitoring และ Logs

#### 1. ดู Logs
```bash
# Real-time logs
make logs

# Logs ของ service เฉพาะ
docker-compose logs -f smlgoapi

# Logs ล่าสุด 100 บรรทัด
docker-compose logs --tail=100 smlgoapi
```

#### 2. Health Check
```bash
# ตรวจสอบ health
curl http://localhost:8008/health

# ตรวจสอบผ่าน docker
docker-compose exec smlgoapi wget -qO- http://localhost:8008/health
```

#### 3. Resource Usage
```bash
# ดู resource usage
docker stats

# ดู disk usage
docker system df
```

### 🔄 Backup และ Restore

#### 1. Backup Image Cache
```bash
# สำรอง image cache
tar -czf image_cache_backup.tar.gz ./image_cache/

# Restore
tar -xzf image_cache_backup.tar.gz
```

#### 2. Export/Import Docker Image
```bash
# Export image
docker save smlgoapi:latest > smlgoapi.tar

# Import image
docker load < smlgoapi.tar
```

### 🎯 Performance Tuning

#### 1. Docker Resource Limits
```yaml
# ใน docker-compose.yml
services:
  smlgoapi:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 1G
        reservations:
          memory: 512M
```

#### 2. Image Cache Optimization
```bash
# ตั้งค่า cache directory size
# ใน Dockerfile หรือ docker-compose.yml
volumes:
  - type: bind
    source: ./image_cache
    target: /root/image_cache
    bind:
      create_host_path: true
```

### 📋 Deployment Checklist

✅ **Pre-deployment:**
- [ ] ตรวจสอบ Docker และ docker-compose ติดตั้งแล้ว
- [ ] ตรวจสอบ .env มีค่าที่ถูกต้อง
- [ ] ทดสอบ build image สำเร็จ
- [ ] ทดสอบเชื่อมต่อ ClickHouse

✅ **Deployment:**
- [ ] Build image: `make build`
- [ ] Run containers: `make run` 
- [ ] ตรวจสอบ health: `make test`
- [ ] ตรวจสอบ logs: `make logs`

✅ **Post-deployment:**
- [ ] ทดสอบ API endpoints ทั้งหมด
- [ ] ทดสอบ image proxy
- [ ] ทดสอบ search functionality
- [ ] ตั้งค่า monitoring (ถ้าจำเป็น)

### 🌟 Advanced Docker Commands

```bash
# Build แบบไม่ใช้ cache
docker-compose build --no-cache

# รัน container แบบ interactive
docker-compose exec smlgoapi sh

# Copy files เข้า/ออก container
docker cp ./config.json smlgoapi:/root/config.json

# ตรวจสอบ network
docker network ls
docker network inspect smlgoapi_smlgoapi-network

# Cleanup containers ที่หยุดแล้ว
docker container prune

# Cleanup images ที่ไม่ใช้
docker image prune -a
```

## 🐳 Docker & GitHub Actions

### 📦 Building with Docker

This project includes a multi-stage Dockerfile for efficient container builds:

#### Local Development with Docker Compose

1. **Start the full stack (API + ClickHouse)**
   ```bash
   docker-compose up -d
   ```

2. **View logs**
   ```bash
   docker-compose logs -f smlgoapi
   ```

3. **Stop services**
   ```bash
   docker-compose down
   ```

#### Manual Docker Build

```bash
# Build the image
docker build -t smlgoapi:latest .

# Run with external ClickHouse
docker run -d \
  --name smlgoapi \
  -p 8080:8080 \
  -e CLICKHOUSE_HOST=your-clickhouse-host \
  -e CLICKHOUSE_USER=your-user \
  -e CLICKHOUSE_PASSWORD=your-password \
  -e CLICKHOUSE_DATABASE=your-database \
  -v $(pwd)/image_cache:/app/image_cache \
  smlgoapi:latest
```

### 🚀 GitHub Actions CI/CD

The project includes automated Docker image building and publishing to GitHub Container Registry (GHCR):

#### Automated Builds
- **Push to `main`**: Builds and pushes `latest` tag
- **Push to `develop`**: Builds and pushes `develop` tag  
- **Git tags (`v*`)**: Builds and pushes semantic version tags
- **Pull Requests**: Builds image without pushing (validation)

#### Using Published Images

Pull the latest image from GHCR:
```bash
# Latest stable version
docker pull ghcr.io/your-username/smlgoapi:latest

# Development version
docker pull ghcr.io/your-username/smlgoapi:develop

# Specific version
docker pull ghcr.io/your-username/smlgoapi:v1.0.0
```

#### Security Features
- Multi-architecture builds (amd64, arm64)
- Vulnerability scanning with Trivy
- Non-root user execution
- Minimal Alpine-based final image
- Security scan results in GitHub Security tab

#### Repository Setup

To enable GitHub Actions deployment to GHCR:

1. **Enable GitHub Actions** (usually enabled by default)
2. **Set repository visibility** to public, or configure package permissions for private repos
3. **Push your code** - GitHub Actions will automatically build and push images

The workflow file is located at `.github/workflows/docker-build.yml`
