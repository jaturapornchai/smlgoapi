-- Table: public.ic_inventory_barcode

-- DROP TABLE IF EXISTS public.ic_inventory_barcode;

CREATE TABLE IF NOT EXISTS public.ic_inventory_barcode
(
    ic_code VARCHAR NOT NULL,
    barcode VARCHAR NOT NULL,
    unit_code VARCHAR,
    row_order_ref integer DEFAULT 0,
    image_url TEXT,
    CONSTRAINT ic_inventory_barcode_pkey PRIMARY KEY (barcode)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.ic_inventory_barcode
    OWNER to postgres;

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_ic_inventory_barcode_row_order_ref ON public.ic_inventory_barcode(row_order_ref);

/*
  ic_inventory_barcode: ฐานข้อมูลบาร์โค้ดของสินค้า
  field:
    ic_code: รหัสสินค้า
    barcode: บาร์โค้ดสินค้า
    unit_code: รหัสหน่วยสินค้า
    row_order_ref: ใช้อ้างอิงจากฐานข้อมูลอื่น ๆ
    image_url: URL รูปภาพสินค้า
*/

-- Trigger function for weaviate batch processing
CREATE OR REPLACE FUNCTION trigger_weaviate_batch_ic_inventory_barcode()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO weaviate_batch (table_id, active_id, row_order_ref) 
        VALUES (3, 2, NEW.row_order_ref);
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO weaviate_batch (table_id, active_id, row_order_ref) 
        VALUES (3, 1, OLD.row_order_ref);
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
DROP TRIGGER IF EXISTS trigger_ic_inventory_barcode_weaviate ON public.ic_inventory_barcode;
CREATE TRIGGER trigger_ic_inventory_barcode_weaviate
    AFTER INSERT OR DELETE ON public.ic_inventory_barcode
    FOR EACH ROW
    EXECUTE FUNCTION trigger_weaviate_batch_ic_inventory_barcode();
