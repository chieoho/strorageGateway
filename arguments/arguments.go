package arguments

import (
	"flag"
	"strings"
)

var ServerPort string
var BackendPathArray []string

func GetArgs() {
	portPtr := flag.String("p", "7788", "端口号，如：7788")
	BackendPath := flag.String("b", "/sgw11/", "后端存储目录，如：/sgw11/,/sgw12/")
	flag.Parse()
	ServerPort = *portPtr
	BackendPathArray = strings.Split(*BackendPath, ",")
}
