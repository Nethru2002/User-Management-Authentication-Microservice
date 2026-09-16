package utils

import (
	"fmt"
	"go.uber.org/zap"
)

type EmailSender interface {
	SendVerificationEmail(to, token string) error
	SendPasswordResetEmail(to, token string) error
}

type ConsoleEmailSender struct {
	logger *zap.Logger
}

func NewConsoleEmailSender(logger *zap.Logger) EmailSender {
	return &ConsoleEmailSender{logger: logger}
}

func (c *ConsoleEmailSender) SendVerificationEmail(to, token string) error {
	link := fmt.Sprintf("http://localhost:8080/api/v1/auth/verify-email?token=%s", token)
	c.logger.Info("DISPATCHING_VERIFICATION_EMAIL",
		zap.String("recipient", to),
		zap.String("verification_link", link),
	)
	return nil
}

func (c *ConsoleEmailSender) SendPasswordResetEmail(to, token string) error {
	link := fmt.Sprintf("http://localhost:8080/api/v1/auth/password/reset?token=%s", token)
	c.logger.Info("DISPATCHING_PASSWORD_RESET_EMAIL",
		zap.String("recipient", to),
		zap.String("reset_link", link),
	)
	return nil
}