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
	ctxGin, ok := ctx.(*gin.Context)
	if ok {
		parentSpan := sentrygin.GetHubFromContext(ctxGin).Scope().GetSpan()
		wrapFc := func() (string, int64) {
			sql, rowsAffected := fc()
			span1 := parentSpan.StartChild("db.sql.execute", sentry.WithDescription(sql))
			defer span1.Finish()
			span1.Description = sql
			span1.Name = sql[:min(len(sql), 100)]
			span1.StartTime = begin
			span1.EndTime = time.Now()
			span1.SetData("db.name", "campwiz")
			span1.SetData("db.system", "mariadb")
			if err != nil {
				span1.SetData("db.error", err.Error())
			}
			span1.SetData("db.active_record", rowsAffected)
			return sql, rowsAffected
		}
		s.Logger.Trace(ctx, begin, wrapFc, err)
	} else {
		s.Logger.Trace(ctx, begin, fc, err)
	}

}
