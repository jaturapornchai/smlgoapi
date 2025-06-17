# SMLGOAPI - Flutter Integration Guide

## 🚀 Quick Start

### Dependencies
Add to your `pubspec.yaml`:
```yaml
dependencies:
  http: ^1.1.0
  flutter: 
    sdk: flutter

dev_dependencies:
  flutter_test:
    sdk: flutter
```

### Base Configuration
```dart
class ApiConfig {
  static const String baseUrl = 'http://localhost:8008';
  static const Map<String, String> headers = {
    'Content-Type': 'application/json',
  };
}
```

## 📊 API Endpoints Summary

| Endpoint | Method | Purpose | Status |
|----------|--------|---------|--------|
| `/` | GET | API information and endpoint list | ✅ Working |
| `/health` | GET | Health check | ✅ Working |
| `/get/provinces` | POST | Get all Thai provinces | ✅ Working |
| `/get/amphures` | POST | Get districts by province | ✅ Working |
| `/get/tambons` | POST | Get sub-districts by amphure | ✅ Working |
| `/get/findbyzipcode` | POST | Find location by postal code | ✅ Working |
| `/select` | POST | Execute SELECT queries | ✅ Working |
| `/command` | POST | Execute SQL commands | ✅ Working |
| `/search` | POST | Product search with vector similarity | ✅ Working |
| `/imgproxy` | GET | Image proxy with caching | ✅ Working |
| `/api/tables` | GET | List available database tables | ✅ Working |
| `/guide` | GET | API documentation | ✅ Working |

## 🏗️ Model Classes

### Thai Administrative Data Models
```dart
class Province {
  final int id;
  final String nameTh;
  final String nameEn;

  Province({
    required this.id,
    required this.nameTh,
    required this.nameEn,
  });

  factory Province.fromJson(Map<String, dynamic> json) {
    return Province(
      id: json['id'],
      nameTh: json['name_th'],
      nameEn: json['name_en'],
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'name_th': nameTh,
    'name_en': nameEn,
  };
}

class Amphure {
  final int id;
  final String nameTh;
  final String nameEn;
  final int provinceId;

  Amphure({
    required this.id,
    required this.nameTh,
    required this.nameEn,
    required this.provinceId,
  });

  factory Amphure.fromJson(Map<String, dynamic> json) {
    return Amphure(
      id: json['id'],
      nameTh: json['name_th'],
      nameEn: json['name_en'],
      provinceId: json['province_id'] ?? 0,
    );
  }
}

class Tambon {
  final int id;
  final String nameTh;
  final String nameEn;
  final int zipCode;
  final int amphureId;

  Tambon({
    required this.id,
    required this.nameTh,
    required this.nameEn,
    required this.zipCode,
    required this.amphureId,
  });

  factory Tambon.fromJson(Map<String, dynamic> json) {
    return Tambon(
      id: json['id'],
      nameTh: json['name_th'],
      nameEn: json['name_en'],
      zipCode: json['zip_code'],
      amphureId: json['amphure_id'] ?? 0,
    );
  }
}

class LocationData {
  final Province province;
  final Amphure amphure;
  final Tambon tambon;

  LocationData({
    required this.province,
    required this.amphure,
    required this.tambon,
  });

  factory LocationData.fromJson(Map<String, dynamic> json) {
    return LocationData(
      province: Province.fromJson(json['province']),
      amphure: Amphure.fromJson(json['amphure']),
      tambon: Tambon.fromJson(json['tambon']),
    );
  }
}
```

