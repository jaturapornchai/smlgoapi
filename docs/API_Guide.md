# SMLGOAPI - คู่มือการใช้งาน API (Flutter Edition)

## ภาพรวม
SMLGOAPI เป็น REST API ที่ให้บริการข้อมูลการปกครองไทย, ระบบค้นหาสินค้า, Universal SQL execution สำหรับ ClickHouse Database, และ Image Proxy

**Base URL:** `http://localhost:8008`  
**Version:** 1.0.0  
**Last Updated:** 17 มิถุนายน 2025  
**Status:** ✅ ทดสอบแล้ว ทุก API ทำงานปกติ

### 🎯 Features ทั้งหมด
- ✅ **Thai Administrative Data API** (จังหวัด, อำเภอ, ตำบล)
- ✅ **Universal SQL Execution** (SELECT, INSERT, UPDATE, DELETE, CREATE)
- ✅ **Product Search** (Multi-step search with vector similarity)
- ✅ **Image Proxy** (Cache & CORS support)
- ✅ **Health Check & Monitoring**
- ✅ **Database Tables Listing**

---

## 🚀 Quick Start Flutter Setup

### 1. Add HTTP package to pubspec.yaml
```yaml
dependencies:
  flutter:
    sdk: flutter
  http: ^0.13.5
  # For JSON handling
  json_annotation: ^4.8.1
```

### 2. Create API Service Class
```dart
import 'dart:convert';
import 'package:http/http.dart' as http;

class SMLGOAPIService {
  static const String baseUrl = 'http://localhost:8008';
  
  // Headers for all requests
  static const Map<String, String> headers = {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  };
}
```

---

## 📍 ข้อมูลการปกครองไทย (Thai Administrative Data)

### 🌟 Find by Zipcode (เร็วที่สุด) - ใช้บ่อยที่สุด

```dart
// Model classes
class Province {
  final int id;
  final String nameTh;
  final String nameEn;
  
  Province({required this.id, required this.nameTh, required this.nameEn});
  
  factory Province.fromJson(Map<String, dynamic> json) {
    return Province(
      id: json['id'],
      nameTh: json['name_th'],
      nameEn: json['name_en'],
    );
  }
}

class Amphure {
  final int id;
  final String nameTh;
  final String nameEn;
  
  Amphure({required this.id, required this.nameTh, required this.nameEn});
  
  factory Amphure.fromJson(Map<String, dynamic> json) {
    return Amphure(
      id: json['id'],
      nameTh: json['name_th'],
      nameEn: json['name_en'],
    );
  }
}

class Tambon {
  final int id;
  final String nameTh;
  final String nameEn;
  final int zipCode;
  
  Tambon({required this.id, required this.nameTh, required this.nameEn, required this.zipCode});
  
  factory Tambon.fromJson(Map<String, dynamic> json) {
    return Tambon(
      id: json['id'],
      nameTh: json['name_th'],
      nameEn: json['name_en'],
      zipCode: json['zip_code'],
    );
  }
}

class LocationData {
  final Province province;
  final Amphure amphure;
  final Tambon tambon;
  
  LocationData({required this.province, required this.amphure, required this.tambon});
  
  factory LocationData.fromJson(Map<String, dynamic> json) {
    return LocationData(
      province: Province.fromJson(json['province']),
      amphure: Amphure.fromJson(json['amphure']),
      tambon: Tambon.fromJson(json['tambon']),
    );
  }
}

// API Service method
class SMLGOAPIService {
  // Find location by zipcode (แนะนำ - เร็วที่สุด)
  static Future<List<LocationData>> findByZipcode(int zipCode) async {
    final response = await http.post(
      Uri.parse('$baseUrl/get/findbyzipcode'),
      headers: headers,
      body: jsonEncode({'zip_code': zipCode}),
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      if (data['success']) {
        return (data['data'] as List)
            .map((item) => LocationData.fromJson(item))
            .toList();
      } else {
        throw Exception(data['error'] ?? 'API Error');
      }
    } else {
      throw Exception('HTTP Error: ${response.statusCode}');
    }
  }
  
  // Get all provinces
  static Future<List<Province>> getProvinces() async {
    final response = await http.post(
      Uri.parse('$baseUrl/get/provinces'),
      headers: headers,
      body: jsonEncode({}),
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      if (data['success']) {
        return (data['data'] as List)
            .map((item) => Province.fromJson(item))
            .toList();
      } else {
        throw Exception(data['error'] ?? 'API Error');
      }
    } else {
      throw Exception('HTTP Error: ${response.statusCode}');
    }
  }
  
  // Get amphures by province
  static Future<List<Amphure>> getAmphures(int provinceId) async {
    final response = await http.post(
      Uri.parse('$baseUrl/get/amphures'),
      headers: headers,
      body: jsonEncode({'province_id': provinceId}),
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      if (data['success']) {
        return (data['data'] as List)
            .map((item) => Amphure.fromJson(item))
            .toList();
      } else {
        throw Exception(data['error'] ?? 'API Error');
      }
    } else {
      throw Exception('HTTP Error: ${response.statusCode}');
    }
  }
  
  // Get tambons by amphure
  static Future<List<Tambon>> getTambons(int amphureId, int provinceId) async {
    final response = await http.post(
      Uri.parse('$baseUrl/get/tambons'),
      headers: headers,
      body: jsonEncode({
        'amphure_id': amphureId,
        'province_id': provinceId,
      }),
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      if (data['success']) {
        return (data['data'] as List)
            .map((item) => Tambon.fromJson(item))
            .toList();
      } else {
        throw Exception(data['error'] ?? 'API Error');
      }
    } else {
      throw Exception('HTTP Error: ${response.statusCode}');
    }
  }
}
```

