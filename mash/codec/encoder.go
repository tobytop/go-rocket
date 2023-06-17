package codec

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/proto"
)

type jsonCodec struct{}
type protoCodec struct{}
type bytesCodec struct{}

var (
	DefaultGRPCCodecs = map[string]encoding.Codec{
		"application/json":         jsonCodec{},
		"application/proto":        protoCodec{},
		"application/protobuf":     protoCodec{},
		"application/octet-stream": protoCodec{},
		"application/grpc":         protoCodec{},
		"application/grpc+json":    jsonCodec{},
		"application/grpc+proto":   protoCodec{},
		"application/grpc+bytes":   protoCodec{},
	}
)

func (protoCodec) Marshal(v any) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (protoCodec) Unmarshal(data []byte, v any) error {
	return proto.Unmarshal(data, v.(proto.Message))
}

func (protoCodec) Name() string {
	return "proto"
}

func (jsonCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (jsonCodec) Name() string {
	return "json"
}
