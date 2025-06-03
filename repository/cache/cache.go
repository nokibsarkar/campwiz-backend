package cache

import (
	"context"
	"fmt"
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func GetCacheDB(ctx context.Context) (db *gorm.DB, close func()) {
	dsn := consts.Config.Database.Cache.DSN
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: NewSentryGinLogger(logger.Default.LogMode(logger.Warn)),
	})
	if err != nil {
		log.Fatal("failed to connect cache database")
	}
	return db.WithContext(ctx), func() {
		raw_db, err := db.DB()
		if err != nil {
			log.Fatal("failed to get cache database on close")
		}
		if err := raw_db.Close(); err != nil {
			log.Fatal("failed to close cache database")
		}
	}
}
func GetTaskCacheDB(ctx context.Context, taskID models.IDType) (db *gorm.DB, close func()) {
	dsn := fmt.Sprintf(consts.Config.Database.Task.DSN, taskID)
	logMode := logger.Warn
	if consts.Config.Database.Task.Debug {
		logMode = logger.Info
	}
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: NewSentryGinLogger(logger.Default.LogMode(logMode)),
	})
	if err != nil {
		log.Fatal("failed to connect cache database")
	}
	if err := db.AutoMigrate(&Evaluation{}); err != nil {
		log.Fatal("failed to migrate cache database")
	}
	return db.WithContext(ctx), func() {
		raw_db, err := db.DB()
		if err != nil {
			log.Fatal("failed to get cache database on close")
		}
		if err := raw_db.Close(); err != nil {
			log.Fatal("failed to close cache database")
		}
		if err := os.Remove(dsn); err != nil {
			log.Fatal("failed to remove cache database")
		}
	}
}
func GetTestCacheDB() (db *gorm.DB, close func()) {
	dsn := consts.Config.Database.Cache.TestDSN
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: NewSentryGinLogger(logger.Default.LogMode(logger.Warn)),
	})
	if err != nil {
		log.Fatal("failed to connect cache database")
	}
	return db, func() {
		raw_db, err := db.DB()
		if err != nil {
			log.Fatal("failed to get cache database on close")
		}
		if err := raw_db.Close(); err != nil {
			log.Fatal("failed to close cache database")
		}
	}
}
func InitCacheDB(ctx context.Context, testing bool) {
	db, close := GetCacheDB(ctx)
	if testing {
		db, close = GetTestCacheDB()
	}
	defer close()
	if err := db.AutoMigrate(&Session{}); err != nil {
		log.Fatal("failed to migrate cache database")
	}
}
