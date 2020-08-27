package protocol

import (
	"../command"
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
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

type Packet struct {
	Header    MsgHeader
	Data      TaskInfo
	DataBytes []byte
}

func (p *Packet) RecvData(conn net.Conn) bool {
	packetLenBytes := make([]byte, PacketLenBytesNum)
	_, err := io.ReadFull(conn, packetLenBytes)
	if !p.checkIOReadErr(err, "read packet length bytes") {
		return false
	}
	msgLength := binary.BigEndian.Uint32(packetLenBytes)
	msgBytes := make([]byte, msgLength)
	_, err = io.ReadFull(conn, msgBytes[PacketLenBytesNum:])
	if !p.checkIOReadErr(err, "read packet body bytes") {
		return false
	}
	_ = p.unmarshal(msgBytes[:MsgHeaderLen], &p.Header)
	p.DataBytes = msgBytes[MsgHeaderLen:]
	return true
}

func (p *Packet) UnmarshalTaskInfo() error {
	err := p.unmarshal(p.DataBytes, &p.Data)
	return err
}

func (p *Packet) unmarshal(inBytes []byte, packet interface{}) error {
	inBytesBuf := bytes.NewBuffer(inBytes)
	err := binary.Read(inBytesBuf, binary.BigEndian, packet)
	return err
}

func (p *Packet) Marshal(packet interface{}) (error, []byte) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, packet)
	return err, buf.Bytes()
}

func (p *Packet) SendAck(ack uint32, conn net.Conn) bool {
	p.Header.AckCode = ack
	return p.sendData(nil, conn)
}

func (p *Packet) SendBlock(dataBytes []byte, conn net.Conn) bool {
	p.Header.Command = command.DownloadBlockRet
	return p.sendData(dataBytes, conn)
}

func (p *Packet) sendData(dataBytes []byte, conn net.Conn) bool {
	p.Header.MsgLength = MsgHeaderLen + uint32(len(dataBytes))
	_, msgHeaderBytes := p.Marshal(&p.Header)
	if _, err := conn.Write(append(msgHeaderBytes, dataBytes...)); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (p *Packet) checkIOReadErr(err error, info string) bool {
	if err != nil {
		if err == io.EOF {
			log.Printf("connection closed when %s\n", info)
			return false
		}
		log.Printf("%s err: %s\n", info, err)
		return false
	}
	return true
}
