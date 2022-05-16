package srv

import (
	"context"
	"os"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/stdlib"
	"go.uber.org/zap"
)

type pgClient struct {
	squirrel.StatementBuilderType
}

func (s *Server) newPGClient(ctx context.Context) *pgClient {
	log := NewLogFromCtx(ctx)
	start := time.Now()
	c := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).RunWith(s.pg)
	log.Debug("acquired postgres conn", zap.Duration("in", time.Since(start)))
	return &pgClient{c}
}

func (s *Server) setupPostgres(ctx context.Context) (err error) {
	log := NewLogFromCtx(ctx)
	start := time.Now()

	connConfig, err := pgx.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Error("", zap.Error(err))
		return
	}
	connConfig.Logger = zapadapter.NewLogger(log)

	s.pg = stdlib.OpenDB(*connConfig)

	log.Info("initiated postgres",
		zap.String("pg", connConfig.ConnString()),
		zap.Duration("in", time.Since(start)),
	)
	return
}
