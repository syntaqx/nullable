package nullable_test

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/syntaqx/nullable"
)

func ExampleValue() {
	n := nullable.Value("hello")
	fmt.Println(n.Or("fallback"))
	// Output: hello
}

func ExampleNull() {
	n := nullable.Null[string]()
	fmt.Println(n.Or("fallback"))
	// Output: fallback
}

func ExampleMap() {
	name := nullable.Value("ada")
	upper := nullable.Map(name, strings.ToUpper)
	fmt.Println(upper.Or(""))
	// Output: ADA
}

func ExampleEqual() {
	fmt.Println(nullable.Equal(nullable.Null[int](), nullable.Null[int]()))
	fmt.Println(nullable.Equal(nullable.Value(1), nullable.Value(2)))
	// Output:
	// true
	// false
}

func ExampleNullable_json() {
	type User struct {
		ID    nullable.Int64  `json:"id"`
		Name  nullable.String `json:"name"`
		Email nullable.String `json:"email,omitzero"`
	}

	// Email is left null and dropped thanks to the omitzero tag option.
	u := User{ID: nullable.Value[int64](1), Name: nullable.Value("Ada")}
	b, _ := json.Marshal(u)
	fmt.Println(string(b))
	// Output: {"id":1,"name":"Ada"}
}
