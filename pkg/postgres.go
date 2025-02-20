package pkg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

type PostgresHasableVectorStore struct {
	conn             *pgx.Conn
	connectionString string
	vs               vectorstores.VectorStore
}

func NewPostgresHasableVectorStore(s vectorstores.VectorStore, connectionString string) (*PostgresHasableVectorStore, error) {
	conn, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		return nil, err
	}
	return &PostgresHasableVectorStore{conn: conn, connectionString: connectionString, vs: s}, nil
}

var _ HasableVectorStore = &PostgresHasableVectorStore{}

func (d *PostgresHasableVectorStore) Has(ctx context.Context, metadata map[string]any) (bool, error) {
	whereQuerys := make([]string, 0)
	for k, v := range metadata {
		whereQuerys = append(whereQuerys, fmt.Sprintf("(langchain_pg_embedding ->> '%s') = '%s'", k, v))
	}
	whereQuery := strings.Join(whereQuerys, " AND ")
	if len(whereQuery) == 0 {
		whereQuery = "TRUE"
	}
	query := fmt.Sprintf("SELECT COUNT(*) FROM langchain_pg_embedding WHERE %s", whereQuery)
	var count int
	err := d.conn.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *PostgresHasableVectorStore) AddDocuments(ctx context.Context, documents []schema.Document, opts ...vectorstores.Option) ([]string, error) {
	return d.vs.AddDocuments(ctx, documents, opts...)
}

func (d *PostgresHasableVectorStore) SimilaritySearch(ctx context.Context, query string, k int, opts ...vectorstores.Option) ([]schema.Document, error) {
	return d.vs.SimilaritySearch(ctx, query, k, opts...)
}

func (d *PostgresHasableVectorStore) Close() error {
	return d.conn.Close(context.Background())
}
