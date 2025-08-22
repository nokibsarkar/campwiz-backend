package cache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm/logger"
)

type SentryGinLogger struct {
	Logger                    logger.Interface
	LogLevel                  logger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	ParameterizedQueries      bool
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
			span1 := s.getChildSpan(parentSpan, sql)
			if span1 == nil {
				return sql, rowsAffected
			}
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

func (s *SentryGinLogger) getChildSpan(parentSpan *sentry.Span, sql string) *sentry.Span {
	if parentSpan == nil {
		return nil
	}
	op := "db.sql.execute"
	if strings.HasPrefix("SELECT ", strings.ToUpper(sql)) {
		op = "db.sql.query"
	}
	name := sql[:min(len(sql), 100)]
	childSpan := parentSpan.StartChild(op, sentry.WithDescription(sql))
	childSpan.Description = sql
	childSpan.Name = name
	childSpan.SetData("db.name", "campwiz")
	childSpan.SetData("db.system", "mariadb")
	return childSpan
}

type contextKey string

const (
	GRPC_HUB_KEY         contextKey = "sentry_hub"
	GRPC_TRANSACTION_KEY contextKey = "sentry_transaction"
)

func GetHubFromContext(ctx context.Context) *sentry.Hub {

	if ctx == nil {
		return nil
	} else if val := reflect.ValueOf(ctx); val.Kind() == reflect.Ptr && val.IsNil() {
		return nil
	}
	if val := ctx.Value(GRPC_HUB_KEY); val != nil {
		if hub, ok := val.(*sentry.Hub); ok {
			return hub
		}
	}
	return nil
}

func GRPCErrorToSpanStatus(err error) sentry.SpanStatus {
	if err == nil {
		return sentry.SpanStatusOK
	}
	if grpcErr, ok := status.FromError(err); ok {
		switch grpcErr.Code() {
		case codes.OK:
			return sentry.SpanStatusOK
		case codes.Canceled:
			return sentry.SpanStatusCanceled
		case codes.Unknown:
			return sentry.SpanStatusUnknown
		case codes.InvalidArgument:
			return sentry.SpanStatusInvalidArgument
		case codes.DeadlineExceeded:
			return sentry.SpanStatusDeadlineExceeded
		case codes.NotFound:
			return sentry.SpanStatusNotFound
		case codes.AlreadyExists:
			return sentry.SpanStatusAlreadyExists
		case codes.PermissionDenied:
			return sentry.SpanStatusPermissionDenied
		case codes.ResourceExhausted:
			return sentry.SpanStatusResourceExhausted
		case codes.FailedPrecondition:
			return sentry.SpanStatusFailedPrecondition
		case codes.Aborted:
			return sentry.SpanStatusAborted
		case codes.OutOfRange:
			return sentry.SpanStatusOutOfRange
		case codes.Unimplemented:
			return sentry.SpanStatusUnimplemented
		case codes.Internal:
			return sentry.SpanStatusInternalError
		case codes.Unavailable:
			return sentry.SpanStatusUnavailable
		case codes.DataLoss:
			return sentry.SpanStatusDataLoss
		case codes.Unauthenticated:
			return sentry.SpanStatusUnauthenticated
		default:
			// For any other status code, we return UnknownError
			return sentry.SpanStatusUnknown
		}
	}
	return sentry.SpanStatusUnknown
}
func GRPCSentryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res any, err error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	hub := GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	if client := hub.Client(); client != nil {
		client.SetSDKIdentifier("grpc-client")
	}
	transactionName := info.FullMethod
	transactionSource := sentry.SourceCustom
	traceparent, ok := ctx.Value(sentry.SentryTraceHeader).(string)
	if !ok {
		traceparent = ""
	}
	baggage, ok := ctx.Value(sentry.SentryBaggageHeader).(string)
	if !ok {
		baggage = ""
	}
	log.Printf("Sentry GRPC Trace ID: %s", traceparent)

	options := []sentry.SpanOption{
		sentry.ContinueTrace(hub, traceparent, baggage),
		sentry.WithOpName("grpc.server"),
		sentry.WithTransactionSource(transactionSource),
		sentry.WithSpanOrigin(sentry.SpanOrigin("auto.grpc.grpc")),
	}

	transaction := sentry.StartTransaction(
		sentry.SetHubOnContext(ctx, hub),
		fmt.Sprintf("%s %s", "GRPC Server", transactionName),
		options...,
	)

	transaction.SetData("grpc.request.method", info.FullMethod)
	hub.Scope().SetSpan(transaction)
	defer func() {
		if err != nil {
			transaction.Status = GRPCErrorToSpanStatus(err)
			transaction.SetData("grpc.response.error", err.Error())
			transaction.SetData("grpc.response.code", status.Code(err).String())
		} else {
			transaction.Status = sentry.SpanStatusOK
		}

		transaction.Finish()
	}()
	// ctx = sentry.SetHubOnContext(ctx, hub)
	ctx = context.WithValue(context.Background(), contextKey(GRPC_HUB_KEY), hub)
	ctx = context.WithValue(ctx, contextKey(sentry.SentryTraceHeader), traceparent)
	ctx = context.WithValue(ctx, contextKey(sentry.SentryBaggageHeader), baggage)
	log.Printf("Sentry GRPC Trace ID: %s", transaction.TraceID.String())
	res, err = handler(ctx, req)
	return res, err
}

func WithGRPCContext(ctx1 context.Context) context.Context {
	if ctx1 == nil {
		return ctx1
	}
	ctx, ok := ctx1.(*gin.Context)
	if !ok {
		return ctx1
	}
	tx := sentry.TransactionFromContext(ctx)
	var traceId string
	var baggage string
	if tx == nil || tx.TraceID.String() == "" {
		tx = sentry.StartTransaction(ctx, "grpc.client")

	}
	traceId = tx.TraceID.String()
	baggage = tx.ToBaggage()
	log.Printf("Sentry Trace ID: %s", traceId)
	ctx2 := context.WithValue(context.Background(), contextKey(sentry.SentryTraceHeader), traceId)
	ctx2 = context.WithValue(ctx2, contextKey(sentry.SentryBaggageHeader), baggage)
	return ctx2
}