### 📱 Flutter Widget Example - Address Form

```dart
import 'package:flutter/material.dart';

class AddressFormWidget extends StatefulWidget {
  @override
  _AddressFormWidgetState createState() => _AddressFormWidgetState();
}

class _AddressFormWidgetState extends State<AddressFormWidget> {
  final TextEditingController _zipCodeController = TextEditingController();
  List<LocationData> _locations = [];
  bool _isLoading = false;
  String? _error;

  // Method 1: Find by Zipcode (แนะนำ)
  Future<void> _searchByZipcode() async {
    final zipCode = _zipCodeController.text.trim();
    if (zipCode.isEmpty || zipCode.length < 5) return;

    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final locations = await SMLGOAPIService.findByZipcode(int.parse(zipCode));
      setState(() {
        _locations = locations;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _isLoading = false;
        _locations = [];
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('ค้นหาที่อยู่จากรหัสไปรษณีย์'),
      ),
      body: Padding(
        padding: EdgeInsets.all(16.0),
        child: Column(
          children: [
            TextField(
              controller: _zipCodeController,
              decoration: InputDecoration(
                labelText: 'รหัสไปรษณีย์',
                hintText: 'เช่น 10200',
                border: OutlineInputBorder(),
                suffixIcon: IconButton(
                  icon: Icon(Icons.search),
                  onPressed: _searchByZipcode,
                ),
              ),
              keyboardType: TextInputType.number,
              onSubmitted: (_) => _searchByZipcode(),
            ),
            SizedBox(height: 16),
            
            if (_isLoading)
              CircularProgressIndicator()
            else if (_error != null)
              Text(
                'Error: $_error',
                style: TextStyle(color: Colors.red),
              )
            else if (_locations.isNotEmpty)
              Expanded(
                child: ListView.builder(
                  itemCount: _locations.length,
                  itemBuilder: (context, index) {
                    final location = _locations[index];
                    return Card(
                      child: ListTile(
                        title: Text(location.tambon.nameTh),
                        subtitle: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('อำเภอ: ${location.amphure.nameTh}'),
                            Text('จังหวัด: ${location.province.nameTh}'),
                            Text('รหัสไปรษณีย์: ${location.tambon.zipCode}'),
                          ],
                        ),
                        onTap: () {
                          // Handle selection
                          print('Selected: ${location.tambon.nameTh}');
                        },
                      ),
                    );
                  },
                ),
              ),
          ],
        ),
      ),
    );
  }

  @override
  void dispose() {
    _zipCodeController.dispose();
    super.dispose();
  }
}
```

---

## 🗄️ Database Operations (Universal SQL)

