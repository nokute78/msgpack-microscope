package msgpack

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestDecode(t *testing.T) {
	type testcase struct {
		casename string
		bytes    []byte
		obj      *MPObject
	}

	typeNameInit()
	extFormatInit()

	cases := []testcase{
		{"p fixint", []byte{0x01}, &MPObject{DataStr: "1", TypeName: "positive fixint"}},
		{"n fixint", []byte{0xff}, &MPObject{DataStr: "-1", TypeName: "negative fixint"}},
		{"nil", []byte{0xc0}, &MPObject{DataStr: "nil", TypeName: "nil"}},
		{"never used", []byte{0xc1}, &MPObject{DataStr: "(never used)", TypeName: "(never used)"}},
		{"true", []byte{0xc3}, &MPObject{DataStr: "true", TypeName: "true"}},
		{"false", []byte{0xc2}, &MPObject{DataStr: "false", TypeName: "false"}},
		{"float32", []byte{0xca, 0x80, 0x00, 0x00, 0x00}, &MPObject{DataStr: "-0.000000", TypeName: "float 32"}},
		{"float64", []byte{0xcb, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, &MPObject{DataStr: "-0.000000", TypeName: "float 64"}},
		{"uint8", []byte{0xcc, 0xff}, &MPObject{DataStr: "255", TypeName: "uint 8"}},
		{"uint16", []byte{0xcd, 0xff, 0x00}, &MPObject{DataStr: "65280", TypeName: "uint 16"}},
		{"uint32", []byte{0xce, 0xff, 0x00, 0xff, 0x00}, &MPObject{DataStr: "4278255360", TypeName: "uint 32"}},
		{"uint64", []byte{0xcf, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00}, &MPObject{DataStr: "18374966859414961920", TypeName: "uint 64"}},
		{"int8", []byte{0xd0, 0xff}, &MPObject{DataStr: "-1", TypeName: "int 8"}},
		{"int16", []byte{0xd1, 0xff, 0x00}, &MPObject{DataStr: "-256", TypeName: "int 16"}},
		{"int32", []byte{0xd2, 0xff, 0x00, 0xff, 0x00}, &MPObject{DataStr: "-16711936", TypeName: "int 32"}},
		{"int64", []byte{0xd3, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00}, &MPObject{DataStr: "-71777214294589696", TypeName: "int 64"}},
		{"str8", []byte{0xd9, 0x0f, 0xe3, 0x81, 0x93, 0xe3, 0x82, 0x93, 0xe3, 0x81, 0xab, 0xe3, 0x81, 0xa1, 0xe3, 0x81, 0xaf}, &MPObject{DataStr: "こんにちは", TypeName: "str 8"}},
		{"bin8", []byte{0xc4, 0x04, 0xde, 0xad, 0xbe, 0xef}, &MPObject{DataStr: "0xdeadbeef", TypeName: "bin 8"}},
		{"fixstr len31", []byte{0xbf, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x30, 0x31}, &MPObject{DataStr: "1234567890123456789012345678901", TypeName: "fixstr"}},
		{"fixarray len2", []byte{0x92, 0x00, 0x01}, &MPObject{DataStr: "(fixarray)", TypeName: "fixarray"}},
		{"fixmap len2", []byte{0x82, 0xa1, 0x30, 0x00, 0xa1, 0x31, 0x01}, &MPObject{DataStr: "(fixmap)", TypeName: "fixmap"}},
		{"array16", []byte{0xdc, 0x00, 0x0f, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01}, &MPObject{DataStr: "(array 16)", TypeName: "array 16"}},
		{"fixext1", []byte{0xd4, 0x01, 0xff}, &MPObject{DataStr: "0xff", TypeName: "fixext 1", ExtType: 1}},
		{"fixext2", []byte{0xd5, 0x01, 0xfe, 0xed}, &MPObject{DataStr: "0xfeed", TypeName: "fixext 2", ExtType: 1}},
		{"fixext4", []byte{0xd6, 0x01, 0xde, 0xad, 0xbe, 0xef}, &MPObject{DataStr: "0xdeadbeef", TypeName: "fixext 4", ExtType: 1}},
		{"fixext8", []byte{0xd7, 0x01, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef}, &MPObject{DataStr: "0xdeadbeefdeadbeef", TypeName: "fixext 8", ExtType: 1}},
		{"ext8", []byte{0xc7, 0x04, 0x01, 0xde, 0xad, 0xbe, 0xef}, &MPObject{DataStr: "0xdeadbeef", TypeName: "ext 8", ExtType: 1}},
	}

	/* str16 */
	strcase := testcase{casename: "str16", obj: &MPObject{DataStr: strings.Repeat("こんにちは", 20), TypeName: "str 16"}}
	strcase.bytes = []byte{0xda, 0x01, 0x2c}
	for i := 0; i < 20; i++ {
		strcase.bytes = append(strcase.bytes, []byte{0xe3, 0x81, 0x93, 0xe3, 0x82, 0x93, 0xe3, 0x81, 0xab, 0xe3, 0x81, 0xa1, 0xe3, 0x81, 0xaf}...)
	}
	cases = append(cases, strcase)

	/* str32 */
	strcase = testcase{casename: "str32", obj: &MPObject{DataStr: strings.Repeat("こんにちは", 4370), TypeName: "str 32"}}
	strcase.bytes = []byte{0xdb, 0x00, 0x01, 0x00, 0x0e}
	for i := 0; i < 4370; i++ {
		strcase.bytes = append(strcase.bytes, []byte{0xe3, 0x81, 0x93, 0xe3, 0x82, 0x93, 0xe3, 0x81, 0xab, 0xe3, 0x81, 0xa1, 0xe3, 0x81, 0xaf}...)
	}
	cases = append(cases, strcase)

	/* bin16 */
	deadbeef := []byte{0xde, 0xad, 0xbe, 0xef}
	strcase = testcase{casename: "bin16", obj: &MPObject{DataStr: fmt.Sprintf("0x%x", bytes.Repeat(deadbeef, 64)), TypeName: "bin 16"}}
	strcase.bytes = []byte{0xc5, 0x01, 0x00}
	for i := 0; i < 64; i++ {
		strcase.bytes = append(strcase.bytes, deadbeef...)
	}
	cases = append(cases, strcase)

	/* bin32 */
	strcase = testcase{casename: "bin32", obj: &MPObject{DataStr: fmt.Sprintf("0x%x", bytes.Repeat(deadbeef, 16384)), TypeName: "bin 32"}}
	strcase.bytes = []byte{0xc6, 0x00, 0x01, 0x00, 0x00}
	for i := 0; i < 16384; i++ {
		strcase.bytes = append(strcase.bytes, deadbeef...)
	}
	cases = append(cases, strcase)

	/* ext16 */
	strcase = testcase{casename: "ext16", obj: &MPObject{DataStr: fmt.Sprintf("0x%x", bytes.Repeat(deadbeef, 64)), TypeName: "ext 16", ExtType: 1}}
	strcase.bytes = []byte{0xc8, 0x01, 0x00, 0x01}
	for i := 0; i < 64; i++ {
		strcase.bytes = append(strcase.bytes, deadbeef...)
	}
	cases = append(cases, strcase)

	/* ext32 */
	strcase = testcase{casename: "ext32", obj: &MPObject{DataStr: fmt.Sprintf("0x%x", bytes.Repeat(deadbeef, 16384)), TypeName: "ext 32", ExtType: 1}}
	strcase.bytes = []byte{0xc9, 0x00, 0x01, 0x00, 0x00, 0x01}
	for i := 0; i < 16384; i++ {
		strcase.bytes = append(strcase.bytes, deadbeef...)
	}
	cases = append(cases, strcase)

	/* array32 */
	strcase = testcase{casename: "array32", obj: &MPObject{DataStr: "(array 32)", TypeName: "array 32"}}
	strcase.bytes = []byte{0xdd, 0x00, 0x01, 0x00, 0x00}
	for i := 0; i < 8192; i++ {
		strcase.bytes = append(strcase.bytes, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}...)
	}
	cases = append(cases, strcase)

	/* map16 */
	strcase = testcase{casename: "map16", obj: &MPObject{DataStr: "(map 16)", TypeName: "map 16"}}
	strcase.bytes = []byte{0xde, 0x00, 0x10}
	for i := 0; i < 4; i++ {
		strcase.bytes = append(strcase.bytes, []byte{0xa1, 0x30, 0x00, 0xa1, 0x31, 0x01, 0xa1, 0x32, 0x02, 0xa1, 0x33, 0x03}...)
	}
	cases = append(cases, strcase)

	/* map32 */
	strcase = testcase{casename: "map32", obj: &MPObject{DataStr: "(map 32)", TypeName: "map 32"}}
	strcase.bytes = []byte{0xdf, 0x00, 0x01, 0x00, 0x00}
	for i := 0; i < 16384; i++ {
		strcase.bytes = append(strcase.bytes, []byte{0xa1, 0x30, 0x00, 0xa1, 0x31, 0x01, 0xa1, 0x32, 0x02, 0xa1, 0x33, 0x03}...)
	}
	cases = append(cases, strcase)

	for _, v := range cases {
		v.obj.Raw = v.bytes
		v.obj.FirstByte = v.bytes[0]
		ret, err := Decode(bytes.NewBuffer(v.bytes))
		if err != nil && err != io.EOF {
			t.Errorf("Decode error %s", err)
		}
		if ret.TypeName != v.obj.TypeName {
			t.Errorf("%s: TypeName mismatch: %s, expect %s", v.casename, ret.TypeName, v.obj.TypeName)
		}
		if ret.DataStr != v.obj.DataStr {
			t.Errorf("%s: DataStr mismatch: %s, expect %s", v.casename, ret.DataStr, v.obj.DataStr)
		}
	}
}

