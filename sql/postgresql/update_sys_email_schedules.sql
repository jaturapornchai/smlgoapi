-- ไฟล์: smlgoapi/sql/postgresql/update_sys_email_schedules.sql
-- Database: rojproject

-- เพิ่ม columns ที่ขาดหายจาก MongoDB schema
ALTER TABLE sys_email_schedules
ADD COLUMN IF NOT EXISTS shop_id TEXT,
ADD COLUMN IF NOT EXISTS report_id TEXT,
ADD COLUMN IF NOT EXISTS report_name TEXT,
ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS days_of_week INTEGER[], -- Array ของวัน (0-6)
ADD COLUMN IF NOT EXISTS times TEXT[], -- Array ของเวลา ("HH:MM")
ADD COLUMN IF NOT EXISTS timezone VARCHAR(50) DEFAULT 'Asia/Bangkok',
ADD COLUMN IF NOT EXISTS recipients TEXT[], -- Array ของอีเมลผู้รับ
ADD COLUMN IF NOT EXISTS cc_recipients TEXT[], -- Array ของอีเมล CC
ADD COLUMN IF NOT EXISTS condition_guid TEXT,
ADD COLUMN IF NOT EXISTS email_subject TEXT,
ADD COLUMN IF NOT EXISTS include_pdf BOOLEAN DEFAULT TRUE;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_sys_email_schedules_shop_report ON sys_email_schedules(shop_id, report_id);
CREATE INDEX IF NOT EXISTS idx_sys_email_schedules_enabled ON sys_email_schedules(enabled);
CREATE INDEX IF NOT EXISTS idx_sys_email_schedules_next_run ON sys_email_schedules(next_run_at);
