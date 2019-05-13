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

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nokute78/msgpack2txt/pkg/msgpack"
	"strings"
	"testing"
)

func TestOutputJSON(t *testing.T) {
	type testcase struct {
		casename string
		msgpdata []byte
		expected string
	}

	cases := []testcase{
		{"fixstr", []byte{0xa2, 0x41, 0x42}, `"AB"`},
		{"fixint", []byte{0x01}, "1"},
		{"fixarray", []byte{0x92, 0x01, 0x02}, "[1,2]"},
		{"array size 0", []byte{0x90}, "[]"},
		{"fixmap", []byte{0x82, 0xa1, 0x41, 0x00, 0xa1, 0x42, 0x01}, `{"A":0,"B":1}`},
		{"map size 0", []byte{0x80}, "{}"},
		{"nested map", []byte{0x82, 0xa1, 0x30, 0xa1, 0x30, 0xa1, 0x31, 0x83, 0xa1, 0x32, 0xa1, 0x32, 0xa1, 0x33, 0xa1, 0x33, 0xa1, 0x34, 0xa1, 0x34}, `{"0":"0","1":{"2":"2","3":"3","4":"4"}}`},
		{"nested array", []byte{0x82, 0xa1, 0x30, 0xa1, 0x30, 0xa1, 0x31, 0x93, 0x00, 0x01, 0x02}, `{"0":"0","1":[0,1,2]}`},

		{"n fixint", []byte{0xff}, "-1"},
		{"nil", []byte{0xc0}, "null"},
		{"never used", []byte{0xc1}, "(never used)"},
		{"true", []byte{0xc3}, "true"},
		{"false", []byte{0xc2}, "false"},
		{"float32", []byte{0xca, 0x80, 0x00, 0x00, 0x00}, "-0.000000"},
		{"float64", []byte{0xcb, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, "-0.000000"},
		{"uint8", []byte{0xcc, 0xff}, "255"},
		{"uint16", []byte{0xcd, 0xff, 0x00}, "65280"},
		{"uint32", []byte{0xce, 0xff, 0x00, 0xff, 0x00}, "4278255360"},
		{"uint64", []byte{0xcf, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00}, "18374966859414961920"},
		{"int8", []byte{0xd0, 0xff}, "-1"},
		{"int16", []byte{0xd1, 0xff, 0x00}, "-256"},
		{"int32", []byte{0xd2, 0xff, 0x00, 0xff, 0x00}, "-16711936"},
		{"int64", []byte{0xd3, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00}, "-71777214294589696"},
		{"str8", []byte{0xd9, 0x0f, 0xe3, 0x81, 0x93, 0xe3, 0x82, 0x93, 0xe3, 0x81, 0xab, 0xe3, 0x81, 0xa1, 0xe3, 0x81, 0xaf}, `"こんにちは"`},
		{"bin8", []byte{0xc4, 0x04, 0xde, 0xad, 0xbe, 0xef}, `"0xdeadbeef"`},
		{"fixstr len31", []byte{0xbf, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x30, 0x31}, `"1234567890123456789012345678901"`},
		{"array16", []byte{0xdc, 0x00, 0x0f, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00}, "[0,1,0,1,0,1,0,1,0,1,0,1,0,1,0]"},
		{"fixext1", []byte{0xd4, 0x01, 0xff}, "0xff"},
		{"fixext2", []byte{0xd5, 0x01, 0xfe, 0xed}, "0xfeed"},
		{"fixext4", []byte{0xd6, 0x01, 0xde, 0xad, 0xbe, 0xef}, "0xdeadbeef"},
		{"fixext8", []byte{0xd7, 0x01, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef}, "0xdeadbeefdeadbeef"},
		{"ext8", []byte{0xc7, 0x04, 0x01, 0xde, 0xad, 0xbe, 0xef}, "0xdeadbeef"},
	}

	/* str16 */
	strcase := testcase{casename: "str16", expected: `"` + strings.Repeat("こんにちは", 20) + `"`}
	strcase.msgpdata = []byte{0xda, 0x01, 0x2c}
	for i := 0; i < 20; i++ {
		strcase.msgpdata = append(strcase.msgpdata, []byte{0xe3, 0x81, 0x93, 0xe3, 0x82, 0x93, 0xe3, 0x81, 0xab, 0xe3, 0x81, 0xa1, 0xe3, 0x81, 0xaf}...)
	}
	cases = append(cases, strcase)

	/* str32 */
	strcase = testcase{casename: "str32", expected: `"` + strings.Repeat("こんにちは", 4370) + `"`}
	strcase.msgpdata = []byte{0xdb, 0x00, 0x01, 0x00, 0x0e}
	for i := 0; i < 4370; i++ {
		strcase.msgpdata = append(strcase.msgpdata, []byte{0xe3, 0x81, 0x93, 0xe3, 0x82, 0x93, 0xe3, 0x81, 0xab, 0xe3, 0x81, 0xa1, 0xe3, 0x81, 0xaf}...)
	}
	cases = append(cases, strcase)

	/* bin16 */
	deadbeef := []byte{0xde, 0xad, 0xbe, 0xef}
	strcase = testcase{casename: "bin16", expected: fmt.Sprintf(`"0x%x"`, bytes.Repeat(deadbeef, 64))}
	strcase.msgpdata = []byte{0xc5, 0x01, 0x00}
	for i := 0; i < 64; i++ {
		strcase.msgpdata = append(strcase.msgpdata, deadbeef...)
	}
	cases = append(cases, strcase)

	/* bin32 */
	strcase = testcase{casename: "bin32", expected: fmt.Sprintf(`"0x%x"`, bytes.Repeat(deadbeef, 16384))}
	strcase.msgpdata = []byte{0xc6, 0x00, 0x01, 0x00, 0x00}
	for i := 0; i < 16384; i++ {
		strcase.msgpdata = append(strcase.msgpdata, deadbeef...)
	}
	cases = append(cases, strcase)

	/* ext16 */
	strcase = testcase{casename: "ext16", expected: fmt.Sprintf("0x%x", bytes.Repeat(deadbeef, 64))}
	strcase.msgpdata = []byte{0xc8, 0x01, 0x00, 0x01}
	for i := 0; i < 64; i++ {
		strcase.msgpdata = append(strcase.msgpdata, deadbeef...)
	}
	cases = append(cases, strcase)

	/* ext32 */
	strcase = testcase{casename: "ext32", expected: fmt.Sprintf("0x%x", bytes.Repeat(deadbeef, 16384))}
	strcase.msgpdata = []byte{0xc9, 0x00, 0x01, 0x00, 0x00, 0x01}
	for i := 0; i < 16384; i++ {
		strcase.msgpdata = append(strcase.msgpdata, deadbeef...)
	}
	cases = append(cases, strcase)

	/* array32 */
	strcase = testcase{casename: "array32", expected: "[" + strings.Repeat("0,1,2,3,4,5,6,7,", 8191) + "0,1,2,3,4,5,6,7]"}
	strcase.msgpdata = []byte{0xdd, 0x00, 0x01, 0x00, 0x00}
	for i := 0; i < 8192; i++ {
		strcase.msgpdata = append(strcase.msgpdata, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}...)
	}
	cases = append(cases, strcase)

	/* map16 */
	strcase = testcase{casename: "map16", expected: "{" + strings.Repeat(`"0":0,"1":1,"2":2,"3":3,`, 3) + `"0":0,"1":1,"2":2,"3":3}`}
	strcase.msgpdata = []byte{0xde, 0x00, 0x10}
	for i := 0; i < 4; i++ {
		strcase.msgpdata = append(strcase.msgpdata, []byte{0xa1, 0x30, 0x00, 0xa1, 0x31, 0x01, 0xa1, 0x32, 0x02, 0xa1, 0x33, 0x03}...)
	}
	cases = append(cases, strcase)

	/* map32 */
	strcase = testcase{casename: "map32", expected: "{" + strings.Repeat(`"0":0,"1":1,"2":2,"3":3,`, 16383) + `"0":0,"1":1,"2":2,"3":3}`}
	strcase.msgpdata = []byte{0xdf, 0x00, 0x01, 0x00, 0x00}
	for i := 0; i < 16384; i++ {
		strcase.msgpdata = append(strcase.msgpdata, []byte{0xa1, 0x30, 0x00, 0xa1, 0x31, 0x01, 0xa1, 0x32, 0x02, 0xa1, 0x33, 0x03}...)
	}
	cases = append(cases, strcase)

	buf := bytes.Buffer{}
	for _, v := range cases {
		buf.Reset()
		ret, err := msgpack.Decode(bytes.NewBuffer(v.msgpdata))
		if err != nil {
			t.Errorf("%s: Decode failed. Error: %s", v.casename, err)
			continue
		}
		outputJSON(ret, &buf, 0)

		if buf.String() != v.expected {
			t.Logf("%s: mismatch. given: %s. expected: %s", v.casename, buf.String(), v.expected)
		}
		given := strings.Replace(buf.String(), " ", "", -1)
		expected := strings.Replace(v.expected, " ", "", -1)

		if given != expected {
			t.Errorf("%s: mismatch. given: %s. expected: %s", v.casename, given, expected)
		}
	}
}

