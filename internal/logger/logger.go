package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	configs "tektmud/internal/config"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Global logger instance
var globalLogger *Logger

type Logger struct {
	mainLogger *slog.Logger

	//Component-specific loggers
	gameLogger   *slog.Logger
	adminLogger  *slog.Logger
	systemLogger *slog.Logger
	errorLogger  *slog.Logger

	//File handles for rotation
	mainRotator  *lumberjack.Logger
	gameRotator  *lumberjack.Logger
	adminRotator *lumberjack.Logger
	errorRotator *lumberjack.Logger
}

func NewLogger() (*Logger, error) {
	c := configs.GetConfig()
	logPath := filepath.Join(c.Paths.RootDataDir, c.Paths.Logs)
	if err := os.MkdirAll(logPath, 0755); err != nil {
		return nil, err
	}

	logger := &Logger{}

	if err := logger.setupMainLogger(); err != nil {
		return nil, err
	}

	return logger, nil
}

func (l *Logger) setupMainLogger() error {
	c := configs.GetConfig()
	logPath := filepath.Join(c.Paths.RootDataDir, c.Paths.Logs, c.Logging.LogFile)
	l.mainRotator = &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    c.Logging.MaxSize,
		MaxBackups: c.Logging.MaxBackups,
		MaxAge:     c.Logging.MaxAge,
		Compress:   c.Logging.Compress,
	}

	var writers []io.Writer
	writers = append(writers, l.mainRotator)

	if c.Logging.EnableConsole {
		writers = append(writers, os.Stdout)
	}

	writer := io.MultiWriter(writers...)

	var handler slog.Handler
	if c.Logging.EnableStructured {
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level:     l.slogLevel(),
			AddSource: true,
		})
	} else {
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{
			Level:     l.slogLevel(),
			AddSource: true,
		})
	}

	l.mainLogger = slog.New(handler)
	return nil
}

