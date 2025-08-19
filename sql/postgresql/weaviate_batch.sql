-- Vector search batch processing
-- Table public.weaviate_batch
-- DROP TABLE IF EXISTS public.weaviate_batch;

CREATE TABLE IF NOT EXISTS public.weaviate_batch (
    roworder SERIAL PRIMARY KEY,
    table_id integer,
    active_id integer,
    row_order_ref integer DEFAULT 0
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_weaviate_batch_table_id ON public.weaviate_batch(table_id);
CREATE INDEX IF NOT EXISTS idx_weaviate_batch_active_id ON public.weaviate_batch(active_id);
CREATE INDEX IF NOT EXISTS idx_weaviate_batch_row_order_ref ON public.weaviate_batch(row_order_ref);

-- Set table owner
ALTER TABLE IF EXISTS public.weaviate_batch OWNER to postgres;

/*
  weaviate_batch: ฐานข้อมูล batch processing สำหรับ vector search
  field:
    roworder: ลำดับการประมวลผล (Primary Key, SERIAL)
    table_id: รหัสตาราง (1=ar_customer,2=ic_inventory_barcode
    active_id: การดำเนินการ (1=delete,2=insert)
    row_order_ref: อ้างอิงไปยัง row_order_ref ในตารางต้นทาง
*/