### Product Search Models
```dart
class Product {
  final String productCode;
  final String productName;
  final String category;
  final double price;
  final double relevanceScore;
  final int searchStep;

  Product({
    required this.productCode,
    required this.productName,
    required this.category,
    required this.price,
    required this.relevanceScore,
    required this.searchStep,
  });

  factory Product.fromJson(Map<String, dynamic> json) {
    return Product(
      productCode: json['product_code'] ?? '',
      productName: json['product_name'] ?? '',
      category: json['category'] ?? '',
      price: (json['price'] ?? 0).toDouble(),
      relevanceScore: (json['relevance_score'] ?? 0).toDouble(),
      searchStep: json['search_step'] ?? 0,
    );
  }
}

class SearchResponse {
  final bool success;
  final String message;
  final List<Product> data;
  final SearchMetadata? metadata;

  SearchResponse({
    required this.success,
    required this.message,
    required this.data,
    this.metadata,
  });

  factory SearchResponse.fromJson(Map<String, dynamic> json) {
    return SearchResponse(
      success: json['success'] ?? false,
      message: json['message'] ?? '',
      data: (json['data'] as List?)
          ?.map((item) => Product.fromJson(item))
          .toList() ?? [],
      metadata: json['metadata'] != null 
          ? SearchMetadata.fromJson(json['metadata']) 
          : null,
    );
  }
}

class SearchMetadata {
  final double durationMs;
  final String query;
  final List<String> searchSteps;
  final int totalFound;

  SearchMetadata({
    required this.durationMs,
    required this.query,
    required this.searchSteps,
    required this.totalFound,
  });

  factory SearchMetadata.fromJson(Map<String, dynamic> json) {
    return SearchMetadata(
      durationMs: (json['duration_ms'] ?? 0).toDouble(),
      query: json['query'] ?? '',
      searchSteps: List<String>.from(json['search_steps'] ?? []),
      totalFound: json['total_found'] ?? 0,
    );
  }
}
```

### API Response Models
```dart
class ApiResponse<T> {
  final bool success;
  final String message;
  final T? data;
  final String? error;

  ApiResponse({
    required this.success,
    required this.message,
    this.data,
    this.error,
  });

  factory ApiResponse.fromJson(Map<String, dynamic> json, T Function(dynamic)? fromJsonT) {
    return ApiResponse<T>(
      success: json['success'] ?? false,
      message: json['message'] ?? '',
      data: json['data'] != null && fromJsonT != null ? fromJsonT(json['data']) : null,
      error: json['error'],
    );
  }
}

class HealthStatus {
  final String status;
  final String timestamp;
  final String version;
  final String database;

  HealthStatus({
    required this.status,
    required this.timestamp,
    required this.version,
    required this.database,
  });

  factory HealthStatus.fromJson(Map<String, dynamic> json) {
    return HealthStatus(
      status: json['status'] ?? '',
      timestamp: json['timestamp'] ?? '',
      version: json['version'] ?? '',
      database: json['database'] ?? '',
    );
  }
}
```

## 🔧 Service Classes

### Main API Service
```dart
import 'dart:convert';
import 'package:http/http.dart' as http;

class SmlGoApiService {
  static const String _baseUrl = ApiConfig.baseUrl;
  static const Map<String, String> _headers = ApiConfig.headers;

  // Health Check
  static Future<HealthStatus?> checkHealth() async {
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/health'),
      );

      if (response.statusCode == 200) {
        final Map<String, dynamic> data = json.decode(response.body);
        return HealthStatus.fromJson(data);
      }
      return null;
    } catch (e) {
      print('Health check error: $e');
      return null;
    }
  }

  // Get API Info
  static Future<Map<String, dynamic>?> getApiInfo() async {
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/'),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('API info error: $e');
      return null;
    }
  }

  // Execute SELECT Query
  static Future<Map<String, dynamic>?> executeSelect(String query) async {
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/select'),
        headers: _headers,
        body: json.encode({'query': query}),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Select query error: $e');
      return null;
    }
  }

  // Execute Command Query
  static Future<Map<String, dynamic>?> executeCommand(String query) async {
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/command'),
        headers: _headers,
        body: json.encode({'query': query}),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Command query error: $e');
      return null;
    }
  }

  // Get Database Tables
  static Future<List<Map<String, dynamic>>?> getTables() async {
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/tables'),
      );

      if (response.statusCode == 200) {
        final Map<String, dynamic> data = json.decode(response.body);
        if (data['success'] == true) {
          return List<Map<String, dynamic>>.from(data['data']);
        }
      }
      return null;
    } catch (e) {
      print('Get tables error: $e');
      return null;
    }
  }

  // Image Proxy
  static String getImageProxyUrl(String imageUrl) {
    return '$_baseUrl/imgproxy?url=${Uri.encodeComponent(imageUrl)}';
  }
}
```

