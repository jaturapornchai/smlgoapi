# MongoDB Atlas API Endpoints

## สรุป
API ทำหน้าที่เป็น **Proxy/Bridge** ระหว่าง Frontend กับ MongoDB Atlas  
Frontend มีอิสระเต็มที่ในการกำหนดโครงสร้างข้อมูล, filter, และ operations

## Base URL
`http://localhost:8108/v1`

## Endpoints

### 1. Insert/Update (Upsert) - `/mongoatlasupdate`

**Method:** POST

**Request Body:**
```json
{
  "collection": "shopping_carts",
  "filter": {
    "shopid": "shop001",
    "email": "user@example.com",
    "cartid": "cart123"
  },
  "data": {
    "items": [
      {"product_id": "P001", "quantity": 2, "price": 100},
      {"product_id": "P002", "quantity": 1, "price": 250}
    ],
    "total": 450,
    "status": "active",
    "updated_at": "2024-01-20T10:30:00Z"
  },
  "upsert": true,
  "replaceone": false
}
```

**Parameters:**
- `collection` (required): ชื่อ collection ใน MongoDB
- `filter` (required): เงื่อนไขในการค้นหา document (กำหนดโครงสร้างเองได้)
- `data` (required): ข้อมูลที่ต้องการบันทึก (กำหนดโครงสร้างเองได้)
- `upsert` (boolean, default: false): `true` = insert ถ้าไม่เจอหรือ update ถ้าเจอ
- `replaceone` (boolean, default: false): `true` = แทนที่ document ทั้งหมด, `false` = update ด้วย `$set`

**Response Success:**
```json
{
  "status": "success",
  "code": 200,
  "matched_count": 1,
  "modified_count": 1,
  "upserted_id": null
}
```

**Note:**
- Frontend ควบคุมโครงสร้าง `filter` และ `data` ทั้งหมด
- ไม่มีการเพิ่ม field อัตโนมัติ (เช่น updated_at) ต้องส่งมาเองถ้าต้องการ
- `replaceone: false` จะใช้ `$set` (ไม่ลบ field เดิมที่ไม่ได้ระบุ)
- `replaceone: true` จะแทนที่ document ทั้งหมด (ลบ field เดิมที่ไม่ได้ระบุ)

---

### 2. Delete - `/mongoatlasdelete`

**Method:** POST

**Request Body:**
```json
{
  "collection": "shopping_carts",
  "filter": {
    "shopid": "shop001",
    "email": "user@example.com",
    "cartid": "cart123"
  },
  "delete_many": false
}
```

**Parameters:**
- `collection` (required): ชื่อ collection
- `filter` (required): เงื่อนไขในการค้นหา document ที่จะลบ (กำหนดเองได้)
- `delete_many` (boolean, default: false): `true` = ลบหลาย documents, `false` = ลบ document เดียว

**Response Success:**
```json
{
  "status": "success",
  "code": 200,
  "deleted_count": 1
}
```

**Note:**
- Frontend ควบคุม filter condition ทั้งหมด
- ระวัง `delete_many: true` กับ filter ที่กว้างเกินไป

---

### 3. Get/Query - `/mongoatlasget`

**Method:** POST

**Request Body:**
```json
{
  "collection": "shopping_carts",
  "filter": {
    "shopid": "shop001",
    "status": "active"
  },
  "projection": {
    "_id": 0,
    "items": 1,
    "total": 1
  },
  "sort": {
    "updated_at": -1
  },
  "limit": 10,
  "skip": 0
}
```

**Parameters:**
- `collection` (required): ชื่อ collection
- `filter` (optional): เงื่อนไขในการค้นหา (กำหนดเองได้, {} = ดึงทั้งหมด)
- `projection` (optional): กำหนด fields ที่ต้องการ (1 = แสดง, 0 = ซ่อน)
- `sort` (optional): การเรียงลำดับ (1 = ascending, -1 = descending)
- `limit` (optional): จำนวนสูงสุดที่จะดึง
- `skip` (optional): จำนวนที่จะข้าม (สำหรับ pagination)

**Response Success:**
```json
{
  "status": "success",
  "code": 200,
  "count": 2,
  "data": [
    {
      "items": [...],
      "total": 450
    }
  ]
}
```

**Note:**
- Frontend ควบคุม filter, projection, sort ทั้งหมด
- สามารถใช้ MongoDB query operators ได้ทั้งหมด (เช่น `$gte`, `$in`, `$regex`)

---

## Error Responses

