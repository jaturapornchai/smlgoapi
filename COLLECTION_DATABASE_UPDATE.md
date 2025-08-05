# Collection-Based Database Search Update

## 📋 Overview

ได้ปรับปรุง endpoint `/search-by-vector` ให้รองรับการใช้ **collection name เป็นชื่อ database** ใน PostgreSQL search แทนที่จะใช้ database เดียวกันเสมอ

## 🔧 Changes Made

### 1. New PostgreSQL Service Method
เพิ่ม method `SearchProductsByBarcodesWithRelevanceAndBarcodeMapInCollection()` ที่:
- รับ `collection` parameter เพิ่มเติม
- แปลง collection name เป็นตัวพิมพ์เล็ก (lowercase)
- ใช้ collection name เป็นชื่อ database ใน PostgreSQL
- สร้าง connection ใหม่ไปยัง database ที่ระบุ

### 2. Handler Update
แก้ไข `SearchProductsByVector()` ให้:
- ตรวจสอบว่ามี collection parameter หรือไม่
- เรียกใช้ method ที่เหมาะสมตาม collection

## 🎯 Logic Flow

```
1. รับ collection จาก request body
2. ถ้ามี collection:
   - แปลงเป็นตัวพิมพ์เล็ก (lowercase)
   - ใช้เป็นชื่อ database ใน PostgreSQL
   - สร้าง connection ใหม่ไปยัง database นั้น
3. ถ้าไม่มี collection:
   - ใช้ database default จาก config
```

## 📖 Usage Examples

### 1. Default Database (ไม่ระบุ collection)
```json
{
  "query": "โตโยต้า คอยล์",
  "limit": 10,
  "offset": 0
}
```
👆 ใช้ database จาก `POSTGRESQL_DATABASE` ใน .env file

### 2. Custom Database via Collection
```json
{
  "query": "toyota coil",
  "limit": 10,
  "offset": 0,
  "collection": "Changthai"
}
```
👆 ใช้ database: `changthai` (lowercase)

### 3. Different Collection/Database
```json
{
  "query": "brake pad",
  "limit": 5,
  "offset": 0,
  "collection": "Product"
}
```
👆 ใช้ database: `product` (lowercase)

## 🔄 Database Connection Logic

### Collection to Database Mapping
- `"Changthai"` → database: `changthai`
- `"Product"` → database: `product`
- `"TOYOTA"` → database: `toyota`
- `""` (empty) → database จาก config default

### Connection Management
- แต่ละ request สร้าง connection ใหม่ไปยัง target database
- Connection จะถูกปิดหลังจาก query เสร็จสิ้น
- มี error handling สำหรับ database ที่ไม่มีอยู่

## 🛡️ Error Handling

### Database Connection Errors
```json
{
  "success": false,
  "message": "failed to connect to database 'nonexistent': database \"nonexistent\" does not exist"
}
```

### Table Not Found Errors
```json
{
  "success": false,
  "message": "table 'ic_inventory' not found in database 'customdb' - please create the table or contact system administrator"
}
```

## 📝 Implementation Details

### Method Signature
```go
func (s *PostgreSQLService) SearchProductsByBarcodesWithRelevanceAndBarcodeMapInCollection(
    ctx context.Context, 
    barcodes []string, 
    relevanceMap map[string]float64, 
    barcodeMap map[string]string, 
    limit, offset int, 
    collection string
) ([]map[string]interface{}, int, error)
```

### Database Connection String
```go
dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
    s.config.PostgreSQL.Host,
    s.config.PostgreSQL.Port,
    s.config.PostgreSQL.User,
    s.config.PostgreSQL.Password,
    strings.ToLower(collection)) // collection เป็น database name
```

## 🧪 Testing Examples

### Test Script
```bash
# Test with default database
curl -X POST "http://localhost:8008/v1/search-by-vector" \
  -H "Content-Type: application/json" \
  -d '{"query": "โตโยต้า", "limit": 5}'

# Test with custom database
curl -X POST "http://localhost:8008/v1/search-by-vector" \
  -H "Content-Type: application/json" \
  -d '{"query": "toyota", "limit": 5, "collection": "Changthai"}'
```

## ⚠️ Important Notes

1. **Database ต้องมีอยู่จริง**: Collection name ที่ระบุต้องตรงกับ database ที่มีอยู่ใน PostgreSQL
2. **Table Structure**: Database ปลายทางต้องมี tables: `ic_inventory`, `ic_price`, `ic_balance`, `ic_inventory_barcode`
3. **Permissions**: User ที่ระบุใน config ต้องมีสิทธิ์เข้าถึง database ที่ระบุ
4. **Case Sensitivity**: Collection name จะถูกแปลงเป็นตัวพิมพ์เล็กเสมอ

## ✅ Verification Checklist

- ✅ Method ใหม่สร้างแล้ว
- ✅ Handler แก้ไขเรียบร้อย
- ✅ Collection แปลงเป็น lowercase
- ✅ Dynamic database connection
- ✅ Error handling ครบถ้วน
- ✅ Backward compatibility
- ✅ Build สำเร็จ