### Thai Administrative Data Service
```dart
class ThaiAdminService {
  static const String _baseUrl = ApiConfig.baseUrl;
  static const Map<String, String> _headers = ApiConfig.headers;

  // Get All Provinces
  static Future<List<Province>?> getProvinces() async {
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/get/provinces'),
        headers: _headers,
        body: json.encode({}),
      );

      if (response.statusCode == 200) {
        final Map<String, dynamic> data = json.decode(response.body);
        if (data['success'] == true) {
          return (data['data'] as List)
              .map((item) => Province.fromJson(item))
              .toList();
        }
      }
      return null;
    } catch (e) {
      print('Get provinces error: $e');
      return null;
    }
  }

  // Get Amphures by Province
  static Future<List<Amphure>?> getAmphures(int provinceId) async {
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/get/amphures'),
        headers: _headers,
        body: json.encode({'province_id': provinceId}),
      );

      if (response.statusCode == 200) {
        final Map<String, dynamic> data = json.decode(response.body);
        if (data['success'] == true) {
          return (data['data'] as List)
              .map((item) => Amphure.fromJson(item))
              .toList();
        }
      }
      return null;
    } catch (e) {
      print('Get amphures error: $e');
      return null;
    }
  }

  // Get Tambons by Amphure
  static Future<List<Tambon>?> getTambons(int provinceId, int amphureId) async {
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/get/tambons'),
        headers: _headers,
        body: json.encode({
          'province_id': provinceId,
          'amphure_id': amphureId,
        }),
      );

      if (response.statusCode == 200) {
        final Map<String, dynamic> data = json.decode(response.body);
        if (data['success'] == true) {
          return (data['data'] as List)
              .map((item) => Tambon.fromJson(item))
              .toList();
        }
      }
      return null;
    } catch (e) {
      print('Get tambons error: $e');
      return null;
    }
  }

  // Find Location by Zip Code
  static Future<List<LocationData>?> findByZipCode(int zipCode) async {
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/get/findbyzipcode'),
        headers: _headers,
        body: json.encode({'zip_code': zipCode}),
      );

      if (response.statusCode == 200) {
        final Map<String, dynamic> data = json.decode(response.body);
        if (data['success'] == true) {
          return (data['data'] as List)
              .map((item) => LocationData.fromJson(item))
              .toList();
        }
      }
      return null;
    } catch (e) {
      print('Find by zip code error: $e');
      return null;
    }
  }
}
```

### Product Search Service
```dart
class ProductSearchService {
  static const String _baseUrl = ApiConfig.baseUrl;
  static const Map<String, String> _headers = ApiConfig.headers;

  // Search Products
  static Future<SearchResponse?> searchProducts(String query, {int limit = 10}) async {
    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/search'),
        headers: _headers,
        body: json.encode({
          'query': query,
          'limit': limit,
        }),
      );

      if (response.statusCode == 200) {
        final Map<String, dynamic> data = json.decode(response.body);
        return SearchResponse.fromJson(data);
      }
      return null;
    } catch (e) {
      print('Product search error: $e');
      return null;
    }
  }
}
```

## 🎨 Flutter Widgets

