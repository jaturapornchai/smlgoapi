-- Table: public.ar_customer

-- DROP TABLE IF EXISTS public.ar_customer;

CREATE TABLE IF NOT EXISTS public.ar_customer
(
    code VARCHAR NOT NULL,
    price_level integer,
    name_1 VARCHAR,
    address_1 VARCHAR,
    telephone VARCHAR,
    row_order_ref integer DEFAULT 0,
    CONSTRAINT ar_customer_pkey PRIMARY KEY (code)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.ar_customer
    OWNER to postgres;

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_ar_customer_row_order_ref ON public.ar_customer(row_order_ref);

/*
  ar_customer: ฐานข้อมูลลูกค้า
  field:
    code: รหัสลูกค้า
    price_level: ระดับราคาสำหรับลูกค้า
    name_1: ชื่อลูกค้า
    address_1: ที่อยู่ลูกค้า
    telephone: เบอร์โทรศัพท์ลูกค้า
    row_order_ref: ใช้อ้างอิงจากฐานข้อมูลอื่น ๆ
*/

-- Trigger function for weaviate batch processing
CREATE OR REPLACE FUNCTION trigger_weaviate_batch_ar_customer()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO weaviate_batch (table_id, active_id, row_order_ref) 
        VALUES (1, 2, NEW.row_order_ref);
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO weaviate_batch (table_id, active_id, row_order_ref) 
        VALUES (1, 1, OLD.row_order_ref);
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
DROP TRIGGER IF EXISTS trigger_ar_customer_weaviate ON public.ar_customer;
CREATE TRIGGER trigger_ar_customer_weaviate
    AFTER INSERT OR DELETE ON public.ar_customer
    FOR EACH ROW
    EXECUTE FUNCTION trigger_weaviate_batch_ar_customer();
