package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"
)

type Config struct {
	HTTPAddr     string
	DatabaseURL  string
	BaseURL      string
	LogLevel     string
	JWTSecret    string
	JWTAccessTTL time.Duration
}

const (
	defaultJWTAccessTTL = 24 * time.Hour
	minJWTSecretLength  = 32
)

func Load() (Config, error) {
	databaseConfig, err := LoadDatabaseConfig()
	if err != nil {
		return Config{}, err
	}

	httpAddr, err := requiredEnv("HTTP_ADDR")
	if err != nil {
		return Config{}, err
	}

	baseURL, err := requiredEnv("BASE_URL")
	if err != nil {
		return Config{}, err
	}

	logLevel, err := requiredEnv("LOG_LEVEL")
	if err != nil {
		return Config{}, err
	}

	jwtSecret, err := loadJWTSecret()
	if err != nil {
		return Config{}, err
	}

	jwtAccessTTL, err := loadJWTAccessTTL()
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTPAddr:     httpAddr,
		DatabaseURL:  databaseConfig.DSN(),
		BaseURL:      baseURL,
		LogLevel:     logLevel,
		JWTSecret:    jwtSecret,
		JWTAccessTTL: jwtAccessTTL,
	}, nil
}

type DatabaseConfig struct {
	URL      string
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func LoadDatabaseConfig() (DatabaseConfig, error) {
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		return DatabaseConfig{URL: databaseURL}, nil
	}

	host, err := requiredEnv("DB_HOST")
	if err != nil {
		return DatabaseConfig{}, err
	}

	port, err := requiredEnv("DB_PORT")
	if err != nil {
		return DatabaseConfig{}, err
	}

	user, err := requiredEnv("DB_USER")
	if err != nil {
		return DatabaseConfig{}, err
	}

	password, err := requiredEnv("DB_PASSWORD")
	if err != nil {
		return DatabaseConfig{}, err
	}

	name, err := requiredEnv("DB_NAME")
	if err != nil {
		return DatabaseConfig{}, err
	}

	sslMode, err := requiredEnv("DB_SSLMODE")
	if err != nil {
		return DatabaseConfig{}, err
	}

	return DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Name:     name,
		SSLMode:  sslMode,
	}, nil
}

func (c DatabaseConfig) DSN() string {
	if c.URL != "" {
		return c.URL
	}

	return BuildDatabaseURL(c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

func BuildDatabaseURL(host, port, user, password, name, sslMode string) string {
	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   net.JoinHostPort(host, port),
		Path:   name,
	}

	query := dsn.Query()
	query.Set("sslmode", sslMode)
	dsn.RawQuery = query.Encode()

	return dsn.String()
}

func requiredEnv(key string) (string, error) {
	if value := os.Getenv(key); value != "" {
		return value, nil
	}

	return "", fmt.Errorf("%w: %s", ErrMissingEnv, key)
}

func loadJWTSecret() (string, error) {
	secret, err := requiredEnv("JWT_SECRET")
	if err != nil {
		return "", err
	}

	if len(secret) < minJWTSecretLength {
		return "", fmt.Errorf("%w: JWT_SECRET must be at least %d characters", ErrInvalidEnv, minJWTSecretLength)
	}

	return secret, nil
}

func loadJWTAccessTTL() (time.Duration, error) {
	value := os.Getenv("JWT_ACCESS_TTL")
	if value == "" {
		return defaultJWTAccessTTL, nil
	}

	ttl, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%w: JWT_ACCESS_TTL", ErrInvalidEnv)
	}

	if ttl <= 0 {
		return 0, fmt.Errorf("%w: JWT_ACCESS_TTL must be positive", ErrInvalidEnv)
	}

	return ttl, nil
}

var (
	ErrMissingEnv = errors.New("missing required environment variable")
	ErrInvalidEnv = errors.New("invalid environment variable")
)