type MPBase struct {
	Format string `json:format`
	Byte   string `json:header`
	Raw    string `json:raw`
}

type MPString struct {
	MPBase
	Value string `json:value`
}

func TestVerboseJSONString(t *testing.T) {
	type testcase struct {
		casename string
		bytes    []byte
		expected string
	}

	cases := []testcase{
		{"fixstr", []byte{0xa2, 0x41, 0x42}, `AB`},
		{"str8", []byte{0xd9, 0x0f, 0xe3, 0x81, 0x93, 0xe3, 0x82, 0x93, 0xe3, 0x81, 0xab, 0xe3, 0x81, 0xa1, 0xe3, 0x81, 0xaf}, `こんにちは`},
		{"bin8", []byte{0xc4, 0x04, 0xde, 0xad, 0xbe, 0xef}, "0xdeadbeef"},
	}

	buf := bytes.Buffer{}
	for _, v := range cases {
		buf.Reset()
		ret, err := msgpack.Decode(bytes.NewBuffer(v.bytes))
		outputVerboseJSON(ret, &buf, 0)

		p := MPString{}
		err = json.Unmarshal(buf.Bytes(), &p)
		if err != nil {
			t.Errorf("%s: Unmarshal Error %s", v.casename, err)
		}
		if v.expected != p.Value {
			t.Errorf("%s: mismatch. given: %s. expected: %s", v.casename, p.Value, v.expected)
		}
	}
}

