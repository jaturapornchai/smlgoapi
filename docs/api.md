## API (POST) สำหรับค้นหาข้อมูลการปกครองไทย

ส่ง paratameter เป็น json post

# จังหวัด
/get/provinces
ดึงข้อมูลจังหวัดทั้งหมด
ผลลัพธ์: JSON (id, name_th, name_en)
id = xprovince_id


# อำเภอ
/get/amphures
province_id={xprovince_id}
ดึงข้อมูลอำเภอทั้งหมดในจังหวัดที่ระบุ
ผลลัพธ์: JSON (id, name_th, name_en)
id = xamphure_id
เงื่อนไข: where province_id = xprovince_id

# ตำบล
/get/tambons
amphure_id={xamphure_id}&province_id={xprovince_id}
ดึงข้อมูลตำบลทั้งหมดในอำเภอที่ระบุ
ผลลัพธ์: JSON (id, name_th, name_en)
id = xtambon_id
เงื่อนไข: where amphure_id = xamphure_id and province_id = xprovince_id

## ลำดับการใช้งาน:

เรียกจังหวัด -> ได้ xprovince_id
เรียกอำเภอโดยส่ง xprovince_id -> ได้ xamphure_id
เรียกตำบลโดยส่ง xamphure_id และ xprovince_id -> ได้ xtambon_id

ชื่อ file json ใน folder `provinces`:
- api_province.json
- api_amphure.json  
- api_tambon.json
- api_province_with_amphure_tambon.json
- api_revert_tambon_with_amphure_province.json