### SELECT Query
```dart
class DatabaseService {
  static Future<List<Map<String, dynamic>>> executeSelect(String query) async {
    final response = await http.post(
      Uri.parse('${SMLGOAPIService.baseUrl}/select'),
      headers: SMLGOAPIService.headers,
      body: jsonEncode({'query': query}),
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      if (data['success']) {
        return List<Map<String, dynamic>>.from(data['data']);
      } else {
        throw Exception(data['error'] ?? 'Query failed');
      }
    } else {
      throw Exception('HTTP Error: ${response.statusCode}');
    }
  }

  static Future<Map<String, dynamic>> executeCommand(String query) async {
    final response = await http.post(
      Uri.parse('${SMLGOAPIService.baseUrl}/command'),
      headers: SMLGOAPIService.headers,
      body: jsonEncode({'query': query}),
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      if (data['success']) {
        return data['result'];
      } else {
        throw Exception(data['error'] ?? 'Command failed');
      }
    } else {
      throw Exception('HTTP Error: ${response.statusCode}');
    }
  }
  
  static Future<List<String>> getTables() async {
    final response = await http.get(
      Uri.parse('${SMLGOAPIService.baseUrl}/api/tables'),
      headers: {'Accept': 'application/json'},
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      if (data['success']) {
        return (data['data'] as List)
            .map((item) => item['name'] as String)
            .toList();
      } else {
        throw Exception(data['error'] ?? 'Failed to get tables');
      }
    } else {
      throw Exception('HTTP Error: ${response.statusCode}');
    }
  }
}

// Usage Examples
class DatabaseExample extends StatefulWidget {
  @override
  _DatabaseExampleState createState() => _DatabaseExampleState();
}

class _DatabaseExampleState extends State<DatabaseExample> {
  List<Map<String, dynamic>> _queryResults = [];
  List<String> _tables = [];
  bool _isLoading = false;

  Future<void> _runQuery() async {
    setState(() => _isLoading = true);
    
    try {
      // Example: Get current time
      final results = await DatabaseService.executeSelect(
        'SELECT 1 as test, now() as current_time LIMIT 1'
      );
      
      setState(() {
        _queryResults = results;
        _isLoading = false;
      });
    } catch (e) {
      setState(() => _isLoading = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error: $e')),
      );
    }
  }

  Future<void> _loadTables() async {
    try {
      final tables = await DatabaseService.getTables();
      setState(() => _tables = tables);
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error loading tables: $e')),
      );
    }
  }

  @override
  void initState() {
    super.initState();
    _loadTables();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Database Operations')),
      body: Column(
        children: [
          ElevatedButton(
            onPressed: _isLoading ? null : _runQuery,
            child: Text('Run Test Query'),
          ),
          
          if (_isLoading) CircularProgressIndicator(),
          
          if (_queryResults.isNotEmpty)
            Expanded(
              child: ListView.builder(
                itemCount: _queryResults.length,
                itemBuilder: (context, index) {
                  final row = _queryResults[index];
                  return ListTile(
                    title: Text(row.toString()),
                  );
                },
              ),
            ),
            
          Text('Available Tables: ${_tables.length}'),
          Expanded(
            child: ListView.builder(
              itemCount: _tables.length,
              itemBuilder: (context, index) {
                return ListTile(
                  title: Text(_tables[index]),
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}
```

---

## 🔍 Product Search

