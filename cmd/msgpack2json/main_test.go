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
		{"fixmap", []byte{0x82, 0xa1, 0x41, 0x00, 0xa1, 0x42, 0x01}, `{"A":0, "B":1}`},
		{"map size 0", []byte{0x80}, "{}"},
		{"nested map", []byte{0x82, 0xa1, 0x30, 0xa1, 0x30, 0xa1, 0x31, 0x83, 0xa1, 0x32, 0xa1, 0x32, 0xa1, 0x33, 0xa1, 0x33, 0xa1, 0x34, 0xa1, 0x34}, `{"0":"0", "1":{"2":"2", "3":"3", "4":"4"}}`},
	}

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
