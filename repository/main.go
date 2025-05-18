package repository

import (
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/query"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-gorm/caches/v4"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// var memoryCache = &cache.MemoryCacher{}

func getLogMode(debug bool) logger.LogLevel {
	if debug {
		return logger.Info
	}
	return logger.Warn
}
func GetDB() (db *gorm.DB, close func(), err error) {
	dsn := consts.Config.Database.Main.DSN
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(getLogMode(consts.Config.Database.Main.Debug || consts.Config.Server.Mode == "debug")),
		// PrepareStmt:            true,
		SkipDefaultTransaction: true,
		// DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		panic("failed to connect database")
	}
	cachesPlugin := &caches.Caches{Conf: &caches.Config{
		Easer: true,
		// Cacher: memoryCache,
	}}
	// Use caches plugin
	if err := db.Use(cachesPlugin); err != nil {
		log.Printf("failed to use caches plugin %s", err.Error())
		return nil, nil, err
	}
	return db, func() {
		raw_db, err := db.DB()
		if err != nil {
			panic("failed to connect database")
		}
		if err := raw_db.Close(); err != nil {
			log.Printf("failed to close database %s", err.Error())
		}
	}, nil
}
func GetDbWithoutDefaultTransaction() (db *gorm.DB, close func()) {
	dsn := consts.Config.Database.Main.DSN
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(getLogMode(consts.Config.Database.Main.Debug || consts.Config.Server.Mode == "debug")),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic("failed to connect database")
	}
	return db, func() {
		raw_db, err := db.DB()
		if err != nil {
			panic("failed to connect database")
		}
		if err := raw_db.Close(); err != nil {
			log.Printf("failed to close database %s", err.Error())
		}
	}
}
func GetTestDB() (db *gorm.DB, mock sqlmock.Sqlmock, close func()) {
	mockDb, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("failed to open sqlmock database: %s", err)
	}
	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"version()"}).AddRow("5.7.31"))
	db, err = gorm.Open(mysql.New(mysql.Config{
		Conn: mockDb,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("failed to connect database")
	}
	return db, mock, func() {
		raw_db, err := db.DB()
		if err != nil {
			panic("failed to connect database")
		}
		raw_db.Close() //nolint:errcheck
	}
}
func GetDBWithGen() (q *query.Query, close func()) {
	dsn := consts.Config.Database.Main.DSN
	// logMode := logger.Warn
	close = func() {}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(getLogMode(consts.Config.Database.Main.Debug || consts.Config.Server.Mode == "debug")),
	})
	if err != nil {
		return nil, close
	}
	q = query.Use(db)
	return q, func() {
		raw_db, err := db.DB()
		if err != nil {
			log.Printf("failed to connect database %s", err.Error())
			return
		}
		raw_db.Close() //nolint:errcheck
	}
}
func GetCommonsReplicaWithGen() (q *query.Query, close func()) {
	dsn := consts.Config.Database.Commons.DSN
	// logMode := logger.Warn
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(getLogMode(consts.Config.Database.Cache.Debug || consts.Config.Server.Mode == "debug")),
	})
	if err != nil {
		panic("failed to connect database")
	}
	q = query.Use(db)
	return q, func() {
		raw_db, err := db.DB()
		if err != nil {
			panic("failed to connect database")
		}
		raw_db.Close() //nolint:errcheck
	}
}

func InitDB(testing bool) {
	conn, err := gorm.Open(mysql.Open(consts.Config.Database.Main.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		// PrepareStmt:            true,
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("failed to connect database" + err.Error())
	}

	db := conn.Begin()
	// set character set to utf8mb4
	db.Exec("ALTER DATABASE campwiz CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci;")
	err = db.AutoMigrate(&models.Project{}, &models.User{}, &models.Campaign{}, &models.Round{},
		&models.Task{}, &models.Role{}, &models.Submission{},
		&models.Evaluation{}, &models.TaskData{}, &models.Category{})
	if err != nil {
		log.Printf("failed to migrate database %s", err.Error())
		db.Rollback()
		return
	}
	db.Commit()
	conn, err = gorm.Open(mysql.Open(consts.Config.Database.Main.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		// PrepareStmt:            true,
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: false,
	})
	if err != nil {
		panic("failed to connect database" + err.Error())
	}
	db = conn.Begin()
	err = db.AutoMigrate(&models.Project{}, &models.User{}, &models.Campaign{}, &models.Round{},
		&models.Task{}, &models.Role{}, &models.Submission{},
		&models.Evaluation{}, &models.TaskData{}, &models.Category{})
	if err != nil {
		log.Printf("failed to migrate database %s", err.Error())
		db.Rollback()
		return
	}
	db.Commit()
}
