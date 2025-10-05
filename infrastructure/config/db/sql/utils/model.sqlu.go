package sqlu

import (
	"database/sql"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/config/assert"
	"github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/config/slicex"
	"github.com/jmoiron/sqlx/reflectx"
)

// Model specifies the name of the table or view
type Model interface {
	ModelName() string
}

// Name returns the [Model] name
func Name[M Model]() string {
	var model M
	return model.ModelName()
}

// Column asserts that the specified column exists in the [Model],
// and returns it
func Column[M Model](name string) string {
	tag := column(reflect.TypeFor[M](), name)
	assert.NotZero(tag)

	return tag
}

// ColumnFull is a proxy for [Column] which fully qualifies the column with the [Model] name
func ColumnFull[M Model](name string) string {
	return serialize(fullSerialization, Name[M](), Column[M](name))
}

// ColumnAliased is a proxy for [Column] which fully qualifies the column with the [Model] name
// and aliases it like [ColumnsAliased]
func ColumnAliased[M Model](name string) string {
	return serialize(aliasedSerialization, Name[M](), Column[M](name))
}

// ColumnMyAliased is a proxy for [Column] which fully qualifies the column with the [Model] name
// and aliases it like [ColumnsMyAliased]
func ColumnMyAliased[M Model](name string) string {
	return serialize(aliasedMySQLSerialization, Name[M](), Column[M](name))
}

// ColumnQuoted is a proxy for [Column] which fully qualifies the column with the [Model] name
// and quotes each part.
func ColumnQuoted[M Model](name string) string {
	return serialize(quotedSerialization, Name[M](), Column[M](name))
}

func column(modelTyp reflect.Type, name string) string {
	for i := range modelTyp.NumField() {
		field := modelTyp.Field(i)
		tag, ok := field.Tag.Lookup("db")
		if !ok {
			if field.Type.Kind() == reflect.Struct {
				if tag := column(field.Type, name); tag != "" {
					return tag
				}
			}

			continue
		}

		if tag != name {
			continue
		}

		return tag
	}

	return ""
}

type serializationStyle int

const (
	simpleSerialization serializationStyle = iota
	fullSerialization
	quotedSerialization
	aliasedSerialization
	aliasedMySQLSerialization
)

func serialize(style serializationStyle, prefix, column string) string {
	switch style {
	case simpleSerialization:
		return column
	case fullSerialization:
		return prefix + "." + column
	case quotedSerialization:
		return fmt.Sprintf(`"%[1]s"."%[2]s"`, prefix, column)
	case aliasedSerialization:
		return fmt.Sprintf(`"%[1]s"."%[2]s" AS "%[1]s.%[2]s"`, prefix, column)
	case aliasedMySQLSerialization:
		return fmt.Sprintf(`%[1]s.%[2]s AS "%[1]s.%[2]s"`, prefix, column)
	}

	panic(fmt.Errorf("unhandled serializationStyle: %v", style))
}

// Columns extracts columns from a Model.
//
// If the varargs are empty, all the [Model]'s columns are returned; otherwise, only the specified subset
// (the existance of these columns will be asserted)
func Columns[M Model](names ...string) []string {
	return columns[M](simpleSerialization, names...)
}

// ColumnsFull extracts columns from a [Model] and serializes them
// with the [Model] name prefix:
//
//	model.column
//
// If the varargs are empty, all the [Model]'s columns are returned; otherwise, only the specified subset
// (the existance of these columns will be asserted)
func ColumnsFull[M Model](names ...string) []string {
	return columns[M](fullSerialization, names...)
}

// ColumnsAliased extracts columns from a [Model] and serializes them
// with the [Model] name prefix as an alias:
//
//	"model"."column" AS "model.column"
//
// If the varargs are empty, all the [Model]'s columns are returned; otherwise, only the specified subset
// (the existance of these columns will be asserted).
//
// The syntax is only compatible with Postgres. For the MySQL version, see `ColumnsMyAliased()`
func ColumnsAliased[M Model](names ...string) []string {
	return columns[M](aliasedSerialization, names...)
}

// ColumnsMyAliased extracts columns from a [Model] and serializes them
// with the [Model] name prefix as an alias:
//
//	model.column AS "model.column"
//
// If the varargs are empty, all the [Model]'s columns are returned; otherwise, only the specified subset
// (the existance of these columns will be asserted).
//
// The syntax is only compatible with MySQL.
func ColumnsMyAliased[M Model](names ...string) []string {
	return columns[M](aliasedMySQLSerialization, names...)
}

// ColumnsQuoted extracts columns from a [Model] and serializes them
// with the [Model] name prefix as an alias:
//
//	"model"."column"
//
// If the varargs are empty, all the [Model]'s columns are returned; otherwise, only the specified subset
// (the existance of these columns will be asserted).
func ColumnsQuoted[M Model](names ...string) []string {
	return columns[M](quotedSerialization, names...)
}

func columns[M Model](style serializationStyle, names ...string) []string {
	columnTags := make([]string, 0)
	collectColumns(&columnTags, reflect.TypeFor[M](), "", simpleSerialization, nil)
	if len(names) > 0 {
		assert.Check(slicex.ContainsAll(columnTags, names))
		columnTags = names
	}

	return slicex.Map(columnTags, func(col string) string {
		return serialize(style, Name[M](), col)
	})
}

func collectColumns(columnTags *[]string, modelTyp reflect.Type, prefix string, style serializationStyle, overrides map[string]string) {
	for i := range modelTyp.NumField() {
		field := modelTyp.Field(i)
		tag, _ := field.Tag.Lookup("db")
		traverseStruct := reflectx.Deref(field.Type).Kind() == reflect.Struct &&
			!reflect.PointerTo(field.Type).Implements(reflect.TypeFor[sql.Scanner]()) &&
			field.Type != reflect.TypeFor[time.Time]()
		if traverseStruct {
			prefixNested := prefix
			if tag != "" {
				prefixNested = tag
			}

			collectColumns(columnTags, reflectx.Deref(field.Type), prefixNested, style, overrides)
			continue
		}

		if prefix != "" {
			switch style {
			case aliasedSerialization, aliasedMySQLSerialization:
				tag = serialize(style, prefix, tag)

			default:
				tag = prefix + "." + tag
			}
		}

		if overrides != nil {
			if override, ok := overrides[tag]; ok {
				tag = override
			}
		}

		*columnTags = append(*columnTags, tag)
	}
}

func Join(columns []string) string {
	return strings.Join(columns, ",\n\t")
}

func Concat(columnSets ...[]string) []string {
	return slices.Concat(columnSets...)
}

// Extract returns all columns with prefixes present in the
// provided type source. It traverses every struct found recursively
//
// This method is meant for a Postgres query; see [ExtractMy] for MySQL
func Extract[T any](overrides ...map[string]string) []string {
	var overridesOpt map[string]string
	if len(overrides) == 1 {
		overridesOpt = overrides[0]
	}

	columns := make([]string, 0)
	collectColumns(&columns, reflect.TypeFor[T](), "", aliasedSerialization, overridesOpt)

	return columns
}

// ExtractMy returns all columns with prefixes present in the
// provided type source. It traverses every struct found recursively
//
// This method is meant for a MySQL query; see [Extract] for Postgres
func ExtractMy[T any](overrides ...map[string]string) []string {
	var overridesOpt map[string]string
	if len(overrides) == 1 {
		overridesOpt = overrides[0]
	}

	columns := make([]string, 0)
	collectColumns(&columns, reflect.TypeFor[T](), "", aliasedMySQLSerialization, overridesOpt)

	return columns
}
