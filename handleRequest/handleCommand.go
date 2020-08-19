package handler

import (
	"bytes"
	"log"
	"net"
	"os"
	"path/filepath"

	"../acknowledge"
	"../command"
	"../protocol"
)

func handleUpload(msgHeader *protocol.MsgHeader, taskInfoBytes []byte, conn net.Conn) bool {
	dataFile, res := handleStartUpload(msgHeader, taskInfoBytes, conn)
	if !res {
		return false
	}

	if !handleUploadBlock(msgHeader, dataFile, conn) {
		return false
	}

	return true
}

func handleDownload(msgHeader *protocol.MsgHeader, taskInfoBytes []byte, conn net.Conn) bool {
	dataFile, res := handleStartDownload(msgHeader, taskInfoBytes, conn)
	if !res {
		return false
	}
	if !handleDownloadBlock(msgHeader, dataFile, conn) {
		return false
	}

	return true
}

func handleStartUpload(msgHeader *protocol.MsgHeader, taskInfoBytes []byte, conn net.Conn) (*os.File, bool) {
	var taskInfo protocol.TaskInfo
	var dataFile *os.File

	if err := protocol.UnmarshalTaskInfo(taskInfoBytes, &taskInfo); err != nil {
		log.Println("failed to Read:", err)
		sendAck(msgHeader, ack.NotFound, conn)
		return dataFile, false
	}
	nameBytes := taskInfo.FileName
	filePath := string(nameBytes[:bytes.Index(nameBytes[:], []byte{0})])
	filePath = "/sgw11/" + filePath
	log.Printf("%s", filePath)
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		log.Printf("create dir err: %s", err)
		sendAck(msgHeader, ack.NotFound, conn)
		return dataFile, false
	}
	dataFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		sendAck(msgHeader, ack.NotFound, conn)
		return dataFile, false
	}
	if !sendAck(msgHeader, ack.OK, conn) {
		return dataFile, false
	}
	return dataFile, true
}

func handleStartDownload(msgHeader *protocol.MsgHeader, taskInfoBytes []byte, conn net.Conn) (*os.File, bool) {
	var taskInfo protocol.TaskInfo
	var dataFile *os.File

	if err := protocol.UnmarshalTaskInfo(taskInfoBytes, &taskInfo); err != nil {
		log.Println("failed to Read:", err)
		sendAck(msgHeader, ack.NotFound, conn)
		return dataFile, false
	}
	nameBytes := taskInfo.FileName
	filePath := string(nameBytes[:bytes.Index(nameBytes[:], []byte{0})])
	filePath = "/sgw11" + filePath
	log.Printf("%s", filePath)
	dataFile, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		log.Println(err)
		sendAck(msgHeader, ack.NotFound, conn)
		return dataFile, false
	}
	fi, _ := dataFile.Stat()
	msgHeader.Total = uint64(fi.Size())
	if !sendAck(msgHeader, ack.OK, conn) {
		return dataFile, false
	}
	return dataFile, true
}

func handleUploadBlock(msgHeader *protocol.MsgHeader, dataFile *os.File, conn net.Conn) bool {
	for {
		packetBytes, res := recvData(conn)
		if !res {
			sendAck(msgHeader, ack.NotFound, conn)
			return false
		}
		if err := protocol.UnmarshalHeader(packetBytes[:protocol.MsgHeaderLen], msgHeader); err != nil {
			log.Println("failed to unmarshal header:", err)
			sendAck(msgHeader, ack.NotFound, conn)
			return false
		}
		if msgHeader.Command == command.UploadBlock {
			if _, err := dataFile.Write(packetBytes[protocol.MsgHeaderLen:]); err != nil {
				sendAck(msgHeader, ack.NotFound, conn)
				return false
			}
			if !sendAck(msgHeader, ack.OK, conn) {
				return false
			}
		} else {
			if err := dataFile.Close(); err != nil {
				sendAck(msgHeader, ack.NotFound, conn)
				return false
			}
			if !sendAck(msgHeader, ack.OK, conn) {
				return false
			}
			break
		}
	}
	return true
}

func handleDownloadBlock(msgHeader *protocol.MsgHeader, dataFile *os.File, conn net.Conn) bool {
	for {
		packetBytes, res := recvData(conn)
		if !res {
			sendAck(msgHeader, ack.NotFound, conn)
			return false
		}
		if err := protocol.UnmarshalHeader(packetBytes[:protocol.MsgHeaderLen], msgHeader); err != nil {
			log.Println("failed to unmarshal header:", err)
			sendAck(msgHeader, ack.NotFound, conn)
			return false
		}
		if msgHeader.Command == command.DownloadBlockEnd {
			if err := dataFile.Close(); err != nil {
				sendAck(msgHeader, ack.NotFound, conn)
				return false
			}
			if !sendAck(msgHeader, ack.OK, conn) {
				return false
			}
			break
		} else {
			buf := make([]byte, msgHeader.Count)
			n, err := dataFile.ReadAt(buf, int64(msgHeader.Offset))
			if err != nil {
				log.Println(err)
				sendAck(msgHeader, ack.NotFound, conn)
				return false
			}
			if !sendBlock(msgHeader, buf[:n], conn) {
				return false
			}
		}
	}
	return true
}

func sendAck(msgHeader *protocol.MsgHeader, ack uint32, conn net.Conn) bool {
	msgHeader.MsgLength = protocol.MsgHeaderLen
	msgHeader.AckCode = ack
	_, msgHeaderBytes := protocol.MarshalHeader(msgHeader)
	if _, err := conn.Write(msgHeaderBytes); err != nil {
		return false
	}
	return true
}

func sendBlock(msgHeader *protocol.MsgHeader, content []byte, conn net.Conn) bool {
	msgHeader.MsgLength = protocol.MsgHeaderLen + uint32(len(content))
	msgHeader.Command = command.DownloadBlockRet
	msgHeader.AckCode = ack.OK
	_, msgHeaderBytes := protocol.MarshalHeader(msgHeader)
	if _, err := conn.Write(append(msgHeaderBytes, content...)); err != nil {
		return false
	}
	return true
}
