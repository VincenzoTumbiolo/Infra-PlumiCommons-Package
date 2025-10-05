package sqlu

import (
	"database/sql"
	"time"
)

func ToNilString(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}

	return nil
}

func ToNilInt64(i sql.NullInt64) *int64 {
	if i.Valid {
		return &i.Int64
	}

	return nil
}

func ToNilInt32(i sql.NullInt32) *int32 {
	if i.Valid {
		return &i.Int32
	}

	return nil
}

func ToNilInt16(i sql.NullInt16) *int16 {
	if i.Valid {
		return &i.Int16
	}

	return nil
}

func ToNilByte(b sql.NullByte) *byte {
	if b.Valid {
		return &b.Byte
	}

	return nil
}

func ToNilFloat64(f sql.NullFloat64) *float64 {
	if f.Valid {
		return &f.Float64
	}

	return nil
}

func ToNilBool(b sql.NullBool) *bool {
	if b.Valid {
		return &b.Bool
	}

	return nil
}

func ToNilTime(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}

	return nil
}

func FromNilString(s *string) sql.NullString {
	if s != nil {
		return sql.NullString{
			String: *s,
			Valid:  true,
		}
	}

	return sql.NullString{Valid: false}
}

func FromNilInt64(i *int64) sql.NullInt64 {
	if i != nil {
		return sql.NullInt64{
			Int64: *i,
			Valid: true,
		}
	}

	return sql.NullInt64{Valid: false}
}

func FromNilInt32(i *int32) sql.NullInt32 {
	if i != nil {
		return sql.NullInt32{
			Int32: *i,
			Valid: true,
		}
	}

	return sql.NullInt32{Valid: false}
}

func FromNilInt16(i *int16) sql.NullInt16 {
	if i != nil {
		return sql.NullInt16{
			Int16: *i,
			Valid: true,
		}
	}

	return sql.NullInt16{Valid: false}
}

func FromNilByte(b *byte) sql.NullByte {
	if b != nil {
		return sql.NullByte{
			Byte:  *b,
			Valid: true,
		}
	}

	return sql.NullByte{Valid: false}
}

func FromNilFloat64(f *float64) sql.NullFloat64 {
	if f != nil {
		return sql.NullFloat64{
			Float64: *f,
			Valid:   true,
		}
	}

	return sql.NullFloat64{Valid: false}
}

func FromNilBool(b *bool) sql.NullBool {
	if b != nil {
		return sql.NullBool{
			Bool:  *b,
			Valid: true,
		}
	}

	return sql.NullBool{Valid: false}
}

func FromNilTime(t *time.Time) sql.NullTime {
	if t != nil {
		return sql.NullTime{
			Time:  *t,
			Valid: true,
		}
	}

	return sql.NullTime{Valid: false}
}
