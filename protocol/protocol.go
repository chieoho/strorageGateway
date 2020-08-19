package protocol

import (
	"bytes"
	"encoding/binary"
)

const (
	Md5Len     = 32
	MaxNameLen = 255

	PacketLenBytesNum = 4
	MsgHeaderLen      = 64
	TaskInfoLen       = 341
)

type MsgHeader struct {
	MsgLength uint32   // 总长度
	Major     uint8    // 协议主版本号
	Minor     uint8    // 协议次版本号
	SrcType   uint8    // 源节点类型
	DstType   uint8    // 目的节点类型
	SrcId     uint32   // 源节点 ID
	DstId     uint32   // 目的节点 ID
	TransId   uint64   // 任务 ID
	Sequence  uint64   // 消息序号
	Command   uint32   // 命令字
	AckCode   uint32   // 响应码
	Total     uint64   // 数据总量
	Offset    uint64   // 偏移量
	Count     uint32   // 数据量
	Padding   [4]uint8 // 填充到 64B 对齐
	Data      [0]uint8
}

type TaskInfo struct {
	Operation uint16
	RegionId  uint16
	SiteId    uint32
	AppId     uint32
	Timestamp uint32

	SgwPort   uint16   // 落地服务的 port
	ProxyPort uint16   // 中转服务的 port
	SgwIp     uint32   // 落地服务的 IP
	ProxyIp   uint32   // 中转服务的 IP
	SgwId     uint32   // 落地服务的 ID
	ProxyId   uint32   // 中转服务的 ID
	Padding   [4]uint8 // 填充到 64B 对齐

	FileLen     uint64
	FileMd5     [Md5Len + 1]byte
	FileName    [MaxNameLen + 1]byte
	MetadataLen uint32
	Metadata    [0]byte
}

func UnmarshalHeader(inBytes []byte, msgHeader *MsgHeader) error {
	inBytesBuf := bytes.NewBuffer(inBytes)
	err := binary.Read(inBytesBuf, binary.BigEndian, msgHeader)
	return err
}

func MarshalHeader(msgHeader *MsgHeader) (error, []byte) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, msgHeader)
	return err, buf.Bytes()
}

func UnmarshalTaskInfo(inBytes []byte, taskInfo *TaskInfo) error {
	inBytesBuf := bytes.NewBuffer(inBytes)
	err := binary.Read(inBytesBuf, binary.BigEndian, taskInfo)
	return err
}

func MarshalTaskInfo(taskInfo *TaskInfo) (error, []byte) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, taskInfo)
	return err, buf.Bytes()
}
