package nullable

import (
	"encoding/json"
	"fmt"
)

// Optional is a Nullable that can also be absent, giving three states:
//
//   - absent: the field was not in the input (the zero value)
//   - null:   the field was present and explicitly null
//   - value:  the field was present with a value
//
// This is what a JSON PATCH-style request needs: "leave it alone", "clear it"
// and "set it" are different instructions, but *T and Nullable[T] both
// collapse the first two.
//
// Absence only survives encoding with the omitzero tag option; without it an
// absent Optional marshals as null:
//
//	type UserPatch struct {
//		Name nullable.Optional[string] `json:"name,omitzero"`
//	}
//
// Optional deliberately does not implement sql.Scanner or driver.Valuer: SQL
// has no notion of an absent value. Resolve it against the current value with
// [Optional.Or], then store the resulting [Nullable].
type Optional[T any] struct {
	value Nullable[T]
	set   bool
}

// Present returns an Optional holding n. Pass the result of [Value] to set a
// value, or of [Null] to set an explicit null.
func Present[T any](n Nullable[T]) Optional[T] {
	return Optional[T]{value: n, set: true}
}

// IsPresent reports whether o was set, as a value or as null.
func (o Optional[T]) IsPresent() bool {
	return o.set
}

// Get returns the held Nullable and whether o is present. When o is absent
// the Nullable is null.
func (o Optional[T]) Get() (Nullable[T], bool) {
	return o.value, o.set
}

// Or returns the held Nullable when o is present, otherwise current. It is the
// usual way to apply a patch field:
//
//	user.Name = patch.Name.Or(user.Name)
func (o Optional[T]) Or(current Nullable[T]) Nullable[T] {
	if o.set {
		return o.value
	}
	return current
}

// Set makes o present with the value v.
func (o *Optional[T]) Set(v T) {
	o.value.Set(v)
	o.set = true
}

// SetNull makes o present and null.
func (o *Optional[T]) SetNull() {
	o.value.SetNull()
	o.set = true
}

// Unset makes o absent.
func (o *Optional[T]) Unset() {
	*o = Optional[T]{}
}

// IsZero reports whether o is absent. It implements the interface used by the
// encoding/json "omitzero" tag option, which is what keeps absent fields out of
// the output.
func (o Optional[T]) IsZero() bool {
	return !o.set
}

// String implements fmt.Stringer.
func (o Optional[T]) String() string {
	if !o.set {
		return "<absent>"
	}
	return o.value.String()
}

// MarshalJSON implements json.Marshaler. An absent Optional encodes as null;
// use the omitzero tag option to leave it out instead.
func (o Optional[T]) MarshalJSON() ([]byte, error) {
	return o.value.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler. It is only called when the field
// is in the input, so a successful decode always marks o present.
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	if err := o.value.UnmarshalJSON(data); err != nil {
		return err
	}
	o.set = true
	return nil
}

// Compile-time checks that Optional satisfies the standard interfaces.
var (
	_ json.Marshaler   = Optional[int]{}
	_ json.Unmarshaler = (*Optional[int])(nil)
	_ fmt.Stringer     = Optional[int]{}
)
