package dbs

import "context"

type noopDatabaseServer struct {
}

var _ DatabaseServer = &noopDatabaseServer{}

func NewNoopDatabaseServer() *noopDatabaseServer {
	return &noopDatabaseServer{}
}

func (s *noopDatabaseServer) QueryContext(ctx context.Context, queryFunc QueryFunc, query string, args ...interface{}) error {
	return nil
}

func (s *noopDatabaseServer) ExecContext(ctx context.Context, query string, args ...interface{}) error {
	return nil
}

func (s *noopDatabaseServer) Name() ServerName {
	return Noop
}

func (s *noopDatabaseServer) Start(ctx context.Context) error {
	return nil
}

func (s *noopDatabaseServer) Stop(ctx context.Context) error {
	return nil
}
