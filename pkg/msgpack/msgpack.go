/*
   Copyright 2019 Takahiro Yamashita

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package msgpack

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// First Byte of each format.
//   https://github.com/msgpack/msgpack/blob/master/spec.md#overview
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

// IsArray reports whether the byte is array format family header.
func IsArray(b byte) bool {
	return (isFixArray(b) || b == Array16Format || b == Array32Format)
}

// IsMap reports whether the byte is map format family header.
func IsMap(b byte) bool {
	return (isFixMap(b) || b == Map16Format || b == Map32Format)
}

// IsString reports whether the byte is str format family header.
func IsString(b byte) bool {
	return (isFixStr(b) || (b >= Str8Format && b <= Str32Format))
}

// IsExt reports whether the byte is ext format family header.
func IsExt(b byte) bool {
	return (isExt(b) || isFixExt(b))
}

// Init initializes internal ext format settings.
func Init() {
	extFormatInit()
}

func typeStr(b byte) string {
	switch b {
	case NilFormat:
		return "nil"
	case NeverUsedFormat:
		return "(never used)"
	case FalseFormat:
		return "false"
	case TrueFormat:
		return "true"
	case Bin8Format:
		return "bin 8"
	case Bin16Format:
		return "bin 16"
	case Bin32Format:
		return "bin 32"
	case Ext8Format:
		return "ext 8"
	case Ext16Format:
		return "ext 16"
	case Ext32Format:
		return "ext 32"
	case Float32Format:
		return "float 32"
	case Float64Format:
		return "float 64"
	case Uint8Format:
		return "uint 8"
	case Uint16Format:
		return "uint 16"
	case Uint32Format:
		return "uint 32"
	case Uint64Format:
		return "uint 64"
	case Int8Format:
		return "int 8"
	case Int16Format:
		return "int 16"
	case Int32Format:
		return "int 32"
	case Int64Format:
		return "int 64"
	case FixExt1Format:
		return "fixext 1"
	case FixExt2Format:
		return "fixext 2"
	case FixExt4Format:
		return "fixext 4"
	case FixExt8Format:
		return "fixext 8"
	case FixExt16Format:
		return "fixext 16"
	case Str8Format:
		return "str 8"
	case Str16Format:
		return "str 16"
	case Str32Format:
		return "str 32"
	case Array16Format:
		return "array 16"
	case Array32Format:
		return "array 32"
	case Map16Format:
		return "map 16"
	case Map32Format:
		return "map 32"
	}
	return ""
}

// MPObject represents message pack object.
// If the object is array or map, MPObject has Child which represents each element.
type MPObject struct {
	FirstByte byte
	TypeName  string
	ExtType   int8   /* for ext family*/
	Length    uint32 /* for map, array and str family*/
	DataStr   string
	Raw       []byte
	Child     []*MPObject
}

// String implements Stringer interface.
func (obj *MPObject) String() string {
	switch {
	case IsArray(obj.FirstByte) || IsMap(obj.FirstByte):
		return fmt.Sprintf(`%s(0x%02x): length=%d`, obj.TypeName, obj.FirstByte, obj.Length)
	case IsString(obj.FirstByte):
		return fmt.Sprintf(`%s(0x%02x): val="%s"`, obj.TypeName, obj.FirstByte, obj.DataStr)
	case IsExt(obj.FirstByte):
		return fmt.Sprintf(`%s(0x%02x): type=%d val=%s`, obj.TypeName, obj.FirstByte, obj.ExtType, obj.DataStr)
	default:
		return fmt.Sprintf(`%s(0x%02x): val=%s`, obj.TypeName, obj.FirstByte, obj.DataStr)
	}
}

func nextWithError(buf *bytes.Buffer, n int) ([]byte, error) {
	bufs := buf.Next(n)
	if len(bufs) != n {
		return bufs, fmt.Errorf("bytes.Buffer Next error")
	}
	return bufs, nil
}

func (obj *MPObject) setLengthFromBytes(size int, buf *bytes.Buffer) error {
	if size != 1 && size != 2 && size != 4 {
		return fmt.Errorf("illegal size %d", size)
	}

	length, err := nextWithError(buf, size)
	if err != nil {
		return err
	}
	obj.Raw = append(obj.Raw, length...)

	switch size {
	case 1:
		obj.Length = uint32(length[0])
	case 2:
		obj.Length = uint32(binary.BigEndian.Uint16(length))
	case 4:
		obj.Length = binary.BigEndian.Uint32(length)
	}
	return nil
}