### MongoDB Not Connected (503)
```json
{
  "status": "error",
  "code": 503,
  "message": "MongoDB Atlas is not connected"
}
```

### Validation Error (400)
```json
{
  "status": "error",
  "code": 400,
  "message": "Collection name is required"
}
```

### Server Error (500)
```json
{
  "status": "error",
  "code": 500,
  "message": "Failed to update document",
  "error": "connection timeout"
}
```

---

## Environment Variables

ต้องตั้งค่าใน `.env`:

```env
MONGODB_ATLAS_URI=mongodb+srv://username:password@cluster.mongodb.net/dbname?retryWrites=true&w=majority
MONGODB_ATLAS_DBNAME=sml_goapi_db
```

---

## ตัวอย่างการใช้งาน

### ตัวอย่าง 1: เพิ่ม/อัพเดทตะกร้าสินค้า
```bash
curl -X POST http://localhost:8108/v1/mongoatlasupdate \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "carts",
    "filter": {"user_id": "U001", "cart_id": "C001"},
    "data": {
      "user_id": "U001",
      "cart_id": "C001",
      "items": [
        {"sku": "A-001", "qty": 2, "price": 100}
      ],
      "total": 200,
      "updated_at": "2024-01-20T10:30:00Z"
    },
    "upsert": true
  }'
```

### ตัวอย่าง 2: Query ด้วย MongoDB operators
```bash
curl -X POST http://localhost:8108/v1/mongoatlasget \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "products",
    "filter": {
      "price": {"$gte": 100, "$lte": 500},
      "category": {"$in": ["electronics", "gadgets"]},
      "name": {"$regex": "iPhone", "$options": "i"}
    },
    "projection": {"name": 1, "price": 1, "stock": 1},
    "sort": {"price": -1},
    "limit": 20
  }'
```

### ตัวอย่าง 3: ลบ documents ด้วย filter
```bash
curl -X POST http://localhost:8108/v1/mongoatlasdelete \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "sessions",
    "filter": {
      "expired_at": {"$lt": "2024-01-01T00:00:00Z"}
    },
    "delete_many": true
  }'
```

### ตัวอย่าง 4: ReplaceOne (แทนที่ document ทั้งหมด)
```bash
curl -X POST http://localhost:8108/v1/mongoatlasupdate \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "user_settings",
    "filter": {"user_id": "U001"},
    "data": {
      "user_id": "U001",
      "theme": "dark",
      "language": "th",
      "notifications": true
    },
    "upsert": true,
    "replaceone": true
  }'
```

### ตัวอย่าง 5: บันทึกเวลาประมวลผลรายงาน (sml_process_time)
```bash
# บันทึก/อัพเดทเวลาประมวลผล (composite key: shopid, reportid, condition_guid)
curl -X POST http://localhost:8108/v1/mongoatlasupdate \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "sml_process_time",
    "filter": {
      "shopid": "rungroj",
      "reportid": "SRR40001",
      "condition_guid": "cond-001"
    },
    "data": {
      "shopid": "rungroj",
      "reportid": "SRR40001",
      "condition_guid": "cond-001",
      "report_name": "รายงานวิเคราะห์ขายขาดทุน",
      "condition_name": "สินค้าทั้งหมด",
      "start_time": "2024-01-20T10:00:00Z",
      "end_time": "2024-01-20T10:05:23Z",
      "duration_seconds": 323,
      "status": "completed",
      "row_count": 1500,
      "updated_at": "2024-01-20T10:05:23Z"
    },
    "upsert": true
  }'
```

### ตัวอย่าง 6: ดึงประวัติเวลาประมวลผลของรายงาน
```bash
# ดึงเวลาประมวลผลทุกเงื่อนไขของรายงาน SRR40001
curl -X POST http://localhost:8108/v1/mongoatlasget \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "sml_process_time",
    "filter": {
      "shopid": "rungroj",
      "reportid": "SRR40001"
    },
    "projection": {
      "_id": 0,
      "condition_guid": 1,
      "condition_name": 1,
      "duration_seconds": 1,
      "row_count": 1,
      "status": 1,
      "updated_at": 1
    },
    "sort": {"updated_at": -1},
    "limit": 50
  }'
```

### ตัวอย่าง 7: ลบประวัติเวลาประมวลผลเงื่อนไขเฉพาะ
```bash
# ลบเงื่อนไขเฉพาะ
curl -X POST http://localhost:8108/v1/mongoatlasdelete \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "sml_process_time",
    "filter": {
      "shopid": "rungroj",
      "reportid": "SRR40001",
      "condition_guid": "cond-001"
    },
    "delete_many": false
  }'
```

