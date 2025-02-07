package dolt

import (
	"go.uber.org/zap"
)

type doltLocallyRunningServer struct {
	port         int
	host         string
	user         string
	password     string
	databaseName string
	logger       *zap.Logger
}

//var _ dbs.DatabaseServer = &doltLocallyRunningServer{}
//
//func NewDoltLocallyRunningServer(ctx context.Context, logger *zap.Logger) (*doltLocallyRunningServer, error) {
//	return &doltLocallyRunningServer{
//		port:         3306,
//		host:         "127.0.0.1",
//		user:         "root",
//		password:     "",
//		databaseName: "robot_blogger_llama3_v1",
//		logger:       logger,
//	}, nil
//}
//
//func (s *doltLocallyRunningServer) GetConnectionString() string {
//	if s.password == "" {
//		return fmt.Sprintf("mysql://%s@%s:%d/%s", s.user, s.host, s.port, s.databaseName)
//	}
//	return fmt.Sprintf("mysql://%s:%s@%s:%d/%s", s.user, s.password, s.host, s.port, s.databaseName)
//}
//
//func (s *doltLocallyRunningServer) newDB() (*sql.DB, error) {
//	return sql.Open("mysql", s.GetConnectionString())
//}
//
//func (s *doltLocallyRunningServer) createSchema(ctx context.Context, db *sql.DB) error {
//	// create metadata table if not exists
//	_, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS models (name varchar(2048) not null, version varchar(2048) not null, dimension int not null, created_at timestamp not null default current_timestamp, primary key (name, version))")
//	if err != nil {
//		return err
//	}
//
//	// create embeddings table if not exists
//	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS dolthub_blog_embeddings (id varchar(2048) not null, model_name_fk varchar(2048) not null, model_version_fk varchar(2048) not null, embedding vector(4096) not null, content_md5 varchar(2048) not null, content longtext not null, created_at timestamp not null default current_timestamp, primary key(id, content_md5), foreign key (model_name_fk, model_version_fk) references models(name, version))")
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (s *doltLocallyRunningServer) insertModelIfNotExists(ctx context.Context, db *sql.DB, model string, version string, dimension int) error {
//	_, err := db.ExecContext(ctx, "INSERT IGNORE INTO models (name, version, dimension) VALUES (?, ?, ?)", model, version, dimension)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (s *doltLocallyRunningServer) Start(ctx context.Context) error {
//	return nil
//}
//
//func (s *doltLocallyRunningServer) Stop(ctx context.Context) error {
//	return nil
//}
//
//func (s *doltLocallyRunningServer) InsertModel(ctx context.Context, model string, version string, dimension int) error {
//	return nil
//}
//
//func (s *doltLocallyRunningServer) InsertEmbedding(ctx context.Context, id, model, version, contentMd5, content string, embedding []float32) error {
//	return nil
//}
//
//func (s *doltLocallyRunningServer) GetContentFromEmbeddings(ctx context.Context, embeddings []float32) (string, error) {
//	return "", nil
//}
