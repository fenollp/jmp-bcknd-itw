package srv

import (
	"context"
	"database/sql"
	"time"

	"go.uber.org/zap"
)

// Server holds connections to services.
type Server struct {
	pg *sql.DB
	rc *redisClient
}

// Close ends connections to services
func (s *Server) Close(ctx context.Context) {
	log := NewLogFromCtx(ctx)
	// Shutdown server's services here
	log.Info("closing postgres conn")
	s.pg.Close()
}

// NewServer opens connections to our services
func NewServer(ctx context.Context) (s *Server, err error) {
	log := NewLogFromCtx(ctx)
	start := time.Now()

	s = &Server{}
	// Start server's services here (PG, Redis, ...)

	if err = s.setupRedis(ctx); err != nil {
		return
	}

	if err = s.setupPostgres(ctx); err != nil {
		return
	}

	log.Info("server ready", zap.Duration("in", time.Since(start)))
	return
}