func (obj *MPObject) setNum(size int, buf *bytes.Buffer, conv func([]byte) string) error {
	bufs, err := nextWithError(buf, size)
	if err != nil {
		return err
	}
	obj.Raw = append(obj.Raw, bufs...)
	obj.DataStr = conv(bufs)

	return nil
}

func (obj *MPObject) setCollection(buf *bytes.Buffer, length int) error {
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

// Decode analyzes buf and convert MPObject.
func Decode(buf *bytes.Buffer) (*MPObject, error) {
	Init()
	return decode(buf)
}

func decode(buf *bytes.Buffer) (*MPObject, error) {
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
		err := obj.setCollection(buf, int(obj.Length)*2)
		if err != nil {
			return obj, err
		}
	case isFixArray(firstbyte):
		obj.TypeName = "fixarray"
		obj.Length = uint32(firstbyte & 0xf)
		obj.DataStr = "(fixarray)"
		err := obj.setCollection(buf, int(obj.Length))
		if err != nil {
			return obj, err
		}
	case isFixStr(firstbyte):
		obj.TypeName = "fixstr"
		obj.Length = uint32(firstbyte & 0x1f)
		bufs, err := nextWithError(buf, int(obj.Length))
		if err != nil {
			return obj, err
		}
		obj.DataStr = string(bufs)
		obj.Raw = append(obj.Raw, bufs...)
	case isFixExt(firstbyte):
		obj.TypeName = typeStr(firstbyte)
		err := obj.setExtType(buf)
		if err != nil {
			return obj, err
		}
		data, err := nextWithError(buf, 1<<uint(firstbyte-FixExt1Format))
		if err != nil {
			return obj, err
		}
		if !obj.setRegisteredExt(data) {
			obj.DataStr = fmt.Sprintf("0x%x", data)
		}
		obj.Raw = append(obj.Raw, data...)
	case isExt(firstbyte):
		obj.TypeName = typeStr(firstbyte)
		/* length */
		switch firstbyte {
		case Ext8Format:
			err := obj.setLengthFromBytes(1, buf)
			if err != nil {
				return obj, err
			}
		case Ext16Format:
			err := obj.setLengthFromBytes(2, buf)
			if err != nil {
				return obj, err
			}
		case Ext32Format:
			err := obj.setLengthFromBytes(4, buf)
			if err != nil {
				return obj, err
			}
		}

		/* type */
		err := obj.setExtType(buf)
		if err != nil {
			return obj, err
		}

		data, err := nextWithError(buf, int(obj.Length))
		if err != nil {
			return obj, err
		}

		if !obj.setRegisteredExt(data) {
			obj.DataStr = fmt.Sprintf("0x%x", data)
		}
		obj.Raw = append(obj.Raw, data...)

	default:
		obj.TypeName = typeStr(firstbyte)
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
			err := obj.setNum(1, buf, func(b []byte) string {
				return fmt.Sprintf("%d", uint8(b[0]))
			})
			if err != nil {
				return obj, err
			}
		case Uint16Format:
			obj.setNum(2, buf, func(b []byte) string {
				return fmt.Sprintf("%d", (binary.BigEndian.Uint16(b)))
			})
			if err != nil {
				return obj, err
			}
		case Uint32Format:
			obj.setNum(4, buf, func(b []byte) string {
				return fmt.Sprintf("%d", (binary.BigEndian.Uint32(b)))
			})
			if err != nil {
				return obj, err
			}
		case Uint64Format:
			obj.setNum(8, buf, func(b []byte) string {
				return fmt.Sprintf("%d", (binary.BigEndian.Uint64(b)))
			})
			if err != nil {
				return obj, err
			}

			/* Int family */
		case Int8Format:
			obj.setNum(1, buf, func(b []byte) string {
				var v int8
				if binary.Read(bytes.NewReader(b), binary.BigEndian, &v) != nil {
					return ""
				}
				return fmt.Sprintf("%d", v)
			})
			if err != nil {
				return obj, err
			}
		case Int16Format:
			obj.setNum(2, buf, func(b []byte) string {
				var v int16
				if binary.Read(bytes.NewReader(b), binary.BigEndian, &v) != nil {
					return ""
				}
				return fmt.Sprintf("%d", v)
			})
			if err != nil {
				return obj, err
			}
		case Int32Format:
			obj.setNum(4, buf, func(b []byte) string {
				var v int32
				if binary.Read(bytes.NewReader(b), binary.BigEndian, &v) != nil {
					return ""
				}
				return fmt.Sprintf("%d", v)
			})
			if err != nil {
				return obj, err
			}
		case Int64Format:
			obj.setNum(8, buf, func(b []byte) string {
				var v int64
				if binary.Read(bytes.NewReader(b), binary.BigEndian, &v) != nil {
					return ""
				}
				return fmt.Sprintf("%d", v)
			})
			if err != nil {
				return obj, err
			}
		case Float32Format:
			obj.setNum(4, buf, func(b []byte) string {
				var v float32
				if binary.Read(bytes.NewReader(b), binary.BigEndian, &v) != nil {
					return ""
				}
				return fmt.Sprintf("%f", v)
			})
			if err != nil {
				return obj, err
			}
		case Float64Format:
			obj.setNum(8, buf, func(b []byte) string {
				var v float64
				if binary.Read(bytes.NewReader(b), binary.BigEndian, &v) != nil {
					return ""
				}
				return fmt.Sprintf("%f", v)
			})
			if err != nil {
				return obj, err
			}
		case Str8Format:
			err := obj.setLengthFromBytes(1, buf)
			if err != nil {
				return obj, err
			}

			str, err := nextWithError(buf, int(obj.Length))
			if err != nil {
				return obj, err
			}
			obj.Raw = append(obj.Raw, str...)
			obj.DataStr = string(str)

		case Str16Format:
			err := obj.setLengthFromBytes(2, buf)
			if err != nil {
				return obj, err
			}
			str, err := nextWithError(buf, int(obj.Length))
			if err != nil {
				return obj, err
			}
			obj.Raw = append(obj.Raw, str...)
			obj.DataStr = string(str)

		case Str32Format:
			err := obj.setLengthFromBytes(4, buf)
			if err != nil {
				return obj, err
			}
			str, err := nextWithError(buf, int(obj.Length))
			if err != nil {
				return obj, err
			}
			obj.Raw = append(obj.Raw, str...)
			obj.DataStr = string(str)

		case Bin8Format:
			err := obj.setLengthFromBytes(1, buf)
			if err != nil {
				return obj, err
			}
			bins, err := nextWithError(buf, int(obj.Length))
			if err != nil {
				return obj, err
			}
			obj.Raw = append(obj.Raw, bins...)
			obj.DataStr = fmt.Sprintf("0x%x", bins)
		case Bin16Format:
			err := obj.setLengthFromBytes(2, buf)
			if err != nil {
				return obj, err
			}
			bins, err := nextWithError(buf, int(obj.Length))
			if err != nil {
				return obj, err
			}
			obj.Raw = append(obj.Raw, bins...)
			obj.DataStr = fmt.Sprintf("0x%x", bins)

		case Bin32Format:
			err := obj.setLengthFromBytes(4, buf)
			if err != nil {
				return obj, err
			}
			bins, err := nextWithError(buf, int(obj.Length))
			if err != nil {
				return obj, err
			}
			obj.Raw = append(obj.Raw, bins...)
			obj.DataStr = fmt.Sprintf("0x%x", bins)
		case Array16Format:
			err := obj.setLengthFromBytes(2, buf)
			if err != nil {
				return obj, err
			}
			obj.DataStr = "(array 16)"
			err = obj.setCollection(buf, int(obj.Length))
			if err != nil {
				return nil, err
			}

		case Array32Format:
			err := obj.setLengthFromBytes(4, buf)
			if err != nil {
				return obj, err
			}
			obj.DataStr = "(array 32)"
			err = obj.setCollection(buf, int(obj.Length))
			if err != nil {
				return nil, err
			}

		case Map16Format:
			err := obj.setLengthFromBytes(2, buf)
			if err != nil {
				return obj, err
			}
			obj.DataStr = "(map 16)"
			err = obj.setCollection(buf, int(obj.Length)*2)
			if err != nil {
				return nil, err
			}

		case Map32Format:
			err := obj.setLengthFromBytes(4, buf)
			if err != nil {
				return obj, err
			}
			obj.DataStr = "(map 32)"
			err = obj.setCollection(buf, int(obj.Length)*2)
			if err != nil {
				return nil, err
			}
		}
	}
	return obj, nil
}
