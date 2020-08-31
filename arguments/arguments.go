package arguments

import (
	"flag"
	"log"
	"strings"
)

var ServerAddr string
var BackendPathArray []string
var RegionId int
var AsmAddr string

func GetArgs() {
	serverAddr := flag.String("l", "0.0.0.0:7788", "监听地址，如：0.0.0.0:7788")
	BackendPath := flag.String("b", "/sgw11", "后端存储目录，如：/sgw11,/sgw12")
	regionId := flag.Int("r", 2018, "region id，如：2020")
	asmAddr := flag.String("a", "127.0.0.1:12345", "ASM服务地址，如：127.0.0.1:12345")
	flag.Parse()
	ServerAddr = *serverAddr
	BackendPathArray = strings.Split(*BackendPath, ",")
	RegionId = *regionId
	AsmAddr = *asmAddr
	log.Printf("serve on %s", ServerAddr)
	log.Printf("backend path list: %v", BackendPathArray)
	log.Printf("region id: %d", RegionId)
	log.Printf("asm addr: %s", AsmAddr)
}
