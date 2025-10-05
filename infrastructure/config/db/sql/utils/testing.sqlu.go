package sqlu

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/go-sql-driver/mysql"
	fuzz "github.com/google/gofuzz"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/taleeus/sqld"
)

type testRepository struct {
	name string

	integrationTests     map[string]func(*testing.T, sqlx.ExtContext)
	integrationFuzzTests map[string]func(*testing.T, *fuzz.Fuzzer, sqlx.ExtContext)
}

// TestRepository creates a repository to contain test callbacks.
// To be used with Query() and LazyQuery()
func TestRepository(name string) testRepository {
	repo := testRepository{name: name}
	repo.integrationTests = make(map[string]func(*testing.T, sqlx.ExtContext))
	repo.integrationFuzzTests = make(map[string]func(*testing.T, *fuzz.Fuzzer, sqlx.ExtContext))

	return repo
}

func registerQuery[I, O any](testRepo testRepository, q query[I, O]) {
	var zeroIn I
	testRepo.integrationTests[q.name] = func(t *testing.T, db sqlx.ExtContext) {
		ctx := context.Background()
		switch reflect.TypeOf(zeroIn).Kind() {
		case reflect.Slice: // batch insert
			inVal := reflect.MakeSlice(reflect.TypeOf(zeroIn), 1, 1)
			t.Run(q.name, func(t *testing.T) {
				if _, err := q.Run(ctx, db, inVal.Interface().(I)); raiseError(err) {
					t.Fatal(err.Error())
				}
			})

		default:
			var params []I
			if _, ok := any(zeroIn).(Void); !ok {
				params = append(params, zeroIn)
			}

			t.Run(testRepo.name+"/"+q.name, func(t *testing.T) {
				if _, err := q.Run(ctx, db, params...); raiseError(err) {
					t.Fatal(err.Error())
				}
			})

			if q.resultType == multipleResultType {
				t.Run(testRepo.name+"/"+q.name+"(paged)", func(t *testing.T) {
					if _, err := q.Paged(ctx, db, Pagination{10, 0}, nil, params...); raiseError(err) {
						t.Fatal(err.Error())
					}
				})
			}
		}
	}

	if _, ok := any(zeroIn).(Void); !ok {
		testRepo.integrationFuzzTests[q.name] = func(t *testing.T, f *fuzz.Fuzzer, db sqlx.ExtContext) {
			ctx := context.Background()
			switch reflect.TypeOf(zeroIn).Kind() {
			case reflect.Array, reflect.Slice: // batch insert
				inVal := reflect.MakeSlice(reflect.TypeOf(zeroIn), 1, 1)
				inStruct := inVal.Index(0).Addr()
				f.Fuzz(inStruct.Interface())

				params := inVal.Interface()
				t.Run(testRepo.name+"/"+q.name+"(fuzzed)", func(t *testing.T) {
					if _, err := q.Run(ctx, db, params.(I)); raiseError(err) {
						t.Fatal(err.Error())
					}
				})

				if q.resultType == multipleResultType {
					t.Run(testRepo.name+"/"+q.name+"(fuzzed,paged)", func(t *testing.T) {
						if _, err := q.Paged(ctx, db, Pagination{10, 0}, nil, params.(I)); raiseError(err) {
							t.Fatal(err.Error())
						}
					})
				}

			default:
				var params I
				f.Fuzz(&params)

				t.Run(testRepo.name+"/"+q.name+"(fuzzed)", func(t *testing.T) {
					if _, err := q.Run(ctx, db, params); raiseError(err) {
						t.Fatal(err.Error())
					}
				})

				if q.resultType == multipleResultType {
					t.Run(testRepo.name+"/"+q.name+"(fuzzed,paged)", func(t *testing.T) {
						if _, err := q.Paged(ctx, db, Pagination{10, 0}, nil, params); raiseError(err) {
							t.Fatal(err.Error())
						}
					})
				}
			}
		}
	}
}

