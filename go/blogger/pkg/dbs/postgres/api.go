package postgres

import (
	"context"

	"github.com/dolthub/robot-blogger/go/blogger/pkg/dbs"
	"github.com/jackc/pgx/v5"
)

type postgresAPI struct {
	ps dbs.DatabaseServer
}

var _ dbs.DatabaseAPI = &postgresAPI{}

func NewPostgresAPI(ctx context.Context, ps dbs.DatabaseServer) *postgresAPI {
	return &postgresAPI{ps: ps}
}

func (p *postgresAPI) newConn(ctx context.Context) (*pgx.Conn, error) {
	return pgx.Connect(ctx, p.ps.GetConnectionString())
}

func (p *postgresAPI) Embed(ctx context.Context, input []float32) error {
	conn, err := p.newConn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	return nil
}