### Health Check Widget
```dart
class HealthCheckWidget extends StatefulWidget {
  @override
  _HealthCheckWidgetState createState() => _HealthCheckWidgetState();
}

class _HealthCheckWidgetState extends State<HealthCheckWidget> {
  HealthStatus? _healthStatus;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _checkHealth();
  }

  Future<void> _checkHealth() async {
    setState(() {
      _loading = true;
    });

    final health = await SmlGoApiService.checkHealth();
    
    setState(() {
      _healthStatus = health;
      _loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'API Health Status',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                IconButton(
                  icon: Icon(Icons.refresh),
                  onPressed: _checkHealth,
                ),
              ],
            ),
            SizedBox(height: 10),
            if (_loading)
              Center(child: CircularProgressIndicator())
            else if (_healthStatus != null)
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildStatusRow('Status', _healthStatus!.status),
                  _buildStatusRow('Database', _healthStatus!.database),
                  _buildStatusRow('Version', _healthStatus!.version),
                  _buildStatusRow('Timestamp', _healthStatus!.timestamp),
                ],
              )
            else
              Text(
                'Unable to connect to API',
                style: TextStyle(color: Colors.red),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildStatusRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4.0),
      child: Row(
        children: [
          Text('$label: ', style: FontWeight.bold),
          Text(value),
          if (label == 'Status' && value == 'healthy')
            Icon(Icons.check_circle, color: Colors.green, size: 16)
          else if (label == 'Status')
            Icon(Icons.error, color: Colors.red, size: 16),
        ],
      ),
    );
  }
}
```

