package config

import (
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// MinioConfig хранит настройки для MinIO
type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

// Config хранит настройки сервиса
type Config struct {
	ServiceHost string
	ServicePort int
	Minio       MinioConfig
}

// NewConfig загружает и возвращает конфигурацию приложения
func NewConfig() (*Config, error) {
	var err error

	// Подгрузка переменных окружения из .env
	_ = godotenv.Load()

	configName := "config"
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	// Настройка Viper
	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	// Чтение конфигурации
	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	// Преобразование данных в структуру Config
	cfg := &Config{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	log.Info("Config successfully parsed")

	return &Config{
		ServiceHost: cfg.ServiceHost,
		ServicePort: cfg.ServicePort,
		Minio: MinioConfig{
			Endpoint:  os.Getenv("MINIO_ENDPOINT"), // только хост:порт
			AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
			SecretKey: os.Getenv("MINIO_SECRET_KEY"),
			Bucket:    os.Getenv("MINIO_BUCKET"),
		},
	}, nil
}
