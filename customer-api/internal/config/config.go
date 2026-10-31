package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	_ "smswhiteinfo/customer-api/internal/pgcompat"
)

type Config struct {
	Address          string
	Environment      string
	JWTSecret        []byte
	FrontendOrigin   string
	TokenTTL         time.Duration
	RememberTokenTTL time.Duration
	CachePrefix      string
	Database         mysql.Config
	DBDriver         string
	DBDSN            string
	RabbitMQURL      string
}

func Load() (Config, error) {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	tokenTTL, err := duration("CUSTOMER_API_TOKEN_TTL", "8h")
	if err != nil {
		return Config{}, err
	}

	rememberTokenTTL, err := duration("CUSTOMER_API_REMEMBER_TOKEN_TTL", "720h")
	if err != nil {
		return Config{}, err
	}

	secret := os.Getenv("CUSTOMER_API_JWT_SECRET")
	if len(secret) < 32 {
		return Config{}, fmt.Errorf("required environment variable CUSTOMER_API_JWT_SECRET is missing or under 32 characters")
	}

	for _, reqKey := range []string{"DB_HOST", "DB_DATABASE", "DB_USERNAME", "APP_KEY"} {
		if os.Getenv(reqKey) == "" {
			return Config{}, fmt.Errorf("required environment variable %s is missing", reqKey)
		}
	}

	databaseConfig := mysql.NewConfig()
	databaseConfig.User = required("DB_USERNAME")
	databaseConfig.Passwd = os.Getenv("DB_PASSWORD")
	databaseConfig.Net = "tcp"
	databaseConfig.Addr = fmt.Sprintf("%s:%s", required("DB_HOST"), value("DB_PORT", "3306"))
	databaseConfig.DBName = required("DB_DATABASE")
	databaseConfig.ParseTime = true
	databaseConfig.Loc = time.FixedZone("Asia/Dhaka", 6*60*60)
	databaseConfig.Params = map[string]string{"charset": "utf8mb4"}

	dbConn := strings.ToLower(value("DB_CONNECTION", "mysql"))
	dbDriver := "mysql"
	var dbDSN string

	if dbConn == "pgsql" || dbConn == "postgres" || dbConn == "postgresql" {
		dbDriver = "pg_compat"
		dbPort := value("DB_PORT", "5432")
		dbDSN = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable search_path=%s,public",
			required("DB_HOST"), dbPort, required("DB_USERNAME"), os.Getenv("DB_PASSWORD"), required("DB_DATABASE"), required("DB_DATABASE"))
	} else {
		dbDriver = "mysql"
		dbDSN = databaseConfig.FormatDSN()
	}

	rabbitURL := value("RABBITMQ_URL", "")

	return Config{
		Address:          value("CUSTOMER_API_ADDR", "127.0.0.1:8082"),
		Environment:      value("CUSTOMER_API_ENV", "development"),
		JWTSecret:        []byte(secret),
		FrontendOrigin:   value("CUSTOMER_API_FRONTEND_ORIGIN", "http://localhost:5173"),
		TokenTTL:         tokenTTL,
		RememberTokenTTL: rememberTokenTTL,
		CachePrefix:      value("CACHE_PREFIX", ""),
		Database:         *databaseConfig,
		DBDriver:         dbDriver,
		DBDSN:            dbDSN,
		RabbitMQURL:      rabbitURL,
	}, nil
}

func duration(key string, fallback string) (time.Duration, error) {
	parsed, err := time.ParseDuration(value(key, fallback))
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}

	return parsed, nil
}

func required(key string) string {
	return os.Getenv(key)
}

func value(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