### Address Form Widget
```dart
class AddressFormWidget extends StatefulWidget {
  final Function(LocationData?)? onLocationSelected;

  const AddressFormWidget({Key? key, this.onLocationSelected}) : super(key: key);

  @override
  _AddressFormWidgetState createState() => _AddressFormWidgetState();
}

class _AddressFormWidgetState extends State<AddressFormWidget> {
  final _zipCodeController = TextEditingController();
  
  Province? _selectedProvince;
  Amphure? _selectedAmphure;
  Tambon? _selectedTambon;
  
  List<Province> _provinces = [];
  List<Amphure> _amphures = [];
  List<Tambon> _tambons = [];
  List<LocationData> _zipCodeResults = [];
  
  bool _loadingProvinces = false;
  bool _loadingAmphures = false;
  bool _loadingTambons = false;
  bool _loadingZipCode = false;

  @override
  void initState() {
    super.initState();
    _loadProvinces();
    _zipCodeController.addListener(_onZipCodeChanged);
  }

  @override
  void dispose() {
    _zipCodeController.dispose();
    super.dispose();
  }

  Future<void> _loadProvinces() async {
    setState(() {
      _loadingProvinces = true;
    });

    final provinces = await ThaiAdminService.getProvinces();
    
    setState(() {
      _provinces = provinces ?? [];
      _loadingProvinces = false;
    });
  }

  Future<void> _loadAmphures() async {
    if (_selectedProvince == null) return;

    setState(() {
      _loadingAmphures = true;
      _amphures = [];
      _selectedAmphure = null;
      _tambons = [];
      _selectedTambon = null;
    });

    final amphures = await ThaiAdminService.getAmphures(_selectedProvince!.id);
    
    setState(() {
      _amphures = amphures ?? [];
      _loadingAmphures = false;
    });
  }

  Future<void> _loadTambons() async {
    if (_selectedProvince == null || _selectedAmphure == null) return;

    setState(() {
      _loadingTambons = true;
      _tambons = [];
      _selectedTambon = null;
    });

    final tambons = await ThaiAdminService.getTambons(
      _selectedProvince!.id,
      _selectedAmphure!.id,
    );
    
    setState(() {
      _tambons = tambons ?? [];
      _loadingTambons = false;
    });
  }

  void _onZipCodeChanged() {
    final zipCode = _zipCodeController.text;
    if (zipCode.length >= 5) {
      _searchByZipCode(int.tryParse(zipCode));
    } else {
      setState(() {
        _zipCodeResults = [];
      });
    }
  }

  Future<void> _searchByZipCode(int? zipCode) async {
    if (zipCode == null) return;

    setState(() {
      _loadingZipCode = true;
    });

    final results = await ThaiAdminService.findByZipCode(zipCode);
    
    setState(() {
      _zipCodeResults = results ?? [];
      _loadingZipCode = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Thai Address Form',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            SizedBox(height: 20),
            
            // Zip Code Search
            TextField(
              controller: _zipCodeController,
              decoration: InputDecoration(
                labelText: 'รหัสไปรษณีย์ (Zip Code)',
                hintText: 'เช่น 10200',
                suffixIcon: _loadingZipCode 
                  ? SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : null,
              ),
              keyboardType: TextInputType.number,
            ),
            
            if (_zipCodeResults.isNotEmpty) ...[
              SizedBox(height: 10),
              Text('ผลการค้นหาจากรหัสไปรษณีย์:'),
              ...(_zipCodeResults.map((location) => ListTile(
                title: Text('${location.tambon.nameTh}'),
                subtitle: Text('${location.amphure.nameTh}, ${location.province.nameTh}'),
                onTap: () {
                  widget.onLocationSelected?.call(location);
                },
              ))),
            ],
            
            SizedBox(height: 20),
            Divider(),
            SizedBox(height: 20),
            
            // Hierarchical Selection
            Text('หรือเลือกที่อยู่แบบทีละขั้น:'),
            SizedBox(height: 10),
            
            // Province Dropdown
            DropdownButtonFormField<Province>(
              value: _selectedProvince,
              decoration: InputDecoration(labelText: 'จังหวัด (Province)'),
              items: _provinces.map((province) {
                return DropdownMenuItem(
                  value: province,
                  child: Text(province.nameTh),
                );
              }).toList(),
              onChanged: (Province? value) {
                setState(() {
                  _selectedProvince = value;
                });
                _loadAmphures();
              },
              isExpanded: true,
            ),
            
            SizedBox(height: 10),
            
            // Amphure Dropdown
            DropdownButtonFormField<Amphure>(
              value: _selectedAmphure,
              decoration: InputDecoration(labelText: 'อำเภอ/เขต (District)'),
              items: _amphures.map((amphure) {
                return DropdownMenuItem(
                  value: amphure,
                  child: Text(amphure.nameTh),
                );
              }).toList(),
              onChanged: _selectedProvince == null ? null : (Amphure? value) {
                setState(() {
                  _selectedAmphure = value;
                });
                _loadTambons();
              },
              isExpanded: true,
            ),
            
            SizedBox(height: 10),
            
            // Tambon Dropdown
            DropdownButtonFormField<Tambon>(
              value: _selectedTambon,
              decoration: InputDecoration(labelText: 'ตำบล/แขวง (Sub-district)'),
              items: _tambons.map((tambon) {
                return DropdownMenuItem(
                  value: tambon,
                  child: Text('${tambon.nameTh} (${tambon.zipCode})'),
                );
              }).toList(),
              onChanged: _selectedAmphure == null ? null : (Tambon? value) {
                setState(() {
                  _selectedTambon = value;
                });
                if (value != null && _selectedProvince != null && _selectedAmphure != null) {
                  final locationData = LocationData(
                    province: _selectedProvince!,
                    amphure: _selectedAmphure!,
                    tambon: value,
                  );
                  widget.onLocationSelected?.call(locationData);
                }
              },
              isExpanded: true,
            ),
            
            if (_loadingProvinces || _loadingAmphures || _loadingTambons)
              Padding(
                padding: const EdgeInsets.all(8.0),
                child: LinearProgressIndicator(),
              ),
          ],
        ),
      ),
    );
  }
}
```

