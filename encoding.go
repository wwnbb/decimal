// Derived from github.com/govalues/decimal v0.1.36 (MIT), see LICENSE.
package decimal

import (
	"bytes"
	"fmt"
)

// MarshalJSON implements [json.Marshaler]. It writes the decimal as a JSON
// number with exactly the digits of [Decimal.String], so 1.50 stays 1.50.
func (d Decimal) MarshalJSON() ([]byte, error) {
	return d.append(make([]byte, 0, 24)), nil
}

// UnmarshalJSON implements [json.Unmarshaler]. It accepts a JSON number or a
// JSON string that holds a decimal; the text is read with [Parse], so
// whitespace inside a string is trimmed. A JSON null is an error matching
// [ErrSyntax]: a Decimal has no null value (use *Decimal for that). On error
// d is left unchanged.
func (d *Decimal) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if string(data) == "null" {
		return fmt.Errorf("unmarshaling %T: %w: null", Decimal{}, ErrSyntax)
	}
	if len(data) > 0 && data[0] == '"' {
		if len(data) < 2 || data[len(data)-1] != '"' || bytes.IndexByte(data, '\\') >= 0 {
			return fmt.Errorf("unmarshaling %T: %w: bad string", Decimal{}, ErrSyntax)
		}
		data = data[1 : len(data)-1]
	}
	f, err := parse(bytes.TrimSpace(data))
	if err != nil {
		return fmt.Errorf("unmarshaling %T: %w", Decimal{}, err)
	}
	*d = f
	return nil
}

// MarshalText implements [encoding.TextMarshaler]. It returns [Decimal.String].
func (d Decimal) MarshalText() ([]byte, error) {
	return d.append(make([]byte, 0, 24)), nil
}

// AppendText implements [encoding.TextAppender]. It appends [Decimal.String]
// to b and never fails; it does not allocate if b has enough capacity.
func (d Decimal) AppendText(b []byte) ([]byte, error) {
	return d.append(b), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler]. It reads the text
// with [Parse], so surrounding whitespace is trimmed. On error d is left
// unchanged.
func (d *Decimal) UnmarshalText(text []byte) error {
	f, err := parse(bytes.TrimSpace(text))
	if err != nil {
		return fmt.Errorf("unmarshaling %T: %w", Decimal{}, err)
	}
	*d = f
	return nil
}
