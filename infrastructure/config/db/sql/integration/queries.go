package sql_integration

import (
	sqlu "github.com/amplifon-x/ax-go-application-layer/v5/db/sql/utils"
	"github.com/google/uuid"
	"github.com/taleeus/sqld"
)

var testRepo = sqlu.TestRepository("integration")

var findModels = sqlu.Query[sqlu.Void, []Model]("findModels", `
SELECT *
FROM model
`, testRepo)

var findModel = sqlu.Query[FindByIDParams, Model]("findModel", `
SELECT *
FROM model
WHERE id = :id
`, testRepo)

var findModelSliceFilter = sqlu.Query[FindByIDsParams, Model]("findModelSliceFilter", `
SELECT *
FROM model
WHERE id IN (:ids)
`, testRepo)

var findModelSliceEmbeddedFilter = sqlu.Query[FindByIDsEmbeddedParams, Model]("findModelSliceEmbeddedFilter", `
SELECT *
FROM model
WHERE id IN (:ids)
`, testRepo)

var createModel = sqlu.Query[ModelParams, Model]("createModel", `
INSERT INTO model (name, typ, cat_uuid)
VALUES (:name, :typ, :cat_uuid)
RETURNING *
`, testRepo)

var createManyModels = sqlu.Query[[]ModelParams, sqlu.Void]("createManyModels", `
INSERT INTO model (name, typ, cat_uuid)
VALUES (:name, :typ, :cat_uuid)
`, testRepo)

var updateModel = sqlu.Query[UpsertModelParams, Model]("updateModel", `
UPDATE model
SET
	name = :name,
	typ = :typ,
	cat_uuid = :cat_uuid
WHERE id = :id
RETURNING *
`, testRepo)

var deleteModel = sqlu.Query[FindByIDParams, Model]("deleteModel", `
DELETE FROM model
WHERE id = :id
`, testRepo)

var createModelReturningUUID = sqlu.Query[UpsertModelParams, uuid.UUID]("createModelReturningUUID", `
INSERT INTO model (name, typ, cat_uuid)
VALUES (:name, :typ, :cat_uuid)
RETURNING cat_uuid
`, testRepo)

var createManyModelInfos = sqlu.Query[[]UpsertModelInfoParams, sqlu.Void]("createManyModelInfos", `
INSERT INTO model_info (model_id, info)
VALUES (:id, :info)
`, testRepo)

var findModelsFiltered = sqlu.LazyQuery[sqlu.Where[sqlu.Where[ModelFilter]], []ModelFull]("findModelsFiltered",
	func(where sqlu.Where[sqlu.Where[ModelFilter]]) (string, sqld.Params) {
		params := make(sqld.Params)
		query := `
		SELECT ` + sqlu.Join(sqlu.Extract[ModelFull]()) + `
		FROM model
		LEFT JOIN model_info ON model.id = model_info.model_id
		` + sqld.Where(where.Parse(&params))

		return query, params
	},
	testRepo,
)
