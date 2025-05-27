package configs

// LogLevel represents log levels
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logs struct {
	// File settings
	LogDir     string `yaml:"log_dir"`
	LogFile    string `yaml:"log_file"`
	MaxSize    int    `yaml:"max_size"`    // megabytes
	MaxBackups int    `yaml:"max_backups"` // number of old log files to keep
	MaxAge     int    `yaml:"max_age"`     // days
	Compress   bool   `yaml:"compress"`    // compress old log files

	// Logging behavior
	Level            LogLevel `yaml:"level"`
	EnableConsole    bool     `yaml:"enable_console"`    // Also log to console
	EnableStructured bool     `yaml:"enable_structured"` // Use structured yaml logging

	// Component-specific logging
	EnableGameLogs   bool `yaml:"enable_game_logs"`   // Player actions, commands
	EnableAdminLogs  bool `yaml:"enable_admin_logs"`  // Admin actions
	EnableSystemLogs bool `yaml:"enable_system_logs"` // System events
	EnableErrorLogs  bool `yaml:"enable_error_logs"`  // Errors and panics
}

func (lc *Logs) Check() {
	if lc.LogDir == `` {
		lc.LogDir = `logs`
	}
	if lc.LogFile == `` {
		lc.LogFile = `tektmud.log`
	}
	if lc.MaxSize == 0 {
		lc.MaxSize = 100 //100 MB
	}
	if lc.MaxAge == 0 {
		lc.MaxAge = 30 // keep 30 days files
	}
	if lc.MaxBackups == 0 {
		lc.MaxBackups = 5 // Keep up to 5 files
	}
	if lc.Level == 0 {
		lc.Level = LevelInfo
	}
}
