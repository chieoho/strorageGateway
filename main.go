package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"strings"

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
	portPtr := flag.String("p", "7788", "端口号，如：7788")
	BackendPath := flag.String("b", "/sgw11/", "后端存储目录，如：/sgw11/,/sgw12/")
	flag.Parse()

	service := "0.0.0.0:" + *portPtr
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	handler.CheckTcpError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	handler.CheckTcpError(err)
	log.Printf("serve on %s", service)
	log.Printf("backend path list: %v", strings.Split(*BackendPath, ","))
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handler.HandleRequest(conn)
	}
}
