package configs

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// AppEnv values.
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

type OrchestratorConfig struct {
	AppEnv  string // "development" | "production"
	AppPort string

	DBHost         string
	DBPort         string
	DBUser         string
	DBPass         string
	DBName         string
	DBSSLMode      string
	DBMaxOpenConns int
	DBMaxIdleConns int
	DBConnMaxIdle  time.Duration
	DBConnMaxLife  time.Duration

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	SessionTTL time.Duration

	AssistantName     string
	AssistantLanguage string // "id" or "en"

	LLMProvider string

	GeminiAPIKey string
	GeminiModel  string // default: gemini-2.5-flash

	OpenRouterAPIKey string
	OpenRouterModel  string // default: google/gemini-2.5-flash

	STTWorkerURL string
	TTSWorkerURL string
}

// LoadConfig loads configuration from environment variables.
//
// A .env file is loaded if present (handy for local development), but its
// absence is NOT fatal: in production the environment is normally injected
// by the container runtime (docker-compose env_file, systemd, k8s, etc.)
// without a physical .env file ever existing on disk.
func LoadConfig() (*OrchestratorConfig, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("[config] no .env file found, relying on process environment")
	}

	cfg := &OrchestratorConfig{
		AppEnv:  getEnv("APP_ENV", EnvDevelopment),
		AppPort: getEnv("APP_PORT", "8009"),

		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPass:         getEnv("DB_PASS", "password"),
		DBName:         getEnv("DB_NAME", "orchestrator"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		DBMaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
		DBConnMaxIdle:  time.Duration(getEnvAsInt("DB_CONN_MAX_IDLE_MIN", 5)) * time.Minute,
		DBConnMaxLife:  time.Duration(getEnvAsInt("DB_CONN_MAX_LIFE_MIN", 30)) * time.Minute,

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		SessionTTL: time.Duration(getEnvAsInt("SESSION_TTL_HOURS", 24)) * time.Hour,

		AssistantName:     getEnv("ASSISTANT_NAME", "LeviathanBolu"),
		AssistantLanguage: getEnv("ASSISTANT_LANGUAGE", "id"),

		LLMProvider: getEnv("LLM_PROVIDER", "gemini"),

		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
		GeminiModel:  getEnv("GEMINI_MODEL", "gemini-2.5-flash"),

		OpenRouterAPIKey: getEnv("OPENROUTER_API_KEY", ""),
		OpenRouterModel:  getEnv("OPENROUTER_MODEL", "google/gemini-2.5-flash"),

		STTWorkerURL: getEnv("STT_WORKER_URL", "ws://localhost:9001/stt"),
		TTSWorkerURL: getEnv("TTS_WORKER_URL", "ws://localhost:9002/tts"),
	}

	return cfg, nil
}

// IsProduction reports whether the app is running in production mode.
func (c *OrchestratorConfig) IsProduction() bool {
	return c.AppEnv == EnvProduction
}

func (c *OrchestratorConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName, c.DBSSLMode, "UTC",
	)
}

// RedisAddr returns the host:port address used by the redis client.
func (c *OrchestratorConfig) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvAsFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
