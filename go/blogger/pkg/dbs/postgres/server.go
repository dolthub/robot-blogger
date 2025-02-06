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

func NewPostgresLocallyRunningServer(ctx context.Context) (*postgresLocallyRunningServer, error) {
	return &postgresLocallyRunningServer{
		port:         5432,
		host:         "127.0.0.1",
		user:         "postgres",
		password:     "",
		databaseName: "robot_blogger_llama3_v1",
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

func (s *postgresLocallyRunningServer) createSchema(ctx context.Context, conn *pgx.Conn) error {
	// create metadata table if not exists
	_, err := conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS models (name text not null, version text not null, dimension int not null, created_at timestamp not null default current_timestamp, primary key (name, version))")
	if err != nil {
		return err
	}
	// create embeddings table if not exists
	_, err = conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS dolthub_blog_embeddings (id text PRIMARY KEY, model_name_fk text not null, model_version_fk text not null, embedding vector(4096) not null, content text not null, created_at timestamp not null default current_timestamp, foreign key (model_name_fk, model_version_fk) references models(name, version))")
	if err != nil {
		return err
	}
	return nil
}

func (s *postgresLocallyRunningServer) checkForVectorExtension(ctx context.Context, conn *pgx.Conn) error {
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

func (s *postgresLocallyRunningServer) Start(ctx context.Context) error {
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

	err = s.createSchema(ctx, conn)
	if err != nil {
		return err
	}

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
