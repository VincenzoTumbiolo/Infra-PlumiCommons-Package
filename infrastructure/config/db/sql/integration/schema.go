package sql_integration

import (
	"time"

	pgtypex "github.com/amplifon-x/ax-go-application-layer/v5/db/sql/pgclient/typex"
	"github.com/amplifon-x/ax-go-application-layer/v5/opt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ModelType string

const (
	ModelTypeGood ModelType = "good"
	ModelTypeMid  ModelType = "mid"
	ModelTypeBad  ModelType = "bad"
)

func (ModelType) Enumerate() []string {
	return []string{
		string(ModelTypeGood),
		string(ModelTypeMid),
		string(ModelTypeBad),
	}
}

type Model struct {
	ID        int                  `db:"id"`
	Name      string               `db:"name"`
	Type      opt.Value[ModelType] `db:"typ"`
	CatUUID   opt.Value[uuid.UUID] `db:"cat_uuid"`
	CreatedAt time.Time            `db:"created_at"`
}

func (Model) ModelName() string {
	return "model"
}

type ModelInfo struct {
	ModelID      int                        `db:"model_id"`
	Info         opt.Value[string]          `db:"info"`
	Tags         pgtypex.FlatArray[string]  `db:"tags"`
	Availability pgtypex.Range[pgtype.Int4] `db:"availability"`
	CreatedAt    time.Time                  `db:"created_at"`
}

func (ModelInfo) ModelName() string {
	return "model_info"
}

type ModelFull struct {
	Model      `db:"model"`
	*ModelInfo `db:"model_info"`
}
