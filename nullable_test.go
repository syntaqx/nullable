package nullable

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestValueAndNull(t *testing.T) {
	v := Value("hello")
	if !v.Valid() || v.IsNull() {
		t.Fatalf("Value should be valid")
	}
	if got, ok := v.Get(); !ok || got != "hello" {
		t.Fatalf("Get() = %q, %v", got, ok)
	}

	n := Null[string]()
	if n.Valid() || !n.IsNull() {
		t.Fatalf("Null should be null")
	}
	var zero String
	if zero != n {
		t.Fatalf("zero value should equal Null()")
	}
}

func TestFromPtr(t *testing.T) {
	if got := FromPtr[int](nil); got.Valid() {
		t.Fatalf("FromPtr(nil) should be null")
	}
	x := 42
	got := FromPtr(&x)
	if v, ok := got.Get(); !ok || v != 42 {
		t.Fatalf("FromPtr(&42) = %d, %v", v, ok)
	}
	// Mutating the source must not affect the stored copy.
	x = 99
	if v, _ := got.Get(); v != 42 {
		t.Fatalf("FromPtr should copy, got %d", v)
	}
}

func TestOrAndOrZero(t *testing.T) {
	if got := Value(5).Or(10); got != 5 {
		t.Fatalf("Or on valid = %d", got)
	}
	if got := Null[int]().Or(10); got != 10 {
		t.Fatalf("Or on null = %d", got)
	}
	if got := Null[int]().OrZero(); got != 0 {
		t.Fatalf("OrZero on null = %d", got)
	}
}

func TestPtr(t *testing.T) {
	if p := Null[int]().Ptr(); p != nil {
		t.Fatalf("Ptr on null should be nil")
	}
	p := Value(7).Ptr()
	if p == nil || *p != 7 {
		t.Fatalf("Ptr on valid = %v", p)
	}
}

func TestSetAndSetNull(t *testing.T) {
	var n Int
	n.Set(3)
	if v, ok := n.Get(); !ok || v != 3 {
		t.Fatalf("after Set = %d, %v", v, ok)
	}
	n.SetNull()
	if n.Valid() {
		t.Fatalf("after SetNull should be null")
	}
}

func TestString(t *testing.T) {
	if got := Null[int]().String(); got != "<null>" {
		t.Fatalf("null String() = %q", got)
	}
	if got := Value(42).String(); got != "42" {
		t.Fatalf("valid String() = %q", got)
	}
}

func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"valid string", Value("test"), `"test"`},
		{"null string", Null[string](), `null`},
		{"valid int", Value[int64](123), `123`},
		{"null int", Null[int64](), `null`},
		{"valid bool", Value(true), `true`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("Marshal() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		var s String
		if err := json.Unmarshal([]byte(`"test"`), &s); err != nil {
			t.Fatal(err)
		}
		if v, ok := s.Get(); !ok || v != "test" {
			t.Fatalf("got %q, %v", v, ok)
		}
	})
	t.Run("null string", func(t *testing.T) {
		s := Value("preset")
		if err := json.Unmarshal([]byte(`null`), &s); err != nil {
			t.Fatal(err)
		}
		if s.Valid() {
			t.Fatalf("null should clear validity")
		}
	})
	t.Run("invalid payload", func(t *testing.T) {
		var i Int64
		if err := json.Unmarshal([]byte(`"nope"`), &i); err == nil {
			t.Fatalf("expected error for bad payload")
		}
		if i.Valid() {
			t.Fatalf("failed unmarshal should not be valid")
		}
	})
}

