package encode

// Encoder interface allows to switch between different encoding types.
type Encoder interface {
	//Encode takes a Go object and converts it into bytes
	Encode(input interface{}) ([]byte, error)
	// Decode takes an input of bytes and decodes it into Go object.
	Decode(input []byte, output interface{}) error
}
