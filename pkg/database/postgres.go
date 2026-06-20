package database

import (
	"log"

	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/configs"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectPostgres(cfg *configs.OrchestratorConfig) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("[database] failed to connect: %v", err)
	}
	log.Println("[database] connected to PostgreSQL")
	return db
}

func MigratePostgres(db *gorm.DB) {
	err := db.AutoMigrate(
		&domain.User{},
		&domain.UserSettings{},
		&domain.Device{},
		&domain.DeviceSession{},
	)
	if err != nil {
		log.Fatalf("[database] failed to migrate: %v", err)
	}
	log.Println("[database] PostgreSQL migration completed")
}