func TestNestedMap(t *testing.T) {
	/* {"0":"0", "1":{"2":"2", "3":"3", "4":"4"}} in JSON */
	b := []byte{0x82, 0xa1, 0x30, 0xa1, 0x30, 0xa1, 0x31, 0x83, 0xa1, 0x32, 0xa1, 0x32, 0xa1, 0x33, 0xa1, 0x33, 0xa1, 0x34, 0xa1, 0x34}

	typeNameInit()
	extFormatInit()

	ret, err := Decode(bytes.NewBuffer(b))
	if err != nil && err != io.EOF {
		t.Errorf("Decode error %s", err)
	}
	if len(ret.Child) != 2*2 {
		t.Errorf("Map Size error %d != 4", len(ret.Child))
	}

	testcase := []string{"0", "0", "1"}
	for i, v := range testcase {
		if ret.Child[i].DataStr != v {
			t.Errorf("Child[%d] error. \"%s\" is not \"%s\"", i, ret.Child[i].DataStr, v)
		}
	}

	if len(ret.Child[3].Child) != 3*2 {
		t.Errorf("Map Size error %d != 6", len(ret.Child[3].Child))
	}
	testcase = []string{"2", "2", "3", "3", "4", "4"}
	for i, v := range testcase {
		if ret.Child[3].Child[i].DataStr != v {
			t.Errorf("Child[%d] error. \"%s\" is not \"%s\"", 3+i, ret.Child[3].Child[i].DataStr, v)
		}
	}

}
