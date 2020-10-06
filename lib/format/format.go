package format

import "fmt"

type Format int

const (
	FormatNone Format = iota
	CSV
	CompactJSON
	PrettyJSON
)

var formatStringMap = map[string]Format{
	"csv": CSV, "compact": CompactJSON, "pretty": PrettyJSON,
}

var formatMap = map[Format]string{
	FormatNone:  "none",
	CSV:         "CSV",
	CompactJSON: "compact JSON",
	PrettyJSON:  "pretty JSON",
}

func (f Format) String() string {
	return formatMap[f]
}

func NewFormat(s string) (Format, error) {
	if f, ok := formatStringMap[s]; ok {
		return f, nil
	}

	err := fmt.Errorf(
		"cannot convert '%s' to format, use 'csv', 'compact' or 'pretty' as input",
		s,
	)
	return FormatNone, err
}
