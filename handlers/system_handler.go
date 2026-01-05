package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthLoginHandler handles user authentication
func (h *APIHandler) AuthLoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "code": "INVALID_PAYLOAD"})
		return
	}

	user, err := h.postgreSQLService.AuthenticateUser(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		log.Printf("Authentication error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "code": "SERVER_ERROR"})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง", "code": "INVALID_CREDENTIALS"})
		return
	}

	// Verify password (using plain text as requested for local environment)
	dbPassword := user["dbPassword"].(string)
	if dbPassword != req.Password {
		h.postgreSQLService.LogUserAction(c.Request.Context(), req.Username, "login", false, c.ClientIP(), c.Request.UserAgent(), "Invalid password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง", "code": "INVALID_CREDENTIALS"})
		return
	}

	// Log success
	h.postgreSQLService.LogUserAction(c.Request.Context(), req.Username, "login", true, c.ClientIP(), c.Request.UserAgent(), "")

	// Remove dbPassword from response
	delete(user, "dbPassword")

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"user":   user,
	})
}

// AuthRegisterHandler handles user registration
func (h *APIHandler) AuthRegisterHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "code": "INVALID_PAYLOAD"})
		return
	}

	// Default permissions for new users: All reports
	// List of all available reports in the system
	defaultReports := `["SRR40001", "SRR20011", "SRR20016", "SRR30003", "SRR40006", "SRR40010", "SRR50003", "B4029"]`

	// Register user with default settings
	// role: "user", is_admin: false, is_active: true, shop_id: "rungroj"
	err := h.postgreSQLService.UpsertUser(
		c.Request.Context(),
		req.Username,
		req.Password,
		"user",
		false,
		"rungroj",
		defaultReports,
		true,
	)

	if err != nil {
		log.Printf("Registration error: %v", err)
		h.postgreSQLService.LogUserAction(c.Request.Context(), req.Username, "register", false, c.ClientIP(), c.Request.UserAgent(), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถสมัครสมาชิกได้: " + err.Error(), "code": "SERVER_ERROR"})
		return
	}

	h.postgreSQLService.LogUserAction(c.Request.Context(), req.Username, "register", true, c.ClientIP(), c.Request.UserAgent(), "")

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "สมัครสมาชิกสำเร็จ",
	})
}

// UserUpsertHandler handles creating or updating a user
func (h *APIHandler) UserUpsertHandler(c *gin.Context) {
	var req struct {
		Username       string `json:"username" binding:"required"`
		Password       string `json:"password"` // Optional if updating and not changing password
		Role           string `json:"role"`
		IsAdmin        bool   `json:"is_admin"`
		IsActive       *bool  `json:"is_active"`
		ShopID         string `json:"shop_id"`
		AllowedReports string `json:"allowed_reports"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Use plain text password as requested for local
	password := req.Password
	if password == "" {
		// If updating, we might need to get existing password if we don't want to overwrite it with empty
		existingUser, _ := h.postgreSQLService.AuthenticateUser(c.Request.Context(), req.Username, "")
		if existingUser != nil {
			password = existingUser["dbPassword"].(string)
		}
	}

	// Default IsActive to true if nil
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	err := h.postgreSQLService.UpsertUser(c.Request.Context(), req.Username, password, req.Role, req.IsAdmin, req.ShopID, req.AllowedReports, isActive)
	if err != nil {
		h.postgreSQLService.LogUserAction(c.Request.Context(), req.Username, "upsert_user", false, c.ClientIP(), c.Request.UserAgent(), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.postgreSQLService.LogUserAction(c.Request.Context(), req.Username, "upsert_user", true, c.ClientIP(), c.Request.UserAgent(), "")
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// UserDeleteHandler handles deleting a user
func (h *APIHandler) UserDeleteHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing username"})
		return
	}

	err := h.postgreSQLService.DeleteUser(c.Request.Context(), req.Username)
	if err != nil {
		h.postgreSQLService.LogUserAction(c.Request.Context(), req.Username, "delete_user", false, c.ClientIP(), c.Request.UserAgent(), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.postgreSQLService.LogUserAction(c.Request.Context(), req.Username, "delete_user", true, c.ClientIP(), c.Request.UserAgent(), "")
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// UserGetHandler handles retrieving users by shop
func (h *APIHandler) UserGetHandler(c *gin.Context) {
	var req struct {
		ShopID string `json:"shopid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing shopid"})
		return
	}

	users, err := h.postgreSQLService.GetUsersByShopID(c.Request.Context(), req.ShopID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   users,
	})
}

