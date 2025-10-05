package sqlu

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/taleeus/sqld"
)

// Condition makes it compatible with sqld
type Condition interface {
	// Parse transpiles the condition to an SQL string
	Parse(params *sqld.Params, subject string) string
}

// String conditions

// StringEq translates to "target = :param"
type StringEq string

func (val StringEq) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Eq(subject))
}

// StringNe translates to "NOT(target = :param)"
type StringNe string

func (val StringNe) Parse(params *sqld.Params, subject string) string {
	return sqld.Not(sqld.IfNotZero(val, params, sqld.Eq(subject)))
}

// StringContains translates to "target ILIKE :param",
// where :param is formatted as %str%
type StringContains string

func (val StringContains) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(sqld.FmtContains(string(val)), params, sqld.ILike(subject))
}

// StringStartsWith translates to "target ILIKE :param",
// where :param is formatted as str%
type StringStartsWith string

func (val StringStartsWith) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(sqld.FmtStartsWith(string(val)), params, sqld.ILike(subject))
}

// StringEndsWith translates to "target ILIKE :param",
// where :param is formatted as %str
type StringEndsWith string

func (val StringEndsWith) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(sqld.FmtEndsWith(string(val)), params, sqld.ILike(subject))
}

// StringIn translates to "target IN(:param)"
type StringIn []string

func (val StringIn) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotEmpty(val, params, sqld.In(subject))
}

// StringNullIn translates to "(target IS NULL OR target IN(:param))"
type StringNullIn []string

func (val StringNullIn) Parse(params *sqld.Params, subject string) string {
	if len(val) == 0 {
		return ""
	}

	return sqld.Or(
		sqld.Null(subject),
		sqld.IfNotEmpty(val, params, sqld.In(subject)),
	)
}

// StringCondition contains all the possible String filters:
//   - Eq
//   - Ne
//   - Contains
//   - Starts
//   - Ends
//   - In
//   - NullIn
type StringCondition struct {
	Eq       StringEq         `json:"eq,omitempty"`
	Ne       StringNe         `json:"ne,omitempty"`
	Contains StringContains   `json:"contains,omitempty"`
	Starts   StringStartsWith `json:"starts,omitempty"`
	Ends     StringEndsWith   `json:"ends,omitempty"`
	In       StringIn         `json:"in,omitempty"`
	NullIn   StringNullIn     `json:"nullIn,omitempty"`
}

func (op StringCondition) Parse(params *sqld.Params, subject string) string {
	return sqld.And(
		op.Eq.Parse(params, subject),
		op.Ne.Parse(params, subject),
		op.Contains.Parse(params, subject),
		op.Starts.Parse(params, subject),
		op.Ends.Parse(params, subject),
		op.In.Parse(params, subject),
		op.NullIn.Parse(params, subject),
	)
}

// EnumEq translates to "target = :param"
type EnumEq string

func (val EnumEq) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Eq(subject))
}

// EnumNe translates to "NOT(target = :param)"
type EnumNe string

func (val EnumNe) Parse(params *sqld.Params, subject string) string {
	return sqld.Not(sqld.IfNotZero(val, params, sqld.Eq(subject)))
}

// EnumIn translates to "target IN(:param)"
type EnumIn []string

func (val EnumIn) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotEmpty(val, params, sqld.In(subject))
}

// EnumNullIn translates to "(target IS NULL OR target IN(:param))"
type EnumNullIn []string

func (val EnumNullIn) Parse(params *sqld.Params, subject string) string {
	if len(val) == 0 {
		return ""
	}

	return sqld.Or(
		sqld.Null(subject),
		sqld.IfNotEmpty(val, params, sqld.In(subject)),
	)
}

// EnumNullIs translates to "(target IS NULL OR target IN(:param))"
type EnumNullIs string

func (val EnumNullIs) Parse(subject string) string {
	return sqld.Null(subject)
}

// EnumNullIs translates to "(target IS NULL OR target IN(:param))"
type EnumNullIsNot string

func (val EnumNullIsNot) Parse(subject string) string {
	return strings.Replace(sqld.Null(subject), "IS NULL", "IS NOT NULL", 1)
}

// EnumCondition contains all the possible Enum filters:
//   - Eq
//   - Ne
//   - In
//   - NullIn
//   - NullIs
//   - NullIsNot
type EnumCondition[E Enumerable] struct {
	Eq        EnumEq        `json:"eq,omitempty"`
	Ne        EnumNe        `json:"ne,omitempty"`
	In        EnumIn        `json:"in,omitempty"`
	NullIn    EnumNullIn    `json:"nullIn,omitempty"`
	NullIs    EnumNullIs    `json:"nullIs,omitempty"`
	NullIsNot EnumNullIsNot `json:"nullIsNot,omitempty"`
}

