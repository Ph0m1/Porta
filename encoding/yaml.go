package encoding

import (
	"io"

	"github.com/go-yaml/yaml"
)

func YAMLDecoder(r io.Reader, v *map[string]interface{}) error {
	return yaml.NewDecoder(r).Decode(v)
}