```dart
class ProductSearchService {
  static Future<List<Map<String, dynamic>>> searchProducts({
    required String query,
    int limit = 10,
  }) async {
    final response = await http.post(
      Uri.parse('${SMLGOAPIService.baseUrl}/search'),
      headers: SMLGOAPIService.headers,
      body: jsonEncode({
        'query': query,
        'limit': limit,
      }),
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      if (data['success']) {
        return List<Map<String, dynamic>>.from(data['data']['data'] ?? []);
      } else {
        throw Exception(data['error'] ?? 'Search failed');
      }
    } else {
      throw Exception('HTTP Error: ${response.statusCode}');
    }
  }
}

// Product Search Widget
class ProductSearchWidget extends StatefulWidget {
  @override
  _ProductSearchWidgetState createState() => _ProductSearchWidgetState();
}

class _ProductSearchWidgetState extends State<ProductSearchWidget> {
  final TextEditingController _searchController = TextEditingController();
  List<Map<String, dynamic>> _products = [];
  bool _isLoading = false;

  Future<void> _searchProducts() async {
    final query = _searchController.text.trim();
    if (query.isEmpty) return;

    setState(() => _isLoading = true);

    try {
      final products = await ProductSearchService.searchProducts(
        query: query,
        limit: 10,
      );
      setState(() {
        _products = products;
        _isLoading = false;
      });
    } catch (e) {
      setState(() => _isLoading = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Search error: $e')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Product Search')),
      body: Padding(
        padding: EdgeInsets.all(16.0),
        child: Column(
          children: [
            TextField(
              controller: _searchController,
              decoration: InputDecoration(
                labelText: 'Search Products',
                hintText: 'Enter product name or code',
                border: OutlineInputBorder(),
                suffixIcon: IconButton(
                  icon: Icon(Icons.search),
                  onPressed: _searchProducts,
                ),
              ),
              onSubmitted: (_) => _searchProducts(),
            ),
            SizedBox(height: 16),
            
            if (_isLoading)
              CircularProgressIndicator()
            else if (_products.isNotEmpty)
              Expanded(
                child: ListView.builder(
                  itemCount: _products.length,
                  itemBuilder: (context, index) {
                    final product = _products[index];
                    return Card(
                      child: ListTile(
                        title: Text(product['product_name'] ?? 'Unknown'),
                        subtitle: Text('Code: ${product['product_code'] ?? 'N/A'}'),
                        trailing: Text('\$${product['price'] ?? '0.00'}'),
                      ),
                    );
                  },
                ),
              ),
          ],
        ),
      ),
    );
  }
}
```

---

## 🖼️ Image Proxy

```dart
class ImageProxyService {
  static String getProxyImageUrl(String originalUrl) {
    final encodedUrl = Uri.encodeComponent(originalUrl);
    return '${SMLGOAPIService.baseUrl}/imgproxy?url=$encodedUrl';
  }
}

// Usage in Widget
class ImageProxyWidget extends StatelessWidget {
  final String imageUrl;
  
  const ImageProxyWidget({Key? key, required this.imageUrl}) : super(key: key);
  
  @override
  Widget build(BuildContext context) {
    return Image.network(
      ImageProxyService.getProxyImageUrl(imageUrl),
      loadingBuilder: (context, child, loadingProgress) {
        if (loadingProgress == null) return child;
        return CircularProgressIndicator();
      },
      errorBuilder: (context, error, stackTrace) {
        return Icon(Icons.error);
      },
    );
  }
}
```

---

## 🩺 Health Check

```dart
class HealthService {
  static Future<Map<String, dynamic>> checkHealth() async {
    final response = await http.get(
      Uri.parse('${SMLGOAPIService.baseUrl}/health'),
      headers: {'Accept': 'application/json'},
    );

    if (response.statusCode == 200) {
      return jsonDecode(response.body);
    } else {
      throw Exception('Health check failed: ${response.statusCode}');
    }
  }
}

// Health Check Widget
class HealthCheckWidget extends StatefulWidget {
  @override
  _HealthCheckWidgetState createState() => _HealthCheckWidgetState();
}

class _HealthCheckWidgetState extends State<HealthCheckWidget> {
  Map<String, dynamic>? _healthData;
  bool _isLoading = false;

  Future<void> _checkHealth() async {
    setState(() => _isLoading = true);
    
    try {
      final health = await HealthService.checkHealth();
      setState(() {
        _healthData = health;
        _isLoading = false;
      });
    } catch (e) {
      setState(() => _isLoading = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Health check failed: $e')),
      );
    }
  }

  @override
  void initState() {
    super.initState();
    _checkHealth();
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: EdgeInsets.all(16.0),
        child: Column(
          children: [
            Text('API Health Status', style: Theme.of(context).textTheme.titleLarge),
            SizedBox(height: 8),
            
            if (_isLoading)
              CircularProgressIndicator()
            else if (_healthData != null)
              Column(
                children: [
                  Text('Status: ${_healthData!['status']}'),
                  Text('Database: ${_healthData!['database']}'),
                  Text('Version: ${_healthData!['version']}'),
                  Text('Time: ${_healthData!['timestamp']}'),
                ],
              ),
              
            ElevatedButton(
              onPressed: _checkHealth,
              child: Text('Refresh'),
            ),
          ],
        ),
      ),
    );
  }
}
```

