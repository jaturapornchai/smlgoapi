-- ไฟล์: smlgoapi/sql/postgresql/sml_process_time.sql
-- Database: rojproject

CREATE TABLE IF NOT EXISTS sml_process_time (
    id SERIAL PRIMARY KEY,
    shop_id VARCHAR(50) NOT NULL,
    report_id VARCHAR(50) NOT NULL,
    condition_guid VARCHAR(255) NOT NULL,
    report_name TEXT,
    condition_name TEXT,
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    duration_seconds INTEGER,
    status VARCHAR(20) DEFAULT 'pending',
    row_count INTEGER,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(shop_id, report_id, condition_guid)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_sml_process_time_shop_report ON sml_process_time(shop_id, report_id);
CREATE INDEX IF NOT EXISTS idx_sml_process_time_condition_guid ON sml_process_time(condition_guid);
CREATE INDEX IF NOT EXISTS idx_sml_process_time_updated_at ON sml_process_time(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_sml_process_time_status ON sml_process_time(status);

-- Trigger สำหรับอัพเดท updated_at อัตโนมัติ
CREATE OR REPLACE FUNCTION update_sml_process_time_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_sml_process_time_updated_at ON sml_process_time;
CREATE TRIGGER trigger_update_sml_process_time_updated_at
BEFORE UPDATE ON sml_process_time
FOR EACH ROW
EXECUTE FUNCTION update_sml_process_time_updated_at();
