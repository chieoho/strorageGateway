package handler

import (
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
	var packet protocol.Packet
	for {
		res := packet.RecvData(conn)
		if !res {
			break
		}
		if !handleCommand(packet, conn) {
			log.Println("handle command failed")
			break
		}
	}
}

func handleCommand(packet protocol.Packet, conn net.Conn) bool {
	switch packet.Header.Command {
	case command.UploadReq:
		return handleUpload(&packet, conn)
	case command.DownloadReq:
		return handleDownload(&packet, conn)
	default:
		log.Println("unknown command")
		return false
	}
}

func CheckTcpError(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
