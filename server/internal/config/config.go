package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Conf holds the application configuration, making it accessible globally.
var Conf *Config

// Config struct is the top-level configuration structure.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig holds server-related settings.
type ServerConfig struct {
	Port string `mapstructure:"port"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

// LoggingConfig holds settings for the logger.
type LoggingConfig struct {
	Directory  string `mapstructure:"directory"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// setDefaults sets the default values for the configuration.
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "5050")

	// Database defaults
	v.SetDefault("database.host", "db")
	v.SetDefault("database.port", "5432")
	v.SetDefault("database.user", "user")
	v.SetDefault("database.password", "password")
	v.SetDefault("database.dbname", "crapp-db")

	// Logging defaults
	v.SetDefault("logging.directory", "logs")
	v.SetDefault("logging.max_size", 10)   // 10 MB
	v.SetDefault("logging.max_backups", 3) // Keep 3 backups
	v.SetDefault("logging.max_age", 7)     // 7 days
	v.SetDefault("logging.compress", true) // Compress old logs
}

// Init initializes the configuration with Viper.
func Init(projectRoot string, log *zap.Logger) error {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// --- File Configuration ---
	v.AddConfigPath(filepath.Join(projectRoot, "config")) // Search for config file in the current directory
	v.SetConfigName("config")                             // Name of config file (without extension)
	v.SetConfigType("yaml")                               // Type of config file

	// --- Environment Variable Binding ---
	v.SetEnvPrefix("CRAPP") // e.g., CRAPP_SERVER_PORT
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read the initial configuration from the file.
	// It's okay if the file doesn't exist; defaults and env vars will be used.
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal the config into our global Conf variable
	if err := v.Unmarshal(&Conf); err != nil {
		return fmt.Errorf("unable to decode config into struct: %w", err)
	}

	// Set up a watch for configuration changes for hot-reloading
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		log.Info("Configuration file changed, reloading.", zap.String("file", e.Name))
		if err := v.Unmarshal(&Conf); err != nil {
			log.Error("Error reloading configuration", zap.Error(err))
		}
	})

	log.Info("Configuration loaded successfully")
	return nil
}
