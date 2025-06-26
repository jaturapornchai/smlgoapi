# VECTOR SEARCH LOG EXPLANATION

## Log Flow การทำงานของ /search-by-vector

### 1. เริ่มต้น (Start)
```
🚀 [VECTOR-SEARCH] === STARTING SEARCH ===
   📝 Query: 'ท่อแอร์'
   📊 Limit: 50, Offset: 0
   =====================================
```

### 2. Weaviate Vector Database
```
🎲 [VECTOR-SEARCH] Weaviate returned 150 products from vector database
```
- ค้นหาใน Weaviate vector database ด้วย BM25
- ได้ผลลัพธ์ 150 สินค้าที่เกี่ยวข้อง

### 3. IC Code Extraction
```
🎯 [VECTOR-SEARCH] Extracting IC codes from Weaviate: 150 codes found
```
- แยก IC Code จากผลลัพธ์ Weaviate
- ได้ IC Code ทั้งหมด 150 รายการ

### 4. PostgreSQL Database Search
```
🔍 [PostgreSQL] Searching by IC/Barcode (with relevance) codes: 150 items, limit=50, offset=0
```
- ค้นหาในฐานข้อมูล PostgreSQL โดยใช้ IC Code
- สร้าง SQL query พร้อม relevance scoring
- จำกัดผลลัพธ์ 50 รายการ

### 5. Price & Balance Data Loading
```
🏷️ [PostgreSQL] Loading price data for 50 products...
✅ [PostgreSQL] Loaded price data for 34 products
📦 [PostgreSQL] Loading balance data for 50 products...  
✅ [PostgreSQL] Loaded balance data for 19 products
```
- โหลดข้อมูลราคาและสต็อกจากตารางแยก
- ราคา: 34/50 สินค้ามีข้อมูลราคา
- สต็อก: 19/50 สินค้ามีข้อมูลสต็อก

### 6. Summary Statistics
```
💰 [PostgreSQL] Price data: 34/50 products have pricing
📦 [PostgreSQL] Balance data: 19/50 products have stock info
✅ [PostgreSQL] Search completed: found 50 results, total count: 150
```

### 7. Final Results
```
🎯 [VECTOR-SEARCH] === SEARCH RESULTS SUMMARY ===
   📝 Query: 'ท่อแอร์'
   🔗 Search Method: IC Code
   🎲 Vector Database: 150 products found
   📊 PostgreSQL Total: 150 records
   📋 Returned Results: 50 products
   📄 Page Info: page 1 (offset: 0, limit: 50)
   ⏱️  Processing Time: 891.4ms
   🏆 Top Results:
     1. [A-88703-F4040] HOSE SUB-ASSY, DISCHARGE (ท่อแอร์) (Relevance: 100.0%)
     2. [A-88704-F4040] HOSE SUB-ASSY, SUCTION (ท่อแอร์) (Relevance: 100.0%)
     3. [TL-43] ชุดบานแป๊ป ท่อแอร์ (AURUKI) (Relevance: 100.0%)
   ===============================
✅ [VECTOR-SEARCH] COMPLETED (891.4ms)
```

## การทำงานไม่ซ้ำซ้อน

### ✅ ปัจจุบัน (หลังปรับปรุง):
1. **Weaviate search** → 1 ครั้ง
2. **IC code extraction** → 1 ครั้ง  
3. **PostgreSQL search** → 1 ครั้ง
4. **Price loading** → 1 ครั้ง (filter เฉพาะที่ค้นเจอ)
5. **Balance loading** → 1 ครั้ง (filter เฉพาะที่ค้นเจอ)

### Search Methods:
- **Primary**: IC Code (รหัสสินค้า)
- **Fallback**: Barcode (ถ้าไม่เจอ IC Code)

### Log Improvements:
- ✅ ลบ SQL query ยาวๆ ออก
- ✅ แสดงสถิติแบบสรุป
- ✅ ลดการ log รายละเอียดสินค้าทีละรายการ
- ✅ แสดงเฉพาะข้อมูลสำคัญ

## Performance:
- **Total Time**: ~891ms
- **Vector Search**: ~300ms  
- **PostgreSQL**: ~400ms
- **Price/Balance**: ~191ms
