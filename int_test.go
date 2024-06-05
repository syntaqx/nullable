package nullable

import (
	"database/sql"
	"encoding/json"
	"testing"
)

func TestInt64_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		ni      Int64
		want    string
		wantErr bool
	}{
		{
			name: "Valid int",
			ni:   Int64{NullInt64: sql.NullInt64{Int64: 123, Valid: true}},
			want: `123`,
		},
		{
			name: "Null int",
			ni:   Int64{NullInt64: sql.NullInt64{Valid: false}},
			want: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.ni)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() got = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestInt64_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    Int64
		wantErr bool
	}{
		{
			name: "Valid int",
			data: `123`,
			want: Int64{NullInt64: sql.NullInt64{Int64: 123, Valid: true}},
		},
		{
			name: "Null int",
			data: `null`,
			want: Int64{NullInt64: sql.NullInt64{Int64: 0, Valid: false}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ni Int64
			if err := json.Unmarshal([]byte(tt.data), &ni); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if ni != tt.want {
				t.Errorf("UnmarshalJSON() got = %v, want %v", ni, tt.want)
			}
		})
	}
}
