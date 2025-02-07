package dolt

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/dbs"
	"go.uber.org/zap"
)

type doltServer struct {
	port         int
	serverName   dbs.ServerName
	host         string
	user         string
	password     string
	databaseName string
	logger       *zap.Logger
}

var _ dbs.DatabaseServer = &doltServer{}

func NewDoltServer(ctx context.Context, logger *zap.Logger) (*doltServer, error) {
	return &doltServer{
		serverName:   dbs.Mysql,
		port:         3306,
		host:         "127.0.0.1",
		user:         "root",
		password:     "",
		databaseName: "robot_blogger_llama3_v2",
		logger:       logger,
	}, nil
}

func (s *doltServer) GetConnectionString() string {
	if s.password == "" {
		return fmt.Sprintf("mysql://%s@%s:%d/%s", s.user, s.host, s.port, s.databaseName)
	}
	return fmt.Sprintf("mysql://%s:%s@%s:%d/%s", s.user, s.password, s.host, s.port, s.databaseName)
}

func (s *doltServer) newDB() (*sql.DB, error) {
	return sql.Open("mysql", s.GetConnectionString())
}

func (s *doltServer) QueryContext(ctx context.Context, queryFunc dbs.QueryFunc, query string, args ...interface{}) error {
	db, err := s.newDB()
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return queryFunc(ctx, rows)
}

func (s *doltServer) ExecContext(ctx context.Context, query string, args ...interface{}) error {
	db, err := s.newDB()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.ExecContext(ctx, query, args)
	return err
}

func (s *doltServer) Name() dbs.ServerName {
	return s.serverName
}

func (s *doltServer) Start(ctx context.Context) error {
	db, err := s.newDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.PingContext(ctx)
}

func (s *doltServer) Stop(ctx context.Context) error {
	return nil
}
