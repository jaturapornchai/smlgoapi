package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// InitSystemTables ensures all system-related tables exist in the database.
func (s *PostgreSQLService) InitSystemTables() error {
	log.Println("Ensuring system tables exist")
	
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	queries := []string{
		// 1. users
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'user',
			is_admin BOOLEAN NOT NULL DEFAULT FALSE,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			shop_id VARCHAR(50) NOT NULL,
			allowed_reports TEXT, -- JSON array string
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			created_by VARCHAR(255),
			updated_by VARCHAR(255)
		)`,
		
		// 2. sys_email_schedules
		`CREATE TABLE IF NOT EXISTS sys_email_schedules (
			schedule_id TEXT PRIMARY KEY,
			schedule_name TEXT,
			shop_id TEXT,
			schedule_pattern TEXT,
			next_run_at TIMESTAMPTZ,
			interval_minutes INTEGER DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// 3. sys_email_settings
		`CREATE TABLE IF NOT EXISTS sys_email_settings (
			shop_id TEXT PRIMARY KEY,
			email_list JSONB DEFAULT '[]',
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// 4. sys_scheduler_logs
		`CREATE TABLE IF NOT EXISTS sys_scheduler_logs (
			log_id SERIAL PRIMARY KEY,
			schedule_id TEXT,
			status TEXT,
			message TEXT,
			payload JSONB,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)`,

		// 5. sys_application_logs
		`CREATE TABLE IF NOT EXISTS sys_application_logs (
			log_id SERIAL PRIMARY KEY,
			shop_id TEXT,
			username TEXT,
			activity_type TEXT,
			target TEXT,
			details TEXT,
			payload JSONB,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)`,

		// 6. user_logs
		`CREATE TABLE IF NOT EXISTS user_logs (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255),
			action VARCHAR(50) NOT NULL, -- login, create_user, update_user, delete_user, toggle_active
			success BOOLEAN NOT NULL,
			ip_address VARCHAR(45),
			user_agent TEXT,
			error_message TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		// 7. sml_process_time
		`CREATE TABLE IF NOT EXISTS sml_process_time (
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
		)`,
	}

	for _, q := range queries {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("execute query: %w", err)
		}
	}

	// Index สำหรับ user_logs
	indices := []string{
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_shop_id ON users(shop_id)`,
		`CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_users_is_admin ON users(is_admin)`,
		`CREATE INDEX IF NOT EXISTS idx_sys_email_schedules_next_run ON sys_email_schedules(next_run_at)`,
		`CREATE INDEX IF NOT EXISTS idx_sys_scheduler_logs_schedule_id ON sys_scheduler_logs(schedule_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_logs_username ON user_logs(username)`,
		`CREATE INDEX IF NOT EXISTS idx_user_logs_action ON user_logs(action)`,
		`CREATE INDEX IF NOT EXISTS idx_user_logs_created_at ON user_logs(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_sml_process_time_shop_report ON sml_process_time(shop_id, report_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sml_process_time_condition_guid ON sml_process_time(condition_guid)`,
		`CREATE INDEX IF NOT EXISTS idx_sml_process_time_updated_at ON sml_process_time(updated_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_sml_process_time_status ON sml_process_time(status)`,
	}
	
	for _, q := range indices {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("execute index query: %w", err)
		}
	}

	// Update sys_email_schedules with missing columns
	alterQueries := []string{
		`ALTER TABLE sys_email_schedules ADD COLUMN IF NOT EXISTS report_id TEXT`,
		`ALTER TABLE sys_email_schedules ADD COLUMN IF NOT EXISTS report_name TEXT`,
		`ALTER TABLE sys_email_schedules ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT TRUE`,
		`ALTER TABLE sys_email_schedules ADD COLUMN IF NOT EXISTS days_of_week INTEGER[]`,
		`ALTER TABLE sys_email_schedules ADD COLUMN IF NOT EXISTS times TEXT[]`,
		`ALTER TABLE sys_email_schedules ADD COLUMN IF NOT EXISTS timezone VARCHAR(50) DEFAULT 'Asia/Bangkok'`,
		`ALTER TABLE sys_email_schedules ADD COLUMN IF NOT EXISTS recipients TEXT[]`,
		`ALTER TABLE sys_email_schedules ADD COLUMN IF NOT EXISTS cc_recipients TEXT[]`,
		`ALTER TABLE sys_email_schedules ADD COLUMN IF NOT EXISTS condition_guid TEXT`,
		`ALTER TABLE sys_email_schedules ADD COLUMN IF NOT EXISTS email_subject TEXT`,
		`ALTER TABLE sys_email_schedules ADD COLUMN IF NOT EXISTS include_pdf BOOLEAN DEFAULT TRUE`,
	}

	for _, q := range alterQueries {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			log.Printf("Warning: Failed to execute alter query: %v", err)
			// Continue, as column might already exist or there's another non-critical issue
		}
	}

	// Trigger for sml_process_time
	triggerSQL := `
	CREATE OR REPLACE FUNCTION update_sml_process_time_updated_at()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trigger_update_sml_process_time_updated_at') THEN
			CREATE TRIGGER trigger_update_sml_process_time_updated_at
			BEFORE UPDATE ON sml_process_time
			FOR EACH ROW
			EXECUTE FUNCTION update_sml_process_time_updated_at();
		END IF;
	END $$;
	`
	if _, err := tx.ExecContext(ctx, triggerSQL); err != nil {
		log.Printf("Warning: Failed to create trigger: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	log.Println("System tables ready")
	
	// Seed default admin if no users exist
	return s.seedDefaultAdmin()
}

func (s *PostgreSQLService) seedDefaultAdmin() error {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return fmt.Errorf("count users: %w", err)
	}

	if count == 0 {
		log.Println("Seeding default admin user")
		// password is 'admin123' as requested for local
		_, err := s.db.Exec(`
			INSERT INTO users (username, password, role, is_admin, is_active, shop_id, allowed_reports, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			"admin", "admin123", "admin", true, true, "rungroj", "[]", "system")
		if err != nil {
			return fmt.Errorf("insert seed user: %w", err)
		}
	}
	return nil
}

// IsAdminUsername checks if a username is in the pre-defined admin list from environment variable
func (s *PostgreSQLService) IsAdminUsername(username string) bool {
	adminUsernamesEnv := os.Getenv("ADMIN_USERNAMES")
	if adminUsernamesEnv == "" {
		// Default hardcoded admins if env is not set (backup)
		defaultAdmins := []string{"admin", "jaturapornchai", "gonut", "thep", "viroonc", "somkid", "roj2543"}
		username = strings.ToLower(strings.TrimSpace(username))
		for _, admin := range defaultAdmins {
			if admin == username {
				return true
			}
		}
		return false
	}

	admins := strings.Split(adminUsernamesEnv, ",")
	username = strings.ToLower(strings.TrimSpace(username))

	for _, admin := range admins {
		if strings.ToLower(strings.TrimSpace(admin)) == username {
			return true
		}
	}
	return false
}

// User methods
func (s *PostgreSQLService) AuthenticateUser(ctx context.Context, username, password string) (map[string]interface{}, error) {
	var user = make(map[string]interface{})
	var dbPassword string
	var role string
	var isAdmin bool
	var isActive bool
	var shopID string
	var allowedReports sql.NullString

	err := s.db.QueryRowContext(ctx,
		"SELECT password, role, is_admin, shop_id, allowed_reports, is_active FROM users WHERE username = $1",
		username).Scan(&dbPassword, &role, &isAdmin, &shopID, &allowedReports, &isActive)

	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, err
	}

	if !isActive {
		return nil, fmt.Errorf("user is inactive")
	}

	// Check if user is admin from environment variable override
	if s.IsAdminUsername(username) {
		isAdmin = true
		role = "admin"
	}

	user["username"] = username
	user["role"] = role
	user["isAdmin"] = isAdmin
	user["shopId"] = shopID
	user["allowed_reports"] = allowedReports.String
	user["dbPassword"] = dbPassword // Return for manual check in handler (plain text vs hashed)

	return user, nil
}

// Schedule methods
func (s *PostgreSQLService) GetDueSchedules(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, 
		"SELECT schedule_id, schedule_name, shop_id, schedule_pattern, next_run_at, interval_minutes FROM sys_email_schedules WHERE next_run_at <= CURRENT_TIMESTAMP")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, name, shop, pattern string
		var nextRun interface{}
		var interval int
		if err := rows.Scan(&id, &name, &shop, &pattern, &nextRun, &interval); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"schedule_id":      id,
			"schedule_name":    name,
			"shop_id":          shop,
			"schedule_pattern": pattern,
			"next_run_at":     nextRun,
			"interval_minutes": interval,
		})
	}
	return results, nil
}

