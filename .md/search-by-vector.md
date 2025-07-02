# 🔍 `/search-by-vector` API Documentation

## Overview

Advanced product search endpoint that combines vector database (Weaviate) with PostgreSQL to deliver comprehensive and relevant search results. Supports both Thai and English languages with semantic search capabilities.

## Endpoint Details

**URL:** `POST /v1/search-by-vector`  
**Method:** `POST` only  
**Content-Type:** `application/json`  
**Base URL:** `http://localhost:8008`

---

## 📋 Request Format

### JSON Request Body

```json
{
  "query": "string",
  "limit": number,
  "offset": number
}
```

### Parameters

| Parameter | Type   | Required | Default | Max | Description                            |
| --------- | ------ | -------- | ------- | --- | -------------------------------------- |
| `query`   | string | ✅ Yes   | -       | -   | Search term (Thai/English supported)   |
| `limit`   | number | ❌ No    | 50      | 500 | Number of results to return            |
| `offset`  | number | ❌ No    | 0       | -   | Number of results to skip (pagination) |

---

## 🚀 Usage Examples

### Basic Search

```bash
curl -X POST "http://localhost:8008/v1/search-by-vector" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "โตโยต้า คอยล์"
  }'
```

### Search with Pagination

```bash
curl -X POST "http://localhost:8008/v1/search-by-vector" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "toyota coil",
    "limit": 20,
    "offset": 0
  }'
```

### PowerShell Example

```powershell
$body = @{
    query = "brake pad"
    limit = 10
    offset = 0
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8008/v1/search-by-vector" `
  -Method POST `
  -Body $body `
  -ContentType "application/json"
```

### JavaScript/Node.js

```javascript
const searchProducts = async (query, limit = 20, offset = 0) => {
  const response = await fetch("http://localhost:8008/v1/search-by-vector", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ query, limit, offset }),
  });

  return await response.json();
};

// Usage
const results = await searchProducts("toyota coil", 10, 0);
console.log(results.data.data); // Product array
```

### Python

```python
import requests

def search_products(query, limit=20, offset=0):
    url = "http://localhost:8008/v1/search-by-vector"
    payload = {
        "query": query,
        "limit": limit,
        "offset": offset
    }

    response = requests.post(url, json=payload)
    return response.json()

# Usage
results = search_products("brake pad", 10, 0)
products = results["data"]["data"]
```

---

## 📊 Response Format

### Success Response (200 OK)

```json
{
  "success": true,
  "data": {
    "data": [
      {
        "id": "AC3006",
        "code": "AC3006",
        "name": "COIL AIRCOOL TOYOTA VIGO",
        "price": 850.0,
        "unit": "ชิ้น",
        "supplier_code": "AC",
        "img_url": "",
        "similarity_score": 0.85,
        "sale_price": 850.0,
        "premium_word": "N/A",
        "discount_price": 0.0,
        "discount_percent": 0.0,
        "final_price": 850.0,
        "sold_qty": 0.0,
        "multi_packing": 1,
        "multi_packing_name": "ชิ้น",
        "barcodes": "1234567890123",
        "barcode": "1234567890123",
        "qty_available": 10.0,
        "balance_qty": 10.0,
        "search_priority": 1
      }
    ],
    "total_count": 60,
    "query": "toyota coil",
    "duration": 1125.5
  },
  "message": "Search completed successfully"
}
```

### Response Fields Explained

#### Product Fields

| Field              | Type   | Description                 |
| ------------------ | ------ | --------------------------- |
| `id`               | string | Product ID (same as code)   |
| `code`             | string | Internal product code       |
| `name`             | string | Product name                |
| `price`            | number | Base price                  |
| `sale_price`       | number | Current selling price       |
| `final_price`      | number | Final price after discounts |
| `similarity_score` | number | Relevance score (0.0-1.0)   |
| `qty_available`    | number | Available quantity          |
| `balance_qty`      | number | Current stock balance       |
| `barcodes`         | string | Product barcode(s)          |
| `supplier_code`    | string | Supplier identifier         |
| `search_priority`  | number | Internal priority ranking   |

#### Metadata Fields

| Field         | Type   | Description                     |
| ------------- | ------ | ------------------------------- |
| `total_count` | number | Total matching results found    |
| `query`       | string | Original search query           |
| `duration`    | number | Processing time in milliseconds |

---

## 🎯 Search Algorithm

The search follows this priority order when `offset=0`:

### 1. Priority Search (offset=0 only)

- **Step 1:** Exact match in `ic_inventory_barcode.barcode`
- **Step 2:** Exact match in `ic_inventory.code`
- **Step 3:** LIKE search in both barcode and code fields (`%query%`)

### 2. Vector Database Search

- Uses Weaviate for semantic similarity matching
- Returns relevant IC codes and barcodes
- Supports multilingual search (Thai/English)

### 3. PostgreSQL Enhancement

- Enriches results with detailed product information
- Adds pricing, inventory, and supplier data
- Supplements with additional results if needed

### 4. Result Combination

- Merges vector results with PostgreSQL data
- Removes duplicates
- Applies relevance scoring
- Returns up to requested limit

---

## 🔧 Advanced Features

### Multilingual Support

```json
// Thai search
{"query": "โตโยต้า คอยล์"}

// English search
{"query": "toyota coil"}

// Mixed language
{"query": "honda เบรค"}
```

### Barcode Search

```json
// Direct barcode lookup
{"query": "1234567890123"}

// Partial barcode
{"query": "123456"}
```

### Product Code Search

```json
// Exact product code
{"query": "AC3006"}

// Partial code
{"query": "AC30"}
```

---

## 🚨 Error Handling

### Error Response (400/500)

```json
{
  "success": false,
  "message": "Query parameter is required"
}
```

### Common Errors

| Status | Error                         | Solution                       |
| ------ | ----------------------------- | ------------------------------ |
| 400    | `Query parameter is required` | Provide non-empty query string |
| 400    | `Invalid JSON format`         | Check JSON syntax              |
| 405    | `Method Not Allowed`          | Use POST method only           |
| 500    | `Database connection error`   | Check server logs              |

---

## 📈 Performance

### Response Times

- **Priority Search:** ~50-100ms
- **Vector + PostgreSQL:** ~800-1200ms
- **Large Result Sets:** ~1000-2000ms

### Best Practices

- ✅ Use specific product names or codes for exact matches
- ✅ Use Thai or English keywords interchangeably
- ✅ Include brand names for better results
- ✅ Use barcodes for precise product lookup
- ✅ Set appropriate limits to balance speed vs. completeness

---

## 🎯 Search Tips

### Recommended Search Patterns

```bash
# Brand + Category (Recommended)
"toyota brake"
"honda oil filter"
"nissan coil"

# Product Categories
"brake pad"
"air filter"
"spark plug"

# Thai Keywords
"โตโยต้า"
"ฮอนด้า"
"เบรค"
"น้ำมัน"

# Product Codes
"AC3006"
"TG447600"

# Barcodes
"1234567890123"
```

---

## 📋 Rate Limits & Guidelines

### Performance Guidelines

- **Recommended Limit:** 20-50 results per request
- **Maximum Limit:** 500 results per request
- **Optimal Query Length:** 2-10 words
- **Pagination:** Use offset for large result sets

### System Limits

- **Request Size:** Max 1MB JSON payload
- **Timeout:** 30 seconds maximum
- **Concurrent Requests:** Recommended max 10 per client
