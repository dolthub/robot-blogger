package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dolthub/robot-blogger/go/blogger/pkg/dbs"
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

func (s *postgresLocallyRunningServer) newDB() (*sql.DB, error) {
	return sql.Open("postgres", s.GetConnectionString())
}

func (s *postgresLocallyRunningServer) Start(ctx context.Context) error {
	// todo: ping server to ensure it is running
	// todo: check pgvector extension is installed
	return nil
}

func (s *postgresLocallyRunningServer) Stop(ctx context.Context) error {
	return nil
}
