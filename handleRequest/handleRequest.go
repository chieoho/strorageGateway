package handler

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"

	"../command"
	"../protocol"
)

const (
	packetLenBytesNum = 4
	msgHeaderLen      = 64
)

func CheckError(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func handleCommand(msgHeader protocol.MsgHeader, taskInfo protocol.TaskInfo) bool {
	var result bool

	switch msgHeader.Command {
	case command.UploadReq:
		result = true
	default:
		fmt.Println("unknown command")
		result = false
	}
	return result
}

func HandleRequest(conn net.Conn) {
	defer conn.Close() // close connection before exit
	//_ = conn.SetReadDeadline(time.Now().Add(2 * time.Minute)) // set 2 minutes timeout
	packetLenBytes := make([]byte, packetLenBytesNum)
	for {
		readLen, err := io.ReadFull(conn, packetLenBytes)
		if err != nil {
			fmt.Println("read packet length bytes err", err)
			break
		}
		if readLen == 0 {
			fmt.Println("connection already closed by client")
			break
		}

		msgLength := binary.BigEndian.Uint32(packetLenBytes)
		msgBytes := make([]byte, msgLength)
		readLen, err = io.ReadFull(conn, msgBytes[packetLenBytesNum:])
		if err != nil {
			fmt.Println("read msg body bytes err", err)
			break
		}
		if readLen == 0 {
			fmt.Println("connection already closed by client")
			break
		}

		var msgHeader protocol.MsgHeader
		err, msgHeader = protocol.UnmarshalHeader(msgBytes[:msgHeaderLen], msgHeader)
		if err != nil {
			fmt.Println("failed to Read:", err)
			break
		}
		msgHeader.MsgLength = msgLength
		fmt.Printf("%v\n", msgHeader)

		var taskInfo = protocol.TaskInfo{}
		if msgLength > msgHeaderLen {
			err, taskInfo = protocol.UnmarshalTaskInfo(msgBytes[msgHeaderLen:], taskInfo)
			if err != nil {
				fmt.Println("failed to Read:", err)
				break
			}
			fmt.Printf("%v\n", taskInfo)
		}
		if !handleCommand(msgHeader, taskInfo) {
			fmt.Println("handle command failed")
			break
		}
	}
}
