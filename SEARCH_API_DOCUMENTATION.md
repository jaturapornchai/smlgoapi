# API Documentation: /search Endpoint

## Overview
The `/search` endpoint provides intelligent product search capabilities using a multi-step algorithm that prioritizes results based on relevance and search method.

## Endpoint Details

**URL:** `POST /search`  
**Content-Type:** `application/json`  
**Response:** JSON

## Search Algorithm

The search uses a 3-step priority system:

### 1. Code Search (Priority 1 - Highest)
- **Method**: Full text search on product codes
- **Use Case**: Exact or partial product code matching
- **Score**: 1.0 for matches
- **Example**: Searching "07-1151" finds products with codes containing "07-1151"

### 2. Name Search (Priority 2 - Medium)  
- **Method**: Full text search on product names
- **Use Case**: Product name matching
- **Score**: 0.8 for matches
- **Example**: Searching "water" finds products with "water" in their names

### 3. Vector Search (Priority 3 - Lowest)
- **Method**: TF-IDF semantic similarity
- **Use Case**: Fuzzy matching and related products
- **Score**: Variable (0.01-1.0) based on similarity
- **Example**: Searching "beverage" finds drinks and related products

## Request Format

```json
{
  "query": "search_term",
  "limit": 10,
  "offset": 0
}
```

### Parameters

| Parameter | Type   | Required | Default | Description |
|-----------|--------|----------|---------|-------------|
| `query`   | string | Yes      | -       | Search term (product code, name, or keywords) |
| `limit`   | int    | No       | 10      | Maximum number of results (1-100) |
| `offset`  | int    | No       | 0       | Pagination offset |

## Response Format

```json
{
  "success": true,
  "message": "Search completed successfully",
  "data": {
    "data": [
      {
        "id": "07-1151",
        "name": "น่ารัก 300มล./แชมพูเด็ก/สีชมพู",
        "similarity_score": 1.0,
        "code": "07-1151",
        "balance_qty": -1,
        "price": 100,
        "supplier_code": "",
        "unit": "ชิ้น",
        "img_url": "",
        "search_priority": 1
      }
    ],
    "total_count": 1,
    "query": "07-1151",
    "duration_ms": 724.2748
  }
}
```

### Response Fields

#### Main Response
| Field     | Type   | Description |
|-----------|--------|-------------|
| `success` | bool   | Request success status |
| `message` | string | Status message |
| `data`    | object | Search results data |

#### Data Object
| Field         | Type  | Description |
|---------------|-------|-------------|
| `data`        | array | Array of search results |
| `total_count` | int   | Total number of results found |
| `query`       | string| Original search query |
| `duration_ms` | float | Search processing time in milliseconds |

#### Search Result Object
| Field             | Type   | Description |
|-------------------|--------|-------------|
| `id`              | string | Product ID (same as code) |
| `name`            | string | Product name |
| `similarity_score`| float  | Relevance score (0.01-1.0) |
| `code`            | string | Product code |
| `balance_qty`     | float  | Current inventory quantity |
| `price`           | float  | Product price |
| `supplier_code`   | string | Supplier code (may be empty) |
| `unit`            | string | Unit of measurement |
| `img_url`         | string | Product image URL (may be empty) |
| `search_priority` | int    | Search method priority (1=code, 2=name, 3=vector) |

## Usage Examples

### 1. Search by Product Code

**Request:**
```bash
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "07-1151",
    "limit": 5
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Search completed successfully", 
  "data": {
    "data": [
      {
        "id": "07-1151",
        "name": "น่ารัก 300มล./แชมพูเด็ก/สีชมพู",
        "similarity_score": 1.0,
        "code": "07-1151",
        "balance_qty": -1,
        "price": 100,
        "supplier_code": "",
        "unit": "ชิ้น",
        "img_url": "",
        "search_priority": 1
      }
    ],
    "total_count": 1,
    "query": "07-1151",
    "duration_ms": 724.2748
  }
}
```

### 2. Search by Product Name

**Request:**
```bash
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "water",
    "limit": 3
  }'
```

### 3. Partial Code Search

**Request:**
```bash
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "1151",
    "limit": 10
  }'
```

### 4. Semantic Search

**Request:**
```bash
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "shampoo",
    "limit": 5
  }'
```

### 5. Pagination

**Request:**
```bash
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "item",
    "limit": 10,
    "offset": 20
  }'
```

## PowerShell Examples

### Basic Search
```powershell
Invoke-RestMethod -Uri "http://localhost:8008/search" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{"query": "07-1151", "limit": 5}'
```

### Get Detailed JSON Response
```powershell
$response = Invoke-RestMethod -Uri "http://localhost:8008/search" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{"query": "water", "limit": 3}'

$response.data.data[0] | ConvertTo-Json -Depth 10
```

## Response Status Codes

| Code | Status | Description |
|------|--------|-------------|
| 200  | OK     | Search completed successfully |
| 400  | Bad Request | Invalid request format or missing required fields |
| 500  | Internal Server Error | Server error during search processing |

## Error Response Format

```json
{
  "success": false,
  "error": "Query field is required in params JSON"
}
```

## Performance Notes

- **Response Time**: Typically 500-1000ms depending on query complexity
- **Limit**: Maximum 100 results per request
- **Caching**: No client-side caching implemented
- **Rate Limiting**: No rate limiting currently implemented

## Search Tips

### For Best Results:

1. **Product Code Search**: Use exact or partial product codes for highest accuracy
   - Example: `"07-1151"`, `"1151"`, `"07-"`

2. **Product Name Search**: Use specific product names or keywords
   - Example: `"water"`, `"shampoo"`, `"น้ำ"`

3. **Semantic Search**: Use general terms for discovering related products
   - Example: `"beverage"`, `"cleaning"`, `"baby care"`

### Result Prioritization:

Results are automatically sorted by:
1. **Search Priority**: Code (1) > Name (2) > Vector (3)
2. **Similarity Score**: Higher scores appear first
3. **Image Availability**: Products with images are prioritized when scores are close
4. **Alphabetical Order**: Final fallback sorting

## Integration Notes

### Frontend Integration
```javascript
// JavaScript example
const searchProducts = async (query, limit = 10, offset = 0) => {
  const response = await fetch('http://localhost:8008/search', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ query, limit, offset })
  });
  
  return await response.json();
};

// Usage
searchProducts('water', 5).then(results => {
  console.log(results.data.data);
});
```

### Mobile App Integration
The API is REST-compliant and can be easily integrated into mobile applications using standard HTTP libraries.

## Version Information

- **API Version**: 1.0
- **Last Updated**: June 16, 2025
- **Server**: SMLGOAPI
- **Database**: ClickHouse 25.5.1.2782

## Support

For technical support or questions about the search API, please contact the development team.
