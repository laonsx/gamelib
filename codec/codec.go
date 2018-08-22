package codec

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

func MsgPack(v interface{}) ([]byte, error) {
	byteBuf := new(bytes.Buffer)
	enc := codec.NewEncoder(byteBuf, &codec.MsgpackHandle)
	err := enc.Encode(v)
	return byteBuf.Bytes(), err
}

func UnMsgPack(data []byte, v interface{}) error {
	dec := codec.NewDecoder(bytes.NewReader(data), &codec.MsgpackHandle)
	return dec.Decode(v)
}