---

## 💡 Best Practices สำหรับ Flutter

### 1. Error Handling
```dart
class APIException implements Exception {
  final String message;
  final int? statusCode;
  
  APIException(this.message, [this.statusCode]);
  
  @override
  String toString() => 'APIException: $message (Code: $statusCode)';
}

// Wrapper for all API calls
static Future<T> handleAPICall<T>(Future<T> Function() apiCall) async {
  try {
    return await apiCall();
  } on SocketException {
    throw APIException('No internet connection');
  } on TimeoutException {
    throw APIException('Request timeout');
  } on FormatException {
    throw APIException('Invalid response format');
  } catch (e) {
    throw APIException('Unknown error: $e');
  }
}
```

### 2. Loading States
```dart
class LoadingState<T> {
  final bool isLoading;
  final T? data;
  final String? error;
  
  const LoadingState({
    this.isLoading = false,
    this.data,
    this.error,
  });
  
  LoadingState<T> loading() => LoadingState(isLoading: true);
  LoadingState<T> success(T data) => LoadingState(data: data);
  LoadingState<T> failure(String error) => LoadingState(error: error);
}
```

### 3. Caching
```dart
import 'package:shared_preferences/shared_preferences.dart';

class CacheService {
  static const Duration cacheDuration = Duration(hours: 24);
  
  static Future<void> cacheData(String key, String data) async {
    final prefs = await SharedPreferences.getInstance();
    final cacheData = {
      'data': data,
      'timestamp': DateTime.now().millisecondsSinceEpoch,
    };
    await prefs.setString(key, jsonEncode(cacheData));
  }
  
  static Future<String?> getCachedData(String key) async {
    final prefs = await SharedPreferences.getInstance();
    final cached = prefs.getString(key);
    
    if (cached != null) {
      final cacheData = jsonDecode(cached);
      final timestamp = DateTime.fromMillisecondsSinceEpoch(cacheData['timestamp']);
      
      if (DateTime.now().difference(timestamp) < cacheDuration) {
        return cacheData['data'];
      }
    }
    
    return null;
  }
}
```

---

## 📋 API Summary

### ✅ All Available Endpoints

| Endpoint | Method | Purpose | Status |
|----------|--------|---------|--------|
| `/` | GET | API information | ✅ Working |
| `/health` | GET | Health check | ✅ Working |
| `/get/provinces` | POST | Get all provinces | ✅ Working |
| `/get/amphures` | POST | Get amphures by province | ✅ Working |
| `/get/tambons` | POST | Get tambons by amphure | ✅ Working |
| `/get/findbyzipcode` | POST | Find location by zipcode | ✅ Working |
| `/select` | POST | Execute SELECT query | ✅ Working |
| `/command` | POST | Execute any SQL command | ✅ Working |
| `/search` | POST | Product search | ✅ Working |
| `/imgproxy` | GET | Image proxy with cache | ✅ Working |
| `/api/tables` | GET | List database tables | ✅ Working |
| `/guide` | GET | Complete API guide | ✅ Working |

### 🎯 แนะนำการใช้งาน

1. **สำหรับ Address Forms**: ใช้ `/get/findbyzipcode` (เร็วที่สุด)
2. **สำหรับ Hierarchical Selection**: ใช้ `/get/provinces` → `/get/amphures` → `/get/tambons`
3. **สำหรับ Database Operations**: ใช้ `/select` และ `/command`
4. **สำหรับ Product Search**: ใช้ `/search`
5. **สำหรับ Monitoring**: ใช้ `/health` และ `/api/tables`

### 🚀 Performance
- **Find by Zipcode**: < 25ms
- **Provinces**: < 50ms (77 provinces)
- **Database Queries**: Depends on query complexity
- **All APIs**: CORS enabled, JSON response

---

**สำหรับ Production**: เพิ่ม authentication, rate limiting, HTTPS, และ error monitoring

### 1. ตรวจสอบสถานะ API
```bash
curl http://localhost:8008/health
```

