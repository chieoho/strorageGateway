package protocol

import (
	"fmt"
	"testing"
)

func TestUnmarshalHeader(t *testing.T) {
	inBytes := make([]byte, MsgHeaderLen)
	var msgHeader MsgHeader

	inBytes[3] = MsgHeaderLen
	//fmt.Println(inBytes)
	err := UnmarshalHeader(inBytes, &msgHeader)
	if err != nil {
		fmt.Println("failed to Read:", err)
	}
	//fmt.Printf("%v\n", msgHeader)
	if msgHeader.MsgLength != MsgHeaderLen {
		t.Error("unmarshal header failed")
	}
}

func TestMarshalHeader(t *testing.T) {
	var msgHeader = MsgHeader{MsgLength: 264}
	var outBytes []byte

	//fmt.Printf("%v\n", msgHeader)
	err, outBytes := MarshalHeader(&msgHeader)
	if err != nil {
		fmt.Println("failed to Write:", err)
	}
	//fmt.Println(outBytes)
	if outBytes[2] != 1 || outBytes[3] != 8 {
		t.Error("marshal header failed")
	}
}

func TestUnmarshalTaskInfo(t *testing.T) {
	inBytes := make([]byte, TaskInfoLen)
	var taskInfo TaskInfo
	const regionId = 8

	inBytes[3] = regionId
	//fmt.Println(inBytes)
	err := UnmarshalTaskInfo(inBytes, &taskInfo)
	if err != nil {
		fmt.Println("failed to Read:", err)
	}
	//fmt.Printf("%v\n", taskInfo)
	if taskInfo.RegionId != regionId {
		t.Error("unmarshal header failed")
	}
}

func TestMarshalTaskInfo(t *testing.T) {
	var taskInfo = TaskInfo{RegionId: 8}
	var outBytes []byte

	//fmt.Printf("%v\n", msgHeader)
	err, outBytes := MarshalTaskInfo(&taskInfo)
	if err != nil {
		fmt.Println("failed to Write:", err)
	}
	//fmt.Println(outBytes)
	if outBytes[3] != 8 {
		t.Error("marshal header failed")
	}
}
