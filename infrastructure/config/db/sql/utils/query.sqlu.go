package sqlu

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/amplifon-x/ax-go-application-layer/v5/assert"
	"github.com/amplifon-x/ax-go-application-layer/v5/slicex"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/taleeus/sqld"
)

// Page contains the list of items with useful page data
type Page[T any] struct {
	Items      T    `json:"items"`
	Total      uint `json:"total"`
	PagesCount uint `json:"pagesCount"`
	PageIndex  uint `json:"pageIndex"`
}

// Pagination instructs how to query a Page
type Pagination struct {
	Size uint `json:"size" db:"size" query:"size"`
	Skip uint `json:"skip" db:"skip" query:"skip"`
}

func buildPage[O any](items O, total uint, pagination Pagination) Page[O] {
	if pagination.Size == 0 {
		pagination.Size = total
	}

	var pagesCount uint
	if pagination.Size > 0 {
		pagesCount = total / pagination.Size
	}

	if total%pagination.Size > 0 {
		pagesCount++
	}

	var pageIndex uint
	if pagination.Size > 0 {
		pageIndex = pagination.Skip / pagination.Size
	}

	if pagination.Skip%pagination.Size > 0 {
		pageIndex++
	}

	return Page[O]{
		Items:      items,
		Total:      total,
		PagesCount: pagesCount,
		PageIndex:  pageIndex,
	}
}

// Sorting is just an alias of any.
//
//   - It should be a struct
//   - Only the fields ot type `sqld.Sorting` and with a `db` tag will be used
type Sorting any

func parseSorting(sorting Sorting) string {
	if sorting == nil {
		return ""
	}

	sortingValue := reflect.Indirect(reflect.ValueOf(sorting))
	assert.Check(sortingValue.Kind() == reflect.Struct)

	sortBuilder := strings.Builder{}
	sortingType := reflect.TypeOf(sqld.ASC)

	multipleSorts := false
	for i := 0; i < sortingValue.NumField(); i++ {
		fieldVal := sortingValue.Field(i)
		field := sortingValue.Type().Field(i)

		tag, ok := field.Tag.Lookup("db")
		if !ok || fieldVal.IsZero() || !reflectx.Deref(fieldVal.Type()).AssignableTo(sortingType) {
			continue
		}

		if multipleSorts {
			sortBuilder.WriteString(",\n\t")
		}

		sortingOrder := reflect.Indirect(fieldVal).Convert(reflect.TypeOf("")).String()
		sortBuilder.WriteString(tag + " " + sortingOrder)

		multipleSorts = true
	}

	return sortBuilder.String()
}

// Void indicates a void generic parameter.
// Use it to indicate if the query has "void" parameters or result
type Void struct{}

// Queryable describes a query with its input and output
type Queryable[I, O any] interface {
	// Name of the query
	Name() string
	// Run the query
	Run(context.Context, sqlx.ExtContext, ...I) (O, error)
	// Paged runs the query and returns a paged result
	Paged(context.Context, sqlx.ExtContext, Pagination, Sorting, ...I) (Page[O], error)
}

type queryError struct {
	inner error
	query string
	args  []any
}

func QueryError(err error, query string, args ...any) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%w (inner: %w)", queryError{err, query, args}, err)
}

func (err queryError) Error() string {
	return fmt.Sprintf("%v\nquery: %s", err.inner, err.query)
}

func (err queryError) Query() string {
	return err.query
}

func (err queryError) Args() []any {
	return err.args
}

type resultType int

const (
	singleResultType resultType = iota
	multipleResultType
)

func buildQuery[I any](baseQuery string, params ...I) (string, []any) {
	var query string
	var args []any

	var paramType I
	switch any(paramType).(type) {
	case Void:
		assert.Check(len(params) == 0)
		query = baseQuery

	default:
		assert.Check(len(params) >= 1)
		sanitize(reflect.Indirect(reflect.ValueOf(&params)))

		var paramsx any
		if len(params) == 1 {
			paramsx = params[0]
		} else {
			paramsx = params
		}

		query, args = prepareQuery(baseQuery, paramsx)
	}

	return query, args
}

func prepareQuery(queryx string, paramsx any) (string, []any) {
	query, args, err := sqlx.Named(queryx, paramsx)
	assert.Must(err)

	query, args, err = sqlx.In(query, args...)
	assert.Must(err)

	slog.Debug("Prepared query",
		"query", query,
		"args", args,
	)

	return query, args
}

// sanitize ensures that:
//   - all the slices in the parameters object are not empty; SQL requires that all IN clauses contain at least one element.
//   - enums are initialised with a valid value; you need to implement sqlu.Enum for this.
//
// If an empty slice is found, a zero value will be pushed.
//
// The function is recursive. It works on both structs and slices of structs.
func sanitize(params reflect.Value) {
	switch params.Kind() {
	case reflect.Struct:
		for i := range params.NumField() {
			field := params.Field(i)

			if _, ok := params.Type().Field(i).Type.MethodByName("Value"); ok || params.Type().Field(i).Type == reflect.TypeOf(time.Time{}) {
				continue
			}

			switch field.Kind() {
			case reflect.Slice:
				if field.Len() > 0 {
					continue
				}

				if field.Type() == reflect.TypeFor[[]byte]() || field.Type() == reflect.TypeFor[json.RawMessage]() {
					field.Set(reflect.ValueOf("{}").Convert(field.Type()))
					continue
				}

				field.Grow(1)
				field.SetLen(1)

			case reflect.String:
				if enumerate := field.MethodByName("Enumerate"); enumerate != (reflect.Value{}) {
					enumVals := enumerate.Call(nil)[0]
					assert.Check(enumVals.Len() > 0)

					if _, i := slicex.FindFirst(enumVals.Interface().([]string), func(it string) bool {
						return it == field.String()
					}); i >= 0 {
						continue
					}

					field.Set(enumVals.Index(0).Convert(field.Type()))
				}

			case reflect.Struct:
				sanitize(field)
			}
		}

	case reflect.Slice:
		if params.Type() == reflect.TypeFor[[]byte]() || params.Type() == reflect.TypeFor[json.RawMessage]() {
			return
		}

		for i := range params.Len() {
			sanitize(params.Index(i))
		}
	}
}