func (s *PostgreSQLService) UpdateScheduleNextRun(ctx context.Context, id string, nextRun interface{}) error {
	_, err := s.db.ExecContext(ctx, 
		"UPDATE sys_email_schedules SET next_run_at = $1, updated_at = CURRENT_TIMESTAMP WHERE schedule_id = $2", 
		nextRun, id)
	return err
}

func (s *PostgreSQLService) LogSchedulerExecution(ctx context.Context, id, name, status string, payload interface{}) error {
	payloadJSON, _ := json.Marshal(payload)
	_, err := s.db.ExecContext(ctx, 
		"INSERT INTO sys_scheduler_logs (schedule_id, status, message, payload) VALUES ($1, $2, $3, $4)", 
		id, status, name, payloadJSON)
	return err
}

// User Management
func (s *PostgreSQLService) UpsertUser(ctx context.Context, username, password, role string, isAdmin bool, shopID string, allowedReports string, isActive bool) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users (username, password, role, is_admin, shop_id, allowed_reports, is_active, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
		ON CONFLICT (username) DO UPDATE SET
			password = CASE WHEN EXCLUDED.password != '' THEN EXCLUDED.password ELSE users.password END,
			role = EXCLUDED.role,
			is_admin = EXCLUDED.is_admin,
			shop_id = EXCLUDED.shop_id,
			allowed_reports = EXCLUDED.allowed_reports,
			is_active = EXCLUDED.is_active,
			updated_at = CURRENT_TIMESTAMP`,
		username, password, role, isAdmin, shopID, allowedReports, isActive)
	return err
}

func (s *PostgreSQLService) DeleteUser(ctx context.Context, username string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE username = $1", username)
	return err
}

func (s *PostgreSQLService) GetUsersByShopID(ctx context.Context, shopID string) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT username, role, is_admin, shop_id, allowed_reports, is_active, created_at, updated_at FROM users WHERE shop_id = $1",
		shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var username, role, shopID string
		var isAdmin, isActive bool
		var allowedReports sql.NullString
		var createdAt, updatedAt interface{}
		if err := rows.Scan(&username, &role, &isAdmin, &shopID, &allowedReports, &isActive, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"username":        username,
			"role":            role,
			"is_admin":        isAdmin,
			"is_active":       isActive,
			"shop_id":         shopID,
			"allowed_reports": allowedReports.String,
			"created_at":      createdAt,
			"updated_at":      updatedAt,
		})
	}
	return results, nil
}

// Email Settings Management
func (s *PostgreSQLService) GetEmailSettings(ctx context.Context, shopID string) (string, error) {
	var emailList []byte
	err := s.db.QueryRowContext(ctx, "SELECT email_list FROM sys_email_settings WHERE shop_id = $1", shopID).Scan(&emailList)
	if err == sql.ErrNoRows {
		return "[]", nil
	}
	if err != nil {
		return "", err
	}
	return string(emailList), nil
}

func (s *PostgreSQLService) UpsertEmailSettings(ctx context.Context, shopID string, emailList string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sys_email_settings (shop_id, email_list, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (shop_id) DO UPDATE SET
			email_list = EXCLUDED.email_list,
			updated_at = CURRENT_TIMESTAMP`,
		shopID, emailList)
	return err
}

