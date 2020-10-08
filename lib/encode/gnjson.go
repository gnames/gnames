package encode

import (
	"bytes"
	"strings"

	"github.com/gnames/gnames/lib/format"
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

func (e GNjson) Output(input interface{}, f format.Format) string {
	switch f {
	case format.CompactJSON:
		e.Pretty = false
	case format.PrettyJSON:
		e.Pretty = true
	default:
		return ""
	}
	resByte, err := e.Encode(input)
	if err != nil {
		return ""
	}
	res := string(resByte)
	res = strings.Replace(res, "\\u0026", "&", -1)
	return res
}
