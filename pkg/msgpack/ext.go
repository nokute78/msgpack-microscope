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
	"time"
)

// ExtFormat represents Ext Format.
type ExtFormat struct {
	FirstByte  byte
	ExtType    int8
	TypeName   string
	DecodeFunc func([]byte) string
}

// String implements Stringer interface.
func (obj *ExtFormat) String() string {
	return fmt.Sprintf(`%s(0x%02x): type=%d func=%v`, obj.TypeName, obj.FirstByte, obj.ExtType, obj.DecodeFunc)
}

/* Decode functions for Timestamp extension type. */
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
	sec := int64(raw & 0x3FFFFFFFF)
	nsec := int64((raw & 0xFFFFFFFC00000000) >> 34)
	return fmt.Sprintf("%v", time.Unix(sec, nsec))
}
func timestamp96(b []byte) string {
	if len(b) != 12 {
		return ""
	}
	nsec := int64(binary.BigEndian.Uint32(b[0:4]))
	var sec int64
	if binary.Read(bytes.NewReader(b[4:]), binary.BigEndian, &sec) != nil {
		return ""
	}
	return fmt.Sprintf("%v", time.Unix(sec, nsec))
}

/* Fluentd EventTime Ext Format */
/* https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1 */
func extEventTimeV1(b []byte) string {
	if len(b) != 8 {
		return ""
	}
	var sec int32
	if binary.Read(bytes.NewReader(b[:4]), binary.BigEndian, &sec) != nil {
		return ""
	}
	var nsec int32
	if binary.Read(bytes.NewReader(b[4:]), binary.BigEndian, &nsec) != nil {
		return ""
	}
	return fmt.Sprintf("%v", time.Unix(int64(sec), int64(nsec)))
}

var extFormats map[byte]([]*ExtFormat)

// RegisterFluentdEventTime registers Fluentd ext timestamp format.
// https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1 */
func RegisterFluentdEventTime() {
	RegisterExt(&ExtFormat{FirstByte: FixExt8Format, ExtType: 0, TypeName: "event time", DecodeFunc: extEventTimeV1})
	RegisterExt(&ExtFormat{FirstByte: Ext8Format, ExtType: 0, TypeName: "event time", DecodeFunc: extEventTimeV1})
}

// RegisterExt register user defined ext format.
func RegisterExt(ext *ExtFormat) {
	extFormats[ext.FirstByte] = append(extFormats[ext.FirstByte], ext)
}

func extFormatInit() {
	if len(extFormats) != 0 {
		return
	}
	extFormats = map[byte]([]*ExtFormat){}

	RegisterExt(&ExtFormat{FirstByte: FixExt4Format, ExtType: -1, TypeName: "timestamp 32", DecodeFunc: timestamp32})
	RegisterExt(&ExtFormat{FirstByte: FixExt8Format, ExtType: -1, TypeName: "timestamp 64", DecodeFunc: timestamp64})
	RegisterExt(&ExtFormat{FirstByte: Ext8Format, ExtType: -1, TypeName: "timestamp 96", DecodeFunc: timestamp96})
}

func (obj *MPObject) setExtType(buf *bytes.Buffer) error {
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

func (obj *MPObject) setRegisteredExt(extData []byte) bool {
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
