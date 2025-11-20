package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// TableResultCreate ensures the result table along with comments and indexes exists.
func TableResultCreate(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}

	log.Println("Ensuring result table exists (result)")

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	createTableQuery := `
        CREATE TABLE IF NOT EXISTS result (
            id SERIAL PRIMARY KEY,
            guid TEXT,
            querynumber INTEGER default 0,
            level INTEGER default 0,
            typejson INTEGER default 0,
            docdatetime TIMESTAMPTZ,
            linenumber INTEGER,
            datajson JSONB
        )`

	if _, err := tx.ExecContext(ctx, createTableQuery); err != nil {
		return fmt.Errorf("create result table: %w", err)
	}

	commentsAndIndexesQuery := `
        COMMENT ON TABLE result IS 'ตารางเก็บผลลัพธ์การประมวลผลข้อมูล';

        COMMENT ON COLUMN result.id IS 'รหัสอัตโนมัติ';
        COMMENT ON COLUMN result.guid IS 'รหัส GUID สำหรับอ้างอิง';
        COMMENT ON COLUMN result.querynumber IS 'ประเภทของผลลัพธ์';
        COMMENT ON COLUMN result.level IS 'ระดับของข้อมูล (0=หัวข้อหลัก, 1=รายละเอียด, 2=รายละเอียดย่อย ,ฯลฯ)';
        COMMENT ON COLUMN result.typejson IS 'ชนิดของข้อมูล JSON (สำรองไว้ในกรณีที่ต้องการแยกประเภทข้อมูล เช่น ข้อมูล,ยอดรวม)';
        COMMENT ON COLUMN result.docdatetime IS 'วันที่และเวลาของเอกสาร';
        COMMENT ON COLUMN result.linenumber IS 'เลขที่บรรทัด';
        COMMENT ON COLUMN result.datajson IS 'ข้อมูล JSONB ของผลลัพธ์ (รองรับการ query และ index)';

        CREATE INDEX IF NOT EXISTS idx_result_docdatetime ON result (docdatetime);
        CREATE INDEX IF NOT EXISTS idx_result_guid_querynumber_linenumber ON result (guid,querynumber,linenumber);
    `

	if _, err := tx.ExecContext(ctx, commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create result comments and indexes: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	log.Println("Result table ready")
	return nil
}
