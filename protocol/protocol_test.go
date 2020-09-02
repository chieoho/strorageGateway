package protocol

import (
	"fmt"
	"testing"
)

func TestUnmarshalHeader(t *testing.T) {
	var packet Packet

	packet.Init()
	packet.Bytes[3] = MsgHeaderLen
	err := packet.unmarshal(packet.Bytes[:MsgHeaderLen], &packet.Header)
	if err != nil {
		fmt.Println("failed to Read:", err)
	}
	if packet.Header.MsgLength != MsgHeaderLen {
		t.Error("unmarshal header failed")
	}
}

func TestUnmarshalTaskInfo(t *testing.T) {
	var packet Packet
	const regionId = 8

	packet.Init()
	packet.Bytes[MsgHeaderLen+3] = regionId
	err := packet.UnmarshalTaskInfo()
	if err != nil {
		fmt.Println("failed to Read:", err)
	}
	if packet.Data.RegionId != regionId {
		t.Error("unmarshal data failed")
	}
}

func TestMarshalHeader(t *testing.T) {
	var packet Packet
	var outBytes []byte

	packet.Header.MsgLength = 264
	err, outBytes := packet.Marshal(&packet.Header)
	if err != nil {
		fmt.Println("failed to Write:", err)
	}
	if outBytes[2] != 1 || outBytes[3] != 8 {
		t.Error("marshal header failed")
	}
}

func TestMarshalTaskInfo(t *testing.T) {
	var taskInfo = TaskInfo{RegionId: 8}
	var packet Packet
	var outBytes []byte

	err, outBytes := packet.Marshal(&taskInfo)
	if err != nil {
		fmt.Println("failed to Write:", err)
	}
	if outBytes[3] != 8 {
		t.Error("marshal data failed")
	}
}
