package repository

import (
	"fmt"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/query"

	"github.com/go-gorm/caches/v4"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func GetDB() (db *gorm.DB, close func()) {
	dsn := consts.Config.Database.Main.DSN
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Warn),
		Logger: logger.Default.LogMode(logger.Info),
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
	}
}
func GetDbWithoutDefaultTransaction() (db *gorm.DB, close func()) {
	dsn := consts.Config.Database.Main.DSN
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Warn),
		Logger:                 logger.Default.LogMode(logger.Info),
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
func GetTestDB() (db *gorm.DB, close func()) {
	dsn := consts.Config.Database.Main.TestDSN
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
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
func GetDBWithGen() (q *query.Query, close func()) {
	dsn := consts.Config.Database.Main.DSN
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
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
func InitGen() {
	g := gen.NewGenerator(gen.Config{
		OutPath: "query",                                                            // output path
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	})
	conn, close := GetDB()
	defer close()
	g.UseDB(conn)
	g.ApplyBasic(models.Project{}, models.User{}, models.Campaign{}, models.Round{}, models.Task{}, models.Role{}, models.Submission{}, models.Evaluation{})
	// Generate the code
	g.Execute()
}
func InitDB(testing bool) {
	conn, close := GetDB()
	if testing {
		conn, close = GetTestDB()
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
	fmt.Println((db))
	db.Commit()
}
