-- ClickHouse Table: ic_trans
-- Transaction header table for inventory movements

-- DROP TABLE IF EXISTS ic_trans;

CREATE TABLE IF NOT EXISTS ic_trans
(
    doc_date Date,
    doc_no String,
    cust_code String,
    trans_type smallint,
    trans_flag smallint,
    sale_code String,
    vat_rate Float64,
    total_value Float64,
    total_after_vat Float64,
    total_discount Float64,
    total_amount Float64
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(doc_date)
ORDER BY (doc_date, doc_no)
SETTINGS index_granularity = 8192;

/*
  ic_trans: ฐานข้อมูลหัวรายการธุรกรรมสินค้า (ClickHouse)
  field:
    - doc_date: วันที่เอกสาร
    - doc_no: หมายเลขเอกสาร
    - cust_code: รหัสลูกค้า
    - trans_type: ประเภทธุรกรรม
    - trans_flag: ธงสถานะ
    - sale_code: รหัสการขาย
    - vat_rate: อัตราภาษีมูลค่าเพิ่ม
    - total_value: มูลค่ารวม
    - total_after_vat: มูลค่าหลังหักภาษี
    - total_discount: มูลค่าหลังหักส่วนลด
    - total_amount: มูลค่ารวมสุทธิ
*/
