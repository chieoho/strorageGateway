package handler

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"

	"../command"
	"../protocol"
)

func HandleRequest(conn net.Conn) {
	defer func() {
		err := conn.Close() // close connection before exit
		log.Println(err)
	}()
	//_ = conn.SetReadDeadline(time.Now().Add(2 * time.Minute)) // set 2 minutes timeout
	for {
		packetBytes, res := recvData(conn)
		if !res {
			break
		}
		if !handleCommand(packetBytes, conn) {
			log.Println("handle command failed")
			break
		}
	}
}

func handleCommand(packetBytes []byte, conn net.Conn) bool {
	var msgHeader protocol.MsgHeader
	err := protocol.UnmarshalHeader(packetBytes[:protocol.MsgHeaderLen], &msgHeader)
	if err != nil {
		log.Println("failed to unmarshal header:", err)
		return false
	}
	msgHeader.MsgLength = uint32(len(packetBytes))
	log.Printf("%v\n", msgHeader)

	switch msgHeader.Command {

	case command.UploadReq:
		return handleUpload(&msgHeader, packetBytes[protocol.MsgHeaderLen:], conn)

	case command.DownloadReq:
		return handleDownload(&msgHeader, packetBytes[protocol.MsgHeaderLen:], conn)

	default:
		log.Println("unknown command")
		return false
	}
}

func recvData(conn net.Conn) (packetBytes []byte, res bool) {
	packetLenBytes := make([]byte, protocol.PacketLenBytesNum)
	_, err := io.ReadFull(conn, packetLenBytes)
	if !checkIOReadErr(err, "read packet length bytes") {
		return packetBytes, false
	}
	msgLength := binary.BigEndian.Uint32(packetLenBytes)
	msgBytes := make([]byte, msgLength)
	_, err = io.ReadFull(conn, msgBytes[protocol.PacketLenBytesNum:])
	if !checkIOReadErr(err, "read packet body bytes") {
		return packetBytes, false
	}
	return msgBytes, true
}

func CheckTcpError(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func checkIOReadErr(err error, info string) bool {
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
