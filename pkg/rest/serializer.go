package rest

import (
	"encoding/json"
	"io"
)

var serializerMap = map[string]Serializer{
	"application/json": NewJSONSerializer(),
}

type Serializer interface {
	Encoder
	Decoder
}

type Encoder interface {
	Encode(v any, w io.Writer) error
}

type Decoder interface {
	Decode(data []byte, v any) error
}

func SerializerForMediaType(mediaType string) (Serializer, bool) {
	s, ok := serializerMap[mediaType]
	if !ok {
		return nil, false
	}
	return s, true
}

type jsonSerializer struct{}

func (s *jsonSerializer) Decode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
func (s *jsonSerializer) Encode(v any, w io.Writer) error {
	return json.NewEncoder(w).Encode(v)
}

func NewJSONSerializer() *jsonSerializer {
	return new(jsonSerializer)

}
