package main

import (
	"RIP/internal/app/ds"
	"RIP/internal/app/dsn"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	err = db.AutoMigrate(
		&ds.Load{},
		&ds.LoadOrder{},
		&ds.LoadLoadOrder{},
		&ds.User{},
	)
	if err != nil {
		panic("cant migrate db")
	}
}
