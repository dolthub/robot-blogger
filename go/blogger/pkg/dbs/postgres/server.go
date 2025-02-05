package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dolthub/robot-blogger/go/blogger/pkg/dbs"
	"github.com/jackc/pgx/v5"
)

type postgresLocallyRunningServer struct {
	db           *sql.DB
	port         int
	host         string
	user         string
	password     string
	databaseName string
}

var _ dbs.DatabaseServer = &postgresLocallyRunningServer{}

func NewPostgresLocallyRunningServer(ctx context.Context, user, password, databaseName string) (*postgresLocallyRunningServer, error) {
	return &postgresLocallyRunningServer{
		port:         5432,
		host:         "127.0.0.1",
		user:         user,
		password:     password,
		databaseName: databaseName,
	}, nil
}

func (s *postgresLocallyRunningServer) GetConnectionString() string {
	if s.password == "" {
		return fmt.Sprintf("postgres://%s@%s:%d/%s", s.user, s.host, s.port, s.databaseName)
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", s.user, s.password, s.host, s.port, s.databaseName)
}

func (s *postgresLocallyRunningServer) newConn(ctx context.Context) (*pgx.Conn, error) {
	return pgx.Connect(ctx, s.GetConnectionString())
}

func (s *postgresLocallyRunningServer) Start(ctx context.Context) error {
	// todo: ping server to ensure it is running
	// todo: check pgvector extension is installed
	// todo: create embeddings table if not exists
	// todo: create metadata table if not exists
	return nil
}

func (s *postgresLocallyRunningServer) Stop(ctx context.Context) error {
	return nil
}

func (s *postgresLocallyRunningServer) Embed(ctx context.Context, input []float32) error {
	conn, err := s.newConn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	// todo: insert ignore embeddings into embeddings table
	return nil
}
