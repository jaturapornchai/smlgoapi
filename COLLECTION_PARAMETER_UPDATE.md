# Collection Parameter Update for `/search-by-vector` Endpoint

## 📋 Overview

ได้ปรับปรุง endpoint `/search-by-vector` ให้รองรับการระบุ **collection** ผ่าน request body แทนที่จะใช้ค่าจาก environment variable เพียงอย่างเดียว

## 🔧 Changes Made

### 1. Model Update (`models/models.go`)
เพิ่ม field `collection` ใน `SearchParameters`:
```go
type SearchParameters struct {
    Query      string `json:"query" binding:"required"`      
    Limit      int    `json:"limit,omitempty"`               
    Offset     int    `json:"offset,omitempty"`              
    AI         int    `json:"ai,omitempty"`                  
    Collection string `json:"collection,omitempty"`          // ใหม่!
}
```

### 2. Weaviate Service Update (`services/weaviate.go`)
เพิ่ม method `SearchProductsWithCollection()`:
```go
func (w *WeaviateService) SearchProductsWithCollection(ctx context.Context, query string, limit int, collection string) ([]Product, error)
```

### 3. Handler Update (`handlers/api.go`)
แก้ไข `SearchProductsByVector()` ให้ใช้ collection จาก request body:
```go
collection := params.Collection
vectorProducts, err := h.weaviateService.SearchProductsWithCollection(ctx, query, limit, collection)
```

## 📖 Usage Examples

### 1. Default Collection (Backward Compatible)
```json
{
  "query": "โตโยต้า คอยล์",
  "limit": 10,
  "offset": 0
}
```
👆 ใช้ collection จาก `WEAVIATE_COLLECTION` ใน .env file

### 2. Custom Collection
```json
{
  "query": "toyota coil",
  "limit": 10,
  "offset": 0,
  "collection": "Changthai"
}
```
👆 ใช้ collection ที่ระบุใน request body

### 3. Different Collection
```json
{
  "query": "brake pad",
  "limit": 5,
  "offset": 0,
  "collection": "Product"
}
```

## 🎯 Priority Logic

1. **ถ้ามี `collection` ใน request body** → ใช้ค่านั้น
2. **ถ้าไม่มี `collection`** → ใช้ค่าจาก `WEAVIATE_COLLECTION` env variable
3. **ถ้าไม่มีทั้งคู่** → ใช้ default value `"Product"`

## 🔄 Backward Compatibility

✅ **การเปลี่ยนแปลงนี้ไม่กระทบกับ API clients เดิม**

- Client เดิมที่ไม่ส่ง `collection` จะยังทำงานได้ปกติ
- Method `SearchProducts()` เดิมยังคงใช้งานได้
- Environment variable `WEAVIATE_COLLECTION` ยังคงใช้เป็น fallback

## 🧪 Testing

### PowerShell Test
```bash
./test_collection_parameter.ps1
```

### cURL Examples
```bash
# Default collection
curl -X POST "http://localhost:8008/v1/search-by-vector" \
  -H "Content-Type: application/json" \
  -d '{"query": "โตโยต้า", "limit": 5}'

# Custom collection
curl -X POST "http://localhost:8008/v1/search-by-vector" \
  -H "Content-Type: application/json" \
  -d '{"query": "toyota", "limit": 5, "collection": "Changthai"}'
```

## ✅ Verification

1. ✅ Build สำเร็จ
2. ✅ Model เพิ่ม field `collection`
3. ✅ Service รองรับ collection parameter
4. ✅ Handler ใช้ collection จาก request body
5. ✅ Backward compatibility
6. ✅ Test files พร้อมใช้งาน

## 📝 Notes

- Collection name ต้องตรงกับชื่อที่มีอยู่ใน Weaviate database
- หาก collection ไม่มีอยู่จริง Weaviate จะ return error
- Collection parameter เป็น optional (ไม่บังคับ)