// Schedule CRUD
func (s *PostgreSQLService) GetSchedulesByShopID(ctx context.Context, shopID string) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, 
		"SELECT schedule_id, schedule_name, shop_id, report_id, schedule_pattern, next_run_at, interval_minutes FROM sys_email_schedules WHERE shop_id = $1", 
		shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, name, shop, report, pattern string
		var nextRun interface{}
		var interval int
		if err := rows.Scan(&id, &name, &shop, &report, &pattern, &nextRun, &interval); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"schedule_id":      id,
			"schedule_name":    name,
			"shop_id":          shop,
			"report_id":        report,
			"schedule_pattern": pattern,
			"next_run_at":     nextRun,
			"interval_minutes": interval,
		})
	}
	return results, nil
}

func (s *PostgreSQLService) GetSchedulesByReportID(ctx context.Context, shopID, reportID string) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, 
		"SELECT schedule_id, schedule_name, shop_id, report_id, schedule_pattern, next_run_at, interval_minutes FROM sys_email_schedules WHERE shop_id = $1 AND report_id = $2", 
		shopID, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, name, shop, report, pattern string
		var nextRun interface{}
		var interval int
		if err := rows.Scan(&id, &name, &shop, &report, &pattern, &nextRun, &interval); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"schedule_id":      id,
			"schedule_name":    name,
			"shop_id":          shop,
			"report_id":        report,
			"schedule_pattern": pattern,
			"next_run_at":     nextRun,
			"interval_minutes": interval,
		})
	}
	return results, nil
}