### 2. ค้นหาที่อยู่จากรหัสไปรษณีย์ (แนะนำ)
```bash
curl -X POST http://localhost:8008/get/findbyzipcode \
  -H "Content-Type: application/json" \
  -d '{"zip_code": 10200}'
```

### 3. ดึงรายชื่อจังหวัดทั้งหมด
```bash
curl -X POST http://localhost:8008/get/provinces \
  -H "Content-Type: application/json" \
  -d '{}'
```

---

## 📍 ข้อมูลการปกครองไทย (Thai Administrative Data)

### 🌟 `/get/findbyzipcode` - ค้นหาจากรหัสไปรษณีย์ (เร็วที่สุด)

**Method:** POST  
**Performance:** < 25ms  
**Use Case:** เหมาะสำหรับ address validation, auto-complete

**Request:**
```json
{
  "zip_code": 10200
}
```

**Response:**
```json
{
  "success": true,
  "message": "Found 12 locations for zip code 10200",
  "data": [
    {
      "province": {
        "id": 1,
        "name_th": "กรุงเทพมหานคร",
        "name_en": "Bangkok"
      },
      "amphure": {
        "id": 1001,
        "name_th": "เขตพระนคร",
        "name_en": "Khet Phra Nakhon"
      },
      "tambon": {
        "id": 100101,
        "name_th": "พระบรมมหาราชวัง",
        "name_en": "Phra Borom Maha Ratchawang",
        "zip_code": 10200
      }
    }
  ]
}
```

### 📋 `/get/provinces` - รายชื่อจังหวัด

**Method:** POST  
**Request:** `{}`

**Response:**
```json
{
  "success": true,
  "message": "Retrieved 77 provinces successfully",
  "data": [
    {
      "id": 1,
      "name_th": "กรุงเทพมหานคร",
      "name_en": "Bangkok"
    },
    {
      "id": 2,
      "name_th": "สมุทรปราการ",
      "name_en": "Samut Prakan"
    }
  ]
}
```

### 🏘️ `/get/amphures` - รายชื่ออำเภอ

**Method:** POST  
**Request:**
```json
{
  "province_id": 1
}
```

### 🏡 `/get/tambons` - รายชื่อตำบล

**Method:** POST  
**Request:**
```json
{
  "amphure_id": 1001,
  "province_id": 1
}
```

---

## 💻 สำหรับ Frontend Developers

### JavaScript/React Examples

#### 1. Basic API Usage
```javascript
const API_BASE = 'http://localhost:8008';

// ตรวจสอบสถานะ API
async function checkHealth() {
  const response = await fetch(API_BASE + '/health');
  const data = await response.json();
  console.log('API Status:', data.status);
  return data;
}

// ค้นหาที่อยู่จากรหัสไปรษณีย์
async function findLocationByZipCode(zipCode) {
  const response = await fetch(API_BASE + '/get/findbyzipcode', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ zip_code: zipCode })
  });
  const data = await response.json();
  return data;
}

// ค้นหาสินค้า
async function searchProducts(query, limit = 10) {
  const response = await fetch(API_BASE + '/search', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query: query, limit: limit })
  });
  const data = await response.json();
  return data;
}

// วิธีใช้งาน
checkHealth().then(console.log);
findLocationByZipCode(10200).then(console.log);
searchProducts('laptop', 5).then(console.log);
```

#### 2. React Component Example
```jsx
import React, { useState, useEffect } from 'react';

const AddressForm = () => {
  const [zipCode, setZipCode] = useState('');
  const [locations, setLocations] = useState([]);
  const [loading, setLoading] = useState(false);
  const API_BASE = 'http://localhost:8008';

  const searchByZipCode = async (zip) => {
    if (!zip || zip.length < 5) return;
    setLoading(true);
    try {
      const response = await fetch(API_BASE + '/get/findbyzipcode', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ zip_code: parseInt(zip) })
      });
      const data = await response.json();
      if (data.success) {
        setLocations(data.data);
      } else {
        setLocations([]);
      }
    } catch (error) {
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    searchByZipCode(zipCode);
  }, [zipCode]);

  return (
    <div>
      <input
        type="text"
        placeholder="รหัสไปรษณีย์ (เช่น 10200)"
        value={zipCode}
        onChange={(e) => setZipCode(e.target.value)}
      />
      {loading && <p>กำลังค้นหา...</p>}
      {locations.map((location, index) => (
        <div key={index}>
          <p>ตำบล: {location.tambon.name_th}</p>
          <p>อำเภอ: {location.amphure.name_th}</p>
          <p>จังหวัด: {location.province.name_th}</p>
        </div>
      ))}
    </div>
  );
};

export default AddressForm;
```

