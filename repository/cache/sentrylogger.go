package cache

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/logger"
)

type SentryGinLogger struct {
	Logger logger.Interface
}

func NewSentryGinLogger(logger logger.Interface) *SentryGinLogger {
	return &SentryGinLogger{
		Logger: logger,
	}
}
func (s *SentryGinLogger) LogMode(level logger.LogLevel) logger.Interface {
	s.Logger = s.Logger.LogMode(level)
	return s
}
func (s *SentryGinLogger) Info(ctx context.Context, msg string, data ...any) {
	s.Logger.Info(ctx, msg, data...)
}
func (s *SentryGinLogger) Warn(ctx context.Context, msg string, data ...any) {
	s.Logger.Warn(ctx, msg, data...)
}
func (s *SentryGinLogger) Error(ctx context.Context, msg string, data ...any) {
	s.Logger.Error(ctx, msg, data...)
}
func (s *SentryGinLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	// Use pointer type assertion if you need gin.Context, otherwise remove this line if unused
	// ctx1, ok := ctx.(*gin.Context)
	// if !ok {
	// 	log.Printf("context is not *gin.Context")
	// }
	ctxGin, ok := ctx.(*gin.Context)
	if ok {
		parentSpan := sentrygin.GetHubFromContext(ctxGin)
		wrapFc := func() (string, int64) {
			sql, rowsAffected := fc()
			span := parentSpan.Scope().GetSpan().StartChild("db.sql.execute", sentry.WithDescription(sql))
			defer span.Finish()
			tx := span.GetTransaction()
			defer tx.Finish()
			tx.Description = sql
			tx.Name = sql[:min(len(sql), 100)]

			tx.StartTime = begin
			tx.EndTime = time.Now()
			tx.SetData("db.name", "campwiz")
			tx.SetData("db.system", "mariadb")
			tx.SetData("db.error", err)
			tx.SetData("db.active_record", rowsAffected)
			return sql, rowsAffected
		}
		s.Logger.Trace(ctx, begin, wrapFc, err)
	} else {
		s.Logger.Trace(ctx, begin, fc, err)
	}

}
