package rpc

import (
	"encoding/json"

	"github.com/golang/protobuf/proto"
)

func Unmarshal(codec CodecType, data []byte, v interface{}) error {

	var err error

	switch codec {
	case CodecType_ProtoBuf:

		err = proto.Unmarshal(data, v.(proto.Message))

	case CodecType_Json:

		err = json.Unmarshal(data, v)

	}

	return err
}

func Marshal(codec CodecType, v interface{}) (data []byte, err error) {

	switch codec {
	case CodecType_ProtoBuf:

		data, err = proto.Marshal(v.(proto.Message))

	case CodecType_Json:

		data, err = json.Marshal(v)

	}

	return data, err
}