func (op EnumCondition[E]) Parse(params *sqld.Params, subject string) string {
	return sqld.And(
		op.Eq.Parse(params, subject),
		op.Ne.Parse(params, subject),
		op.In.Parse(params, subject),
		op.NullIn.Parse(params, subject),
	)
}

// SelectCondition is like EnumCondition, but it's meant for TEXT columns
type SelectCondition struct {
	Eq     EnumEq     `json:"eq,omitempty"`
	Ne     EnumNe     `json:"ne,omitempty"`
	In     EnumIn     `json:"in,omitempty"`
	NullIn EnumNullIn `json:"nullIn,omitempty"`
}

func (op SelectCondition) Parse(params *sqld.Params, subject string) string {
	return sqld.And(
		op.Eq.Parse(params, subject),
		op.Ne.Parse(params, subject),
		op.In.Parse(params, subject),
		op.NullIn.Parse(params, subject),
	)
}

// Int conditions

// IntEq translates to "target = :param"
type IntEq int

func (val IntEq) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Eq(subject))
}

// IntNe translates to "NOT(target = :param)"
type IntNe int

func (val IntNe) Parse(params *sqld.Params, subject string) string {
	return sqld.Not(sqld.IfNotZero(val, params, sqld.Eq(subject)))
}

// IntGt translates to "target > :param"
type IntGt int

func (val IntGt) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Gt(subject))
}

// IntGte translates to "target >= :param"
type IntGte int

func (val IntGte) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Gte(subject))
}

// IntLt translates to "target < :param"
type IntLt int

func (val IntLt) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Lt(subject))
}

// IntLte translates to "target <= :param"
type IntLte int

func (val IntLte) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Lte(subject))
}

// IntCondition contains all the possible Enum filters:
//   - Eq
//   - Ne
//   - Gt
//   - Gte
//   - Lt
//   - Lte
type IntCondition struct {
	Eq  IntEq  `json:"eq,omitempty"`
	Ne  IntNe  `json:"ne,omitempty"`
	Gt  IntGt  `json:"gt,omitempty"`
	Gte IntGte `json:"gte,omitempty"`
	Lt  IntLt  `json:"lt,omitempty"`
	Lte IntLte `json:"lte,omitempty"`
}

func (op IntCondition) Parse(params *sqld.Params, subject string) string {
	return sqld.And(
		op.Eq.Parse(params, subject),
		op.Ne.Parse(params, subject),
		op.Gt.Parse(params, subject),
		op.Gte.Parse(params, subject),
		op.Lt.Parse(params, subject),
		op.Lte.Parse(params, subject),
	)
}

// Float conditions

// FloatEq translates to "target = :param"
type FloatEq float64

func (val FloatEq) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Eq(subject))
}

// FloatNe translates to "NOT(target = :param)"
type FloatNe float64

func (val FloatNe) Parse(params *sqld.Params, subject string) string {
	return sqld.Not(sqld.IfNotZero(val, params, sqld.Eq(subject)))
}

// FloatGt translates to "target > :param"
type FloatGt float64

func (val FloatGt) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Gt(subject))
}

// FloatGte translates to "target >= :param"
type FloatGte float64

func (val FloatGte) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Gte(subject))
}

// FloatLt translates to "target < :param"
type FloatLt float64

func (val FloatLt) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Lt(subject))
}

// FloatLte translates to "target <= :param"
type FloatLte float64

func (val FloatLte) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Lte(subject))
}

// FloatCondition contains all the possible Enum filters:
//   - Eq
//   - Ne
//   - Gt
//   - Gte
//   - Lt
//   - Lte
type FloatCondition struct {
	Eq  FloatEq  `json:"eq,omitempty"`
	Ne  FloatNe  `json:"ne,omitempty"`
	Gt  FloatGt  `json:"gt,omitempty"`
	Gte FloatGte `json:"gte,omitempty"`
	Lt  FloatLt  `json:"lt,omitempty"`
	Lte FloatLte `json:"lte,omitempty"`
}

func (op FloatCondition) Parse(params *sqld.Params, subject string) string {
	return sqld.And(
		op.Eq.Parse(params, subject),
		op.Ne.Parse(params, subject),
		op.Gt.Parse(params, subject),
		op.Gte.Parse(params, subject),
		op.Lt.Parse(params, subject),
		op.Lte.Parse(params, subject),
	)
}

// Time conditions

type TimeOp interface {
	TimeNe | TimeGt | TimeLt | TimeGte | TimeLte | TimeEq
}

