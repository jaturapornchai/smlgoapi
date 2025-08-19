-- ClickHouse Table: ic_trans_detail
-- Transaction detail table for inventory item movements

-- DROP TABLE IF EXISTS ic_trans_detail;

CREATE TABLE IF NOT EXISTS ic_trans_detail
(
    doc_date Date,
    doc_no String,
    cust_code String,
    trans_type smallint,
    trans_flag smallint,
    item_code String,
    unit_code String,
    wh_code String,
    shelf_code String,
    qty Float64,
    price Float64,
    discount String,
    sum_amount Float64,
    stand_value Float64,
    divide_value Float64,
    qty_calc Float64,
    -- Secondary data skipping index for queries filtering by item_code and doc_date
    INDEX idx_item_code_doc_date (item_code, doc_date) TYPE set(0) GRANULARITY 1
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(doc_date)
ORDER BY (doc_date, item_code)
SETTINGS index_granularity = 8192;

/*
  ic_trans_detail: ฐานข้อมูลรายละเอียดธุรกรรมสินค้า (ClickHouse)
  field:
    - doc_date: วันที่เอกสาร
    - doc_no: หมายเลขเอกสาร
    - cust_code: รหัสลูกค้า
    - trans_type: ประเภทธุรกรรม
    - trans_flag: ธงสถานะ
    - item_code: รหัสสินค้า
    - unit_code: รหัสหน่วย
    - qty: จำนวน
    - price: ราคาต่อหน่วย
    - discount: ส่วนลด
    - sum_amount: มูลค่ารวม
    - wh_code: รหัสคลังสินค้า
    - shelf_code: รหัสชั้นวาง
    - stand_value: มูลค่าสแตนดาร์ด
    - divide_value: มูลค่าหลังหักส่วนแบ่ง
*/
