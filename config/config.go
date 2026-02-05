package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port      string
	Domain    string
	AppSecret string
	MacroAPI  string

	// Database
	DBDSN string

	// JWT Auth
	JWTSecret     string
	JWTExpiry     int // minutes
	RefreshExpiry int // minutes

	// Admin credentials (used for initial seed)
	AdminUsername string
	AdminPassword string

	// File uploads
	UploadDir       string
	MaxUploadSizeMB int

	// Image processing
	WebPQuality  int
	WebPLossless bool
	WebPExact    bool
}

func Load() *Config {
	dbDSN := getEnvFirst("DATABASE_URL", "DB_DSN")
	if dbDSN == "" {
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "eman")
		password := getEnv("DB_PASSWORD", "eman")
		name := getEnv("DB_NAME", "eman")
		sslmode := getEnv("DB_SSLMODE", "disable")
		timezone := getEnv("DB_TIMEZONE", "UTC")
		dbDSN = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
			host,
			user,
			password,
			name,
			port,
			sslmode,
			timezone,
		)
	}

	return &Config{
		Port:      getEnv("PORT", "8080"),
		Domain:    getEnv("MACRO_DOMAIN", "eman-riverside.vercel.app"),
		AppSecret: getEnv("MACRO_APP_SECRET", "zUHxHqwGhPcvy39QD2r3huFCnK3UuKW26C9E"),
		MacroAPI:  getEnv("MACRO_API_URL", "https://api.macroserver.uz"),

		// Database
		DBDSN: dbDSN,

		// JWT Auth
		JWTSecret:     getEnv("JWT_SECRET", "eman-super-secret-jwt-key-change-in-production"),
		JWTExpiry:     15,    // 15 minutes
		RefreshExpiry: 10080, // 7 days

		// Admin credentials
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),

		// File uploads
		UploadDir:       getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadSizeMB: getEnvInt("MAX_UPLOAD_SIZE_MB", 200),

		// Image processing
		WebPQuality:  getEnvInt("WEBP_QUALITY", 85),
		WebPLossless: getEnvBool("WEBP_LOSSLESS", false),
		WebPExact:    getEnvBool("WEBP_EXACT", false),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvFirst(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
