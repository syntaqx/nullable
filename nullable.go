// Package nullable provides a single generic type for values that may be
// absent, with first-class support for JSON and database/sql.
//
// Instead of a separate wrapper per underlying type, use Nullable[T]:
//
//	type User struct {
//		ID    nullable.Int64  `json:"id"`
//		Name  nullable.String `json:"name"`
//		Email nullable.String `json:"email,omitzero"`
//	}
//
// A Nullable is null by its zero value, so a freshly declared field marshals
// to JSON null and stores as SQL NULL without any setup.
//
// # Extending to other formats
//
// The core package deliberately ships only the encodings that can represent a
// true null: encoding/json and database/sql. To support a format it does not
// (YAML, TOML, BSON, ...), embed Nullable in your own type and add the method
// that format looks for, bridging through the public accessors:
//
//	type YAMLString struct{ nullable.String }
//
//	func (s YAMLString) MarshalYAML() (any, error) {
//		if v, ok := s.Get(); ok {
//			return v, nil
//		}
//		return nil, nil // null
//	}
//
// This keeps third-party encoders out of the dependency graph while letting
// callers opt in to any representation, including how null should look.
package nullable

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Nullable holds a value of type T that may be null.
//
// The zero value is a valid, null Nullable. Use Value to construct a non-null
// one. Fields are unexported so the null/non-null invariant can only be changed
// through the provided methods.
type Nullable[T any] struct {
	value T
	valid bool
}

// Value returns a non-null Nullable holding v.
func Value[T any](v T) Nullable[T] {
	return Nullable[T]{value: v, valid: true}
}

// Null returns a null Nullable[T]. It is equivalent to the zero value and
// exists mainly to make intent explicit and to aid type inference.
func Null[T any]() Nullable[T] {
	return Nullable[T]{}
}

// FromPtr returns a null Nullable when p is nil, otherwise a non-null Nullable
// holding a copy of *p.
func FromPtr[T any](p *T) Nullable[T] {
	if p == nil {
		return Nullable[T]{}
	}
	return Nullable[T]{value: *p, valid: true}
}

// Valid reports whether n holds a value (i.e. is not null).
func (n Nullable[T]) Valid() bool {
	return n.valid
}

// IsNull reports whether n is null.
func (n Nullable[T]) IsNull() bool {
	return !n.valid
}

// Get returns the underlying value and whether it is valid. When n is null the
// value is the zero value of T.
func (n Nullable[T]) Get() (T, bool) {
	return n.value, n.valid
}

// Or returns the underlying value when n is valid, otherwise fallback.
func (n Nullable[T]) Or(fallback T) T {
	if n.valid {
		return n.value
	}
	return fallback
}

// OrZero returns the underlying value when n is valid, otherwise the zero value
// of T.
func (n Nullable[T]) OrZero() T {
	return n.value
}

// Ptr returns a pointer to a copy of the underlying value, or nil when n is
// null.
func (n Nullable[T]) Ptr() *T {
	if !n.valid {
		return nil
	}
	v := n.value
	return &v
}

// Set assigns v and marks n as valid.
func (n *Nullable[T]) Set(v T) {
	n.value = v
	n.valid = true
}

// SetNull clears n back to null.
func (n *Nullable[T]) SetNull() {
	var zero T
	n.value = zero
	n.valid = false
}

// IsZero reports whether n is null. It implements the interface used by the
// encoding/json "omitzero" tag option so null fields can be omitted.
func (n Nullable[T]) IsZero() bool {
	return !n.valid
}

// String implements fmt.Stringer.
func (n Nullable[T]) String() string {
	if !n.valid {
		return "<null>"
	}
	return fmt.Sprint(n.value)
}

var jsonNull = []byte("null")

// MarshalJSON implements json.Marshaler. A null Nullable encodes as JSON null.
func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if !n.valid {
		return jsonNull, nil
	}
	return json.Marshal(n.value)
}

// UnmarshalJSON implements json.Unmarshaler. JSON null decodes to a null
// Nullable; any other value is decoded into T and marks n valid.
func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, jsonNull) {
		n.SetNull()
		return nil
	}
	if err := json.Unmarshal(data, &n.value); err != nil {
		return err
	}
	n.valid = true
	return nil
}

// Scan implements sql.Scanner, delegating conversion to the standard library so
// every driver-supported type behaves consistently.
func (n *Nullable[T]) Scan(src any) error {
	var sn sql.Null[T]
	if err := sn.Scan(src); err != nil {
		return err
	}
	n.value = sn.V
	n.valid = sn.Valid
	return nil
}

// Value implements driver.Valuer. A null Nullable produces a nil driver.Value.
func (n Nullable[T]) Value() (driver.Value, error) {
	if !n.valid {
		return nil, nil
	}
	return driver.DefaultParameterConverter.ConvertValue(n.value)
}

// Map applies f to the value of n when it is non-null and returns a Nullable of
// the result. A null input yields a null output and f is not called.
func Map[T, U any](n Nullable[T], f func(T) U) Nullable[U] {
	if !n.valid {
		return Nullable[U]{}
	}
	return Nullable[U]{value: f(n.value), valid: true}
}

// Equal reports whether a and b are both null, or both non-null with equal
// values.
func Equal[T comparable](a, b Nullable[T]) bool {
	if a.valid != b.valid {
		return false
	}
	return !a.valid || a.value == b.value
}

// Common aliases for convenience. Nullable[T] works with any type; these just

// save keystrokes for the usual suspects.
type (
	Bool    = Nullable[bool]
	Bytes   = Nullable[[]byte]
	Float32 = Nullable[float32]
	Float64 = Nullable[float64]
	Int     = Nullable[int]
	Int32   = Nullable[int32]
	Int64   = Nullable[int64]
	String  = Nullable[string]
	Time    = Nullable[time.Time]
)

// Compile-time checks that Nullable satisfies the standard interfaces.
var (
	_ json.Marshaler   = Nullable[int]{}
	_ json.Unmarshaler = (*Nullable[int])(nil)
	_ driver.Valuer    = Nullable[int]{}
	_ sql.Scanner      = (*Nullable[int])(nil)
	_ fmt.Stringer     = Nullable[int]{}
)