### Product Search Widget
```dart
class ProductSearchWidget extends StatefulWidget {
  @override
  _ProductSearchWidgetState createState() => _ProductSearchWidgetState();
}

class _ProductSearchWidgetState extends State<ProductSearchWidget> {
  final _searchController = TextEditingController();
  SearchResponse? _searchResponse;
  bool _loading = false;
  Timer? _debounceTimer;

  @override
  void initState() {
    super.initState();
    _searchController.addListener(_onSearchChanged);
  }

  @override
  void dispose() {
    _searchController.dispose();
    _debounceTimer?.cancel();
    super.dispose();
  }

  void _onSearchChanged() {
    _debounceTimer?.cancel();
    _debounceTimer = Timer(Duration(milliseconds: 500), () {
      final query = _searchController.text;
      if (query.isNotEmpty) {
        _searchProducts(query);
      } else {
        setState(() {
          _searchResponse = null;
        });
      }
    });
  }

  Future<void> _searchProducts(String query) async {
    setState(() {
      _loading = true;
    });

    final response = await ProductSearchService.searchProducts(query, limit: 20);
    
    setState(() {
      _searchResponse = response;
      _loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Product Search',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            SizedBox(height: 10),
            
            TextField(
              controller: _searchController,
              decoration: InputDecoration(
                labelText: 'ค้นหาสินค้า',
                hintText: 'เช่น laptop, ข้าว, เสื้อ',
                prefixIcon: Icon(Icons.search),
                suffixIcon: _loading 
                  ? SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : null,
              ),
            ),
            
            SizedBox(height: 20),
            
            if (_searchResponse != null) ...[
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    'ผลการค้นหา: ${_searchResponse!.data.length} รายการ',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  if (_searchResponse!.metadata != null)
                    Text(
                      '${_searchResponse!.metadata!.durationMs.toStringAsFixed(1)}ms',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                ],
              ),
              SizedBox(height: 10),
              
              Expanded(
                child: ListView.builder(
                  itemCount: _searchResponse!.data.length,
                  itemBuilder: (context, index) {
                    final product = _searchResponse!.data[index];
                    return Card(
                      child: ListTile(
                        leading: CircleAvatar(
                          child: Text('${product.searchStep}'),
                          backgroundColor: _getStepColor(product.searchStep),
                        ),
                        title: Text(product.productName),
                        subtitle: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('Code: ${product.productCode}'),
                            Text('Category: ${product.category}'),
                            Text('Score: ${product.relevanceScore.toStringAsFixed(2)}'),
                          ],
                        ),
                        trailing: Text(
                          '฿${product.price.toStringAsFixed(2)}',
                          style: Theme.of(context).textTheme.titleMedium,
                        ),
                        isThreeLine: true,
                      ),
                    );
                  },
                ),
              ),
              
              if (_searchResponse!.metadata != null) ...[
                SizedBox(height: 10),
                Text(
                  'Search Steps: ${_searchResponse!.metadata!.searchSteps.join(', ')}',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ],
            ] else if (!_loading && _searchController.text.isNotEmpty)
              Center(child: Text('ไม่พบสินค้าที่ตรงกับการค้นหา'))
            else if (_searchController.text.isEmpty)
              Center(child: Text('กรุณาพิมพ์คำค้นหา')),
          ],
        ),
      ),
    );
  }

  Color _getStepColor(int step) {
    switch (step) {
      case 1: return Colors.green;    // Code search
      case 2: return Colors.orange;   // Name search
      case 3: return Colors.blue;     // Vector search
      default: return Colors.grey;
    }
  }
}
```

