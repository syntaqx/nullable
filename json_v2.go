//go:build go1.27

package nullable

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
)

// MarshalJSONTo implements the encoding/json/v2 MarshalerTo interface. A null
// Nullable encodes as JSON null. The value is written straight to the encoder
// using the caller's options, so v2 settings such as StringifyNumbers or
// Deterministic apply to the wrapped value.
//
// encoding/json (v1) is implemented on top of v2 and prefers this method over
// MarshalJSON, so v1 callers benefit too.
func (n Nullable[T]) MarshalJSONTo(enc *jsontext.Encoder) error {
	if !n.valid {
		return enc.WriteToken(jsontext.Null)
	}
	return jsonv2.MarshalEncode(enc, n.value)
}

// UnmarshalJSONFrom implements the encoding/json/v2 UnmarshalerFrom interface.
// JSON null decodes to a null Nullable; any other value is decoded into T with
// the caller's options and marks n valid.
func (n *Nullable[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if dec.PeekKind() == 'n' {
		if _, err := dec.ReadToken(); err != nil {
			return err
		}
		n.SetNull()
		return nil
	}
	if err := jsonv2.UnmarshalDecode(dec, &n.value); err != nil {
		return err
	}
	n.valid = true
	return nil
}

// MarshalJSONTo implements the encoding/json/v2 MarshalerTo interface. An
// absent Optional encodes as null; use the omitzero tag option to leave it out.
func (o Optional[T]) MarshalJSONTo(enc *jsontext.Encoder) error {
	return o.value.MarshalJSONTo(enc)
}

// UnmarshalJSONFrom implements the encoding/json/v2 UnmarshalerFrom interface.
// It is only called when the field is in the input, so a successful decode
// always marks o present.
func (o *Optional[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if err := o.value.UnmarshalJSONFrom(dec); err != nil {
		return err
	}
	o.set = true
	return nil
}

// Compile-time checks that Nullable and Optional satisfy the v2 interfaces.
var (
	_ jsonv2.MarshalerTo     = Nullable[int]{}
	_ jsonv2.UnmarshalerFrom = (*Nullable[int])(nil)
	_ jsonv2.MarshalerTo     = Optional[int]{}
	_ jsonv2.UnmarshalerFrom = (*Optional[int])(nil)
)
