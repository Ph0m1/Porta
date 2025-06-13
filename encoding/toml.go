package encoding

import (
	"fmt"
	"io"

	"github.com/BurntSushi/toml"
)

//func TOMLDecoder(r io.Reader, v *map[string]interface{}) error {
//	_, err := toml.NewDecoder(r).Decode(v)
//	return err
//}

func TOMLDecoder(r io.Reader, v *map[string]interface{}) error {
	md, err := toml.NewDecoder(r).Decode(v)
	if err != nil {
		return err
	}

	// 检查是否有未解码的键（多余的配置项）
	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		return fmt.Errorf("unknown configuration keys: %v", undecoded)
	}

	return nil
}
