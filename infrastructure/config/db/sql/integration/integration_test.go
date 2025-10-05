package sql_integration

import (
	"context"
	"embed"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/config/assert"
	"github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/config/db/sql/pgclient"
	sqlu "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/config/db/sql/utils"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed schema.sql
var schemaFile embed.FS

var ctx = context.Background()
var db sqlx.ExtContext

func TestMain(m *testing.M) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	pgContainer := assert.MustRes(postgres.Run(ctx,
		"postgres:15.3-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	))
	connStr := assert.MustRes(pgContainer.ConnectionString(ctx, "sslmode=disable"))
	pool := assert.MustRes(pgclient.New(connStr, pgclient.DefaultConfig()))

	db = sqlx.NewDb(stdlib.OpenDBFromPool(pool), "pgx")
	schema := assert.MustRes(schemaFile.ReadFile("schema.sql"))
	assert.MustRes(db.ExecContext(ctx, string(schema)))

	// run tests
	exitVal := m.Run()

	// cleanup
	pool.Close()
	assert.Must(pgContainer.Terminate(ctx))

	os.Exit(exitVal)
}

func TestIntegrations(t *testing.T) {
	testRepo.Test(t, db)
}

func FuzzQueries(f *testing.F) {
	testRepo.Fuzz(f, db, sqlu.FuzzGenerator(
		sqlu.EnumConditionFuzzer[ModelType](),
	))
}