// EmailSettingsGetHandler handles retrieving email settings
func (h *APIHandler) EmailSettingsGetHandler(c *gin.Context) {
	var req struct {
		ShopID string `json:"shopid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing shopid"})
		return
	}

	emailList, err := h.postgreSQLService.GetEmailSettings(c.Request.Context(), req.ShopID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"email_list": emailList,
	})
}

// EmailSettingsUpsertHandler handles updating email settings
func (h *APIHandler) EmailSettingsUpsertHandler(c *gin.Context) {
	var req struct {
		ShopID    string `json:"shopid" binding:"required"`
		EmailList string `json:"email_list" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing shopid or email_list"})
		return
	}

	err := h.postgreSQLService.UpsertEmailSettings(c.Request.Context(), req.ShopID, req.EmailList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// ScheduleGetHandler handles retrieving schedules by shop
func (h *APIHandler) ScheduleGetHandler(c *gin.Context) {
	var req struct {
		ShopID string `json:"shopid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing shopid"})
		return
	}

	schedules, err := h.postgreSQLService.GetSchedulesByShopID(c.Request.Context(), req.ShopID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   schedules,
	})
}

// GetSchedulesByReportHandler handles retrieving schedules by shop and report
func (h *APIHandler) GetSchedulesByReportHandler(c *gin.Context) {
	var req struct {
		ShopID   string `json:"shopid" binding:"required"`
		ReportID string `json:"reportid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing shopid or reportid"})
		return
	}

	schedules, err := h.postgreSQLService.GetSchedulesByReportID(c.Request.Context(), req.ShopID, req.ReportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   schedules,
	})
}

// ScheduleUpsertHandler handles creating or updating a schedule
func (h *APIHandler) ScheduleUpsertHandler(c *gin.Context) {
	var req struct {
		ScheduleID      string    `json:"schedule_id" binding:"required"`
		ScheduleName    string    `json:"schedule_name"`
		ShopID          string    `json:"shop_id"`
		ReportID        string    `json:"report_id"`
		SchedulePattern string    `json:"schedule_pattern"`
		NextRunAt       time.Time `json:"next_run_at"`
		IntervalMinutes int       `json:"interval_minutes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	err := h.postgreSQLService.UpsertSchedule(c.Request.Context(), req.ScheduleID, req.ScheduleName, req.ShopID, req.ReportID, req.SchedulePattern, req.NextRunAt, req.IntervalMinutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// ScheduleDeleteHandler handles deleting a schedule
func (h *APIHandler) ScheduleDeleteHandler(c *gin.Context) {
	var req struct {
		ScheduleID string `json:"schedule_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing schedule_id"})
		return
	}

	err := h.postgreSQLService.DeleteSchedule(c.Request.Context(), req.ScheduleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// ActivityLogHandler handles logging application activity
func (h *APIHandler) ActivityLogHandler(c *gin.Context) {
	var req struct {
		ShopID       string      `json:"shopid"`
		Username     string      `json:"username"`
		ActivityType string      `json:"activity_type"`
		Target       string      `json:"target"`
		Details      string      `json:"details"`
		Payload      interface{} `json:"payload"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	err := h.postgreSQLService.LogActivity(c.Request.Context(), req.ShopID, req.Username, req.ActivityType, req.Target, req.Details, req.Payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// ActivityGetHandler handles retrieving activity logs
func (h *APIHandler) ActivityGetHandler(c *gin.Context) {
	var req struct {
		ShopID string `json:"shopid" binding:"required"`
		Target string `json:"target"`
		Limit  int    `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing shopid"})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 50
	}

	logs, err := h.postgreSQLService.GetActivityLogsByShopID(c.Request.Context(), req.ShopID, req.Target, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   logs,
	})
}

// ProcessTimeUpsertHandler handles creating or updating process time
func (h *APIHandler) ProcessTimeUpsertHandler(c *gin.Context) {
	var req struct {
		ShopID          string      `json:"shopid" binding:"required"`
		ReportID        string      `json:"reportid" binding:"required"`
		ConditionGuid   string      `json:"condition_guid" binding:"required"`
		ReportName      string      `json:"report_name"`
		ConditionName   string      `json:"condition_name"`
		StartTime       interface{} `json:"start_time"`
		EndTime         interface{} `json:"end_time"`
		DurationSeconds int         `json:"duration_seconds"`
		Status          string      `json:"status"`
		RowCount        int         `json:"row_count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	err := h.postgreSQLService.UpsertProcessTime(c.Request.Context(), req.ShopID, req.ReportID, req.ConditionGuid, req.ReportName, req.ConditionName, req.StartTime, req.EndTime, req.DurationSeconds, req.RowCount, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// ProcessTimeGetHandler handles retrieving process time
func (h *APIHandler) ProcessTimeGetHandler(c *gin.Context) {
	var req struct {
		ShopID        string `json:"shopid" binding:"required"`
		ReportID      string `json:"reportid"`
		ConditionGuid string `json:"condition_guid"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if req.ConditionGuid != "" {
		// Get single record
		result, err := h.postgreSQLService.GetProcessTime(c.Request.Context(), req.ShopID, req.ReportID, req.ConditionGuid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "data": result})
	} else {
		// Get list
		results, err := h.postgreSQLService.GetProcessTimeList(c.Request.Context(), req.ShopID, req.ReportID, 50)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "data": results})
	}
}

// ProcessTimeDeleteHandler handles deleting process time
func (h *APIHandler) ProcessTimeDeleteHandler(c *gin.Context) {
	var req struct {
		ShopID        string `json:"shopid" binding:"required"`
		ReportID      string `json:"reportid" binding:"required"`
		ConditionGuid string `json:"condition_guid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	err := h.postgreSQLService.DeleteProcessTime(c.Request.Context(), req.ShopID, req.ReportID, req.ConditionGuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
