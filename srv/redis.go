package srv

import (
	"context"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	// neverExpire = 0 // don't use this
	verylongExpire = 3 * 30 * 24 * time.Hour // ~3 months //FIXME
	defaultExpire  = 24 * time.Hour
)

// Alias so as to not import version-specific package and risk messing up
const redisNil = redis.Nil

type redisClient struct {
	*redis.Client
}

func (s *Server) red(ctx context.Context) *redisClient {
	return &redisClient{s.rc.WithContext(ctx)}
}

func (s *Server) setupRedis(ctx context.Context) (err error) {
	log := NewLogFromCtx(ctx)
	redisHost := os.Getenv("REDIS_HOST")
	s.rc = &redisClient{redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "", // no password set
		DB:       0,  // use default DB
	})}
	log.Info("connecting to redis", zap.String("host", redisHost))
	start := time.Now()
	if _, err = s.red(ctx).Ping(ctx).Result(); err != nil {
		log.Error("", zap.Error(err))
		return
	}
	log.Info("connected to redis", zap.Duration("in", time.Since(start)))
	return
}
