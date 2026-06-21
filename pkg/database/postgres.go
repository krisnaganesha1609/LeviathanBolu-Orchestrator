package database

import (
	"context"
	"log"
	"time"

	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/configs"
	"github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectPostgres(cfg *configs.OrchestratorConfig) *gorm.DB {
	logLevel := logger.Info
	if cfg.IsProduction() {
		// Avoid flooding production logs (and leaking query params) with
		// every single SQL statement; only surface slow queries & errors.
		logLevel = logger.Warn
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("[database] failed to connect: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("[database] failed to access underlying *sql.DB: %v", err)
	}

	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxIdleTime(cfg.DBConnMaxIdle)
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLife)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		log.Fatalf("[database] failed to ping PostgreSQL: %v", err)
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
