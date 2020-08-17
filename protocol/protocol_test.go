package protocol

import (
	"fmt"
	"testing"
)

func TestUnmarshalHeader(t *testing.T) {
	inBytes := make([]byte, 64)
	var msgHeader MsgHeader

	inBytes[3] = 64
	//fmt.Println(inBytes)
	err, msgHeader := UnmarshalHeader(inBytes, msgHeader)
	if err != nil {
		fmt.Println("failed to Read:", err)
	}
	//fmt.Printf("%v\n", msgHeader)
	if msgHeader.MsgLength != 64 {
		t.Error("unmarshal header failed")
	}
}

func TestMarshalHeader(t *testing.T) {
	var msgHeader = MsgHeader{MsgLength: 264}
	var outBytes []byte

	//fmt.Printf("%v\n", msgHeader)
	err, outBytes := MarshalHeader(msgHeader)
	if err != nil {
		fmt.Println("failed to Write:", err)
	}
	//fmt.Println(outBytes)
	if outBytes[2] != 1 || outBytes[3] != 8 {
		t.Error("marshal header failed")
	}
}
