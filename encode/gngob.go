package encode

import (
	"bytes"
	"encoding/gob"
)

type GNgob struct{}

func (e GNgob) Encode(input interface{}) ([]byte, error) {
	var respBytes bytes.Buffer
	enc := gob.NewEncoder(&respBytes)
	if err := enc.Encode(input); err != nil {
		return nil, err
	}
	return respBytes.Bytes(), nil
}

func (e GNgob) Decode(input []byte, output interface{}) error {
	b := bytes.NewBuffer(input)
	dec := gob.NewDecoder(b)
	return dec.Decode(output)
}