type MPExt struct {
	MPBase
	Type  int8   `json:type`
	Value string `json:value`
}

func TestVerboseJSONExt(t *testing.T) {
	type testcase struct {
		casename string
		bytes    []byte
		expected string
	}

	cases := []testcase{
		{"fixext1", []byte{0xd4, 0x01, 0xff}, "0xff"},
		{"fixext2", []byte{0xd5, 0x01, 0xfe, 0xed}, "0xfeed"},
		{"fixext4", []byte{0xd6, 0x01, 0xde, 0xad, 0xbe, 0xef}, "0xdeadbeef"},
		{"fixext8", []byte{0xd7, 0x01, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef}, "0xdeadbeefdeadbeef"},
	}

	buf := bytes.Buffer{}
	for _, v := range cases {
		buf.Reset()
		ret, err := msgpack.Decode(bytes.NewBuffer(v.bytes))
		outputVerboseJSON(ret, &buf, 0)

		p := MPExt{}
		err = json.Unmarshal(buf.Bytes(), &p)
		if err != nil {
			t.Errorf("%s: Unmarshal Error %s", v.casename, err)
		}
		if v.expected != p.Value {
			t.Errorf("%s: mismatch. given: %s. expected: %s", v.casename, p.Value, v.expected)
		}
	}
}

type MPBool struct {
	MPBase
	Value bool `json:value`
}

func TestVerboseJSONBool(t *testing.T) {
	type testcase struct {
		casename string
		bytes    []byte
		expected bool
	}

	cases := []testcase{
		{"true", []byte{0xc3}, true},
		{"false", []byte{0xc2}, false},
	}

	buf := bytes.Buffer{}
	for _, v := range cases {
		buf.Reset()
		ret, err := msgpack.Decode(bytes.NewBuffer(v.bytes))
		outputVerboseJSON(ret, &buf, 0)

		p := MPBool{}
		err = json.Unmarshal(buf.Bytes(), &p)
		if err != nil {
			t.Errorf("%s: Unmarshal Error %s", v.casename, err)
		}
		if v.expected != p.Value {
			t.Errorf("%s: mismatch. given: %t. expected: %t", v.casename, p.Value, v.expected)
		}
	}
}

type MPNil struct {
	MPBase
	Value *bool `json:value`
}

func TestVerboseJSONNil(t *testing.T) {
	b := []byte{0xc0}
	buf := bytes.Buffer{}

	ret, err := msgpack.Decode(bytes.NewBuffer(b))
	outputVerboseJSON(ret, &buf, 0)

	p := MPNil{}
	err = json.Unmarshal(buf.Bytes(), &p)
	if err != nil {
		t.Errorf("Nil: Unmarshal Error %s", err)
	}
	if p.Value != nil {
		t.Errorf("Nil: Value is not nil")
	}

}
