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
	})

	if err != nil {
		panic("failed to connect database")
	}
	cachesPlugin := &caches.Caches{Conf: &caches.Config{
		Easer: true,
	}}
	// Use caches plugin
	db.Use(cachesPlugin)
	return db, func() {
		raw_db, err := db.DB()
		if err != nil {
			panic("failed to connect database")
		}
		raw_db.Close()
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
		raw_db.Close()
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
		// Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("failed to connect database")
	}
	return db, mock, func() {
		raw_db, err := db.DB()
		if err != nil {
			panic("failed to connect database")
		}
		raw_db.Close()
	}
}
func GetDBWithGen() (q *query.Query, close func()) {
	dsn := consts.Config.Database.Main.DSN
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
		raw_db.Close()
	}
}

func InitDB(testing bool) {
	conn, close, err := GetDB()
	if err != nil {
		panic("failed to connect database" + err.Error())
	}
	if testing {
		conn, _, close = GetTestDB()
		conn.Exec("DROP DATABASE IF EXISTS campwiz_test;")
		conn.Exec("CREATE DATABASE campwiz_test;")
		conn.Exec("USE campwiz_test;")
	}
	defer close()

	db := conn.Begin()
	// set character set to utf8mb4
	db.Exec("ALTER DATABASE campwiz CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci;")
	db.AutoMigrate(&models.Project{})
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Campaign{})
	db.AutoMigrate(&models.Round{})
	db.AutoMigrate(&models.Task{})
	db.AutoMigrate(&models.Role{})
	db.AutoMigrate(&models.Submission{})
	db.AutoMigrate(&models.Evaluation{})
	log.Println((db))
	db.Commit()
}