func execQuery[O any](ctx context.Context, db sqlx.ExtContext,
	resultTyp resultType,
	query string,
	args ...any,
) (O, error) {
	var zeroOut O
	query = db.Rebind(query)

	switch any(zeroOut).(type) {
	case Void:
		_, err := db.ExecContext(ctx, query, args...)
		return zeroOut, QueryError(err, query, args...)

	default:
		if resultTyp == singleResultType {
			var result O
			err := sqlx.GetContext(ctx, db, &result, query, args...)

			return result, QueryError(err, query, args...)
		} else {
			var result O
			err := sqlx.SelectContext(ctx, db, &result, query, args...)

			return result, QueryError(err, query, args...)
		}
	}
}

func execPagedQuery[O any](ctx context.Context, db sqlx.ExtContext,
	query string,
	pagination Pagination,
	sorting Sorting,
	args ...any,
) (Page[O], error) {
	countQuery := `
	SELECT COUNT(*)
	FROM (
		` + query + `
	) AS query
	`

	total, err := execQuery[uint](ctx, db, singleResultType, countQuery, args...)
	if err != nil {
		return Page[O]{}, fmt.Errorf("couldn't execute count query: %w", err)
	}

	pagedQuery := query
	if sorting != nil {
		if sortedColumns := parseSorting(sorting); sortedColumns != "" {
			pagedQuery += "\nORDER BY " + parseSorting(sorting)
		}
	}

	pagedQuery += `
	LIMIT 	?
	OFFSET 	?
	`

	items, err := execQuery[O](ctx, db, multipleResultType, pagedQuery, append(args,
		pagination.Size,
		pagination.Skip,
	)...)
	if err != nil {
		return Page[O]{}, fmt.Errorf("couldn't execute paged query: %w", err)
	}

	return buildPage(items, total, pagination), nil
}

type query[I, O any] struct {
	name       string
	body       string
	resultType resultType
}

// Name of the query
func (q query[I, O]) Name() string {
	return q.name
}

// Run the query
func (q query[I, O]) Run(ctx context.Context, db sqlx.ExtContext, params ...I) (O, error) {
	query, args := buildQuery(q.body, params...)
	return execQuery[O](ctx, db, q.resultType, query, args...)
}

// Paged runs the query and returns a paged result
func (q query[I, O]) Paged(ctx context.Context, db sqlx.ExtContext,
	pagination Pagination,
	sorting Sorting,
	params ...I,
) (Page[O], error) {
	assert.Check(q.resultType == multipleResultType)

	query, args := buildQuery(q.body, params...)
	return execPagedQuery[O](ctx, db, query, pagination, sorting, args...)
}

type lazyQuery[I, O any] struct {
	name       string
	generator  func(I) (string, sqld.Params)
	resultType resultType
}

// Name of the query
func (q lazyQuery[I, O]) Name() string {
	return q.name
}

// Run the query
func (q lazyQuery[I, O]) Run(ctx context.Context, db sqlx.ExtContext, params ...I) (O, error) {
	assert.Check(len(params) == 1)

	query, args := prepareQuery(q.generator(params[0]))
	return execQuery[O](ctx, db, q.resultType, query, args...)
}

// Paged runs the query and returns a paged result
func (q lazyQuery[I, O]) Paged(ctx context.Context, db sqlx.ExtContext,
	pagination Pagination,
	sorting Sorting,
	params ...I,
) (Page[O], error) {
	assert.Check(len(params) == 1)
	assert.Check(q.resultType == multipleResultType)

	query, args := prepareQuery(q.generator(params[0]))
	return execPagedQuery[O](ctx, db, query, pagination, sorting, args...)
}

// Query builds a new query with the given body, parameters and result type
func Query[I, O any](name, body string, testRepo ...testRepository) query[I, O] {
	var zeroOut O

	resultTyp := singleResultType
	zeroOutType := reflect.TypeOf(zeroOut)
	resultKind := zeroOutType.Kind()

	if (resultKind == reflect.Array || resultKind == reflect.Slice) &&
		zeroOutType.Elem().Kind() != reflect.Uint8 && // exclude []byte, since may be represented by a single type (e.g. uuid.UUID)
		!strings.Contains(strings.ToLower(body), "insert") {
		resultTyp = multipleResultType
	}

	result := query[I, O]{name, body, resultTyp}
	if len(testRepo) == 1 {
		registerQuery(testRepo[0], result)
	}

	return result
}

// LazyQuery is a query that is generated "lazily";
// an example would be a dynamic query, that changes based on inputs
func LazyQuery[I, O any](
	name string,
	generator func(I) (string, sqld.Params),
	testRepo ...testRepository,
) lazyQuery[I, O] {
	var zeroIn I
	_, ok := any(zeroIn).(Void)
	assert.Check(!ok)

	var zeroOut O

	resultTyp := singleResultType
	resultKind := reflect.TypeOf(zeroOut).Kind()

	if resultKind == reflect.Array || resultKind == reflect.Slice {
		resultTyp = multipleResultType
	}

	result := lazyQuery[I, O]{name, generator, resultTyp}
	if len(testRepo) == 1 {
		registerLazyQuery(testRepo[0], result)
	}

	return result
}
