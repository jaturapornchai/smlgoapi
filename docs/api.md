## API (POST) สำหรับค้นหาข้อมูลการปกครองไทย

ส่ง parameter เป็น JSON POST

# จังหวัด
POST /get/provinces
ดึงข้อมูลจังหวัดทั้งหมด
Request Body: {} (empty JSON object)
ผลลัพธ์: JSON (id, name_th, name_en)
id = xprovince_id

ตัวอย่าง Response:
```json
{
  "success": true,
  "message": "Retrieved 77 provinces successfully",
  "data": [
    {"id": 1, "name_th": "กรุงเทพมหานคร", "name_en": "Bangkok"},
    {"id": 2, "name_th": "สมุทรปราการ", "name_en": "Samut Prakan"}
  ]
}
```

# อำเภอ
POST /get/amphures
Request Body: {"province_id": xprovince_id}
ดึงข้อมูลอำเภอทั้งหมดในจังหวัดที่ระบุ
ผลลัพธ์: JSON (id, name_th, name_en)
id = xamphure_id
เงื่อนไข: where province_id = xprovince_id

ตัวอย่าง Request:
```json
{"province_id": 1}
```

ตัวอย่าง Response:
```json
{
  "success": true,
  "message": "Retrieved 50 amphures for province_id 1",
  "data": [
    {"id": 1001, "name_th": "เขตพระนคร", "name_en": "Khet Phra Nakhon"},
    {"id": 1002, "name_th": "เขตดุสิต", "name_en": "Khet Dusit"}
  ]
}
```

# ตำบล
POST /get/tambons
Request Body: {"amphure_id": xamphure_id, "province_id": xprovince_id}
ดึงข้อมูลตำบลทั้งหมดในอำเภอที่ระบุ
ผลลัพธ์: JSON (id, name_th, name_en)
id = xtambon_id
เงื่อนไข: where amphure_id = xamphure_id and province_id = xprovince_id

ตัวอย่าง Request:
```json
{"amphure_id": 1001, "province_id": 1}
```

ตัวอย่าง Response:
```json
{
  "success": true,
  "message": "Retrieved 12 tambons for amphure_id 1001 in province_id 1",
  "data": [
    {"id": 100101, "name_th": "พระบรมมหาราชวัง", "name_en": "Phra Borom Maha Ratchawang"},
    {"id": 100102, "name_th": "วังบูรพาภิรมย์", "name_en": "Wang Burapha Phirom"}
  ]
}
```

## ลำดับการใช้งาน:

1. เรียกจังหวัด POST /get/provinces -> ได้ xprovince_id
2. เรียกอำเภอโดยส่ง {"province_id": xprovince_id} -> ได้ xamphure_id
3. เรียกตำบลโดยส่ง {"amphure_id": xamphure_id, "province_id": xprovince_id} -> ได้ xtambon_id

## ข้อมูลที่ใช้:
ชื่อ file JSON ใน folder `provinces`:
- api_province.json - ข้อมูลจังหวัด 77 จังหวัด
- api_amphure.json - ข้อมูลอำเภอทั้งหมด (~1000 อำเภอ)
- api_tambon.json - ข้อมูลตำบลทั้งหมด (~7000+ ตำบล)
- api_province_with_amphure_tambon.json - ข้อมูลแบบครบถ้วน
- api_revert_tambon_with_amphure_province.json - ข้อมูลแบบย้อนกลับ

## สถานะการพัฒนา: ✅ เสร็จสมบูรณ์
- ✅ สร้าง API endpoints ตามเอกสาร
- ✅ ทดสอบการทำงานกับข้อมูลจริง
- ✅ เพิ่มลงใน /guide endpoint
- ✅ สร้าง Postman collection
- ✅ สร้าง Python demo script
- ✅ Upload ทั้งหมดไปยัง GitHub

## การใช้งาน:
เหมาะสำหรับ address forms, location selectors, ระบบจัดการที่อยู่, และแอปพลิเคชันที่ต้องการข้อมูลการปกครองไทย

