package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"time"

	"../acknowledge"
	"../arguments"
	"../command"
)

const (
	Md5Len     = 32
	MaxNameLen = 255

	PacketLenBytesNum = 4
	MsgHeaderLen      = 64
	TaskInfoLen       = 341

	HeartBeatHeaderLen = 8
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
	Conn      net.Conn
}

func (p *Packet) RecvData() bool {
	packetLenBytes := make([]byte, PacketLenBytesNum)
	_, err := io.ReadFull(p.Conn, packetLenBytes)
	if !p.checkIOReadErr(err, "read packet length bytes") {
		return false
	}
	msgLength := binary.BigEndian.Uint32(packetLenBytes)
	msgBytes := make([]byte, msgLength)
	_, err = io.ReadFull(p.Conn, msgBytes[PacketLenBytesNum:])
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

func (p *Packet) SendAck(cmd uint32, ack uint32) bool {
	p.Header.Command = cmd
	p.Header.AckCode = ack
	return p.sendData(nil)
}

func (p *Packet) SendBlock(dataBytes []byte) bool {
	p.Header.Command = command.DownloadBlockRet
	p.Header.AckCode = ackCode.OK
	return p.sendData(dataBytes)
}

func (p *Packet) sendData(dataBytes []byte) bool {
	p.Header.MsgLength = MsgHeaderLen + uint32(len(dataBytes))
	_, msgHeaderBytes := p.Marshal(&p.Header)
	if _, err := p.Conn.Write(append(msgHeaderBytes, dataBytes...)); err != nil {
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

type HeartBeat struct {
	MsgLength uint32
	Command   uint32
}

func (h *HeartBeat) SendHB() {
	h.Command = command.AamHb
	jsonBytes, _ := json.Marshal(map[string]int{"region_id": arguments.RegionId})
	h.MsgLength = uint32(HeartBeatHeaderLen + len(jsonBytes))
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, h)
	heartBeatBytes := append(buf.Bytes(), jsonBytes...)
	conn, err := net.Dial("tcp", arguments.AsmAddr)
	if err != nil {
		log.Println(err)
	}
	for {
		_, err = conn.Write(heartBeatBytes)
		if err != nil {
			log.Println(err)
			conn, _ = net.Dial("tcp", arguments.AsmAddr)
		}
		time.Sleep(3 * time.Second)
	}
}
