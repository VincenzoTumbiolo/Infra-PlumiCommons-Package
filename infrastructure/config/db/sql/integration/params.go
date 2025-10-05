package sql_integration

import (
	sqlu "github.com/amplifon-x/ax-go-application-layer/v5/db/sql/utils"
	"github.com/google/uuid"
	"github.com/taleeus/sqld"
)

type FindByIDParams struct {
	ID int8 `db:"id"`
}

type FindByIDsParams struct {
	IDs []int8 `db:"ids"`
}

type FindByIDsEmbeddedParams struct {
	FindByIDsParams
}

type ModelParams struct {
	Name    string    `db:"name"`
	Type    ModelType `db:"typ"`
	CatUUID uuid.UUID `db:"cat_uuid"`
}

type UpsertModelParams struct {
	FindByIDParams
	ModelParams
}

type ModelInfoParams struct {
	Info string `db:"info"`
}

type UpsertModelInfoParams struct {
	FindByIDParams
	ModelInfoParams
}

type ModelFilter struct {
	Name      sqlu.StringCondition
	Type      sqlu.EnumCondition[ModelType]
	CreatedAt sqlu.TimeCondition
}

func (f ModelFilter) Parse(params *sqld.Params) string {
	return sqld.And(
		f.Name.Parse(params, sqlu.ColumnFull[Model]("name")),
		f.Type.Parse(params, sqlu.Column[Model]("typ")),
		f.CreatedAt.Parse(params, "created_at"),
	)
}
