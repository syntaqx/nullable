package nullable

import (
	"database/sql"
	"encoding/json"
)

type String struct {
	sql.NullString
}

func (ns String) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

func (ns *String) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ns.String = ""
		ns.Valid = false
		return nil
	}
	err := json.Unmarshal(data, &ns.String)
	if err == nil {
		ns.Valid = true
	}
	return err
}