func registerLazyQuery[I, O any](testRepo testRepository, q lazyQuery[I, O]) {
	testRepo.integrationTests[q.name] = func(t *testing.T, db sqlx.ExtContext) {
		ctx := context.Background()
		var params I

		t.Run(testRepo.name+"/"+q.name, func(t *testing.T) {
			if _, err := q.Run(ctx, db, params); raiseError(err) {
				t.Fatal(err.Error())
			}
		})

		if q.resultType == multipleResultType {
			t.Run(testRepo.name+"/"+testRepo.name+"/"+q.name+"(paged)", func(t *testing.T) {
				if _, err := q.Paged(ctx, db, Pagination{10, 0}, nil, params); raiseError(err) {
					t.Fatal(err.Error())
				}
			})
		}
	}

	testRepo.integrationFuzzTests[q.name] = func(t *testing.T, f *fuzz.Fuzzer, db sqlx.ExtContext) {
		var params I
		ctx := context.Background()

		f.Fuzz(&params)
		t.Run(testRepo.name+"/"+q.name+"(fuzzed)", func(t *testing.T) {
			if _, err := q.Run(ctx, db, params); raiseError(err) {
				t.Fatal(err.Error())
			}
		})
		if q.resultType == multipleResultType {
			t.Run(testRepo.name+"/"+q.name+"(fuzzed,paged)", func(t *testing.T) {
				if _, err := q.Paged(ctx, db, Pagination{10, 0}, nil, params); raiseError(err) {
					t.Fatal(err.Error())
				}
			})
		}
	}
}

type fuzzGenerator func([]byte) *fuzz.Fuzzer

// Test executes all integration tests
func (repo testRepository) Test(t *testing.T, db sqlx.ExtContext) {
	t.Helper()
	for _, it := range repo.integrationTests {
		it(t, db)
	}
}

// Fuzz executes all fuzzing tests
func (repo testRepository) Fuzz(f *testing.F, db sqlx.ExtContext, fuzzGenerator fuzzGenerator) {
	f.Helper()
	f.Fuzz(func(t *testing.T, data []byte) {
		f := fuzzGenerator(data)
		for _, ift := range repo.integrationFuzzTests {
			ift(t, f, db)
		}
	})
}

func FuzzGenerator(funcs ...any) func(data []byte) *fuzz.Fuzzer {
	return func(data []byte) *fuzz.Fuzzer {
		return fuzz.NewFromGoFuzz(data).NilChance(.3).Funcs(
			func(op *sqld.Op, c fuzz.Continue) {
				switch c.RandBool() {
				case true:
					*op = sqld.AND
				case false:
					*op = sqld.OR
				}
			},
		).Funcs(funcs...)
	}
}

func EnumConditionFuzzer[E Enumerable]() func(*EnumCondition[E], fuzz.Continue) {
	return func(cond *EnumCondition[E], c fuzz.Continue) {
		var enum E
		values := enum.Enumerate()

		if c.RandBool() {
			cond.Eq = EnumEq(values[c.Intn(len(values))])
		}
		if c.RandBool() {
			cond.Ne = EnumNe(values[c.Intn(len(values))])
		}
		if c.RandBool() {
			cond.In = EnumIn(values[c.Intn(len(values)/2):])
		}
		if c.RandBool() {
			cond.NullIn = EnumNullIn(values[c.Intn(len(values)/2):])
		}
	}
}

func raiseError(err error) bool {
	if err == nil {
		return false
	}

	var queryErr queryError
	if errors.As(err, &queryErr) {
		err = queryErr.inner
	}

	if errors.Is(err, sql.ErrNoRows) {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ForeignKeyViolation:
			return false
		case pgerrcode.UniqueViolation:
			return false
		case pgerrcode.InvalidTextRepresentation:
			return false
		case pgerrcode.CheckViolation:
			return false
		case pgerrcode.FeatureNotSupported:
			return false
		}
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1216, 1451, 1452:
			return false
		case 1062:
			return false
		}
	}

	return true
}
