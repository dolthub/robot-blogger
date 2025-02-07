package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/dbs"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type postgresServer struct {
	db           *sql.DB
	serverName   string
	port         int
	host         string
	user         string
	password     string
	databaseName string
	logger       *zap.Logger
}

func (s *postgresServer) Name() string {
	return s.serverName
}

var _ dbs.DatabaseServer = &postgresServer{}

func NewPostgresLocallyRunningServer(ctx context.Context, logger *zap.Logger) (*postgresServer, error) {
	return &postgresServer{
		serverName: dbs.Postgres,
		port:       5432,
		host:       "127.0.0.1",
		user:       "postgres",
		password:   "",
		//databaseName: "robot_blogger_llama3_v1", // this has full blog as content
		//databaseName: "robot_blogger_llama3_v2", // this has chunked blog as content
		databaseName: "robot_blogger_llama3_v3", // this is to test my refactor branch
		logger:       logger,
	}, nil
}

func (s *postgresServer) GetConnectionString() string {
	if s.password == "" {
		return fmt.Sprintf("postgres://%s@%s:%d/%s", s.user, s.host, s.port, s.databaseName)
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", s.user, s.password, s.host, s.port, s.databaseName)
}

func (s *postgresServer) newConn(ctx context.Context) (*pgx.Conn, error) {
	return pgx.Connect(ctx, s.GetConnectionString())
}

func (s *postgresServer) insertModelIfNotExists(ctx context.Context, conn *pgx.Conn, model string, version string, dimension int) error {
	_, err := conn.Exec(ctx, "INSERT INTO models (name, version, dimension) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", model, version, dimension)
	if err != nil {
		return err
	}
	return nil
}

func (s *postgresServer) checkForVectorExtension(ctx context.Context, conn *pgx.Conn) error {
	res, err := conn.Query(ctx, "SELECT * FROM pg_extension WHERE extname = 'vector';")
	if err != nil {
		return err
	}
	defer res.Close()
	found := 0
	for res.Next() {
		found++
	}
	if found == 0 {
		return fmt.Errorf("could not find vector extension")
	}
	return res.Err()
}

func (s *postgresServer) Start(ctx context.Context) error {
	conn, err := s.newConn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	err = conn.Ping(ctx)
	if err != nil {
		return err
	}

	err = s.checkForVectorExtension(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func (s *postgresServer) Stop(ctx context.Context) error {
	return nil
}

func (s *postgresServer) QueryContext(ctx context.Context, queryFunc dbs.QueryFunc, query string, args ...interface{}) error {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return queryFunc(ctx, rows)
}

func (s *postgresServer) ExecContext(ctx context.Context, query string, args ...interface{}) error {
	_, err := s.db.ExecContext(ctx, query, args)
	return err
}
