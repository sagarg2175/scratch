package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppName           string
	AppEnv            string
	DBConnection      string
	TokenSymmetricKey string
	HttpUrl           string
	HttpPort          string
	DBHost            string
	DBPort            string
	DBdatabase        string
	DBUsername        string
	DBPassword        string
	JWTSecretKey      string
	JWTTokenDuration  time.Duration
}

func Load() Config {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	viper.AddConfigPath(".")
	viper.AddConfigPath("/app")

	c := Config{}
	//log := logger.New("DEBUG")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Print("failed to read config file")
		os.Exit(1)
	}

	if err := viper.Unmarshal(&c); err != nil {
		fmt.Print("not marshaled error")
	}

	durationStr := viper.GetString("jwt_token_duration")
	if durationStr == "" {
		durationStr = "1h" // default fallback
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		log.Fatalf("❌ Invalid JWT_TOKEN_DURATION in config: %v", err)
	}
	c.JWTTokenDuration = duration

	os.Setenv("APP_ENV", c.AppEnv)
	os.Setenv("DB_CONNECTION", c.DBConnection)
	os.Setenv("TOKEN_SYMMETRIC_KEY", c.TokenSymmetricKey)
	os.Setenv("HTTP_URL", c.HttpUrl)
	os.Setenv("HTTP_PORT", c.HttpPort)
	os.Setenv("DB_HOST", c.DBHost)
	os.Setenv("DB_PORT", c.DBPort)
	os.Setenv("DB_DATABASE", c.DBdatabase)
	os.Setenv("DB_USERNAME", c.DBUsername)
	os.Setenv("DB_PASSWORD", c.DBPassword)
	os.Setenv("JWT_SECRET_KEY", c.JWTSecretKey)
	os.Setenv("JWT_TOKEN_DURATION", durationStr)

	return c
}
