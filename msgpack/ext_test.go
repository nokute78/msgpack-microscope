package msgpack

import (
	"fmt"
	"testing"
	"time"
)

func TestTimestampExt(t *testing.T) {
	typeNameInit()
	extFormatInit()

	tm32 := &MPObject{FirstByte: 0xd6, ExtType: -1}
	if !tm32.SetRegisteredExt([]byte{0x00, 0x00, 0x00, 0x01}) {
		t.Error("tm32.SetRegisteredExt Failed")
	}
	timestr := fmt.Sprintf("%v", time.Unix(1, 0))
	if tm32.DataStr != timestr {
		t.Errorf("tm32 error. \"%s\" is not %s", tm32.DataStr, timestr)
	}

	tm64 := &MPObject{FirstByte: 0xd7, ExtType: -1}
	if !tm64.SetRegisteredExt([]byte{0x00, 0x00, 0x01, 0x90, 0x00, 0x00, 0x00, 0x01}) {
		t.Error("tm64.SetRegisteredExt Failed")
	}
	timestr = fmt.Sprintf("%v", time.Unix(1, 100))
	if tm64.DataStr != timestr {
		t.Errorf("tm64 error. \"%s\" is not %s", tm64.DataStr, timestr)
	}

	tm96 := &MPObject{FirstByte: 0xc7, ExtType: -1}
	if !tm96.SetRegisteredExt([]byte{0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}) {
		t.Error("tm96.SetRegisteredExt Failed")
	}
	timestr = fmt.Sprintf("%v", time.Unix(1, 100))
	if tm96.DataStr != timestr {
		t.Errorf("tm96 error. \"%s\" is not %s", tm96.DataStr, timestr)
	}

}
