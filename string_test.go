package nullable

import (
	"database/sql"
	"encoding/json"
	"testing"
)

func TestString_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		ns      String
		want    string
		wantErr bool
	}{
		{
			name: "Valid string",
			ns:   String{NullString: sql.NullString{String: "test", Valid: true}},
			want: `"test"`,
		},
		{
			name: "Null string",
			ns:   String{NullString: sql.NullString{Valid: false}},
			want: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.ns)
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

func TestString_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    String
		wantErr bool
	}{
		{
			name: "Valid string",
			data: `"test"`,
			want: String{NullString: sql.NullString{String: "test", Valid: true}},
		},
		{
			name: "Null string",
			data: `null`,
			want: String{NullString: sql.NullString{String: "", Valid: false}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ns String
			if err := json.Unmarshal([]byte(tt.data), &ns); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if ns != tt.want {
				t.Errorf("UnmarshalJSON() got = %v, want %v", ns, tt.want)
			}
		})
	}
}
