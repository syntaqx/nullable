package nullable

import (
	"database/sql"
	"encoding/json"
)

type Int64 struct {
	sql.NullInt64
}

func (ni Int64) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ni.Int64)
}

func (ni *Int64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ni.Int64 = 0
		ni.Valid = false
		return nil
	}
	err := json.Unmarshal(data, &ni.Int64)
	if err == nil {
		ni.Valid = true
	}
	return err
}

type Float64 struct {
	sql.NullFloat64
}

func (nf Float64) MarshalJSON() ([]byte, error) {
	if !nf.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nf.Float64)
}

func (nf *Float64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nf.Float64 = 0
		nf.Valid = false
		return nil
	}
	err := json.Unmarshal(data, &nf.Float64)
	if err == nil {
		nf.Valid = true
	}
	return err
}