func (s *PostgreSQLService) UpsertSchedule(ctx context.Context, id, name, shop, report, pattern string, nextRun interface{}, interval int) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sys_email_schedules (schedule_id, schedule_name, shop_id, report_id, schedule_pattern, next_run_at, interval_minutes, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
		ON CONFLICT (schedule_id) DO UPDATE SET
			schedule_name = EXCLUDED.schedule_name,
			shop_id = EXCLUDED.shop_id,
			report_id = EXCLUDED.report_id,
			schedule_pattern = EXCLUDED.schedule_pattern,
			next_run_at = EXCLUDED.next_run_at,
			interval_minutes = EXCLUDED.interval_minutes,
			updated_at = CURRENT_TIMESTAMP`,
		id, name, shop, report, pattern, nextRun, interval)
	return err
}

func (s *PostgreSQLService) DeleteSchedule(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM sys_email_schedules WHERE schedule_id = $1", id)
	return err
}

// LogUserAction logs user-related activity to user_logs table
func (s *PostgreSQLService) LogUserAction(ctx context.Context, username, action string, success bool, ipAddress, userAgent, errorMessage string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_logs (username, action, success, ip_address, user_agent, error_message)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		username, action, success, ipAddress, userAgent, errorMessage)
	return err
}

// Activity Logs
func (s *PostgreSQLService) LogActivity(ctx context.Context, shopID, username, activityType, target, details string, payload interface{}) error {
	payloadJSON, _ := json.Marshal(payload)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sys_application_logs (shop_id, username, activity_type, target, details, payload)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		shopID, username, activityType, target, details, payloadJSON)
	return err
}

func (s *PostgreSQLService) GetActivityLogsByShopID(ctx context.Context, shopID, target string, limit int) ([]map[string]interface{}, error) {
	query := `SELECT username, activity_type, target, details, payload, created_at 
	          FROM sys_application_logs 
	          WHERE shop_id = $1`
	args := []interface{}{shopID}
	
	if target != "" {
		query += " AND target = $2"
		args = append(args, target)
	}
	
	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprint(len(args)+1)
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var username, activityType, targetRes, details string
		var payload []byte
		var createdAt interface{}
		if err := rows.Scan(&username, &activityType, &targetRes, &details, &payload, &createdAt); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"username":      username,
			"activity_type": activityType,
			"target":        targetRes,
			"details":       details,
			"payload":       string(payload),
			"created_at":    createdAt,
		})
	}
	return results, nil
}

// Process Time Management

// UpsertProcessTime creates or updates process time record
func (s *PostgreSQLService) UpsertProcessTime(ctx context.Context, shopID, reportID, conditionGuid, reportName, conditionName string, startTime, endTime interface{}, durationSeconds, rowCount int, status string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sml_process_time (shop_id, report_id, condition_guid, report_name, condition_name, start_time, end_time, duration_seconds, status, row_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (shop_id, report_id, condition_guid) DO UPDATE SET
			report_name = EXCLUDED.report_name,
			condition_name = EXCLUDED.condition_name,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			duration_seconds = EXCLUDED.duration_seconds,
			status = EXCLUDED.status,
			row_count = EXCLUDED.row_count,
			updated_at = CURRENT_TIMESTAMP`,
		shopID, reportID, conditionGuid, reportName, conditionName, startTime, endTime, durationSeconds, status, rowCount)
	return err
}

