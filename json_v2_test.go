//go:build go1.27

package nullable

import (
	jsonv2 "encoding/json/v2"
	"testing"
)

func TestJSONv2Marshal(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"valid string", Value("test"), `"test"`},
		{"null string", Null[string](), `null`},
		{"valid int", Value[int64](123), `123`},
		{"null int", Null[int64](), `null`},
		{"valid empty string", Value(""), `""`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonv2.Marshal(tt.in)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("Marshal() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestJSONv2OptionsPropagate checks that caller options reach the wrapped
// value, which the v1 MarshalJSON path cannot do.
func TestJSONv2OptionsPropagate(t *testing.T) {
	got, err := jsonv2.Marshal(Value[int64](5), jsonv2.StringifyNumbers(true))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `"5"` {
		t.Fatalf("Marshal(StringifyNumbers) = %s, want \"5\"", got)
	}

	var n Int64
	if err := jsonv2.Unmarshal([]byte(`"7"`), &n, jsonv2.StringifyNumbers(true)); err != nil {
		t.Fatal(err)
	}
	if v, ok := n.Get(); !ok || v != 7 {
		t.Fatalf("Unmarshal(StringifyNumbers) = %d, %v", v, ok)
	}
}

func TestJSONv2Unmarshal(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		var s String
		if err := jsonv2.Unmarshal([]byte(`"test"`), &s); err != nil {
			t.Fatal(err)
		}
		if v, ok := s.Get(); !ok || v != "test" {
			t.Fatalf("got %q, %v", v, ok)
		}
	})
	t.Run("null clears", func(t *testing.T) {
		s := Value("preset")
		if err := jsonv2.Unmarshal([]byte(`null`), &s); err != nil {
			t.Fatal(err)
		}
		if s.Valid() {
			t.Fatalf("null should clear validity")
		}
	})
	t.Run("invalid payload", func(t *testing.T) {
		var i Int64
		if err := jsonv2.Unmarshal([]byte(`"nope"`), &i); err == nil {
			t.Fatalf("expected error for bad payload")
		}
		if i.Valid() {
			t.Fatalf("failed unmarshal should not be valid")
		}
	})
	t.Run("truncated null", func(t *testing.T) {
		s := Value("preset")
		if err := jsonv2.Unmarshal([]byte(`nul`), &s); err == nil {
			t.Fatalf("expected error for truncated null")
		}
	})
}

func TestJSONv2Struct(t *testing.T) {
	type User struct {
		ID    Int64  `json:"id"`
		Name  String `json:"name"`
		Email String `json:"email,omitzero"`
	}
	b, err := jsonv2.Marshal(User{ID: Value[int64](1)})
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"id":1,"name":null}`; string(b) != want {
		t.Fatalf("Marshal = %s, want %s", b, want)
	}

	var back User
	if err := jsonv2.Unmarshal([]byte(`{"id":2,"name":null,"email":"a@b.co"}`), &back); err != nil {
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

func TestOptionalJSONv2(t *testing.T) {
	var p userPatch
	if err := jsonv2.Unmarshal([]byte(`{"name":"Ann","email":null}`), &p); err != nil {
		t.Fatal(err)
	}
	checkPatch(t, p)

	b, err := jsonv2.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"name":"Ann","email":null}`; string(b) != want {
		t.Fatalf("Marshal = %s, want %s", b, want)
	}

	var bad userPatch
	if err := jsonv2.Unmarshal([]byte(`{"age":"old"}`), &bad); err == nil {
		t.Fatalf("expected error for bad payload")
	}
	if bad.Age.IsPresent() {
		t.Fatalf("failed decode should leave Age absent")
	}
}
