package main

import (
	"io"
	"log"
	"net"
	"os"

	"./arguments"
	"./handleRequest"
	"./protocol"
)

// init is called prior to main.
func init() {
	logFile, err := os.OpenFile("sgw.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	arguments.GetArgs()
}

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", arguments.ServerAddr)
	handler.CheckTcpError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	handler.CheckTcpError(err)
	var heartBeat protocol.HeartBeat
	go heartBeat.SendHB()
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handler.HandleRequest(conn)
	}
}
