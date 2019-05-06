package msgpack

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	NilFormat byte = 0xc0 + iota
	NeverUsedFormat
	FalseFormat
	TrueFormat
	Bin8Format
	Bin16Format
	Bin32Format
	Ext8Format
	Ext16Format
	Ext32Format
	Float32Format
	Float64Format
	Uint8Format
	Uint16Format
	Uint32Format
	Uint64Format
	Int8Format
	Int16Format
	Int32Format
	Int64Format
	FixExt1Format
	FixExt2Format
	FixExt4Format
	FixExt8Format
	FixExt16Format
	Str8Format
	Str16Format
	Str32Format
	Array16Format
	Array32Format
	Map16Format
	Map32Format
)

func isPositiveFixInt(b byte) bool {
	return (b >= 0x00 && b <= 0x7f)
}

func isNegativeFixInt(b byte) bool {
	return (b >= 0xe0 && b <= 0xff)
}

func isFixMap(b byte) bool {
	return (b >= 0x80 && b <= 0x8f)
}

func isFixArray(b byte) bool {
	return (b >= 0x90 && b <= 0x9f)
}

func isFixStr(b byte) bool {
	return (b >= 0xa0 && b <= 0xbf)
}

func isFixExt(b byte) bool {
	return (b >= 0xd4 && b <= 0xd8)
}
func isExt(b byte) bool {
	return (b >= 0xc7 && b <= 0xc9)
}

var typeStr map[byte]string

func typeNameInit() {
	typeStr = map[byte]string{}

	typeStr[NilFormat] = "nil"
	typeStr[NeverUsedFormat] = "(never used)"
	typeStr[FalseFormat] = "false"
	typeStr[TrueFormat] = "true"
	typeStr[Bin8Format] = "bin 8"
	typeStr[Bin16Format] = "bin 16"
	typeStr[Bin32Format] = "bin 32"
	typeStr[Ext8Format] = "ext 8"
	typeStr[Ext16Format] = "ext 16"
	typeStr[Ext32Format] = "ext 32"
	typeStr[Float32Format] = "float 32"
	typeStr[Float64Format] = "float 64"
	typeStr[Uint8Format] = "uint 8"
	typeStr[Uint16Format] = "uint 16"
	typeStr[Uint32Format] = "uint 32"
	typeStr[Uint64Format] = "uint 64"
	typeStr[Int8Format] = "int 8"
	typeStr[Int16Format] = "int 16"
	typeStr[Int32Format] = "int 32"
	typeStr[Int64Format] = "int 64"
	typeStr[FixExt1Format] = "fixext 1"
	typeStr[FixExt2Format] = "fixext 2"
	typeStr[FixExt4Format] = "fixext 4"
	typeStr[FixExt8Format] = "fixext 8"
	typeStr[FixExt16Format] = "fixext 16"
	typeStr[Str8Format] = "str 8"
	typeStr[Str16Format] = "str 16"
	typeStr[Str32Format] = "str 32"
	typeStr[Array16Format] = "array 16"
	typeStr[Array32Format] = "array 32"
	typeStr[Map16Format] = "map 16"
	typeStr[Map32Format] = "map 32"
}

type MPObject struct {
	FirstByte byte
	TypeName  string
	ExtType   int8   /* for ext family*/
	Length    uint32 /* for map, array and str family*/
	DataStr   string
	Raw       []byte
	Child     []*MPObject
}

func (obj *MPObject) SetNum(size int, buf *bytes.Buffer, conv func([]byte) string) {
	bufs := buf.Next(size)
	obj.Raw = append(obj.Raw, bufs...)
	obj.DataStr = conv(bufs)
}

func (obj *MPObject) SetCollection(buf *bytes.Buffer, length int) error {
	obj.Child = make([]*MPObject, length)

	for i := 0; i < length; i++ {
		mpobj, err := Decode(buf)
		if err != nil {
			return err
		}
		obj.Child[i] = mpobj
		obj.Raw = append(obj.Raw, mpobj.Raw...)
	}
	return nil
}

