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
	"fmt"
	"io"
	"runtime/debug"
	"strings"
	"testing"
)

func TestDecode(t *testing.T) {
	type testcase struct {
		casename string
		bytes    []byte
		obj      *MPObject
	}

	Init()

	cases := []testcase{
		{"p fixint", []byte{0x01}, &MPObject{DataStr: "1", TypeName: "positive fixint"}},
		{"n fixint", []byte{0xff}, &MPObject{DataStr: "-1", TypeName: "negative fixint"}},
		{"nil", []byte{0xc0}, &MPObject{DataStr: "nil", TypeName: "nil"}},
		{"never used", []byte{0xc1}, &MPObject{DataStr: "(never used)", TypeName: "(never used)"}},
		{"true", []byte{0xc3}, &MPObject{DataStr: "true", TypeName: "true"}},
		{"false", []byte{0xc2}, &MPObject{DataStr: "false", TypeName: "false"}},
		{"float32", []byte{0xca, 0x80, 0x00, 0x00, 0x00}, &MPObject{DataStr: "-0.000000", TypeName: "float 32"}},
		{"float64", []byte{0xcb, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, &MPObject{DataStr: "-0.000000", TypeName: "float 64"}},
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
		{"array16", []byte{0xdc, 0x00, 0x0f, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00}, &MPObject{DataStr: "(array 16)", TypeName: "array 16"}},
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
		if bytes.Compare(ret.Raw, v.obj.Raw) != 0 {
			t.Errorf("%s: Raw data is different.", v.casename)
			t.Errorf(" given : %x", ret.Raw)
			t.Errorf(" expect: %x", v.obj.Raw)
		}

	}
}

func TestNestedMap(t *testing.T) {
	/* {"0":"0", "1":{"2":"2", "3":"3", "4":"4"}} in JSON */
	b := []byte{0x82, 0xa1, 0x30, 0xa1, 0x30, 0xa1, 0x31, 0x83, 0xa1, 0x32, 0xa1, 0x32, 0xa1, 0x33, 0xa1, 0x33, 0xa1, 0x34, 0xa1, 0x34}

	Init()

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

func BenchmarkDecodeNestedMap(b *testing.B) {
	/* {"0":"0", "1":{"2":"2", "3":"3", "4":"4"}, "5":"5" } in JSON */
	benchData := []byte{0x83, 0xa1, 0x30, 0xa1, 0x30, 0xa1, 0x31, 0x83, 0xa1, 0x32, 0xa1, 0x32, 0xa1, 0x33, 0xa1, 0x33, 0xa1, 0x34, 0xa1, 0x34, 0xa1, 0x35, 0xa1, 0x35}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Decode(bytes.NewBuffer(benchData))
		if err != nil && err != io.EOF {
			b.Errorf("Decode error %s", err)
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	/* {"val":12.3, "str":"hoge"} in JSON */
	benchData := []byte{0x82, 0xa3, 0x76, 0x61, 0x6c, 0xcb, 0x40, 0x28, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9a, 0xa3, 0x73, 0x74, 0x72, 0xa4, 0x68, 0x6f, 0x67, 0x65}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Decode(bytes.NewBuffer(benchData))
		if err != nil && err != io.EOF {
			b.Errorf("Decode error %s", err)
		}
	}
}

func TestIsArray(t *testing.T) {
	var i uint

	for i = 0; i <= 0xff; i++ {
		ok := IsArray(byte(i))
		if ok && ((i < 0x90 && i > 0x9f) && i != 0xdc && i != 0xdd) {
			t.Errorf("IsArray Error! 0x%x is not Array", i)
		} else if !ok && ((i >= 0x90 && i <= 0x9f) || i == 0xdc || i == 0xdd) {
			t.Errorf("IsArray Error! 0x%x is Array", i)
		}
	}
}

func TestIsMap(t *testing.T) {
	var i uint

	for i = 0; i <= 0xff; i++ {
		ok := IsMap(byte(i))
		if ok && ((i < 0x80 && i > 0x8f) && i != 0xde && i != 0xdf) {
			t.Errorf("IsMap Error! 0x%x is not Map", i)
		} else if !ok && ((i >= 0x80 && i <= 0x8f) || i == 0xde || i == 0xdf) {
			t.Errorf("IsMap Error! 0x%x is Map", i)
		}
	}
}

func TestIsString(t *testing.T) {
	var i uint

	for i = 0; i <= 0xff; i++ {
		ok := IsString(byte(i))
		if ok && ((i < 0xa0 && i > 0xbf) && (i < 0xd9 && i > 0xdb)) {
			t.Errorf("IsString Error! 0x%x is not String", i)
		} else if !ok && ((i >= 0xa0 && i <= 0xbf) || (i >= 0xd9 && i <= 0xdb)) {
			t.Errorf("IsString Error! 0x%x is String", i)
		}
	}
}

func TestIsExt(t *testing.T) {
	var i uint

	for i = 0; i <= 0xff; i++ {
		ok := IsExt(byte(i))
		if ok && ((i < 0xc7 && i > 0xc9) && (i < 0xd4 && i > 0xd8)) {
			t.Errorf("IsExt Error! 0x%x is not Ext", i)
		} else if !ok && ((i >= 0xc7 && i <= 0xc9) || (i >= 0xd4 && i <= 0xd8)) {
			t.Errorf("IsExt Error! 0x%x is Ext", i)
		}
	}
}

func TestShortenData(t *testing.T) {
	currentcase := ""

	type testcase struct {
		casename string
		bytes    []byte
	}

	defer func(str *string) {
		err := recover()
		if err != nil {
			t.Errorf("%s: panic occured. %s", *str, err)
			debug.PrintStack()
		}
	}(&currentcase)

	cases := []testcase{
		{"FixMap Only Header", []byte{0x82}},
		{"FixArray Only Header", []byte{0x92}},
		{"FixStr Only Header", []byte{0xa2}},
		{"Ext8 Only Header", []byte{0xc7}},
		{"Ext16 Only Header", []byte{0xc8}},
		{"Ext32 Only Header", []byte{0xc9}},
		{"FixExt1 Only Header", []byte{0xd4}},
		{"FixExt2 Only Header", []byte{0xd5}},
		{"FixExt4 Only Header", []byte{0xd6}},
		{"FixExt8 Only Header", []byte{0xd7}},
		{"FixExt16 Only Header", []byte{0xd8}},
		/* ----- no length -----*/
		{"Bin8 No Length", []byte{0xc4}},
		{"Bin16 No Length", []byte{0xc5}},
		{"Bin32 No Length", []byte{0xc6}},
		{"Str8 No Length", []byte{0xd9}},
		{"Str16 No Length", []byte{0xda}},
		{"Str32 No Length", []byte{0xdb}},
		{"Array16 No Length", []byte{0xdc}},
		{"Array32 No Length", []byte{0xdd}},
		{"Map16 No Length", []byte{0xde}},
		{"Map32 No Length", []byte{0xdf}},
		/* ----- shorten length -----*/
		/*  partial length field     */
		{"Bin16 Shorten Length", []byte{0xc5, 0x01}},
		{"Bin32 Shorten Length", []byte{0xc6, 0x01}},
		{"Ext16 Shorten Length", []byte{0xc8, 0x01}},
		{"Ext32 Shorten Length", []byte{0xc9, 0x01}},
		{"Str16 Shorten Length", []byte{0xda, 0x01}},
		{"Str32 Shorten Length", []byte{0xdb, 0x01}},
		{"Array16 Shorten Length", []byte{0xdc, 0x01}},
		{"Array32 Shorten Length", []byte{0xdd, 0x01}},
		{"Map16 Shorten Length", []byte{0xde, 0x01}},
		{"Map32 Shorten Length", []byte{0xdf, 0x01}},
		/* ----- shorten data -----*/
		{"Shorten FixMap KV", []byte{0x82, 0xa1, 0x60, 0x01 }},
		{"Shorten FixMap Key", []byte{0x82, 0xa1, 0x60}},
		{"Shorten FixArray", []byte{0x92, 0x01}},
		{"Shorten FixStr", []byte{0xa2, 0x10}},
		{"Shorten Bin8", []byte{0xc4, 0x03, 0x01}},
		{"Shorten Bin16", []byte{0xc5, 0x00, 0x10, 0x1}},
		{"Shorten Bin32", []byte{0xc6, 0x00, 0x01, 0x00, 0x00, 0x01}},
		{"Shorten Ext8", []byte{0xc7, 0x04, 0x01, 0x01}},
		{"Ext8 No Data", []byte{0xc8, 0x01, 0x01}},
		{"Shorten Ext16", []byte{0xc8, 0x00, 0x04, 0x01, 0x01}},
		{"Ext16 No Data", []byte{0xc8, 0x00, 0x01, 0x01}},
		{"Shorten Ext32", []byte{0xc9, 0x00, 0x00, 0x00, 0x04, 0x01, 0x01}},
		{"Ext32 No Data", []byte{0xc9, 0x00, 0x00, 0x00, 0x01, 0x01}},
		{"FixExt1 No Data", []byte{0xd4, 0x01}},
		{"Shorten FixExt2", []byte{0xd5, 0x01, 0x01}},
		{"FixExt2 No Data", []byte{0xd5, 0x01}},
		{"Shorten FixExt4", []byte{0xd6, 0x01, 0x03, 0x01}},
		{"FixExt4 No Data", []byte{0xd6, 0x01}},
		{"Shorten FixExt8", []byte{0xd7, 0x01, 0x01, 0x02}},
		{"FixExt8 No Data", []byte{0xd7, 0x01}},
		{"Shorten FixExt16", []byte{0xd8, 0x01, 0x01, 0x02}},
		{"FixExt16 No Data", []byte{0xd8, 0x01}},
		{"Shorten Str8", []byte{0xd9, 0x03, 0x01}},
		{"Shorten Str16", []byte{0xda, 0x00, 0x03, 0x01}},
		{"Shorten Str32", []byte{0xdb, 0x00, 0x00, 0x00, 0x04, 0x01}},
		{"Shorten Array16", []byte{0xdc, 0x00, 0x04, 0x01}},
		{"Shorten Array32", []byte{0xdd, 0x00, 0x00, 0x00, 0x04, 0x01}},
		{"Shorten Map16 KV", []byte{0xde, 0x00, 0x04, 0xa1, 0x60, 0x01}},
		{"Shorten Map16 Key", []byte{0xde, 0x00, 0x04, 0xa1, 0x60}},
		{"Shorten Map32 KV", []byte{0xdf, 0x00, 0x00, 0x00, 0x04, 0xa1, 0x60, 0x01}},
		{"Shorten Map32 Key", []byte{0xdf, 0x00, 0x00, 0x00, 0x04, 0xa1, 0x60}},
	}

	Init()
	for _, v := range cases {
		currentcase = v.casename
		_, err := Decode(bytes.NewBuffer(v.bytes))
		if err == nil {
			t.Errorf("%s: No error is detected", v.casename)
		}
	}
}

func TestNextWithError(t *testing.T) {

	errorcase := []byte{0x0, 0x1, 0x2, 0x3}
	buf := bytes.NewBuffer(errorcase)
	_, err := nextWithError(buf, len(errorcase)+1)
	if err == nil {
		t.Errorf("NextWithError successed")
	}

	okcase := []byte{0x0, 0x1, 0x2, 0x3}
	buf = bytes.NewBuffer(okcase)
	_, err = nextWithError(buf, len(okcase))
	if err != nil {
		t.Errorf("NextWithError is not successed")
	}
}
