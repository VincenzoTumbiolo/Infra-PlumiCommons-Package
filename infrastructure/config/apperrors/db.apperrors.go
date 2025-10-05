package apperrors

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

type DbError uint

const (
	GENERIC_QUERY_ERROR DbError = iota
	NOT_FOUND
	ALREADY_EXISTS
	RELATION_CONSTRAINT
)

func (DbError) Subject() string {
	return "DBE"
}

func (err DbError) Code() uint {
	return uint(err)
}

func (err DbError) StatusCode() int {
	switch err {
	case GENERIC_QUERY_ERROR:
		return http.StatusInternalServerError
	case NOT_FOUND:
		return http.StatusNotFound
	case ALREADY_EXISTS:
		return http.StatusConflict
	case RELATION_CONSTRAINT:
		return http.StatusBadRequest
	}

	return http.StatusInternalServerError
}

func (err DbError) Error() string {
	switch err {
	case GENERIC_QUERY_ERROR:
		return "Generic query error"
	case NOT_FOUND:
		return "Entity not found"
	case ALREADY_EXISTS:
		return "An entity with these unique identifiers already exists"
	case RELATION_CONSTRAINT:
		return "This operation is blocked by a relation; operate on the related entity first"
	}

	return UNKNOWN_ERROR_MESSAGE
}

func FromSQLErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return NOT_FOUND
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ForeignKeyViolation:
			return errors.Join(RELATION_CONSTRAINT, err)
		case pgerrcode.UniqueViolation:
			return errors.Join(ALREADY_EXISTS, err)
		}
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1216:
			return errors.Join(RELATION_CONSTRAINT, err)
		case 1062:
			return errors.Join(ALREADY_EXISTS, err)
		}
	}

	return errors.Join(GENERIC_QUERY_ERROR, err)
}