// GetProcessTime retrieves a single process time record
func (s *PostgreSQLService) GetProcessTime(ctx context.Context, shopID, reportID, conditionGuid string) (map[string]interface{}, error) {
	var result = make(map[string]interface{})
	var startTime, endTime, createdAt, updatedAt interface{}
	var durationSeconds, rowCount int
	var sId, rId, cGuid, rName, cName, status string

	err := s.db.QueryRowContext(ctx,
		`SELECT shop_id, report_id, condition_guid, report_name, condition_name, start_time, end_time, duration_seconds, status, row_count, created_at, updated_at
		 FROM sml_process_time
		 WHERE shop_id = $1 AND report_id = $2 AND condition_guid = $3`,
		shopID, reportID, conditionGuid).Scan(
		&sId, &rId, &cGuid, &rName, &cName, &startTime, &endTime, &durationSeconds, &status, &rowCount, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	result["shop_id"] = sId
	result["report_id"] = rId
	result["condition_guid"] = cGuid
	result["report_name"] = rName
	result["condition_name"] = cName
	result["start_time"] = startTime
	result["end_time"] = endTime
	result["duration_seconds"] = durationSeconds
	result["status"] = status
	result["row_count"] = rowCount
	result["created_at"] = createdAt
	result["updated_at"] = updatedAt

	return result, nil
}

// DeleteProcessTime deletes a process time record
func (s *PostgreSQLService) DeleteProcessTime(ctx context.Context, shopID, reportID, conditionGuid string) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM sml_process_time WHERE shop_id = $1 AND report_id = $2 AND condition_guid = $3",
		shopID, reportID, conditionGuid)
	return err
}

// GetProcessTimeList retrieves a list of process time records
func (s *PostgreSQLService) GetProcessTimeList(ctx context.Context, shopID, reportID string, limit int) ([]map[string]interface{}, error) {
	query := `SELECT shop_id, report_id, condition_guid, report_name, condition_name, start_time, end_time, duration_seconds, status, row_count, updated_at
			  FROM sml_process_time WHERE shop_id = $1`
	args := []interface{}{shopID}

	if reportID != "" {
		query += " AND report_id = $2"
		args = append(args, reportID)
	}

	query += " ORDER BY updated_at DESC"

	if limit > 0 {
		query += " LIMIT $" + fmt.Sprint(len(args)+1)
		args = append(args, limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var sId, rId, cGuid, rName, cName, status string
		var startTime, endTime, updatedAt interface{}
		var durationSeconds, rowCount int

		if err := rows.Scan(&sId, &rId, &cGuid, &rName, &cName, &startTime, &endTime, &durationSeconds, &status, &rowCount, &updatedAt); err != nil {
			return nil, err
		}

		results = append(results, map[string]interface{}{
			"shop_id":          sId,
			"report_id":        rId,
			"condition_guid":   cGuid,
			"report_name":      rName,
			"condition_name":   cName,
			"start_time":       startTime,
			"end_time":         endTime,
			"duration_seconds": durationSeconds,
			"status":           status,
			"row_count":        rowCount,
			"updated_at":       updatedAt,
		})
	}

	return results, nil
}
