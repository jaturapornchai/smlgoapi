# API Documentation - Search Endpoint

## 🔍 Vector Search API

### Endpoint
```
POST /search
```

### Description
ค้นหาสินค้าโดยใช้ TF-IDF Vector Similarity Search ที่รองรับทั้งภาษาไทยและภาษาอังกฤษ

### Request Format

#### Headers
```http
Content-Type: application/json
```

#### Request Body
```json
{
  "query": "คำค้นหา",
  "limit": 10,
  "offset": 0
}
```

#### Parameters
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `query` | string | ✅ Yes | - | คำค้นหาสินค้า |
| `limit` | integer | ❌ No | 10 | จำนวนผลลัพธ์สูงสุดที่ต้องการ (max: 100) |
| `offset` | integer | ❌ No | 0 | เลขเริ่มต้นสำหรับ pagination |

### Response Format

#### Success Response (200 OK)
```json
{
  "success": true,
  "data": {
    "data": [
      {
        "id": "ITEM001",
        "name": "ชื่อสินค้า",
        "similarity_score": 0.8945,
        "img_url": "https://example.com/image.jpg",
        "metadata": {
          "code": "ITEM001",
          "unit": "PCS",
          "balance_qty": 100.50,
          "supplier_code": "SUP001",
          "price": 150.00,
          "img_url": "https://example.com/image.jpg"
        }
      }
    ],
    "total_count": 25,
    "query": "คำค้นหา",
    "duration_ms": 45.2
  },
  "message": "Search completed successfully"
}
```

#### Error Response (400 Bad Request)
```json
{
  "success": false,
  "error": "Query field is required in params JSON"
}
```

#### Error Response (500 Internal Server Error)
```json
{
  "success": false,
  "error": "Search failed: database connection error"
}
```

## 📋 Examples

### 1. Basic Search
```bash
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "motor",
    "limit": 5
  }'
```

### 2. Search with Pagination
```bash
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "สว่าน",
    "limit": 10,
    "offset": 20
  }'
```

### 3. Thai Language Search
```bash
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "เครื่องมือช่าง",
    "limit": 15
  }'
```

### 4. English Language Search
```bash
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "drill machine",
    "limit": 8
  }'
```

## 💻 Programming Examples

### JavaScript/Node.js
```javascript
async function searchProducts(query, limit = 10, offset = 0) {
  const response = await fetch('http://localhost:8008/search', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      query: query,
      limit: limit,
      offset: offset
    })
  });
  
  const result = await response.json();
  return result;
}

// Usage
searchProducts('motor', 5).then(results => {
  console.log('Found products:', results.data.data);
});
```

### Python
```python
import requests
import json

def search_products(query, limit=10, offset=0):
    url = "http://localhost:8008/search"
    
    payload = {
        "query": query,
        "limit": limit,
        "offset": offset
    }
    
    headers = {
        "Content-Type": "application/json"
    }
    
    response = requests.post(url, data=json.dumps(payload), headers=headers)
    return response.json()

# Usage
results = search_products("เครื่องมือ", limit=15)
print(f"Found {results['data']['total_count']} products")
for product in results['data']['data']:
    print(f"- {product['name']} (Score: {product['similarity_score']:.3f})")
```

### PHP
```php
<?php
function searchProducts($query, $limit = 10, $offset = 0) {
    $url = 'http://localhost:8008/search';
    
    $data = [
        'query' => $query,
        'limit' => $limit,
        'offset' => $offset
    ];
    
    $options = [
        'http' => [
            'header' => "Content-Type: application/json\r\n",
            'method' => 'POST',
            'content' => json_encode($data)
        ]
    ];
    
    $context = stream_context_create($options);
    $result = file_get_contents($url, false, $context);
    
    return json_decode($result, true);
}

// Usage
$results = searchProducts('สว่าน', 10);
echo "Found " . $results['data']['total_count'] . " products\n";
?>
```

### C#
```csharp
using System;
using System.Net.Http;
using System.Text;
using System.Threading.Tasks;
using Newtonsoft.Json;

public class SearchRequest
{
    public string query { get; set; }
    public int limit { get; set; } = 10;
    public int offset { get; set; } = 0;
}

public async Task<dynamic> SearchProducts(string query, int limit = 10, int offset = 0)
{
    using (var client = new HttpClient())
    {
        var request = new SearchRequest 
        { 
            query = query, 
            limit = limit, 
            offset = offset 
        };
        
        var json = JsonConvert.SerializeObject(request);
        var content = new StringContent(json, Encoding.UTF8, "application/json");
        
        var response = await client.PostAsync("http://localhost:8008/search", content);
        var responseString = await response.Content.ReadAsStringAsync();
        
        return JsonConvert.DeserializeObject(responseString);
    }
}

// Usage
var results = await SearchProducts("motor", 5);
Console.WriteLine($"Found {results.data.total_count} products");
```

## 🔧 Advanced Features

### 1. Image Prioritization
- สินค้าที่มีรูปภาพจะถูกจัดอันดับให้อยู่ด้านบนเสมอ
- ระบบจะดึงรูปภาพจากฐานข้อมูลอัตโนมัติ

### 2. Multi-language Support
- รองรับการค้นหาภาษาไทย (ใช้ GSE segmentation)
- รองรับการค้นหาภาษาอังกฤษ (ใช้ stemming)

### 3. Performance Optimization
- ใช้ 2-round vector search algorithm
- Caching และ optimization สำหรับประสิทธิภาพ

### 4. Similarity Scoring
- คะแนนความคล้ายคลึง (0.0 - 1.0)
- คะแนนสูงกว่า = เกี่ยวข้องมากกว่า

## 🚨 Error Handling

### Common Errors
1. **Empty Query**: ต้องระบุ query ที่ไม่ว่าง
2. **Invalid Limit**: limit ต้องไม่เกิน 100
3. **Database Error**: ปัญหาการเชื่อมต่อฐานข้อมูล

### Best Practices
1. ตรวจสอบ `success` field ก่อนใช้ข้อมูล
2. จัดการ error cases ที่เหมาะสม
3. ใช้ pagination สำหรับผลลัพธ์จำนวนมาก

## 📊 Response Fields

### Product Object
| Field | Type | Description |
|-------|------|-------------|
| `id` | string | รหัสสินค้า |
| `name` | string | ชื่อสินค้า |
| `similarity_score` | number | คะแนนความคล้ายคลึง (0.0-1.0) |
| `img_url` | string | URL รูปภาพสินค้า |
| `metadata.code` | string | รหัสสินค้า |
| `metadata.unit` | string | หน่วยนับ |
| `metadata.balance_qty` | number | จำนวนคงเหลือ |
| `metadata.supplier_code` | string | รหัสผู้จำหน่าย |
| `metadata.price` | number | ราคา |

### Response Metadata
| Field | Type | Description |
|-------|------|-------------|
| `total_count` | integer | จำนวนผลลัพธ์ทั้งหมด |
| `query` | string | คำค้นหาที่ใช้ |
| `duration_ms` | number | เวลาที่ใช้ในการค้นหา (มิลลิวินาที) |

## 🔗 Related Endpoints

- `GET /health` - ตรวจสอบสถานะระบบ
- `GET /imgproxy` - Proxy สำหรับรูปภาพ
- `GET /api/tables` - ดูตารางในฐานข้อมูล

---

📝 **Last Updated**: June 16, 2025  
🔗 **API Base URL**: `http://localhost:8008`
