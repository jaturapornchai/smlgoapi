# Quick Reference: /search API

## 🚀 Quick Start

**URL:** `POST http://localhost:8008/search`  
**Content-Type:** `application/json`

```json
{
  "query": "search_term",
  "limit": 10,
  "offset": 0
}
```

## 📊 Response Structure (Flattened)

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
    "duration_ms": 724.27
  }
}
```

## 🎯 Search Priorities

| Priority | Method | Example | Score |
|----------|--------|---------|-------|
| **1** | Code Search | `"07-1151"` | 1.0 |
| **2** | Name Search | `"water"` | 0.8 |
| **3** | Vector Search | `"beverage"` | 0.01-1.0 |

## 💡 Quick Examples

### Search by Code
```bash
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{"query": "07-1151", "limit": 5}'
```

### Search by Name
```bash
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{"query": "water", "limit": 5}'
```

### Pagination
```bash
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{"query": "item", "limit": 10, "offset": 20}'
```

## 🔧 PowerShell Examples

```powershell
# Basic search
Invoke-RestMethod -Uri "http://localhost:8008/search" `
  -Method POST -ContentType "application/json" `
  -Body '{"query": "07-1151", "limit": 5}'

# Get first result details
$response = Invoke-RestMethod -Uri "http://localhost:8008/search" `
  -Method POST -ContentType "application/json" `
  -Body '{"query": "water", "limit": 3}'

$response.data.data[0]
```

## 📋 Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Product ID |
| `name` | string | Product name |
| `similarity_score` | float | Relevance (0.01-1.0) |
| `code` | string | Product code |
| `balance_qty` | float | Inventory quantity |
| `price` | float | Product price |
| `supplier_code` | string | Supplier code |
| `unit` | string | Unit of measurement |
| `img_url` | string | Product image URL |
| `search_priority` | int | Search method (1-3) |

## ⚡ Performance Tips

- **Response Time:** ~500-1000ms
- **Max Limit:** 100 results
- **Best Practice:** Use specific codes for fastest results
- **Pagination:** Use offset for large result sets

## 🚨 Error Codes

| Code | Reason |
|------|--------|
| 400 | Missing/empty query |
| 500 | Server error |

## 🔍 Search Strategy

1. **Known Code:** Use exact product code
2. **Product Name:** Search by Thai/English name
3. **Discovery:** Use general terms for related products
4. **Pagination:** Use offset for browsing large result sets
