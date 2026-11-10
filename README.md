# nullable

[![Go Reference](https://pkg.go.dev/badge/github.com/syntaqx/nullable.svg)](https://pkg.go.dev/github.com/syntaqx/nullable)
[![codecov](https://codecov.io/gh/syntaqx/nullable/graph/badge.svg?token=ksm1V45rxe)](https://codecov.io/gh/syntaqx/nullable)

A single generic type for values that may be null, with first-class support for
JSON and `database/sql`.

No more one wrapper struct per underlying type. Everything is `Nullable[T]`.

## Install

```bash
go get github.com/syntaqx/nullable
```

Requires Go 1.26+.

## Usage

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/syntaqx/nullable"
)

type User struct {
	ID    nullable.Int64  `json:"id"`
	Name  nullable.String `json:"name"`
	Email nullable.String `json:"email,omitzero"` // omitted when null
}

func main() {
	u := User{
		ID:   nullable.Value[int64](1),
		Name: nullable.Value("John Doe"),
		// Email left as its zero value -> null
	}

	b, _ := json.Marshal(u)
	fmt.Println(string(b)) // {"id":1,"name":"John Doe"}
}
```

## Constructing values

```go
nullable.Value("hi")        // non-null Nullable[string]
nullable.Null[string]()     // explicit null
nullable.FromPtr(ptr)       // nil pointer -> null, otherwise a copy

var x nullable.Int          // the zero value is a valid null
x.Set(42)                   // now non-null
x.SetNull()                 // back to null
```

## Reading values

```go
v, ok := n.Get()   // value and whether it is valid
n.Or(fallback)     // value if valid, else fallback
n.OrZero()         // value if valid, else the zero value of T
n.Ptr()            // *T, or nil when null
n.Valid()          // bool
n.IsNull()         // bool
```

## Transforming and comparing

```go
// Map runs the function only when non-null; a null stays null.
upper := nullable.Map(name, strings.ToUpper)

// Equal treats two nulls as equal.
nullable.Equal(a, b)
```

## Null: emit or omit

Both behaviors are first-class. Choose per field via the struct tag:

```go
type Patch struct {
	// Emits `"bio":null` when null (the default).
	Bio nullable.String `json:"bio"`

	// Dropped entirely when null (Go 1.24+ omitzero).
	Nickname nullable.String `json:"nickname,omitzero"`
}
```

A valid zero value (e.g. `nullable.Value("")`) is *not* treated as null, so an
explicit empty string still serializes as `""` rather than being omitted.

## encoding/json/v2

On Go 1.27+, `Nullable` and `Optional` also implement the `encoding/json/v2`
`MarshalerTo` and `UnmarshalerFrom` interfaces. They write to and read from the
v2 encoder and decoder directly, and the caller's options, such as
`StringifyNumbers`, apply to the wrapped value. `encoding/json` is built on v2
and uses these methods too, so v1 callers get the same speedup without changing
any code.

## Absent vs. null: `Optional`

`Nullable` has two states, so a missing field and an explicit `null` decode the
same way. For PATCH-style input, where "leave it alone" and "clear it" mean
different things, use `Optional[T]`. It has three states:

| JSON input           | `Optional` state | `IsPresent()` | Held `Nullable` |
| -------------------- | ---------------- | ------------- | --------------- |
| field missing        | absent (zero)    | `false`       | null            |
| `"email": null`      | null             | `true`        | null            |
| `"email": "a@b.co"`  | value            | `true`        | `"a@b.co"`      |

```go
type UserPatch struct {
	Name  nullable.Optional[string] `json:"name,omitzero"`
	Email nullable.Optional[string] `json:"email,omitzero"`
}

var p UserPatch
json.Unmarshal([]byte(`{"email":null}`), &p)

user.Name = p.Name.Or(user.Name)    // not in the request: unchanged
user.Email = p.Email.Or(user.Email) // explicit null: cleared
```

Use the `omitzero` tag option on `Optional` fields. Without it, an absent field
marshals as `null`. `Optional` does not implement the `database/sql` interfaces
because SQL has no way to store "absent". Use `Or` to resolve it to a
`Nullable` first.

Build one with `nullable.Present(nullable.Value(v))` or
`nullable.Present(nullable.Null[T]())`, or change it with `Set`, `SetNull` and
`Unset`.

## Why not `*T` or `sql.Null[T]`?

The standard library covers each piece separately, but no single type covers
them all:

|                                | `*T` | `sql.Null[T]` | `Nullable[T]` |
| ------------------------------ | ---- | ------------- | ------------- |
| JSON `null`                    | yes  | no\*          | yes           |
| SQL `NULL`                     | yes  | yes           | yes           |
| No aliasing or nil dereference | no   | yes           | yes           |
| `omitzero` drops null          | yes  | no            | yes           |
| Helpers (`Or`, `Map`, `Equal`) | no   | no            | yes           |

\* `sql.Null[T]` encodes as `{"V":…,"Valid":…}` under both JSON v1 and v2.

If you only need JSON, `*T` with `omitzero` is a fine choice. This package is
for when you want one type all the way from the database to the API. JSON v2
doesn't add a nullable type, so this hasn't changed.

## database/sql

`Nullable[T]` implements `driver.Valuer` and `sql.Scanner`, delegating
conversion to the standard library, so it works with any type your driver
supports:

```go
var name nullable.String
row := db.QueryRow("SELECT name FROM users WHERE id = ?", id)
if err := row.Scan(&name); err != nil {
	// ...
}

_, err := db.Exec("INSERT INTO users (name) VALUES (?)", name)
```

## Extending to other formats

The core package ships only the encodings that can represent a real null:
`encoding/json` and `database/sql`. Text formats can't tell `null` apart from an
empty value, and YAML/BSON decoders need their own interfaces, so baking them in
would either corrupt data or drag extra dependencies into every dependent.

Instead, it's designed to be extended. Embed the type and add the method your
format looks for, bridging through the public accessors so you decide how null
is represented:

```go
type YAMLString struct {
	nullable.String
}

func (s YAMLString) MarshalYAML() (any, error) {
	if v, ok := s.Get(); ok {
		return v, nil
	}
	return nil, nil // null
}
```

The embedded type's JSON and SQL behavior is promoted for free, so `YAMLString`
still works everywhere `nullable.String` does. For imperative codecs, drive
everything from `Get`, `Value`, `Null`, and `Set` with no wrapper required.

## Any type works

The aliases below are just conveniences; `nullable.Value(anything)` gives you a
`Nullable` of that type.

| Alias              | Underlying type |
| ------------------ | --------------- |
| `nullable.Bool`    | `bool`          |
| `nullable.Bytes`   | `[]byte`        |
| `nullable.Float32` | `float32`       |
| `nullable.Float64` | `float64`       |
| `nullable.Int`     | `int`           |
| `nullable.Int32`   | `int32`         |
| `nullable.Int64`   | `int64`         |
| `nullable.String`  | `string`        |
| `nullable.Time`    | `time.Time`     |

For anything else, use `nullable.Nullable[YourType]` directly.
