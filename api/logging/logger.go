package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type Loggers struct {
	App    *slog.Logger
	Audit  *slog.Logger
	Access *slog.Logger
}

func InitLoggers(logDir string) (*Loggers, error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// 1. Application Log (stdout + application.log)
	appFile, err := os.OpenFile(filepath.Join(logDir, "application.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return nil, fmt.Errorf("failed to open application.log: %w", err)
	}
	appWriter := io.MultiWriter(os.Stdout, appFile)
	appLogger := slog.New(slog.NewTextHandler(appWriter, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// 2. Audit Log (audit.log - tracks actions)
	auditFile, err := os.OpenFile(filepath.Join(logDir, "audit.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit.log: %w", err)
	}
	auditLogger := slog.New(slog.NewJSONHandler(auditFile, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// 3. Access Log (access.log - tracks incoming API traffic)
	accessFile, err := os.OpenFile(filepath.Join(logDir, "access.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return nil, fmt.Errorf("failed to open access.log: %w", err)
	}
	accessLogger := slog.New(slog.NewJSONHandler(accessFile, &slog.HandlerOptions{Level: slog.LevelInfo}))

	return &Loggers{
		App:    appLogger,
		Audit:  auditLogger,
		Access: accessLogger,
	}, nil
}
