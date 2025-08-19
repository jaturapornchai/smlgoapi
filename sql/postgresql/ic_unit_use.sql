-- Table: public.ic_unit_use

-- DROP TABLE IF EXISTS public.ic_unit_use;

CREATE TABLE IF NOT EXISTS public.ic_unit_use
(
    ic_code VARCHAR NOT NULL,
    code VARCHAR NOT NULL,
    line_number integer DEFAULT 0,
    stand_value numeric DEFAULT 0,
    divide_value numeric DEFAULT 0,
    row_order_ref integer DEFAULT 0
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.ic_unit_use
    OWNER to postgres;

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_ic_unit_use_row_order_ref ON public.ic_unit_use(row_order_ref);
CREATE INDEX IF NOT EXISTS idx_ic_unit_use_ic_code ON public.ic_unit_use(ic_code);
CREATE INDEX IF NOT EXISTS idx_ic_unit_use_code ON public.ic_unit_use(code);

/*  
    ic_unit_use: ฐานข้อมูลการใช้งานหน่วยสินค้า
    field:
        ic_code: รหัสสินค้า
        line_number: หมายเลขบรรทัด
        stand_value: ค่ามาตรฐาน
        divide_value: ค่าที่ใช้ในการแบ่ง
        row_order_ref: ใช้อ้างอิงจากฐานข้อมูลอื่น ๆ
*/
