package nullable_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/syntaqx/nullable"
)

// yamlMarshaler mirrors the gopkg.in/yaml.v3 Marshaler interface. Declaring it
// locally lets us prove the extension pattern satisfies that contract without
// pulling YAML into the module.
type yamlMarshaler interface {
	MarshalYAML() (any, error)
}

// YAMLString is a user-side extension: embed the core type, then add the method
// the format needs, bridging through the public accessors.
type YAMLString struct {
	nullable.String
}

func (s YAMLString) MarshalYAML() (any, error) {
	if v, ok := s.Get(); ok {
		return v, nil
	}
	return nil, nil
}

func TestExtend_EmbedForForeignFormat(t *testing.T) {
	// The embedding type satisfies the foreign interface...
	var _ yamlMarshaler = YAMLString{}

	// ...and still carries the core behavior via promotion.
	got := YAMLString{nullable.Value("hi")}
	if v, ok := got.Get(); !ok || v != "hi" {
		t.Fatalf("promoted Get() = %q, %v", v, ok)
	}

	node, err := got.MarshalYAML()
	if err != nil {
		t.Fatal(err)
	}
	if node != "hi" {
		t.Fatalf("valid MarshalYAML() = %v, want hi", node)
	}

	node, err = YAMLString{nullable.Null[string]()}.MarshalYAML()
	if err != nil {
		t.Fatal(err)
	}
	if node != nil {
		t.Fatalf("null MarshalYAML() = %v, want nil", node)
	}
}

// TestExtend_ValueAPIBridge shows the imperative escape hatch: any codec can be
// driven purely from the exported API, with the caller deciding how null looks.
func TestExtend_ValueAPIBridge(t *testing.T) {
	// Encode: turn a Nullable into a wire token.
	encode := func(n nullable.Int) string {
		if v, ok := n.Get(); ok {
			return strconv.Itoa(v)
		}
		return "~" // this fictional format spells null as "~"
	}
	// Decode: rebuild a Nullable from a wire token.
	decode := func(tok string) nullable.Int {
		if tok == "~" {
			return nullable.Null[int]()
		}
		n, _ := strconv.Atoi(tok)
		return nullable.Value(n)
	}

	for _, in := range []nullable.Int{nullable.Value(42), nullable.Null[int]()} {
		if got := decode(encode(in)); !nullable.Equal(got, in) {
			t.Fatalf("round trip mismatch: %v -> %q -> %v", in, encode(in), got)
		}
	}
}

func Example_extending() {
	// A format the core package does not ship: represent null however you like.
	fmt.Println(mustYAML(YAMLString{nullable.Value("hello")}))
	fmt.Println(mustYAML(YAMLString{nullable.Null[string]()}))
	// Output:
	// hello
	// <nil>
}

func mustYAML(m yamlMarshaler) any {
	v, err := m.MarshalYAML()
	if err != nil {
		panic(err)
	}
	return v
}
