package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dolthub/robot-blogger/go/blogger/pkg/dbs"
	"github.com/jackc/pgx/v5"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
)

type postgresLocallyRunningServer struct {
	db           *sql.DB
	port         int
	host         string
	user         string
	password     string
	databaseName string
	logger       *zap.Logger
}

var _ dbs.DatabaseServer = &postgresLocallyRunningServer{}

func NewPostgresLocallyRunningServer(ctx context.Context, logger *zap.Logger) (*postgresLocallyRunningServer, error) {
	return &postgresLocallyRunningServer{
		port:         5432,
		host:         "127.0.0.1",
		user:         "postgres",
		password:     "",
		databaseName: "robot_blogger_llama3_v1",
		logger:       logger,
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
	_, err = conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS dolthub_blog_embeddings (id text, model_name_fk text not null, model_version_fk text not null, embedding vector(4096) not null, content_md5 text not null, content text not null, created_at timestamp not null default current_timestamp, primary key(id, content_md5), foreign key (model_name_fk, model_version_fk) references models(name, version))")
	if err != nil {
		return err
	}
	return nil
}

func (s *postgresLocallyRunningServer) insertModelIfNotExists(ctx context.Context, conn *pgx.Conn, model string, version string, dimension int) error {
	_, err := conn.Exec(ctx, "INSERT INTO models (name, version, dimension) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", model, version, dimension)
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

func (s *postgresLocallyRunningServer) InsertModel(ctx context.Context, model string, version string, dimension int) error {
	start := time.Now()
	defer func() {
		s.logger.Info("postgres locally running server insert model", zap.String("model", model), zap.String("version", version), zap.Int("dimension", dimension), zap.Duration("duration", time.Since(start)))
	}()

	conn, err := s.newConn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	return s.insertModelIfNotExists(ctx, conn, model, version, dimension)
}

func (s *postgresLocallyRunningServer) InsertEmbedding(ctx context.Context, id, model, version, contentMd5, content string, embedding []float32) error {
	start := time.Now()
	defer func() {
		s.logger.Info("postgres locally running server insert embedding", zap.String("id", id), zap.String("model", model), zap.String("version", version), zap.Duration("duration", time.Since(start)))
	}()

	conn, err := s.newConn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	// check if embedding already exists
	res, err := conn.Query(ctx, "SELECT * FROM dolthub_blog_embeddings WHERE id = $1 and content_md5 = $2 and model_name_fk = $3 and model_version_fk = $4;", id, contentMd5, model, version)
	if err != nil {
		return err
	}
	defer res.Close()
	found := 0
	for res.Next() {
		found++
	}
	if found > 0 {
		return nil
	}

	_, err = conn.Exec(ctx, "INSERT INTO dolthub_blog_embeddings (id, model_name_fk, model_version_fk, embedding, content_md5, content) VALUES ($1, $2, $3, $4, $5, $6)", id, model, version, pgvector.NewVector(embedding), contentMd5, content)
	if err != nil {
		return err
	}

	return nil
}

type Result struct {
	id      string
	content string
}

func (s *postgresLocallyRunningServer) GetContentFromEmbeddings(ctx context.Context, embeddings []float32) (string, error) {
	start := time.Now()
	defer func() {
		s.logger.Info("postgres locally running server get content from embeddings", zap.Duration("duration", time.Since(start)))
	}()

	conn, err := s.newConn(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Close(ctx)

	results := make([]Result, 0)
	res, err := conn.Query(ctx, "SELECT id, content FROM dolthub_blog_embeddings ORDER BY embedding <-> $1 LIMIT 10", pgvector.NewVector(embeddings))
	if err != nil {
		return "", err
	}
	defer res.Close()

	for res.Next() {
		var result Result
		err = res.Scan(&result.id, &result.content)
		if err != nil {
			return "", err
		}
		results = append(results, result)
	}
	if res.Err() != nil {
		return "", res.Err()
	}

	combinedContent := ""
	for _, result := range results {
		s.logger.Info("postgres locally running server get content from embeddings using id:", zap.String("id", result.id))
		combinedContent += result.content + "\n\n"
	}

	return combinedContent, nil
}