### ตัวอย่าง 8: หาเวลาประมวลผลเฉลี่ยของรายงาน
```bash
# Query รายงานที่ใช้เวลานานกว่า 5 นาที (300 วินาที)
curl -X POST http://localhost:8108/v1/mongoatlasget \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "sml_process_time",
    "filter": {
      "shopid": "rungroj",
      "duration_seconds": {"$gte": 300},
      "status": "completed"
    },
    "projection": {
      "reportid": 1,
      "report_name": 1,
      "condition_name": 1,
      "duration_seconds": 1,
      "row_count": 1
    },
    "sort": {"duration_seconds": -1},
    "limit": 10
  }'
```

### ตัวอย่าง 9: บันทึกตารางส่งอีเมลอัตโนมัติ (email_schedules)
```bash
# สร้าง/อัพเดทตารางส่งอีเมลรายงาน
curl -X POST http://localhost:8108/v1/mongoatlasupdate \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "email_schedules",
    "filter": {
      "shopid": "rungroj",
      "reportid": "SRR40001",
      "schedule_id": "schedule-001"
    },
    "data": {
      "shopid": "rungroj",
      "reportid": "SRR40001",
      "schedule_id": "schedule-001",
      "schedule_name": "ส่งรายงานขายขาดทุนประจำวัน",
      "report_name": "รายงานวิเคราะห์ขายขาดทุน",
      "enabled": true,
      "days_of_week": [1, 2, 3, 4, 5],
      "times": ["08:00", "17:00"],
      "timezone": "Asia/Bangkok",
      "recipients": [
        "manager@example.com",
        "sales@example.com",
        "director@example.com"
      ],
      "cc_recipients": [
        "accounting@example.com",
        "warehouse@example.com",
        "support@example.com"
      ],
      "condition_guid": "cond-001",
      "email_subject": "รายงานวิเคราะห์ขายขาดทุนประจำวัน",
      "include_pdf": true,
      "created_at": "2024-01-20T10:00:00Z",
      "updated_at": "2024-01-20T10:00:00Z"
    },
    "upsert": true
  }'
```

**โครงสร้าง email_schedules:**
- `days_of_week`: วันที่ต้องการส่ง (0=อาทิตย์, 1=จันทร์, ..., 6=เสาร์) - สามารถเลือกหลายวัน
- `times`: เวลาที่ต้องการส่งในแต่ละวัน (format: "HH:MM") - สามารถหลายครั้งต่อวัน
- `timezone`: timezone สำหรับการคำนวณเวลา
- `recipients`: รายชื่ออีเมลผู้รับ (TO) - array สามารถใส่ได้หลายคน
- `cc_recipients`: รายชื่ออีเมล CC - array สามารถใส่ได้หลายคน
- `enabled`: เปิด/ปิดการส่งอีเมล

### ตัวอย่าง 10: ตารางส่งอีเมลทุกวัน เวลา 09:00
```bash
curl -X POST http://localhost:8108/v1/mongoatlasupdate \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "email_schedules",
    "filter": {
      "shopid": "rungroj",
      "reportid": "SRR40002",
      "schedule_id": "schedule-daily-morning"
    },
    "data": {
      "shopid": "rungroj",
      "reportid": "SRR40002",
      "schedule_id": "schedule-daily-morning",
      "schedule_name": "ส่งรายงานยอดขายทุกเช้า",
      "enabled": true,
      "days_of_week": [0, 1, 2, 3, 4, 5, 6],
      "times": ["09:00"],
      "timezone": "Asia/Bangkok",
      "recipients": ["owner@example.com"],
      "updated_at": "2024-01-20T10:00:00Z"
    },
    "upsert": true
  }'
```

### ตัวอย่าง 11: ตารางส่งอีเมลเฉพาะวันธรรมดา 3 รอบ
```bash
curl -X POST http://localhost:8108/v1/mongoatlasupdate \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "email_schedules",
    "filter": {
      "shopid": "rungroj",
      "reportid": "SRR40003",
      "schedule_id": "schedule-weekday-3times"
    },
    "data": {
      "shopid": "rungroj",
      "reportid": "SRR40003",
      "schedule_id": "schedule-weekday-3times",
      "schedule_name": "ส่งรายงานสต็อกวันธรรมดา 3 รอบ",
      "enabled": true,
      "days_of_week": [1, 2, 3, 4, 5],
      "times": ["08:00", "13:00", "18:00"],
      "timezone": "Asia/Bangkok",
      "recipients": ["warehouse@example.com", "purchasing@example.com"],
      "updated_at": "2024-01-20T10:00:00Z"
    },
    "upsert": true
  }'
```

