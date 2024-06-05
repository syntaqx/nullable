# nullable

[![Go Reference](https://pkg.go.dev/badge/github.com/syntaqx/nullable.svg)](https://pkg.go.dev/github.com/syntaqx/nullable)
[![Report card](https://goreportcard.com/badge/github.com/syntaqx/nullable)](https://goreportcard.com/report/github.com/syntaqx/nullable)

A way to represent nullable types in Go with JSON serialization.

## Usage

```bash
go get github.com/syntaqx/nullable
```

## Example

```go
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/syntaqx/nullable"
)

type User struct {
	ID   nullable.Int64
	Name nullable.String
}

func main() {
	u := User{
		Name: nullable.String{
			sql.NullString{String: "John Doe", Valid: true},
		},
	}

	b, err := json.Marshal(u)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}

```