func TestJSONRoundTripStruct(t *testing.T) {
	type User struct {
		ID    Int64  `json:"id"`
		Name  String `json:"name"`
		Email String `json:"email,omitzero"`
	}
	u := User{ID: Value[int64](1), Name: Value("John")}
	b, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	// Email is null -> omitted via omitzero; Name present.
	want := `{"id":1,"name":"John"}`
	if string(b) != want {
		t.Fatalf("Marshal = %s, want %s", b, want)
	}

	var back User
	if err := json.Unmarshal([]byte(`{"id":2,"name":null,"email":"a@b.co"}`), &back); err != nil {
		t.Fatal(err)
	}
	if v, ok := back.ID.Get(); !ok || v != 2 {
		t.Fatalf("ID = %d, %v", v, ok)
	}
	if back.Name.Valid() {
		t.Fatalf("Name should be null")
	}
	if v, _ := back.Email.Get(); v != "a@b.co" {
		t.Fatalf("Email = %q", v)
	}
}

func TestValueDriver(t *testing.T) {
	got, err := Value[int64](5).Value()
	if err != nil {
		t.Fatal(err)
	}
	if got != int64(5) {
		t.Fatalf("Value() = %v", got)
	}

	got, err = Null[int64]().Value()
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("null Value() = %v, want nil", got)
	}
}

func TestScan(t *testing.T) {
	tests := []struct {
		name      string
		src       any
		wantValid bool
		wantValue string
	}{
		{"string", "hello", true, "hello"},
		{"bytes", []byte("world"), true, "world"},
		{"nil", nil, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s String
			if err := s.Scan(tt.src); err != nil {
				t.Fatal(err)
			}
			if s.Valid() != tt.wantValid {
				t.Fatalf("Valid() = %v, want %v", s.Valid(), tt.wantValid)
			}
			if v, _ := s.Get(); tt.wantValid && v != tt.wantValue {
				t.Fatalf("Get() = %q, want %q", v, tt.wantValue)
			}
		})
	}
}

func TestScanNumeric(t *testing.T) {
	var i Int64
	if err := i.Scan(int64(64)); err != nil {
		t.Fatal(err)
	}
	if v, ok := i.Get(); !ok || v != 64 {
		t.Fatalf("Get() = %d, %v", v, ok)
	}
}

func TestScanError(t *testing.T) {
	i := Value[int64](7)
	if err := i.Scan("not-a-number"); err == nil {
		t.Fatalf("expected error scanning non-numeric string into int64")
	}
	// A failed Scan must leave the prior value untouched.
	if v, ok := i.Get(); !ok || v != 7 {
		t.Fatalf("value changed after failed Scan: %d, %v", v, ok)
	}
}

func TestScanTime(t *testing.T) {
	now := time.Now()
	var tm Time
	if err := tm.Scan(now); err != nil {
		t.Fatal(err)
	}
	if v, ok := tm.Get(); !ok || !v.Equal(now) {
		t.Fatalf("Get() = %v, %v", v, ok)
	}
}

func TestMap(t *testing.T) {
	got := Map(Value(3), func(i int) string { return fmt.Sprint(i * 2) })
	if v, ok := got.Get(); !ok || v != "6" {
		t.Fatalf("Map valid = %q, %v", v, ok)
	}

	called := false
	out := Map(Null[int](), func(i int) string { called = true; return "" })
	if out.Valid() {
		t.Fatalf("Map of null should stay null")
	}
	if called {
		t.Fatalf("f must not be called for a null input")
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b Int
		want bool
	}{
		{"both null", Null[int](), Null[int](), true},
		{"both same value", Value(1), Value(1), true},
		{"different values", Value(1), Value(2), false},
		{"null vs value", Null[int](), Value(1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Equal(tt.a, tt.b); got != tt.want {
				t.Fatalf("Equal = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNullEmitVsOmit documents the two ways a null field is serialized: emitted
// as JSON null by default, or omitted entirely with the omitzero tag option.
func TestNullEmitVsOmit(t *testing.T) {
	type emit struct {
		Name String `json:"name"`
	}
	type omit struct {
		Name String `json:"name,omitzero"`
	}

	b, err := json.Marshal(emit{})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"name":null}` {
		t.Fatalf("emit = %s, want {\"name\":null}", b)
	}

	b, err = json.Marshal(omit{})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{}` {
		t.Fatalf("omit = %s, want {}", b)
	}
}

// Ensure the interface assertions hold for a driver.Valuer round trip.
var _ driver.Valuer = String{}