### ตัวอย่าง 12: ดึงตารางส่งอีเมลทั้งหมดของ shop
```bash
curl -X POST http://localhost:8108/v1/mongoatlasget \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "email_schedules",
    "filter": {
      "shopid": "rungroj",
      "enabled": true
    },
    "projection": {
      "_id": 0,
      "reportid": 1,
      "schedule_name": 1,
      "days_of_week": 1,
      "times": 1,
      "recipients": 1
    },
    "sort": {"reportid": 1}
  }'
```

### ตัวอย่าง 13: ปิดการส่งอีเมลอัตโนมัติชั่วคราว
```bash
curl -X POST http://localhost:8108/v1/mongoatlasupdate \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "email_schedules",
    "filter": {
      "shopid": "rungroj",
      "reportid": "SRR40001",
      "schedule_id": "schedule-001"
    },
    "data": {
      "enabled": false,
      "updated_at": "2024-01-20T15:00:00Z"
    },
    "upsert": false
  }'
```

### ตัวอย่าง 14: หาตารางที่ต้องส่งในวันนี้
```bash
# Query ตารางที่เปิดใช้งาน และกำหนดส่งในวันจันทร์ (day 1)
curl -X POST http://localhost:8108/v1/mongoatlasget \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "email_schedules",
    "filter": {
      "shopid": "rungroj",
      "enabled": true,
      "days_of_week": {"$in": [1]}
    },
    "projection": {
      "reportid": 1,
      "schedule_name": 1,
      "times": 1,
      "recipients": 1,
      "cc_recipients": 1,
      "condition_guid": 1
    }
  }'
```

### ตัวอย่าง 15: เพิ่มผู้รับอีเมลและ CC เพิ่มเติม
```bash
# อัพเดท recipients และ cc_recipients โดยไม่กระทบข้อมูลอื่น
curl -X POST http://localhost:8108/v1/mongoatlasupdate \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "email_schedules",
    "filter": {
      "shopid": "rungroj",
      "reportid": "SRR40001",
      "schedule_id": "schedule-001"
    },
    "data": {
      "recipients": [
        "manager@example.com",
        "sales@example.com",
        "director@example.com",
        "owner@example.com",
        "ceo@example.com"
      ],
      "cc_recipients": [
        "accounting@example.com",
        "warehouse@example.com",
        "support@example.com",
        "it@example.com",
        "marketing@example.com",
        "hr@example.com"
      ],
      "updated_at": "2024-01-20T16:00:00Z"
    },
    "upsert": false
  }'
```

### ตัวอย่าง 16: ส่งรายงานให้หลายทีม (TO และ CC เยอะ)
```bash
curl -X POST http://localhost:8108/v1/mongoatlasupdate \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "email_schedules",
    "filter": {
      "shopid": "rungroj",
      "reportid": "SRR40004",
      "schedule_id": "schedule-all-teams"
    },
    "data": {
      "shopid": "rungroj",
      "reportid": "SRR40004",
      "schedule_id": "schedule-all-teams",
      "schedule_name": "ส่งรายงานสรุปทุกทีม",
      "enabled": true,
      "days_of_week": [1, 3, 5],
      "times": ["09:00"],
      "timezone": "Asia/Bangkok",
      "recipients": [
        "ceo@example.com",
        "cfo@example.com",
        "coo@example.com"
      ],
      "cc_recipients": [
        "sales.manager@example.com",
        "warehouse.manager@example.com",
        "accounting.manager@example.com",
        "it.manager@example.com",
        "marketing.manager@example.com",
        "hr.manager@example.com",
        "purchasing.manager@example.com",
        "operations.manager@example.com"
      ],
      "updated_at": "2024-01-20T10:00:00Z"
    },
    "upsert": true
  }'
```

---

## สรุป

API ทำหน้าที่เป็น **Pure Proxy** ไม่มีการบังคับโครงสร้างใดๆ:

✅ Frontend กำหนด collection structure เอง  
✅ Frontend กำหนด filter conditions เอง  
✅ Frontend กำหนด data schema เอง  
✅ รองรับ MongoDB query operators ทั้งหมด  
✅ ไม่เพิ่ม/ลบ fields อัตโนมัติ  
✅ ยืดหยุ่นและใช้งานง่าย

