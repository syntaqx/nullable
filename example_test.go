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

func ExampleOptional_patch() {
	type UserPatch struct {
		Name  nullable.Optional[string] `json:"name,omitzero"`
		Email nullable.Optional[string] `json:"email,omitzero"`
	}

	name := nullable.Value("Ada")
	email := nullable.Value("ada@example.com")

	// Email is cleared explicitly; Name is not mentioned, so it is left alone.
	var p UserPatch
	_ = json.Unmarshal([]byte(`{"email":null}`), &p)

	name = p.Name.Or(name)
	email = p.Email.Or(email)
	fmt.Println(name, email)
	// Output: Ada <null>
}