### Database Query Widget
```dart
class DatabaseQueryWidget extends StatefulWidget {
  @override
  _DatabaseQueryWidgetState createState() => _DatabaseQueryWidgetState();
}

class _DatabaseQueryWidgetState extends State<DatabaseQueryWidget> with SingleTickerProviderStateMixin {
  final _queryController = TextEditingController();
  late TabController _tabController;
  
  Map<String, dynamic>? _queryResult;
  List<Map<String, dynamic>>? _tables;
  bool _loading = false;
  String _selectedQueryType = 'SELECT';

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _loadTables();
  }

  @override
  void dispose() {
    _queryController.dispose();
    _tabController.dispose();
    super.dispose();
  }

  Future<void> _loadTables() async {
    final tables = await SmlGoApiService.getTables();
    setState(() {
      _tables = tables;
    });
  }

  Future<void> _executeQuery() async {
    final query = _queryController.text.trim();
    if (query.isEmpty) return;

    setState(() {
      _loading = true;
      _queryResult = null;
    });

    Map<String, dynamic>? result;
    
    if (_selectedQueryType == 'SELECT') {
      result = await SmlGoApiService.executeSelect(query);
    } else {
      result = await SmlGoApiService.executeCommand(query);
    }
    
    setState(() {
      _queryResult = result;
      _loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          children: [
            TabBar(
              controller: _tabController,
              tabs: [
                Tab(text: 'Query'),
                Tab(text: 'Tables'),
              ],
            ),
            
            Expanded(
              child: TabBarView(
                controller: _tabController,
                children: [
                  // Query Tab
                  Column(
                    children: [
                      SizedBox(height: 20),
                      
                      Row(
                        children: [
                          Text('Query Type: '),
                          DropdownButton<String>(
                            value: _selectedQueryType,
                            items: ['SELECT', 'COMMAND'].map((type) {
                              return DropdownMenuItem(
                                value: type,
                                child: Text(type),
                              );
                            }).toList(),
                            onChanged: (value) {
                              setState(() {
                                _selectedQueryType = value!;
                              });
                            },
                          ),
                        ],
                      ),
                      
                      SizedBox(height: 10),
                      
                      TextField(
                        controller: _queryController,
                        decoration: InputDecoration(
                          labelText: 'SQL Query',
                          hintText: _selectedQueryType == 'SELECT' 
                            ? 'SELECT * FROM products LIMIT 10'
                            : 'CREATE TABLE test (id UInt32) ENGINE = MergeTree() ORDER BY id',
                          border: OutlineInputBorder(),
                        ),
                        maxLines: 3,
                      ),
                      
                      SizedBox(height: 10),
                      
                      ElevatedButton(
                        onPressed: _loading ? null : _executeQuery,
                        child: _loading 
                          ? CircularProgressIndicator() 
                          : Text('Execute Query'),
                      ),
                      
                      SizedBox(height: 20),
                      
                      if (_queryResult != null)
                        Expanded(
                          child: SingleChildScrollView(
                            child: Container(
                              width: double.infinity,
                              padding: EdgeInsets.all(12),
                              decoration: BoxDecoration(
                                border: Border.all(color: Colors.grey),
                                borderRadius: BorderRadius.circular(4),
                              ),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    'Result:',
                                    style: Theme.of(context).textTheme.titleMedium,
                                  ),
                                  SizedBox(height: 8),
                                  Text(
                                    JsonEncoder.withIndent('  ').convert(_queryResult),
                                    style: TextStyle(fontFamily: 'monospace'),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        ),
                    ],
                  ),
                  
                  // Tables Tab
                  Column(
                    children: [
                      SizedBox(height: 20),
                      Text(
                        'Available Tables',
                        style: Theme.of(context).textTheme.titleLarge,
                      ),
                      SizedBox(height: 10),
                      
                      if (_tables != null)
                        Expanded(
                          child: ListView.builder(
                            itemCount: _tables!.length,
                            itemBuilder: (context, index) {
                              final table = _tables![index];
                              return ListTile(
                                leading: Icon(Icons.table_chart),
                                title: Text(table['name'] ?? ''),
                                subtitle: Text('Engine: ${table['engine'] ?? ''}, Rows: ${table['rows'] ?? 0}'),
                                onTap: () {
                                  _queryController.text = 'SELECT * FROM ${table['name']} LIMIT 10';
                                  _tabController.animateTo(0);
                                },
                              );
                            },
                          ),
                        )
                      else
                        CircularProgressIndicator(),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
```

## 🎯 Complete Usage Examples

