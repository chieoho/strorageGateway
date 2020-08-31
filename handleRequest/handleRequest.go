package handler

import (
	"log"
	"net"
	"os"

	"../command"
	"../protocol"
)

type CmdHandler interface {
	Handle(packet *protocol.Packet) bool
}

func HandleRequest(conn net.Conn) {
	defer func() {
		err := conn.Close() // close connection before exit
		if err != nil {
			log.Println(err)
		}
	}()
	//_ = conn.SetReadDeadline(time.Now().Add(2 * time.Minute)) // set 2 minutes timeout
	var packet = protocol.Packet{Conn: conn}
	for {
		res := packet.RecvData()
		if !res {
			break
		}
		if !handleCommand(packet) {
			log.Println("handle command failed")
			break
		}
	}
}

func handleCommand(packet protocol.Packet) bool {
	var cmdHandler CmdHandler

	switch packet.Header.Command {
	case command.UploadReq:
		cmdHandler = &UploadHandler{}
	case command.DownloadReq:
		cmdHandler = &DownloadHandler{}
	default:
		log.Println("unknown command")
		return false
	}
	return cmdHandler.Handle(&packet)
}

func CheckTcpError(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
