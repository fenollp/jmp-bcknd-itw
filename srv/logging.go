package srv

import (
	"context"
	"os"

	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

// ctx keys types
type (
	requestIDKey struct{}
)

var baseLog *zap.Logger

// SetupLogging sets up logging
func SetupLogging() (err error) {
	if baseLog == nil {
		cfg := zap.NewProductionConfig()
		cfg.Level = map[string]zap.AtomicLevel{
			"debug": zap.NewAtomicLevelAt(zap.DebugLevel),
			"":      zap.NewAtomicLevelAt(zap.InfoLevel),
			"info":  zap.NewAtomicLevelAt(zap.InfoLevel),
			"warn":  zap.NewAtomicLevelAt(zap.WarnLevel),
			"error": zap.NewAtomicLevelAt(zap.ErrorLevel),
		}[os.Getenv("LOG_LEVEL")]
		baseLog, err = cfg.Build(
			zap.AddCaller(),
		)
	}
	return
}

// NewLogFromCtx attaches to the set up logging value
func NewLogFromCtx(ctx context.Context) *zap.Logger {
	l := baseLog
	if ctx != nil {
		var rqID string
		var ok bool
		if rqID, ok = ctx.Value(requestIDKey{}).(string); !ok {
			rqID = ksuid.New().String()
		}
		l = l.With(zap.String("", rqID))
	}
	return l
}