### Main App Integration
```dart
class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'SMLGOAPI Flutter Demo',
      theme: ThemeData(
        primarySwatch: Colors.blue,
        visualDensity: VisualDensity.adaptivePlatformDensity,
      ),
      home: MainScreen(),
    );
  }
}

class MainScreen extends StatefulWidget {
  @override
  _MainScreenState createState() => _MainScreenState();
}

class _MainScreenState extends State<MainScreen> with TickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 4, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('SMLGOAPI Flutter Demo'),
        bottom: TabBar(
          controller: _tabController,
          tabs: [
            Tab(icon: Icon(Icons.health_and_safety), text: 'Health'),
            Tab(icon: Icon(Icons.location_on), text: 'Address'),
            Tab(icon: Icon(Icons.search), text: 'Search'),
            Tab(icon: Icon(Icons.storage), text: 'Database'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          HealthCheckWidget(),
          AddressFormWidget(
            onLocationSelected: (location) {
              if (location != null) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text('Selected: ${location.tambon.nameTh}, ${location.amphure.nameTh}, ${location.province.nameTh}'),
                  ),
                );
              }
            },
          ),
          ProductSearchWidget(),
          DatabaseQueryWidget(),
        ],
      ),
    );
  }
}
```

## 📱 Best Practices

### Error Handling
```dart
class ApiErrorHandler {
  static void handleError(BuildContext context, dynamic error) {
    String message;
    
    if (error is http.ClientException) {
      message = 'เกิดข้อผิดพลาดในการเชื่อมต่อ';
    } else if (error is FormatException) {
      message = 'ข้อมูลที่ได้รับไม่ถูกต้อง';
    } else {
      message = 'เกิดข้อผิดพลาดที่ไม่คาดคิด: ${error.toString()}';
    }
    
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: Colors.red,
        action: SnackBarAction(
          label: 'ปิด',
          textColor: Colors.white,
          onPressed: () {},
        ),
      ),
    );
  }
}
```

### Caching
```dart
class CacheManager {
  static const String _provincesKey = 'cached_provinces';
  static const Duration _cacheExpiry = Duration(hours: 24);

  static Future<void> cacheProvinces(List<Province> provinces) async {
    final prefs = await SharedPreferences.getInstance();
    final cacheData = {
      'timestamp': DateTime.now().millisecondsSinceEpoch,
      'data': provinces.map((p) => p.toJson()).toList(),
    };
    await prefs.setString(_provincesKey, json.encode(cacheData));
  }

  static Future<List<Province>?> getCachedProvinces() async {
    final prefs = await SharedPreferences.getInstance();
    final cacheString = prefs.getString(_provincesKey);
    
    if (cacheString == null) return null;
    
    final cacheData = json.decode(cacheString);
    final timestamp = DateTime.fromMillisecondsSinceEpoch(cacheData['timestamp']);
    
    if (DateTime.now().difference(timestamp) > _cacheExpiry) {
      return null;
    }
    
    return (cacheData['data'] as List)
        .map((item) => Province.fromJson(item))
        .toList();
  }
}
```

### Loading States
```dart
class LoadingStateWidget extends StatelessWidget {
  final bool isLoading;
  final Widget child;
  final String loadingText;

  const LoadingStateWidget({
    Key? key,
    required this.isLoading,
    required this.child,
    this.loadingText = 'กำลังโหลด...',
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: [
        child,
        if (isLoading)
          Container(
            color: Colors.black.withOpacity(0.3),
            child: Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  CircularProgressIndicator(),
                  SizedBox(height: 16),
                  Text(
                    loadingText,
                    style: TextStyle(color: Colors.white),
                  ),
                ],
              ),
            ),
          ),
      ],
    );
  }
}
```

## 🔧 Performance Tips

1. **Use caching** for province/amphure/tambon data
2. **Implement debouncing** for search inputs
3. **Use pagination** for large result sets
4. **Cache images** using the image proxy endpoint
5. **Monitor API response times** using metadata
6. **Handle offline scenarios** gracefully

## 📞 Support

- API Base URL: `http://localhost:8008`
- Health Check: `http://localhost:8008/health`
- Documentation: `http://localhost:8008/guide`

## 📄 License

This API integration guide is provided as-is for Flutter development purposes.
