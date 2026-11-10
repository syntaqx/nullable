package nullable

import (
	"encoding/json"
	"testing"
)

type userPatch struct {
	Name  Optional[string] `json:"name,omitzero"`
	Email Optional[string] `json:"email,omitzero"`
	Age   Optional[int]    `json:"age,omitzero"`
}

func checkPatch(t *testing.T, p userPatch) {
	t.Helper()
	// name: present with a value
	if n, ok := p.Name.Get(); !ok || n.OrZero() != "Ann" {
		t.Fatalf("Name = %v, %v; want present Ann", n, ok)
	}
	// email: present and null
	if n, ok := p.Email.Get(); !ok || n.Valid() {
		t.Fatalf("Email = %v, %v; want present null", n, ok)
	}
	// age: absent
	if p.Age.IsPresent() {
		t.Fatalf("Age should be absent")
	}
}

func TestOptionalJSONThreeStates(t *testing.T) {
	var p userPatch
	if err := json.Unmarshal([]byte(`{"name":"Ann","email":null}`), &p); err != nil {
		t.Fatal(err)
	}
	checkPatch(t, p)

	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"name":"Ann","email":null}`; string(b) != want {
		t.Fatalf("Marshal = %s, want %s", b, want)
	}
}

func TestOptionalAbsentWithoutOmitzero(t *testing.T) {
	type s struct {
		Name Optional[string] `json:"name"`
	}
	b, err := json.Marshal(s{})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"name":null}` {
		t.Fatalf("Marshal = %s, want {\"name\":null}", b)
	}
}

func TestOptionalUnmarshalError(t *testing.T) {
	var p userPatch
	if err := json.Unmarshal([]byte(`{"age":"old"}`), &p); err == nil {
		t.Fatalf("expected error for bad payload")
	}
	if p.Age.IsPresent() {
		t.Fatalf("failed decode should leave Age absent")
	}
}

func TestOptionalOr(t *testing.T) {
	current := Value("old")
	var absent Optional[string]
	if got := absent.Or(current); !Equal(got, current) {
		t.Fatalf("absent.Or = %v, want %v", got, current)
	}
	if got := Present(Null[string]()).Or(current); got.Valid() {
		t.Fatalf("null.Or = %v, want null", got)
	}
	if got := Present(Value("new")).Or(current); got.OrZero() != "new" {
		t.Fatalf("value.Or = %v, want new", got)
	}
}

func TestOptionalMutators(t *testing.T) {
	var o Optional[int]
	if !o.IsZero() || o.IsPresent() || o.String() != "<absent>" {
		t.Fatalf("zero Optional should be absent, got %v", o)
	}
	o.Set(3)
	if n, ok := o.Get(); !ok || n.OrZero() != 3 || o.String() != "3" {
		t.Fatalf("after Set = %v, %v", n, ok)
	}
	o.SetNull()
	if n, ok := o.Get(); !ok || n.Valid() || o.String() != "<null>" {
		t.Fatalf("after SetNull = %v, %v", n, ok)
	}
	o.Unset()
	if o.IsPresent() || !o.IsZero() {
		t.Fatalf("after Unset should be absent")
	}
}

// TestOptionalJSONv1Methods calls the v1 methods directly; on Go 1.27+
// encoding/json routes through the v2 methods instead.
func TestOptionalJSONv1Methods(t *testing.T) {
	b, err := Present(Value(1)).MarshalJSON()
	if err != nil || string(b) != `1` {
		t.Fatalf("MarshalJSON = %s, %v", b, err)
	}
	var o Optional[int]
	if err := o.UnmarshalJSON([]byte(`null`)); err != nil || !o.IsPresent() {
		t.Fatalf("UnmarshalJSON(null) present=%v err=%v", o.IsPresent(), err)
	}
	var bad Optional[int]
	if err := bad.UnmarshalJSON([]byte(`"x"`)); err == nil || bad.IsPresent() {
		t.Fatalf("UnmarshalJSON(bad) present=%v err=%v", bad.IsPresent(), err)
	}
}
