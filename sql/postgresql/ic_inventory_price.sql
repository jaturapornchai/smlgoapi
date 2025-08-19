-- Table: public.ic_inventory_price

-- DROP TABLE IF EXISTS public.ic_inventory_price;

-- Create sequence if not exists
CREATE SEQUENCE IF NOT EXISTS public.ic_inventory_price_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 2147483647
    CACHE 1;

ALTER SEQUENCE public.ic_inventory_price_id_seq
    OWNER TO postgres;

CREATE TABLE IF NOT EXISTS public.ic_inventory_price
(
    id integer NOT NULL DEFAULT nextval('ic_inventory_price_id_seq'::regclass),
    row_order_ref integer DEFAULT 0,
    ic_code VARCHAR NOT NULL,
    unit_code VARCHAR,
    from_qty numeric DEFAULT 0,
    to_qty numeric DEFAULT 0,
    from_date date,
    to_date date,
    sale_type integer DEFAULT 0,
    sale_price1 numeric DEFAULT 0,
    status_price integer DEFAULT 0,
    price_type integer DEFAULT 0,
    cust_code VARCHAR,
    sale_price2 numeric DEFAULT 0,
    cust_group_1 VARCHAR,
    price_mode VARCHAR,
    CONSTRAINT ic_inventory_price_pkey PRIMARY KEY (id)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.ic_inventory_price
    OWNER to postgres;

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_ic_inventory_price_row_order_ref ON public.ic_inventory_price(row_order_ref);

/*
  ic_inventory_price: ฐานข้อมูลราคาสินค้า
  field:
    id: รหัสราคา
    row_order_ref: ใช้อ้างอิงจากฐานข้อมูลอื่น ๆ
    ic_code: รหัสสินค้า
    unit_code: รหัสหน่วยสินค้า
    from_qty: จำนวนเริ่มต้น
    to_qty: จำนวนสิ้นสุด
    from_date: วันที่เริ่มต้น
    to_date: วันที่สิ้นสุด
    sale_type: ประเภทการขาย
    sale_price1: ราคาขาย 1
    status_price: สถานะราคา (active/inactive)
    price_type: ประเภทราคา
    cust_code: รหัสลูกค้า
    sale_price2: ราคาขาย 2
    cust_group_1: กลุ่มลูกค้า 1
    price_mode: โหมดราคา (fixed/variable)
*/