func Decode(buf *bytes.Buffer) (*MPObject, error) {
	firstbyte, err := buf.ReadByte()
	if err != nil {
		/* io.EOF ?*/
		return nil, err
	}
	obj := &MPObject{FirstByte: firstbyte, Raw: []byte{firstbyte}}

	switch {
	case isPositiveFixInt(firstbyte):
		obj.TypeName = "positive fixint"
		obj.DataStr = fmt.Sprintf("%d", int8(firstbyte))
	case isNegativeFixInt(firstbyte):
		obj.TypeName = "negative fixint"
		obj.DataStr = fmt.Sprintf("%d", int8(firstbyte))
	case isFixMap(firstbyte):
		obj.TypeName = "fixmap"
		obj.Length = uint32(firstbyte & 0xf)
		obj.DataStr = "(fixmap)"
		err := obj.SetCollection(buf, int(obj.Length)*2)
		if err != nil {
			return nil, err
		}
	case isFixArray(firstbyte):
		obj.TypeName = "fixarray"
		obj.Length = uint32(firstbyte & 0xf)
		obj.DataStr = "(fixarray)"
		err := obj.SetCollection(buf, int(obj.Length))
		if err != nil {
			return nil, err
		}
	case isFixStr(firstbyte):
		obj.TypeName = "fixstr"
		obj.Length = uint32(firstbyte & 0x1f)
		bufs := buf.Next(int(obj.Length))
		if len(bufs) < int(obj.Length) {
			return nil, io.EOF
		}
		obj.DataStr = string(bufs)
		obj.Raw = append(obj.Raw, bufs...)
	case isFixExt(firstbyte):
		obj.TypeName = typeStr[firstbyte]
		err := obj.SetExtType(buf)
		if err != nil {
			return nil, err
		}
		data := buf.Next(1 << uint(firstbyte-FixExt1Format))
		if !obj.SetRegisteredExt(data) {
			obj.DataStr = fmt.Sprintf("0x%x", data)
		}
		obj.Raw = append(obj.Raw, data...)
	case isExt(firstbyte):
		obj.TypeName = typeStr[firstbyte]
		var length []byte
		/* length */
		switch firstbyte {
		case Ext8Format:
			length = buf.Next(1)
			obj.Length = uint32(length[0])
		case Ext16Format:
			length = buf.Next(2)
			obj.Length = uint32(binary.BigEndian.Uint16(length))
		case Ext32Format:
			length = buf.Next(4)
			obj.Length = binary.BigEndian.Uint32(length)
		}
		obj.Raw = append(obj.Raw, length...)

		/* type */
		err := obj.SetExtType(buf)
		if err != nil {
			return nil, err
		}

		data := buf.Next(int(obj.Length))
		if !obj.SetRegisteredExt(data) {
			obj.DataStr = fmt.Sprintf("0x%x", data)
		}
		obj.Raw = append(obj.Raw, data...)

	default:
		obj.TypeName = typeStr[firstbyte]
		switch firstbyte {
		case NilFormat:
			obj.DataStr = "nil"
		case NeverUsedFormat:
			obj.DataStr = "(never used)"
		case TrueFormat:
			obj.DataStr = "true"
		case FalseFormat:
			obj.DataStr = "false"

			/* Uint family*/
		case Uint8Format:
			obj.SetNum(1, buf, func(b []byte) string {
				return fmt.Sprintf("%d", uint8(b[0]))
			})
		case Uint16Format:
			obj.SetNum(2, buf, func(b []byte) string {
				return fmt.Sprintf("%d", (binary.BigEndian.Uint16(b)))
			})
		case Uint32Format:
			obj.SetNum(4, buf, func(b []byte) string {
				return fmt.Sprintf("%d", (binary.BigEndian.Uint32(b)))
			})
		case Uint64Format:
			obj.SetNum(8, buf, func(b []byte) string {
				return fmt.Sprintf("%d", (binary.BigEndian.Uint64(b)))
			})

			/* Int family */
		case Int8Format:
			obj.SetNum(1, buf, func(b []byte) string {
				var v int8
				if binary.Read(bytes.NewReader(b), binary.BigEndian, &v) != nil {
					return ""
				}
				return fmt.Sprintf("%d", v)
			})
		case Int16Format:
			obj.SetNum(2, buf, func(b []byte) string {
				var v int16
				if binary.Read(bytes.NewReader(b), binary.BigEndian, &v) != nil {
					return ""
				}
				return fmt.Sprintf("%d", v)
			})
		case Int32Format:
			obj.SetNum(4, buf, func(b []byte) string {
				var v int32
				if binary.Read(bytes.NewReader(b), binary.BigEndian, &v) != nil {
					return ""
				}
				return fmt.Sprintf("%d", v)
			})
		case Int64Format:
			obj.SetNum(8, buf, func(b []byte) string {
				var v int64
				if binary.Read(bytes.NewReader(b), binary.BigEndian, &v) != nil {
					return ""
				}
				return fmt.Sprintf("%d", v)
			})
		case Float32Format:
			obj.SetNum(4, buf, func(b []byte) string {
				var v float32
				if binary.Read(bytes.NewReader(b), binary.BigEndian, &v) != nil {
					return ""
				}
				return fmt.Sprintf("%f", v)
			})
		case Float64Format:
			obj.SetNum(8, buf, func(b []byte) string {
				var v float64
				if binary.Read(bytes.NewReader(b), binary.BigEndian, &v) != nil {
					return ""
				}
				return fmt.Sprintf("%f", v)
			})
		case Str8Format:
			length := buf.Next(1)
			obj.Raw = append(obj.Raw, length...)

			obj.Length = uint32(length[0])
			str := buf.Next(int(obj.Length))
			obj.Raw = append(obj.Raw, str...)
			obj.DataStr = string(str)

		case Str16Format:
			length := buf.Next(2)
			obj.Raw = append(obj.Raw, length...)

			obj.Length = uint32(binary.BigEndian.Uint16(length))
			str := buf.Next(int(obj.Length))
			obj.Raw = append(obj.Raw, str...)
			obj.DataStr = string(str)

		case Str32Format:
			length := buf.Next(4)
			obj.Raw = append(obj.Raw, length...)

			obj.Length = binary.BigEndian.Uint32(length)
			str := buf.Next(int(obj.Length))
			obj.Raw = append(obj.Raw, str...)
			obj.DataStr = string(str)

		case Bin8Format:
			length := buf.Next(1)
			obj.Raw = append(obj.Raw, length...)

			obj.Length = uint32(length[0])
			bins := buf.Next(int(obj.Length))
			obj.Raw = append(obj.Raw, bins...)
			obj.DataStr = fmt.Sprintf("0x%x", bins)
		case Bin16Format:
			length := buf.Next(2)
			obj.Raw = append(obj.Raw, length...)

			obj.Length = uint32(binary.BigEndian.Uint16(length))
			bins := buf.Next(int(obj.Length))
			obj.Raw = append(obj.Raw, bins...)
			obj.DataStr = fmt.Sprintf("0x%x", bins)

		case Bin32Format:
			length := buf.Next(4)
			obj.Raw = append(obj.Raw, length...)

			obj.Length = binary.BigEndian.Uint32(length)
			bins := buf.Next(int(obj.Length))
			obj.Raw = append(obj.Raw, bins...)
			obj.DataStr = fmt.Sprintf("0x%x", bins)
		case Array16Format:
			length := buf.Next(2)
			obj.Raw = append(obj.Raw, length...)

			obj.Length = uint32(binary.BigEndian.Uint16(length))
			obj.DataStr = "(array 16)"
			err := obj.SetCollection(buf, int(obj.Length))
			if err != nil {
				return nil, err
			}

		case Array32Format:
			length := buf.Next(4)
			obj.Raw = append(obj.Raw, length...)
			obj.Length = binary.BigEndian.Uint32(length)
			obj.DataStr = "(array 32)"
			err := obj.SetCollection(buf, int(obj.Length))
			if err != nil {
				return nil, err
			}

		case Map16Format:
			length := buf.Next(2)
			obj.Raw = append(obj.Raw, length...)

			obj.Length = uint32(binary.BigEndian.Uint16(length))
			obj.DataStr = "(map 16)"
			err := obj.SetCollection(buf, int(obj.Length)*2)
			if err != nil {
				return nil, err
			}

		case Map32Format:
			length := buf.Next(4)
			obj.Raw = append(obj.Raw, length...)

			obj.Length = binary.BigEndian.Uint32(length)
			obj.DataStr = "(map 32)"
			err := obj.SetCollection(buf, int(obj.Length)*2)
			if err != nil {
				return nil, err
			}
		}
	}
	return obj, nil
}
