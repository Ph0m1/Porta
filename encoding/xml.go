package encoding

import (
	"encoding/xml"
	"io"
)

func XMLDecoder(r io.Reader, v *map[string]interface{}) error {
	return xml.NewDecoder(r).Decode(v)
}
