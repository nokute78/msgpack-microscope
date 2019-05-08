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
	"fmt"
	"testing"
	"time"
)

func TestTimestampExt(t *testing.T) {
	Init()

	tm32 := &MPObject{FirstByte: 0xd6, ExtType: -1}
	if !tm32.setRegisteredExt([]byte{0x00, 0x00, 0x00, 0x01}) {
		t.Error("tm32.setRegisteredExt Failed")
	}
	timestr := fmt.Sprintf("%v", time.Unix(1, 0))
	if tm32.DataStr != timestr {
		t.Errorf("tm32 error. \"%s\" is not %s", tm32.DataStr, timestr)
	}

	tm64 := &MPObject{FirstByte: 0xd7, ExtType: -1}
	if !tm64.setRegisteredExt([]byte{0x00, 0x00, 0x01, 0x90, 0x00, 0x00, 0x00, 0x01}) {
		t.Error("tm64.setRegisteredExt Failed")
	}
	timestr = fmt.Sprintf("%v", time.Unix(1, 100))
	if tm64.DataStr != timestr {
		t.Errorf("tm64 error. \"%s\" is not %s", tm64.DataStr, timestr)
	}

	tm96 := &MPObject{FirstByte: 0xc7, ExtType: -1}
	if !tm96.setRegisteredExt([]byte{0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}) {
		t.Error("tm96.setRegisteredExt Failed")
	}
	timestr = fmt.Sprintf("%v", time.Unix(1, 100))
	if tm96.DataStr != timestr {
		t.Errorf("tm96 error. \"%s\" is not %s", tm96.DataStr, timestr)
	}
}

func TestFluentdEventTime(t *testing.T) {
	Init()
	RegisterFluentdEventTime()

	fixext8 := &MPObject{FirstByte: 0xd7, ExtType: 0}
	if !fixext8.setRegisteredExt([]byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01}) {
		t.Errorf("fixext8.setRegisteredExt Failed")
	}
	timestr := fmt.Sprintf("%v", time.Unix(1, 1))
	if fixext8.DataStr != timestr {
		t.Errorf("fixext8 error. \"%s\" is not %s", fixext8.DataStr, timestr)
	}

	ext8 := &MPObject{FirstByte: 0xc7, ExtType: 0}
	if !ext8.setRegisteredExt([]byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01}) {
		t.Errorf("ext8.setRegisteredExt Failed")
	}
	timestr = fmt.Sprintf("%v", time.Unix(1, 1))
	if ext8.DataStr != timestr {
		t.Errorf("ext8 error. \"%s\" is not %s", ext8.DataStr, timestr)
	}
}