#### 3. Error Handling
```javascript
// วิธีจัดการ Error ที่ถูกต้อง
async function apiCall(endpoint, data = {}) {
  try {
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    });
    const result = await response.json();
    if (result.success) {
      return { success: true, data: result.data };
    } else {
      return { success: false, error: result.error || 'Unknown error' };
    }
  } catch (error) {
    return { success: false, error: error.message };
  }
}
```

#### 4. Data Caching
```javascript
// การ Cache ข้อมูลใน localStorage
const CACHE_KEY = 'thai_provinces';
const CACHE_DURATION = 24 * 60 * 60 * 1000; // 24 ชั่วโมง

async function getProvinces() {
  const cached = localStorage.getItem(CACHE_KEY);
  if (cached) {
    const { data, timestamp } = JSON.parse(cached);
    if (Date.now() - timestamp < CACHE_DURATION) {
      return data;
    }
  }
  const response = await fetch('/get/provinces', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({})
  });
  const result = await response.json();
  localStorage.setItem(CACHE_KEY, JSON.stringify({
    data: result.data,
    timestamp: Date.now()
  }));
  return result.data;
}
```

---

## 🎯 แนวทางการใช้งานข้อมูลการปกครองไทย

### วิธีที่ 1: Hierarchical Selection (เลือกทีละขั้น)
**เหมาะสำหรับ:** Form ที่ให้ user เลือกทีละขั้น

**ขั้นตอน:**
1. เรียก `/get/provinces` เพื่อดึงรายชื่อจังหวัดทั้งหมด
2. เมื่อ user เลือกจังหวัด เรียก `/get/amphures` พร้อม `province_id`
3. เมื่อ user เลือกอำเภอ เรียก `/get/tambons` พร้อม `amphure_id` และ `province_id`

**ข้อดี:**
- UX ที่ชัดเจน
- ข้อมูลไม่ซ้ำซ้อน
- ใช้ bandwidth น้อย

**ข้อเสีย:**
- ต้องรอ user เลือกทีละขั้น
- หลาย API calls

### วิธีที่ 2: Zip Code Search (ค้นหาจากรหัสไปรษณีย์)
**เหมาะสำหรับ:** Auto-complete และ validation

**ขั้นตอน:**
1. User ป้อนรหัสไปรษณีย์
2. เรียก `/get/findbyzipcode` พร้อม `zip_code`
3. ได้ข้อมูลครบทั้งจังหวัด อำเภอ ตำบล ในครั้งเดียว

**ข้อดี:**
- เร็วมาก (< 25ms)
- ได้ข้อมูลครบในครั้งเดียว
- เหมาะสำหรับ validation

**ข้อเสีย:**
- User ต้องรู้รหัสไปรษณีย์
- อาจได้หลาย location สำหรับ zip code เดียว

### 💡 คำแนะนำ
**ใช้ทั้งสองวิธีร่วมกัน:** zipcode สำหรับ quick search, hierarchical สำหรับ manual selection

---

## 🗄️ Database Operations

### `/command` - Universal SQL Execution
**Method:** POST  
**Purpose:** Execute any SQL command (CREATE, INSERT, UPDATE, DELETE, etc.)

**Request:**
```json
{
  "query": "CREATE TABLE IF NOT EXISTS test (id UInt32, name String, price Float64) ENGINE = MergeTree() ORDER BY id"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Command executed successfully",
  "result": {
    "query": "CREATE TABLE...",
    "rows_affected": 0,
    "status": "success"
  },
  "command": "CREATE TABLE...",
  "duration_ms": 150.5
}
```

### `/select` - Data Retrieval
**Method:** POST  
**Purpose:** Execute SELECT queries

