package pkg

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

type DoltHasableVectorStore struct {
	db               *sql.DB
	connectionString string
	vs               vectorstores.VectorStore
}

func NewDoltHasableVectorStore(s vectorstores.VectorStore, connectionString string) (HasableVectorStore, error) {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}
	return &DoltHasableVectorStore{db: db, connectionString: connectionString, vs: s}, nil
}

var _ HasableVectorStore = &DoltHasableVectorStore{}

func (d *DoltHasableVectorStore) Has(ctx context.Context, metadata map[string]any) (bool, error) {
	whereQuerys := make([]string, 0)
	for k, v := range metadata {
		whereQuerys = append(whereQuerys, fmt.Sprintf("JSON_UNQUOTE(JSON_EXTRACT(langchain_dolt_embedding.cmetadata, '$.%s')) = '%s'", k, v))
	}
	whereQuery := strings.Join(whereQuerys, " AND ")
	if len(whereQuery) == 0 {
		whereQuery = "TRUE"
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM langchain_dolt_embedding WHERE %s", whereQuery)
	var count int
	err := d.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *DoltHasableVectorStore) AddDocuments(ctx context.Context, documents []schema.Document, opts ...vectorstores.Option) ([]string, error) {
	return d.vs.AddDocuments(ctx, documents, opts...)
}

func (d *DoltHasableVectorStore) SimilaritySearch(ctx context.Context, query string, k int, opts ...vectorstores.Option) ([]schema.Document, error) {
	return d.vs.SimilaritySearch(ctx, query, k, opts...)
}

func (d *DoltHasableVectorStore) Close() error {
	return d.db.Close()
}
