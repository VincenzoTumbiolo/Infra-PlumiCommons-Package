package migrators

import (
	"embed"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	st "github.com/golang-migrate/migrate/v4/source/testing"
)

//go:embed testdata/migrations/*.sql
var migrations embed.FS

func TestIOFSEnvDriver(t *testing.T) {
	first, second := "it", "works!"
	os.Setenv("ENV1", first)
	os.Setenv("ENV2", second)

	d, err := iofs.New(migrations, "testdata/migrations")
	if err != nil {
		t.Fatal(err)
	}

	d = NewEnvDriver(d)
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

	t.Logf("Migration:\n%s", migration)

	assert.Equal(t, strings.Contains(string(migration), first), true)
	assert.Equal(t, strings.Contains(string(migration), second), true)

	escapedStr := "$escapeTest"
	assert.Equal(t, strings.Contains(string(migration), escapedStr), true)
}