**Request:**
```json
{
  "query": "SELECT * FROM test ORDER BY id LIMIT 10"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Query executed successfully, 10 rows returned",
  "data": [
    {"id": 1, "name": "Product 1", "price": 99.99}
  ],
  "query": "SELECT * FROM test...",
  "row_count": 10,
  "duration_ms": 45.2
}
```

---

## 🔍 Product Search

### `/search` - Multi-step Product Search
**Method:** POST  
**Search Priority:** Code → Name → Vector Similarity

**Request:**
```json
{
  "query": "laptop gaming",
  "limit": 5
}
```

**Response:**
```json
{
  "success": true,
  "message": "Search completed successfully",
  "data": [
    {
      "product_code": "LAP001",
      "product_name": "Gaming Laptop RTX 4080",
      "price": 1299.99,
      "category": "Electronics",
      "search_step": 1,
      "relevance_score": 1.0
    }
  ],
  "metadata": {
    "query": "laptop gaming",
    "total_found": 1,
    "search_steps": ["code_search", "name_search", "vector_search"],
    "duration_ms": 156.7
  }
}
```

---

## 🖼️ Image Proxy

### `/imgproxy` - Image Proxy & Caching
**Method:** GET  
**Purpose:** Proxy external images with caching

**Request:**
```
GET /imgproxy?url=https://example.com/image.jpg
```

**Features:**
- Image caching for performance
- CORS headers for frontend use
- Support for various image formats

---

## 📊 Health Check & System

### `/health` - API Health Status
**Method:** GET

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-06-17T05:10:16.7356603+07:00",
  "version": "25.5.1.2782",
  "database": "connected"
}
```

### `/api/tables` - Database Tables
**Method:** GET  
**Purpose:** List all available database tables

---

## 💡 Best Practices สำหรับ Frontend

1. **ใช้ async/await** แทน .then() สำหรับการอ่านโค้ดที่ง่ายขึ้น
2. **ตรวจสอบ response.success** ก่อนใช้ข้อมูล
3. **ใช้ try-catch** เพื่อจัดการ network errors
4. **Cache ข้อมูล** ที่ไม่เปลี่ยนแปลงบ่อย (เช่น รายชื่อจังหวัด)
5. **ใช้ debounce** สำหรับ search ที่ user พิมพ์
6. **แสดง loading state** ขณะรอข้อมูล
7. **Validate input** ก่อนส่ง request
8. **ใช้ environment variables** สำหรับ API URL

---

## 🚨 Troubleshooting

### ปัญหาที่พบบ่อย
- **Connection errors:** ตรวจสอบว่า server รันที่ port 8008
- **CORS issues:** API รองรับ CORS สำหรับ localhost และ * origins
- **Query errors:** ตรวจสอบ SQL syntax และชื่อ table/column
- **Performance issues:** ดู duration_ms ใน response

### Logs
ตรวจสอบ server console สำหรับข้อมูล error แบบละเอียด

---

## 🎮 Testing Examples

### cURL Examples
```bash
# Health Check
curl http://localhost:8008/health

# Find by Zip Code
curl -X POST http://localhost:8008/get/findbyzipcode \
  -H "Content-Type: application/json" \
  -d '{"zip_code": 10200}'

# Get Provinces
curl -X POST http://localhost:8008/get/provinces \
  -H "Content-Type: application/json" \
  -d '{}'

# Search Products
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{"query": "laptop", "limit": 5}'
```

### JavaScript Fetch Examples
```javascript
// Health Check
fetch('http://localhost:8008/health')
  .then(r => r.json())
  .then(console.log);

// Find by Zip Code
fetch('http://localhost:8008/get/findbyzipcode', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ zip_code: 10200 })
}).then(r => r.json()).then(console.log);
```

---

## 📝 Notes

- API ออกแบบให้ใช้งานง่ายสำหรับ frontend developers
- รองรับ CORS สำหรับการพัฒนา local
- ข้อมูลการปกครองไทยครบถ้วน 77 จังหวัด
- Performance optimized สำหรับการค้นหาจากรหัสไปรษณีย์
- Compatible กับ React, Vue, Angular และ vanilla JavaScript

**สำหรับ Production:** ควรเพิ่ม authentication, rate limiting, และ HTTPS
