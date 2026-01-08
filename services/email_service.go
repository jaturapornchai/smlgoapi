package services

import (
	"fmt"
	"log"

	"gopkg.in/mail.v2"
	"smlgoapi/config"
)

// EmailService handles sending emails via SMTP
type EmailService struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
}

// NewEmailService สร้าง EmailService จาก Config
func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{
		smtpHost:     cfg.SMTP.Host,
		smtpPort:     cfg.SMTP.Port,
		smtpUsername: cfg.SMTP.Username,
		smtpPassword: cfg.SMTP.Password,
		fromEmail:    cfg.SMTP.FromEmail,
		fromName:     cfg.SMTP.FromName,
	}
}

// SendEmailWithAttachment ส่ง Email พร้อม Attachment ผ่าน SMTP
func (s *EmailService) SendEmailWithAttachment(
	to []string, cc []string, bcc []string,
	subject, htmlContent, attachmentPath, senderName string,
) error {
	// Logging - ก่อนส่ง
	log.Printf("[EMAIL] Sending email to: %v", to)
	log.Printf("[EMAIL] Subject: %s", subject)
	log.Printf("[EMAIL] Attachment: %s", attachmentPath)
	log.Printf("[EMAIL] SMTP Host: %s:%d", s.smtpHost, s.smtpPort)

	// สร้าง Message
	m := mail.NewMessage()

	// ตั้งค่า Sender Name
	if senderName == "" {
		senderName = s.fromName
	}
	m.SetAddressHeader("From", s.fromEmail, senderName)

	// ตั้งค่า Recipients
	m.SetHeader("To", to...)
	if len(cc) > 0 {
		m.SetHeader("Cc", cc...)
	}
	if len(bcc) > 0 {
		m.SetHeader("Bcc", bcc...)
	}

	// ตั้งค่า Content
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlContent)

	// แนบไฟล์
	if attachmentPath != "" {
		m.Attach(attachmentPath)
	}

	// สร้าง SMTP Dialer และส่ง
	d := mail.NewDialer(s.smtpHost, s.smtpPort, s.smtpUsername, s.smtpPassword)

	err := d.DialAndSend(m)

	// Logging - ผลลัพธ์
	if err != nil {
		log.Printf("[EMAIL] Failed: %v", err)
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	log.Printf("[EMAIL] Success: sent to %v", to)
	return nil
}
