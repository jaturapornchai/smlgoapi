package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// LogLevel types
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	SUCCESS
	WARN
	ERROR
	FATAL
)

// Development Log Configuration
var (
	// Development default: แสดงทุก level
	CurrentLogLevel = DEBUG

	// Map string to LogLevel
	logLevelMap = map[string]LogLevel{
		"DEBUG":   DEBUG,
		"INFO":    INFO,
		"SUCCESS": SUCCESS,
		"WARN":    WARN,
		"ERROR":   ERROR,
		"FATAL":   FATAL,
	}
)

func init() {
	// Development mode: แสดง timestamp
	log.SetFlags(log.Ldate | log.Ltime)
	log.SetOutput(os.Stdout)

	// อ่าน LOG_LEVEL จาก environment (default: DEBUG for dev)
	logLevelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	if level, exists := logLevelMap[logLevelStr]; exists {
		CurrentLogLevel = level
	} else {
		CurrentLogLevel = DEBUG // Dev default: แสดงทุกอย่าง
	}

	log.Println("[LOGGER] Development mode - logging ENABLED with", GetLogLevel(), "level")
}

// getCallerInfo - ดึงข้อมูล file และ line ของ caller
func getCallerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}
	// แปลง path ให้เป็น forward slash และแสดง full path
	file = filepath.ToSlash(file)
	return fmt.Sprintf("%s:%d", file, line)
}

// shouldLog - ตรวจสอบว่าควรแสดง log หรือไม่
func shouldLog(level LogLevel) bool {
	return level >= CurrentLogLevel
}

// Debug - log เฉพาะตอน level >= DEBUG
func Debug(format string, args ...interface{}) {
	if !shouldLog(DEBUG) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: [DEBUG] "+format, append([]interface{}{caller}, args...)...)
}

// Info - log ข้อมูลทั่วไป
func Info(format string, args ...interface{}) {
	if !shouldLog(INFO) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: [INFO] "+format, append([]interface{}{caller}, args...)...)
}

// Success - log เมื่อสำเร็จ
func Success(format string, args ...interface{}) {
	if !shouldLog(SUCCESS) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: ✓ "+format, append([]interface{}{caller}, args...)...)
}

// Error - log error
func Error(format string, args ...interface{}) {
	if !shouldLog(ERROR) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: ❌ "+format, append([]interface{}{caller}, args...)...)
}

// Warn - log warning
func Warn(format string, args ...interface{}) {
	if !shouldLog(WARN) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: ⚠️  "+format, append([]interface{}{caller}, args...)...)
}

// Fatal - log error และหยุดโปรแกรม
func Fatal(format string, args ...interface{}) {
	caller := getCallerInfo(2)
	log.Fatalf("%s: 💥 "+format, append([]interface{}{caller}, args...)...)
}

// Printf - backward compatibility
func Printf(format string, args ...interface{}) {
	if !shouldLog(INFO) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: "+format, append([]interface{}{caller}, args...)...)
}

// SetLogLevel - เปลี่ยน log level ใน runtime
func SetLogLevel(levelStr string) {
	if level, exists := logLevelMap[strings.ToUpper(levelStr)]; exists {
		CurrentLogLevel = level
		log.Println("[LOGGER] Log level set to", levelStr)
	} else {
		log.Println("[LOGGER] Invalid log level:", levelStr, "keeping current level")
	}
}

// GetLogLevel - ดึง log level ปัจจุบัน
func GetLogLevel() string {
	for levelStr, level := range logLevelMap {
		if level == CurrentLogLevel {
			return levelStr
		}
	}
	return "UNKNOWN"
}

// InitDebugMode - reload DEBUG_MODE setting (legacy support)
func InitDebugMode() {
	logLevelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	if logLevelStr == "" {
		// Legacy compatibility: DEBUG_MODE=true -> DEBUG level, false -> INFO level
		if os.Getenv("DEBUG_MODE") == "true" {
			CurrentLogLevel = DEBUG
		} else {
			CurrentLogLevel = INFO // Default to INFO if DEBUG_MODE=false or not set
		}
	} else {
		// Use new LOG_LEVEL if provided
		if level, exists := logLevelMap[logLevelStr]; exists {
			CurrentLogLevel = level
		}
	}

	if CurrentLogLevel <= INFO {
		log.Println("[LOGGER] ⚙️  Debug mode ENABLED with", GetLogLevel(), "level")
	} else {
		log.Println("[LOGGER] ⚙️  Debug mode DISABLED (minimal logging) -", GetLogLevel(), "level")
	}
}
