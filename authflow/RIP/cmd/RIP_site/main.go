package main

import (
	_ "RIP/docs"
	"RIP/internal/app/config"
	"RIP/internal/app/dsn"
	"RIP/internal/app/handler"
	redisclient "RIP/internal/app/redis"
	"RIP/internal/app/repository"
	"RIP/internal/pkg"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	router := gin.Default()

	// Загружаем конфигурацию
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	// Получаем строку подключения к PostgreSQL
	postgresString := dsn.FromEnv()
	fmt.Println("Postgres DSN:", postgresString)

	// Инициализируем репозиторий (Postgres + MinIO)
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

	// Создаём хендлер с репозиторием и конфигом
	// Создаём redis client
	redisCli, err := redisclient.New(context.Background(), conf.Redis)
	if err != nil {
		logrus.Fatalf("error initializing redis: %v", err)
	}

	hand := handler.NewHandler(rep, conf, redisCli)

	// Инициализируем приложение и запускаем сервер
	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
