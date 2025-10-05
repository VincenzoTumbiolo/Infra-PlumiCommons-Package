package migrators

import (
	"io"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4/source"
)

type envDriver struct {
	source.Driver
}

// NewEnvDriver returns a new EnvDriver wrapper for an existing Driver
func NewEnvDriver(d source.Driver) source.Driver {
	return &envDriver{d}
}

func expandEnvReader(r io.ReadCloser) (io.ReadCloser, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	content := os.ExpandEnv(string(raw))
	content = strings.ReplaceAll(content, "$\\", "$")

	return io.NopCloser(strings.NewReader(content)), nil
}

// ReadUp is a facade of the underlying Driver implementation,
// which additionally expands all envs in the reader
func (d *envDriver) ReadUp(version uint) (io.ReadCloser, string, error) {
	r, identifier, err := d.Driver.ReadUp(version)
	if err != nil {
		return nil, identifier, err
	}
	defer r.Close()

	envR, err := expandEnvReader(r)
	if err != nil {
		return nil, identifier, err
	}

	return envR, identifier, nil
}

// ReadDown is a facade of the underlying Driver implementation,
// which additionally expands all envs in the reader
func (d *envDriver) ReadDown(version uint) (io.ReadCloser, string, error) {
	r, identifier, err := d.Driver.ReadDown(version)
	if err != nil {
		return nil, identifier, err
	}
	defer r.Close()

	envR, err := expandEnvReader(r)
	if err != nil {
		return nil, identifier, err
	}

	return envR, identifier, nil
}
