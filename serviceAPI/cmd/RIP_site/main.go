package main

import (
	"RIP/internal/app/config"
	"RIP/internal/app/dsn"
	"RIP/internal/app/handler"
	"RIP/internal/app/repository"
	"RIP/internal/pkg"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	router := gin.Default()

	// Загрузка конфигурации приложения
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	// Получение строки подключения к PostgreSQL
	postgresString := dsn.FromEnv()
	fmt.Println("Postgres DSN:", postgresString)

	// Инициализация репозитория с MinIO
	rep, err := repository.New(
		postgresString,
		conf.Minio.Endpoint,
		conf.Minio.AccessKey,
		conf.Minio.SecretKey,
		conf.Minio.Bucket,
	)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	// Создание хендлера с подключённым репозиторием
	hand := handler.NewHandler(rep)

	// Инициализация приложения и запуск сервера
	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