// slogLevel converts our LogLevel to slog.Level
func (l *Logger) slogLevel() slog.Level {

	switch configs.GetConfig().Logging.Level {
	case configs.LevelDebug:
		return slog.LevelDebug
	case configs.LevelInfo:
		return slog.LevelInfo
	case configs.LevelWarn:
		return slog.LevelWarn
	case configs.LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// setupComponentLoggers creates specialized loggers for different components
func (l *Logger) setupComponentLoggers() error {
	c := configs.GetConfig()
	// Game logger (player actions, commands)
	if c.Logging.EnableGameLogs {
		l.gameRotator = &lumberjack.Logger{
			Filename:   filepath.Join(c.Paths.RootDataDir, c.Paths.Logs, "game.log"),
			MaxSize:    c.Logging.MaxSize / 2, // Smaller files for game logs
			MaxBackups: c.Logging.MaxBackups,
			MaxAge:     c.Logging.MaxAge,
			Compress:   c.Logging.Compress,
		}

		var gameWriters []io.Writer
		gameWriters = append(gameWriters, l.gameRotator)
		if c.Logging.EnableConsole {
			gameWriters = append(gameWriters, os.Stdout)
		}

		gameHandler := slog.NewJSONHandler(io.MultiWriter(gameWriters...), &slog.HandlerOptions{
			Level: l.slogLevel(),
		})
		l.gameLogger = slog.New(gameHandler).With("component", "game")
	}

	// Admin logger (administrative actions)
	if c.Logging.EnableAdminLogs {
		l.adminRotator = &lumberjack.Logger{
			Filename:   filepath.Join(c.Paths.RootDataDir, c.Paths.Logs, "admin.log"),
			MaxSize:    c.Logging.MaxSize / 4,    // Even smaller for admin logs
			MaxBackups: c.Logging.MaxBackups * 2, // Keep more admin logs
			MaxAge:     c.Logging.MaxAge * 2,     // Keep admin logs longer
			Compress:   c.Logging.Compress,
		}

		var adminWriters []io.Writer
		adminWriters = append(adminWriters, l.adminRotator)
		if c.Logging.EnableConsole {
			adminWriters = append(adminWriters, os.Stdout)
		}

		adminHandler := slog.NewJSONHandler(io.MultiWriter(adminWriters...), &slog.HandlerOptions{
			Level: slog.LevelInfo, // Always log admin actions
		})
		l.adminLogger = slog.New(adminHandler).With("component", "admin")
	}

	// Error logger (errors and panics)
	if c.Logging.EnableErrorLogs {
		l.errorRotator = &lumberjack.Logger{
			Filename:   filepath.Join(c.Paths.RootDataDir, c.Paths.Logs, "errors.log"),
			MaxSize:    c.Logging.MaxSize,
			MaxBackups: c.Logging.MaxBackups * 3, // Keep lots of error logs
			MaxAge:     c.Logging.MaxAge * 3,     // Keep error logs very long
			Compress:   c.Logging.Compress,
		}

		var errorWriters []io.Writer
		errorWriters = append(errorWriters, l.errorRotator)
		errorWriters = append(errorWriters, os.Stderr) // Always show errors on console

		errorHandler := slog.NewJSONHandler(io.MultiWriter(errorWriters...), &slog.HandlerOptions{
			Level:     slog.LevelWarn, // Warnings and above
			AddSource: true,
		})
		l.errorLogger = slog.New(errorHandler).With("component", "error")
	}

	// System logger uses main logger with component tag
	if c.Logging.EnableSystemLogs {
		l.systemLogger = l.mainLogger.With("component", "system")
	}

	return nil
}

// Main logging methods
func (l *Logger) Debug(msg string, args ...any) {
	if l.mainLogger != nil {
		l.mainLogger.Debug(msg, args...)
	}
}

func (l *Logger) Info(msg string, args ...any) {
	if l.mainLogger != nil {
		l.mainLogger.Info(msg, args...)
	}
}

func (l *Logger) Warn(msg string, args ...any) {
	if l.mainLogger != nil {
		l.mainLogger.Warn(msg, args...)
	}
}

func (l *Logger) Error(msg string, args ...any) {
	if l.mainLogger != nil {
		l.mainLogger.Error(msg, args...)
	}
	if l.errorLogger != nil {
		l.errorLogger.Error(msg, args...)
	}
}

// Game-specific logging methods
func (l *Logger) LogPlayerAction(playerID, playerName, action string, args ...any) {
	if l.gameLogger != nil {
		allArgs := append([]any{
			"player_id", playerID,
			"player_name", playerName,
			"action", action,
			"timestamp", time.Now(),
		}, args...)
		l.gameLogger.Info("Player action", allArgs...)
	}
}

func (l *Logger) LogPlayerConnect(playerID, playerName, ip string) {
	if l.gameLogger != nil {
		l.gameLogger.Info("Player connected",
			"player_id", playerID,
			"player_name", playerName,
			"ip", ip,
			"timestamp", time.Now(),
		)
	}
}

func (l *Logger) LogPlayerDisconnect(playerID, playerName string, reason string) {
	if l.gameLogger != nil {
		l.gameLogger.Info("Player disconnected",
			"player_id", playerID,
			"player_name", playerName,
			"reason", reason,
			"timestamp", time.Now(),
		)
	}
}

// Admin-specific logging methods
func (l *Logger) LogAdminAction(adminID, adminName, action string, target string, args ...any) {
	if l.adminLogger != nil {
		allArgs := append([]any{
			"admin_id", adminID,
			"admin_name", adminName,
			"action", action,
			"target", target,
			"timestamp", time.Now(),
		}, args...)
		l.adminLogger.Warn("Admin action", allArgs...) // Use Warn level for visibility
	}
}

func (l *Logger) LogRoomCreation(adminID, adminName, areaID, roomID, roomTitle string) {
	l.LogAdminAction(adminID, adminName, "create_room", roomID,
		"area_id", areaID,
		"room_title", roomTitle,
	)
}

func (l *Logger) LogAreaCreation(adminID, adminName, areaID, areaName string) {
	l.LogAdminAction(adminID, adminName, "create_area", areaID,
		"area_name", areaName,
	)
}

func (l *Logger) LogRoomEdit(adminID, adminName, areaID, roomID, field, oldValue, newValue string) {
	l.LogAdminAction(adminID, adminName, "edit_room", roomID,
		"area_id", areaID,
		"field", field,
		"old_value", oldValue,
		"new_value", newValue,
	)
}

// System logging methods
func (l *Logger) LogSystemStart(version string, config map[string]interface{}) {
	if l.systemLogger != nil {
		l.systemLogger.Info("MUD server starting",
			"version", version,
			"config", config,
			"timestamp", time.Now(),
		)
	}
}

func (l *Logger) LogSystemShutdown(reason string) {
	if l.systemLogger != nil {
		l.systemLogger.Warn("MUD server shutting down",
			"reason", reason,
			"timestamp", time.Now(),
		)
	}
}

func (l *Logger) LogTickStats(tickCount int64, queueSize int, playerCount int) {
	if l.systemLogger != nil && configs.GetConfig().Logging.Level == configs.LevelDebug {
		l.systemLogger.Debug("Tick statistics",
			"tick_count", tickCount,
			"queue_size", queueSize,
			"player_count", playerCount,
			"timestamp", time.Now(),
		)
	}
}

// Error logging with stack traces
func (l *Logger) LogError(err error, context string, args ...any) {
	if l.errorLogger != nil {
		allArgs := append([]any{
			"error", err.Error(),
			"context", context,
			"timestamp", time.Now(),
		}, args...)
		l.errorLogger.Error("Application error", allArgs...)
	}
}

func (l *Logger) LogPanic(recovered interface{}, stack []byte) {
	if l.errorLogger != nil {
		l.errorLogger.Error("Panic recovered",
			"panic", recovered,
			"stack", string(stack),
			"timestamp", time.Now(),
		)
	}
}

// Utility methods
func (l *Logger) Rotate() error {
	var errs []error

	if l.mainRotator != nil {
		if err := l.mainRotator.Rotate(); err != nil {
			errs = append(errs, err)
		}
	}

	if l.gameRotator != nil {
		if err := l.gameRotator.Rotate(); err != nil {
			errs = append(errs, err)
		}
	}

	if l.adminRotator != nil {
		if err := l.adminRotator.Rotate(); err != nil {
			errs = append(errs, err)
		}
	}

	if l.errorRotator != nil {
		if err := l.errorRotator.Rotate(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0] // Return first error
	}

	return nil
}

func (l *Logger) Close() error {
	// Lumberjack doesn't need explicit closing, but we could add cleanup here
	return nil
}

// InitGlobalLogger initializes the global logger
func InitGlobalLogger() error {
	logger, err := NewLogger()
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// Global logging functions for backward compatibility
func Debug(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Debug(msg, args...)
	}
}

func Info(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Info(msg, args...)
	}
}

func Warn(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Warn(msg, args...)
	}
}

func Error(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Error(msg, args...)
	}
}

func Printf(format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(fmt.Sprintf(format, v...))
	}
}

func Println(v ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(fmt.Sprint(v...))
	}
}

// Specialized logging functions
func LogPlayerAction(playerID, playerName, action string, args ...any) {
	if globalLogger != nil {
		globalLogger.LogPlayerAction(playerID, playerName, action, args...)
	}
}

func LogAdminAction(adminID, adminName, action string, target string, args ...any) {
	if globalLogger != nil {
		globalLogger.LogAdminAction(adminID, adminName, action, target, args...)
	}
}

func LogSystemStart(version string, config map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.LogSystemStart(version, config)
	}
}

func GetLogger() *Logger {
	return globalLogger
}
