# SMLGOAPI Configuration

## การตั้งค่าฐานข้อมูล ClickHouse

SMLGOAPI สามารถอ่านการตั้งค่าจากไฟล์ JSON ได้ เพื่อให้ง่ายต่อการจัดการ

### วิธีการใช้งาน

1. **สร้างไฟล์ `smlgoapi.json`** ในโฟลเดอร์รูท:

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": "8008"
  },
  "clickhouse": {
    "host": "your_clickhouse_host",
    "port": "9000",
    "user": "your_username",
    "password": "your_password",
    "database": "your_database",
    "secure": false
  }
}
```

2. **ลำดับการอ่านค่า Configuration:**
   - อ่านจากไฟล์ `smlgoapi.json` ก่อน (แนะนำ)
   - ถ้าไม่มีไฟล์ จะใช้ Environment Variables
   - ถ้าไม่มี Environment Variables จะใช้ค่า Default

### ไฟล์ Configuration ตัวอย่าง

- `smlgoapi.json` - ค่าเริ่มต้น (localhost)
- `smlgoapi.production.json` - ตัวอย่างสำหรับ production (ไม่ถูกส่งขึ้น Git)

### ความปลอดภัย

- ไฟล์ที่มี `production` หรือ `local` ในชื่อจะไม่ถูกส่งขึ้น Git
- แก้ไข `.gitignore` เพื่อป้องกันไฟล์ที่มีข้อมูลสำคัญ

### การใช้งานจริง

1. Copy `smlgoapi.production.json` เป็น `smlgoapi.json`
2. แก้ไขค่า `host`, `user`, `password`, `database` 
3. เริ่มเซิร์ฟเวอร์: `go run main.go`

เซิร์ฟเวอร์จะแสดงข้อความยืนยันการอ่านจากไฟล์ JSON:
```
✅ Successfully loaded configuration from smlgoapi.json
📄 Loading configuration from smlgoapi.json
```
