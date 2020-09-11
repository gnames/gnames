package encode

import (
	"bytes"

	jsoniter "github.com/json-iterator/go"
)

type GNjson struct {
	Pretty bool
}

func (e GNjson) Encode(input interface{}) ([]byte, error) {
	if e.Pretty {
		return jsoniter.MarshalIndent(input, "", "  ")
	}
	return jsoniter.Marshal(input)
}

func (e GNjson) Decode(input []byte, output interface{}) error {
	r := bytes.NewReader(input)
	err := jsoniter.NewDecoder(r).Decode(output)
	return err
}
