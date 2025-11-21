package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	brevo "github.com/getbrevo/brevo-go/lib"
)

type EmailService struct {
	client *brevo.APIClient
	apiKey string
}

func NewEmailService(apiKey string) *EmailService {
	cfg := brevo.NewConfiguration()
	cfg.AddDefaultHeader("api-key", apiKey)
	client := brevo.NewAPIClient(cfg)
	return &EmailService{
		client: client,
		apiKey: apiKey,
	}
}

func (s *EmailService) SendEmailWithAttachment(to []string, cc []string, bcc []string, subject, htmlContent, attachmentPath, senderName string) error {
	// Read attachment file
	fileContent, err := os.ReadFile(attachmentPath)
	if err != nil {
		return fmt.Errorf("failed to read attachment file: %w", err)
	}

	// Encode content to base64
	encodedContent := base64.StdEncoding.EncodeToString(fileContent)
	fileName := filepath.Base(attachmentPath)

	// Create attachment object
	attachment := brevo.SendSmtpEmailAttachment{
		Name:    fileName,
		Content: encodedContent,
	}

	// Create sender with custom name
	if senderName == "" {
		senderName = "SML Email Service"
	}
	sender := &brevo.SendSmtpEmailSender{
		Name:  senderName,
		Email: "noreply@bcaccount.com",
	}

	// Create recipients
	toList := make([]brevo.SendSmtpEmailTo, len(to))
	for i, email := range to {
		toList[i] = brevo.SendSmtpEmailTo{Email: email}
	}

	ccList := make([]brevo.SendSmtpEmailCc, len(cc))
	for i, email := range cc {
		ccList[i] = brevo.SendSmtpEmailCc{Email: email}
	}

	bccList := make([]brevo.SendSmtpEmailBcc, len(bcc))
	for i, email := range bcc {
		bccList[i] = brevo.SendSmtpEmailBcc{Email: email}
	}

	// Create email content
	emailContent := brevo.SendSmtpEmail{
		Sender:      sender,
		To:          toList,
		Cc:          ccList,
		Bcc:         bccList,
		Subject:     subject,
		HtmlContent: htmlContent,
		Attachment:  []brevo.SendSmtpEmailAttachment{attachment},
	}

	// Send email
	_, resp, err := s.client.TransactionalEmailsApi.SendTransacEmail(context.Background(), emailContent)
	if err != nil {
		return fmt.Errorf("failed to send email: %w (response: %v)", err, resp)
	}

	return nil
}
