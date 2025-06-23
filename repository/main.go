package repository

import (
	"context"
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/query"
	"nokib/campwiz/repository/cache"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/go-gorm/caches/v4"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// var memoryCache = &cache.MemoryCacher{}

var sentrylogger = logger.Default.LogMode(logger.Info) // cache.NewSentryGinLogger(logger.Default.LogMode(getLogMode(consts.Config.Database.Main.Debug || consts.Config.Server.Mode == "debug")))

func GetDB(ctx context.Context) (db *gorm.DB, close func(), err error) {
	dsn := consts.Config.Database.Main.DSN
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: sentrylogger,
		// PrepareStmt:            true,
		SkipDefaultTransaction: true,
		// DisableForeignKeyConstraintWhenMigrating: true,

	})
	ctxgin, ok := ctx.(*gin.Context)
	var hub *sentry.Hub
	if ok {
		hub = sentrygin.GetHubFromContext(ctxgin)
	} else {
		hub = cache.GetHubFromContext(ctx)
	}
	if err != nil {
		if hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelFatal)
				// will be tagged with my-tag="my value"
				sentry.CaptureException(err)
				scope.GetSpan().Description = "Failed to connect to Main Database"
			})
		}
		panic("failed to connect database")
	}
	cachesPlugin := &caches.Caches{Conf: &caches.Config{
		Easer: true,
		// Cacher: memoryCache,
	}}
	// Use caches plugin
	if err := db.Use(cachesPlugin); err != nil {
		log.Printf("failed to use caches plugin %s", err.Error())
		if hub != nil {
			// will be tagged with my-tag="my value"
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelFatal)
				scope.GetSpan().Description = "Failed to use caches plugin"
				sentry.CaptureException(err)
			})
		}
		return nil, nil, err
	}

	return db.WithContext(ctx), func() {
		raw_db, err := db.DB()

		if err != nil {
			// if hub != nil {
			// 	hub.WithScope(func(scope *sentry.Scope) {

			// 		scope.SetLevel(sentry.LevelFatal)
			// 		// will be tagged with my-tag="my value"
			// 		sentry.CaptureException(err)
			// 	})
			// }
			panic("failed to connect database")
		}

		if err := raw_db.Close(); err != nil {
			// if hub != nil {
			// 	hub.WithScope(func(scope *sentry.Scope) {
			// 		scope.SetLevel(sentry.LevelFatal)
			// 		sentry.CaptureException(err)
			// 	})
			// }

			// log.Printf("Trace ID: %s", hub.Scope().GetSpan().TraceID)
			log.Printf("failed to close database %s", err.Error())
		}

	}, nil
}
func GetDbWithoutDefaultTransaction(ctx context.Context) (db *gorm.DB, close func()) {
	dsn := consts.Config.Database.Main.DSN
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                 sentrylogger,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic("failed to connect database")
	}
	return db.WithContext(ctx), func() {
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
		Logger: cache.NewSentryGinLogger(logger.Default.LogMode(logger.Info)),
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
func GetDBWithGen(ctx1 context.Context) (q *query.Query, close func()) {
	dsn := consts.Config.Database.Main.DSN
	// logMode := logger.Warn
	close = func() {}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: sentrylogger,
	})
	if err != nil {
		return nil, close
	}
	q = query.Use(db.WithContext(ctx1))
	return q, func() {
		raw_db, err := db.DB()
		if err != nil {
			log.Printf("failed to connect database %s", err.Error())
			return
		}
		raw_db.Close() //nolint:errcheck
	}
}
func GetCommonsReplicaWithGen(ctx1 context.Context) (q *query.Query, close func()) {
	dsn := consts.Config.Database.Commons.DSN
	// logMode := logger.Warn
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: sentrylogger,
	})
	ctxgin, ok := ctx1.(*gin.Context)
	var hub *sentry.Hub
	if ok {
		hub = sentrygin.GetHubFromContext(ctxgin)
	} else {
		hub = cache.GetHubFromContext(ctx1)
	}
	if err != nil {
		if hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelError)
				scope.GetSpan().Description = "Failed to connect to Commons Replica Database"
				// will be tagged with my-tag="my value"
				sentry.CaptureException(err)
			})
		}
		panic("failed to connect database")
	}
	q = query.Use(db.WithContext(ctx1))
	return q, func() {
		raw_db, err := db.DB()
		if err != nil {
			if hub != nil {
				// will be tagged with my-tag="my value"
				hub.WithScope(func(scope *sentry.Scope) {
					scope.SetLevel(sentry.LevelError)
					scope.GetSpan().Description = "Failed to get Commons Replica Database"
					sentry.CaptureException(err)
				})
			}
			panic("failed to connect database")
		}
		if err := raw_db.Close(); err != nil {
			sentry.WithScope(func(scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelError)
				scope.GetSpan().Description = "Failed to close Commons Replica Database"
				// will be tagged with my-tag="my value"
				sentry.CaptureException(err)
			})
			log.Printf("failed to close database %s", err.Error())
		}
	}
}

func InitDB(ctx context.Context, testing bool) {
	conn, err := gorm.Open(mysql.Open(consts.Config.Database.Main.DSN), &gorm.Config{
		Logger: sentrylogger,
		// PrepareStmt:            true,
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("failed to connect database" + err.Error())
	}

	db := conn.WithContext(ctx).Begin()
	// set character set to utf8mb4
	db.Exec("ALTER DATABASE campwiz CHARACTER SET = utf8mb4 COLLATE = utf8mb4_bin;")
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
		Logger:                                   logger.Default.LogMode(logger.Info),
		PrepareStmt:                              true,
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
