# New Multi-Step Search Logic

## Overview

The SMLGOAPI now implements a sophisticated multi-step search algorithm that prioritizes results based on search method relevance:

## Search Steps (in Priority Order)

### 1. Full Text Search by Code (Highest Priority)
- **Search Method**: Direct string matching against product codes
- **Priority**: 1 (Highest)
- **Score**: 1.0 for exact matches
- **Use Case**: When searching for specific product codes like "06-1151", "A001", etc.

### 2. Full Text Search by Name (Medium Priority)  
- **Search Method**: Direct string matching against product names
- **Priority**: 2 (Medium)
- **Score**: 0.8 for matches
- **Use Case**: When searching for product names like "น้ำตาล", "water", "shampoo", etc.

### 3. Vector Search (Lowest Priority)
- **Search Method**: TF-IDF vector similarity calculation
- **Priority**: 3 (Lowest)
- **Score**: Variable (0.01 - 1.0) based on similarity
- **Use Case**: Semantic search, fuzzy matching, related products

## Result Processing

### Deduplication
- Results from all three search methods are combined
- Duplicate products are removed (keeping highest priority match)
- Each product appears only once in final results

### Sorting Algorithm
1. **Primary Sort**: Search priority (1 → 2 → 3)
2. **Secondary Sort**: For similar similarity scores (within 0.1 difference):
   - Products with images are prioritized
3. **Tertiary Sort**: Similarity score (descending)
4. **Fallback Sort**: Product code (alphabetical)

## API Usage

### Request Format
```json
POST /search
Content-Type: application/json

{
  "query": "search_term",
  "limit": 10,
  "offset": 0
}
```

### Response Format
```json
{
  "success": true,
  "message": "Search completed successfully",
  "data": {
    "data": [
      {
        "id": "06-1151",
        "name": "พีเจ้น จุกนมซิลิคอน คลาสสิค /ไซส์M/แพ็ค3",
        "similarity_score": 1.0,
        "metadata": {
          "search_priority": 1,
          "balance_qty": 7,
          "code": "06-1151",
          "img_url": "https://imageapi-dev.dedepos.com/api/productimage/...",
          "price": 100,
          "supplier_code": "AP288",
          "unit": "ชิ้น"
        },
        "img_url": "https://imageapi-dev.dedepos.com/api/productimage/..."
      }
    ],
    "total_count": 1,
    "query": "06-1151",
    "duration_ms": 928.4445
  }
}
```

## Search Examples

### 1. Code Search (Highest Priority)
```bash
# Search for exact product code
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{"query": "06-1151", "limit": 5}'

# Result: Exact match "06-1151" appears first with priority 1
```

### 2. Partial Code Search
```bash
# Search for code fragment
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{"query": "1151", "limit": 5}'

# Result: All products with codes containing "1151" (priority 1)
```

### 3. Name Search (Medium Priority)
```bash
# Search for product name
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{"query": "water", "limit": 5}'

# Result: Products with "water" in name (priority 2 if no code matches)
```

### 4. Vector Search (Lowest Priority)
```bash
# Semantic/fuzzy search
curl -X POST http://localhost:8008/search \
  -H "Content-Type: application/json" \
  -d '{"query": "beverage", "limit": 5}'

# Result: Related products based on TF-IDF similarity (priority 3)
```

## Performance Optimizations

1. **2-Round Processing**: Initial filtering followed by refined calculation
2. **Limited Candidates**: Each search method returns max 2x limit for efficiency
3. **Smart Deduplication**: Map-based approach prevents duplicate processing
4. **Lazy Data Loading**: Additional product data fetched only for final results

## Metadata Fields

Each search result includes metadata with:
- `search_priority`: 1 (code), 2 (name), or 3 (vector)
- `balance_qty`: Current inventory quantity
- `price`: Product price
- `supplier_code`: Supplier information
- `unit`: Unit of measurement
- `img_url`: Product image URL

## Implementation Details

The new search logic is implemented in `services/vector_db.go` with these key functions:

- `SearchProducts()`: Main search orchestrator
- `searchByCode()`: Full text code search
- `searchByName()`: Full text name search  
- `performVectorSearch()`: TF-IDF vector search
- `combineSearchResults()`: Result deduplication
- `sortResultsByPriority()`: Multi-criteria sorting

This approach ensures that users get the most relevant results first, with exact code matches taking precedence over fuzzy semantic matches.