func timeScanner[T TimeOp](dst *T, value any) error {
	var t time.Time

	switch v := value.(type) {
	case nil:
		t = time.Time{}
	case time.Time:
		t = v
	case []byte:
		parsed, err := time.Parse(time.RFC3339, string(v))
		if err != nil {
			return fmt.Errorf("cannot scan []byte into %T: %w", dst, err)
		}
		t = parsed
	case string:
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return fmt.Errorf("cannot scan string into %T: %w", dst, err)
		}
		t = parsed
	default:
		return fmt.Errorf("unsupported type %T for scanning into %T", value, dst)
	}

	*dst = T(t)
	return nil
}

func timeValuer[T TimeOp](src T) (driver.Value, error) {
	return time.Time(src).Format(time.RFC3339), nil
}

// TimeEq translates to "target = :param"
type TimeEq time.Time

func (t *TimeEq) Scan(src any) error {
	return timeScanner(t, src)
}

func (t TimeEq) Value() (driver.Value, error) {
	return timeValuer(t)
}

func (val TimeEq) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Eq(subject))
}

func (t *TimeEq) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")

	tt, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	*t = TimeEq(tt)
	return nil
}

// TimeNe translates to "NOT(target = :param)"
type TimeNe time.Time

func (t *TimeNe) Scan(src any) error {
	return timeScanner(t, src)
}

func (t TimeNe) Value() (driver.Value, error) {
	return timeValuer(t)
}

func (val TimeNe) Parse(params *sqld.Params, subject string) string {
	return sqld.Not(sqld.IfNotZero(val, params, sqld.Eq(subject)))
}

func (t *TimeNe) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")

	tt, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	*t = TimeNe(tt)
	return nil
}

// TimeGt translates to "target > :param"
type TimeGt time.Time

func (t *TimeGt) Scan(src any) error {
	return timeScanner(t, src)
}

func (t TimeGt) Value() (driver.Value, error) {
	return timeValuer(t)
}

func (val TimeGt) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Gt(subject))
}

func (t *TimeGt) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")

	tt, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	*t = TimeGt(tt)
	return nil
}

// TimeGte translates to "target >= :param"
type TimeGte time.Time

func (t *TimeGte) Scan(src any) error {
	return timeScanner(t, src)
}

func (t TimeGte) Value() (driver.Value, error) {
	return timeValuer(t)
}

func (val TimeGte) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Gte(subject))
}

func (t *TimeGte) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")

	tt, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	*t = TimeGte(tt)
	return nil
}

// TimeLt translates to "target < :param"
type TimeLt time.Time

func (t *TimeLt) Scan(src any) error {
	return timeScanner(t, src)
}

func (t TimeLt) Value() (driver.Value, error) {
	return timeValuer(t)
}

func (val TimeLt) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Lt(subject))
}

func (t *TimeLt) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")

	tt, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	*t = TimeLt(tt)
	return nil
}

// TimeLte translates to "target <= :param"
type TimeLte time.Time

func (t *TimeLte) Scan(src any) error {
	return timeScanner(t, src)
}

func (t TimeLte) Value() (driver.Value, error) {
	return timeValuer(t)
}

func (val TimeLte) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Lte(subject))
}

func (t *TimeLte) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")

	tt, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	*t = TimeLte(tt)
	return nil
}

// TimeCondition contains all the possible Enum filters:
//   - Eq
//   - Ne
//   - Gt
//   - Gte
//   - Lt
//   - Lte
type TimeCondition struct {
	Eq  TimeEq  `json:"eq,omitempty"`
	Ne  TimeNe  `json:"ne,omitempty"`
	Gt  TimeGt  `json:"gt,omitempty"`
	Gte TimeGte `json:"gte,omitempty"`
	Lt  TimeLt  `json:"lt,omitempty"`
	Lte TimeLte `json:"lte,omitempty"`
}

func (op TimeCondition) Parse(params *sqld.Params, subject string) string {
	return sqld.And(
		op.Eq.Parse(params, subject),
		op.Ne.Parse(params, subject),
		op.Gt.Parse(params, subject),
		op.Gte.Parse(params, subject),
		op.Lt.Parse(params, subject),
		op.Lte.Parse(params, subject),
	)
}

// Bool conditions

// BoolEq translates to "target = :param"
type BoolEq bool

func (val BoolEq) Parse(params *sqld.Params, subject string) string {
	return sqld.IfNotZero(val, params, sqld.Eq(subject))
}

// TimeCondition contains all the possible Enum filters:
//   - Eq
type BoolCondition struct {
	Eq BoolEq `json:"eq,omitempty"`
}

func (op BoolCondition) Parse(params *sqld.Params, subject string) string {
	return sqld.And(
		op.Eq.Parse(params, subject),
	)
}
