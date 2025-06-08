package encoding

import (
	"github.com/go-yaml/yaml"
	"io"
)

func YAMLDecoder(r io.Reader, v *map[string]interface{}) error {
	return yaml.NewDecoder(r).Decode(v)
}
