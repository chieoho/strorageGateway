package main

import (
	"net"
	"os"

	"./handleRequest"
)

func main() {
	var service string
	if len(os.Args) >= 2 {
		port := os.Args[1]
		service = "0.0.0.0:" + port
	} else {
		service = "0.0.0.0:9200"
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	handler.CheckError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	handler.CheckError(err)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handler.HandleRequest(conn)
	}
}
