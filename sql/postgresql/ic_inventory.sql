-- Table: public.ic_inventory

-- DROP TABLE IF EXISTS public.ic_inventory;

CREATE TABLE IF NOT EXISTS public.ic_inventory
(
    ic_code VARCHAR NOT NULL,
    name_1 VARCHAR,
    unit_standard_code VARCHAR,
    item_type integer DEFAULT 0,
    row_order_ref integer DEFAULT 0,
    unit_type integer DEFAULT 0,    
    balance_qty numeric DEFAULT 0,
    image_url TEXT,
    CONSTRAINT ic_inventory_pkey PRIMARY KEY (ic_code)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.ic_inventory
    OWNER to postgres;

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_ic_inventory_row_order_ref ON public.ic_inventory(row_order_ref);

/*  
    ic_inventory: ฐานข้อมูลสินค้าคงคลัง
    field:
        ic_code: รหัสสินค้า
        name_1: ชื่อสินค้า
        unit_standard_code: รหัสหน่วยมาตรฐาน
        item_type: ประเภทสินค้า (0 = สินค้าทั่วไป, 1 = วัตถุดิบ, 2 = สินค้าสำเร็จรูป)
        row_order_ref: ใช้อ้างอิงจากฐานข้อมูลอื่น ๆ
        unit_type: ประเภทหน่วย (0 = หน่วยมาตรฐาน, 1 = หน่วยย่อย)
        balance_qty: จำนวนคงเหลือในคลัง
        image_url: URL รูปภาพสินค้า
*/
    