package s3migrator

import (
	"embed"
	"io"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	st "github.com/golang-migrate/migrate/v4/source/testing"
)

const testMigrationJSON = `[
	{
		"action": "UPLOAD",
		"path": "/path/to/file",
		"filename": "file.mp4"
	},
	{
		"action": "DELETE",
		"filename": "file.mp4"
	}
]`

func TestActionsUnmarshaling(t *testing.T) {
	stmts, err := UnmarshalMigrationJSON([]byte(testMigrationJSON))
	if err != nil {
		t.Fatalf("failed to unmarshal migration JSON: %v", err)
	}

	assert.Equal(t, len(stmts), 2)
	assert.Equal(t, stmts[0].Action(), UploadAction)
	assert.Equal(t, stmts[1].Action(), DeleteAction)
}

//go:embed testdata/migrations/*.json
var migrations embed.FS

func TestIOFSDriver(t *testing.T) {
	d, err := iofs.New(migrations, "testdata/migrations")
	if err != nil {
		t.Fatal(err)
	}

	st.Test(t, d)

	v, err := d.First()
	if err != nil {
		t.Fatal(err)
	}

	r, _, err := d.ReadUp(v)
	if err != nil {
		t.Fatal(err)
	}

	migration, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	stmts, err := UnmarshalMigrationJSON(migration)
	if err != nil {
		t.Fatalf("failed to unmarshal migration JSON: %v", err)
	}

	assert.Equal(t, len(stmts), 1)
	assert.Equal(t, stmts[0].Action(), UploadAction)
}
