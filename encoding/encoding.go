package encoding

import "io"

// Read from r, into map of interfaces
type Decoder func(r io.Reader, v *map[string]interface{}) error
