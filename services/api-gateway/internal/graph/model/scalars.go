// Package model содержит модели данных для GraphQL
package model

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

// Date представляет дату без времени
type Date struct {
	time.Time
}

// MarshalGQL реализует интерфейс graphql.Marshaler
func (d Date) MarshalGQL(w io.Writer) {
	if d.IsZero() {
		_, _ = w.Write([]byte("null"))
		return
	}
	_, _ = w.Write([]byte(strconv.Quote(d.Format("2006-01-02"))))
}

// UnmarshalGQL реализует интерфейс graphql.Unmarshaler
func (d *Date) UnmarshalGQL(v interface{}) error {
	switch v := v.(type) {
	case string:
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return err
		}
		d.Time = t
		return nil
	case nil:
		return nil
	default:
		return fmt.Errorf("date must be a string, got %T", v)
	}
}

// NewDate создает новый Date из time.Time
func NewDate(t time.Time) *Date {
	if t.IsZero() {
		return nil
	}
	return &Date{Time: t}
}

// DateFromPtr создает Date из указателя на time.Time
func DateFromPtr(t *time.Time) *Date {
	if t == nil || t.IsZero() {
		return nil
	}
	return &Date{Time: *t}
}

// MarshalDate маршалит Date для gqlgen
func MarshalDate(t Date) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		t.MarshalGQL(w)
	})
}

// UnmarshalDate демаршалит Date для gqlgen
func UnmarshalDate(v interface{}) (Date, error) {
	var d Date
	err := d.UnmarshalGQL(v)
	return d, err
}

