# Relevance-Based Sorting Implementation

## ✅ การปรับปรุงที่ทำไป (What was improved)

### 1. **Weaviate Service Enhancement** (`services/weaviate.go`)
เพิ่ม methods ใหม่เพื่อส่งคืน relevance scores พร้อมกับ IC codes และ barcodes:

```go
// GetICCodesWithRelevance - ดึง IC codes พร้อม relevance scores
func (w *WeaviateService) GetICCodesWithRelevance(products []Product) ([]string, map[string]float64)

// GetBarcodesWithRelevance - ดึง barcodes พร้อม relevance scores  
func (w *WeaviateService) GetBarcodesWithRelevance(products []Product) ([]string, map[string]float64)
```

### 2. **PostgreSQL Service Enhancement** (`services/postgresql.go`)
เพิ่ม method ใหม่ที่รองรับ relevance-based sorting:

```go
// SearchProductsByBarcodesWithRelevance - ค้นหาพร้อมเรียงตาม relevance
func (s *PostgreSQLService) SearchProductsByBarcodesWithRelevance(
    ctx context.Context, 
    barcodes []string, 
    relevanceMap map[string]float64, 
    limit, offset int
) ([]map[string]interface{}, int, error)
```

#### Key Features:
- **Dynamic ORDER BY**: สร้าง CASE statement เพื่อเรียงตาม relevance score
- **Relevance Score Integration**: ใส่ relevance score จาก Weaviate ลงใน `similarity_score` field
- **Fallback Ordering**: ถ้าไม่มี relevance map จะเรียงตาม name

### 3. **API Handler Update** (`handlers/api.go`)
ปรับปรุง `SearchProductsByVector` เพื่อใช้ relevance scores:

```go
// ใช้ GetICCodesWithRelevance แทน GetICCodes
icCodes, relevanceMap := h.weaviateService.GetICCodesWithRelevance(vectorProducts)

// เรียก SearchProductsByBarcodesWithRelevance พร้อม relevanceMap
searchResults, totalCount, err = h.postgreSQLService.SearchProductsByBarcodesWithRelevance(
    ctx, icCodes, relevanceMap, limit, offset
)
```

## 🔄 วิธีการทำงานใหม่ (How it works now)

### การเรียงลำดับแบบใหม่:
```
1. Weaviate Search → ได้ IC codes + relevance scores
2. PostgreSQL Query with CASE statement:
   ORDER BY 
     CASE 
       WHEN code = 'SP-RAF1425' THEN 100.0
       WHEN code = '403AA-8' THEN 89.7
       WHEN code = '507AA-6' THEN 85.3
       ELSE 0 
     END DESC, 
     name ASC
3. Result → เรียงตาม relevance สูงสุดก่อน
```

### SQL Query ที่สร้างขึ้น:
```sql
SELECT code, name, unit_standard_code, item_type, row_order_ref, 5 as search_priority
FROM ic_inventory 
WHERE CAST(code AS TEXT) IN ($1,$2,$3,...)
ORDER BY 
  CASE 
    WHEN CAST(code AS TEXT) = '403AA-8' THEN 100.000000
    WHEN CAST(code AS TEXT) = '403AA-6' THEN 100.000000  
    WHEN CAST(code AS TEXT) = 'SP-RAF1425' THEN 69.322330
    ELSE 0 
  END DESC, 
  name ASC
LIMIT 50 OFFSET 0
```

## 📊 ผลลัพธ์ที่คาดหวัง (Expected Results)

### ก่อนการปรับปรุง:
```
1. [507AA-8] 1/2" ทีเมียข้างกลางเติมน้ำยา 134A โตโยต้า (Score: 5.0)
2. [403AA-8] 1/2" เมียเก่า x 1/2" ผู้โตโยต้า (Score: 5.0)  
3. [998MC-6] 3/8 134A โตโยต้า (โอริง) ทุุกรุ่นยกเว้น TIGER (Score: 5.0)
```
*เรียงตาม name ASC*

### หลังการปรับปรุง:
```
1. [403AA-8] 1/2" เมียเก่า x 1/2" ผู้โตโยต้า (Score: 100.0)
2. [403AA-6] 3/8" เมียเก่า x 3/8" ผู้โตโยต้า (Score: 100.0)
3. [507AA-8] 1/2" ทีเมียข้างกลางเติมน้ำยา 134A โตโยต้า (Score: 89.7)
```
*เรียงตาม relevance score DESC → name ASC*

## 🎯 ประโยชน์ที่ได้รับ (Benefits)

### 1. **ผลลัพธ์ที่แม่นยำกว่า**
- สินค้าที่มี relevance สูงขึ้นก่อน
- ตรงกับความต้องการของผู้ใช้มากขึ้น

### 2. **เรียงลำดับอัจฉริยะ**
- ใช้ AI-powered relevance จาก Weaviate
- Fallback เป็นการเรียงตามชื่อถ้าไม่มี relevance

### 3. **Backward Compatibility**
- Method เดิม `SearchProductsByBarcodes` ยังใช้งานได้
- ไม่กระทบต่อ API อื่นๆ

## 🧪 การทดสอบ (Testing)

### API Call:
```bash
POST /v1/search-by-vector
{
  "query": "โตโยต้า สายพาน",
  "limit": 10
}
```

### Expected Response Format:
```json
{
  "success": true,
  "data": {
    "data": [
      {
        "code": "403AA-8",
        "name": "1/2\" เมียเก่า x 1/2\" ผู้โตโยต้า",
        "similarity_score": 100.0,
        "sale_price": 100.00,
        "qty_available": 5.00
      }
    ],
    "total_count": 150,
    "query": "โตโยต้า สายพาน",
    "duration_ms": 750
  }
}
```

## 🔧 Debug Information

### Log ที่จะเห็น:
```
🔍 [vector-search] Found 150 IC codes from vector database: [403AA-8, 403AA-6, ...]
🔍 Barcode Search SQL Query: ... ORDER BY CASE WHEN ... THEN 100.0 ... END DESC, name ASC
💰 Found price for 403AA-8: sale_price=100.00
📦 Found balance for 403AA-8: qty_available=5.00
✅ [vector-search] Found 50 results using IC codes
```

การปรับปรุงนี้ทำให้ API `/search-by-vector` ส่งคืนผลลัพธ์ที่เรียงตาม relevance percentage จาก Weaviate แบบถูกต้องแล้ว! 🎉
