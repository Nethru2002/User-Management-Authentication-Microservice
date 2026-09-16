package logger

import (
	"time"

	"go.uber.org/zap"
)

type AuditEvent string

const (
	EventRegisterSuccess AuditEvent = "AUTH_REGISTER_SUCCESS"
	EventLoginSuccess    AuditEvent = "AUTH_LOGIN_SUCCESS"
	EventLoginFailed     AuditEvent = "AUTH_LOGIN_FAILURE"
	EventTokenRefreshed  AuditEvent = "AUTH_TOKEN_REFRESH"
	EventSessionRevoked  AuditEvent = "AUTH_LOGOUT"
)

type AuditLogger struct {
	logger *zap.Logger
}

func NewAuditLogger(zapLogger *zap.Logger) *AuditLogger {
	return &AuditLogger{
		logger: zapLogger.Named("audit"),
	}
}

func (a *AuditLogger) Log(event AuditEvent, actorEmail, actorIP, detail string) {
	a.logger.Info("SECURITY_AUDIT_TRAIL",
		zap.String("event_type", string(event)),
		zap.String("actor_email", actorEmail),
		zap.String("actor_ip", actorIP),
		zap.String("detail", detail),
		zap.Time("timestamp", time.Now().UTC()),
	)
}