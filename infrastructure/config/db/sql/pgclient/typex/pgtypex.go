// This package provides wrappers to make the std scanning
// system work with [pgtype].
//
// This package will become deprecated when [the relative CL]
// will be allowed in the next Go release (1.25, I think)
//
// [the relative CL]: https://go-review.googlesource.com/c/go/+/588435
package pgtypex

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

var PgxMapper = pgtype.NewMap()

func scan[pgT any](src any) (pgT, error) {
	var original pgT
	if err := PgxMapper.SQLScanner(&original).Scan(src); err != nil {
		return original, err
	}

	return original, nil
}

func value[pgT any](val pgT, buf []byte) (driver.Value, error) {
	typ, ok := PgxMapper.TypeForValue(val)
	if !ok {
		return nil, fmt.Errorf("pgx type not found for %T", val)
	}

	buf, err := PgxMapper.Encode(typ.OID, pgtype.TextFormatCode, val, buf)
	return string(buf), err
}

type FlatArray[T any] pgtype.FlatArray[T]

func (arr *FlatArray[T]) Scan(src any) error {
	val, err := scan[pgtype.FlatArray[T]](src)
	if err != nil {
		return err
	}

	*arr = (FlatArray[T])(val)
	return nil
}

func (arr FlatArray[T]) Value() (driver.Value, error) {
	buf := make([]byte, 0, 64*len(arr))
	return value(pgtype.FlatArray[T](arr), buf)
}

func (rng *FlatArray[T]) UnmarshalJSON(data []byte) error {
	return rng.Scan(string(data))
}

type Range[T any] pgtype.Range[T]

func (rng *Range[T]) Scan(src any) error {

	if src == nil {
		*rng = Range[T]{}
		return nil
	}
	if src == "empty" {
		*rng = Range[T]{}
		return nil
	}
	if s, ok := src.(string); ok && strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		src = strings.Trim(s, "\"")
	}

	val, err := scan[pgtype.Range[T]](src)
	if err != nil {
		return err
	}

	*rng = (Range[T])(val)
	return nil
}

func (rng Range[T]) Value() (driver.Value, error) {
	buf := make([]byte, 0, 64)
	return value(pgtype.Range[T](rng), buf)
}

func (rng *Range[T]) UnmarshalJSON(data []byte) error {
	return rng.Scan(string(data))
}
