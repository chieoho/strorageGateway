package main

import (
	"io"
	"log"
	"net"
	"os"

	"./handleRequest"
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
}

func main() {
	var service string
	if len(os.Args) >= 2 {
		port := os.Args[1]
		service = "0.0.0.0:" + port
	} else {
		service = "0.0.0.0:9200"
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	handler.CheckTcpError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	handler.CheckTcpError(err)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handler.HandleRequest(conn)
	}
}
