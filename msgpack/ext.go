package msgpack

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

type extFormat struct {
	ExtType    int8
	TypeName   string
	DecodeFunc func([]byte) string
}

/* https://github.com/msgpack/msgpack/blob/master/spec.md#extension-types */
func timestamp32(b []byte) string {
	if len(b) != 4 {
		return ""
	}
	return fmt.Sprintf("%v", time.Unix(int64(binary.BigEndian.Uint32(b)), 0))
}
func timestamp64(b []byte) string {
	if len(b) != 8 {
		return ""
	}
	raw := binary.BigEndian.Uint64(b)
	var sec int64 = int64(raw & 0x3FFFFFFFF)
	var nsec int64 = int64((raw & 0xFFFFFFFC00000000) >> 34)
	return fmt.Sprintf("%v", time.Unix(sec, nsec))
}
func timestamp96(b []byte) string {
	if len(b) != 12 {
		return ""
	}
	var nsec int64 = int64(binary.BigEndian.Uint32(b[0:4]))
	var sec int64
	if binary.Read(bytes.NewReader(b[4:]), binary.BigEndian, &sec) != nil {
		return ""
	}
	return fmt.Sprintf("%v", time.Unix(sec, nsec))
}

var extFormats map[byte]([]*extFormat)

func RegisterExt(b byte, ext *extFormat) {
	extFormats[b] = append(extFormats[b], ext)
}

func extFormatInit() {
	extFormats = map[byte]([]*extFormat){}

	RegisterExt(FixExt4Format, &extFormat{ExtType: -1, TypeName: "timestamp 32", DecodeFunc: timestamp32})
	RegisterExt(FixExt8Format, &extFormat{ExtType: -1, TypeName: "timestamp 64", DecodeFunc: timestamp64})
	RegisterExt(Ext8Format, &extFormat{ExtType: -1, TypeName: "timestamp 96", DecodeFunc: timestamp96})
}

func (obj *MPObject) SetExtType(buf *bytes.Buffer) error {
	types := buf.Next(1)
	obj.Raw = append(obj.Raw, types...)

	var v int8
	err := binary.Read(bytes.NewReader(types), binary.BigEndian, &v)
	if err != nil {
		return err
	}
	obj.ExtType = v
	return nil
}

func (obj *MPObject) SetRegisteredExt(extData []byte) bool {
	list, ok := extFormats[obj.FirstByte]
	if ok && len(list) > 0 {
		for _, v := range list {
			if v.ExtType == obj.ExtType {
				obj.TypeName = v.TypeName
				obj.DataStr = v.DecodeFunc(extData)
				return true
			}
		}
	}
	return false
}
