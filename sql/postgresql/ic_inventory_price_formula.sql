-- Table: public.ic_inventory_price_formula

-- DROP TABLE IF EXISTS public.ic_inventory_price_formula;

-- Create sequence if not exists
CREATE SEQUENCE IF NOT EXISTS public.ic_inventory_price_formula_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 2147483647
    CACHE 1;

ALTER SEQUENCE public.ic_inventory_price_formula_id_seq
    OWNER TO postgres;

CREATE TABLE IF NOT EXISTS public.ic_inventory_price_formula
(
    id integer NOT NULL DEFAULT nextval('ic_inventory_price_formula_id_seq'::regclass),
    row_order_ref integer DEFAULT 0,
    ic_code VARCHAR NOT NULL,
    unit_code VARCHAR NOT NULL,
    sale_type smallint NOT NULL DEFAULT 0,
    price_0 VARCHAR,
    price_1 VARCHAR,
    price_2 VARCHAR,
    price_3 VARCHAR,
    price_4 VARCHAR,
    price_5 VARCHAR,
    price_6 VARCHAR,
    price_7 VARCHAR,
    price_8 VARCHAR,
    price_9 VARCHAR,
    tax_type smallint NOT NULL DEFAULT 0,
    price_currency smallint DEFAULT 0,
    currency_code VARCHAR,
    CONSTRAINT ic_inventory_price_formula_pkey PRIMARY KEY (id)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.ic_inventory_price_formula
    OWNER to postgres;

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_ic_inventory_price_formula_row_order_ref ON public.ic_inventory_price_formula(row_order_ref);

/*
    ic_inventory_price_formula: ฐานข้อมูลสูตรการคำนวณราคาสินค้า
    field:
        id: รหัสสูตรการคำนวณราคา
        row_order_ref: ใช้อ้างอิงจากฐานข้อมูลอื่น ๆ
        ic_code: รหัสสินค้า
        unit_code: รหัสหน่วยสินค้า
        sale_type: ประเภทการขาย (0 = ขายปลีก, 1 = ขายส่ง)
        price_0 - price_9: ราคาสำหรับแต่ละระดับ
        tax_type: ประเภทภาษี (0 = ไม่มีภาษี, 1 = มีภาษี)
        price_currency: สกุลเงินของราคา
        currency_code: รหัสสกุลเงิน
*/