package handler

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"log"
	"net"
	"os"
	"path/filepath"

	"../acknowledge"
	"../arguments"
	"../command"
	"../protocol"
)

func handleUpload(packet *protocol.Packet, conn net.Conn) bool {
	dataFiles, res := handleStartUpload(packet, conn)
	if !res {
		return false
	}
	if !handleUploadBlock(packet, dataFiles, conn) {
		return false
	}
	return true
}

func handleStartUpload(packet *protocol.Packet, conn net.Conn) ([]*os.File, bool) {
	var dataFiles []*os.File

	if err := packet.UnmarshalTaskInfo(); err != nil {
		log.Println("failed to Read:", err)
		packet.SendAck(ack.NotFound, conn)
		return dataFiles, false
	}
	nameBytes := packet.Data.FileName
	fileRelativePath := string(nameBytes[:bytes.Index(nameBytes[:], []byte{0})])
	for _, backendPath := range arguments.BackendPathArray {
		filePath := filepath.Join(backendPath, fileRelativePath)
		log.Printf("%s", filePath)
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			log.Printf("create dir err: %s", err)
			packet.SendAck(ack.NotFound, conn)
			return dataFiles, false
		}
		dataFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			packet.SendAck(ack.NotFound, conn)
			return dataFiles, false
		}
		dataFiles = append(dataFiles, dataFile)
	}
	if !packet.SendAck(ack.OK, conn) {
		return dataFiles, false
	}
	return dataFiles, true
}

func handleUploadBlock(packet *protocol.Packet, dataFiles []*os.File, conn net.Conn) bool {
	md5Ctx := md5.New()
	for {
		res := packet.RecvData(conn)
		if !res {
			packet.SendAck(ack.NotFound, conn)
			return false
		}
		if packet.Header.Command == command.UploadBlock {
			for _, dataFile := range dataFiles {
				if _, err := dataFile.Write(packet.DataBytes); err != nil {
					packet.SendAck(ack.NotFound, conn)
					return false
				}
				if !packet.SendAck(ack.OK, conn) {
					return false
				}
			}
			md5Ctx.Write(packet.DataBytes)
		} else {
			for _, dataFile := range dataFiles {
				if err := dataFile.Close(); err != nil {
					packet.SendAck(ack.NotFound, conn)
					return false
				}
			}
			cipherStr := md5Ctx.Sum(nil)
			md5Str := hex.EncodeToString(cipherStr)
			if err := packet.UnmarshalTaskInfo(); err != nil {
				log.Println("failed to Read:", err)
				packet.SendAck(ack.NotFound, conn)
				return false
			}
			md5Bytes := packet.Data.FileMd5
			if md5Str != string(md5Bytes[:32]) {
				return false
			}
			nameBytes := packet.Data.FileName
			fileRelativePath := string(nameBytes[:bytes.Index(nameBytes[:], []byte{0})])
			for _, backendPath := range arguments.BackendPathArray {
				filePath := filepath.Join(backendPath, fileRelativePath)
				hashPath := filepath.Dir(filePath) + "/.hash"
				hashFile, _ := os.OpenFile(hashPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
				_, _ = hashFile.Write([]byte(md5Str + "\n"))
			}

			if !packet.SendAck(ack.OK, conn) {
				return false
			}
			break
		}
	}
	return true
}

func handleDownload(packet *protocol.Packet, conn net.Conn) bool {
	dataFile, res := handleStartDownload(packet, conn)
	if !res {
		return false
	}
	if !handleDownloadBlock(packet, dataFile, conn) {
		return false
	}
	return true
}

func handleStartDownload(packet *protocol.Packet, conn net.Conn) (*os.File, bool) {
	var dataFile *os.File

	if err := packet.UnmarshalTaskInfo(); err != nil {
		log.Println("failed to Read:", err)
		packet.SendAck(ack.NotFound, conn)
		return dataFile, false
	}
	nameBytes := packet.Data.FileName
	fileRelativePath := string(nameBytes[:bytes.Index(nameBytes[:], []byte{0})])
	filePath := filepath.Join(arguments.BackendPathArray[0], fileRelativePath)
	log.Printf("%s", filePath)
	dataFile, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		log.Println(err)
		packet.SendAck(ack.NotFound, conn)
		return dataFile, false
	}
	fi, _ := dataFile.Stat()
	packet.Header.Total = uint64(fi.Size())
	if !packet.SendAck(ack.OK, conn) {
		return dataFile, false
	}
	return dataFile, true
}

func handleDownloadBlock(packet *protocol.Packet, dataFile *os.File, conn net.Conn) bool {
	for {
		res := packet.RecvData(conn)
		if !res {
			packet.SendAck(ack.NotFound, conn)
			return false
		}
		if packet.Header.Command == command.DownloadBlockEnd {
			if err := dataFile.Close(); err != nil {
				packet.SendAck(ack.NotFound, conn)
				return false
			}
			if err := packet.UnmarshalTaskInfo(); err != nil {
				log.Println("failed to Read:", err)
				packet.SendAck(ack.NotFound, conn)
				return false
			}
			md5Bytes := packet.Data.FileMd5
			if md5Bytes[0] != 0 {
				nameBytes := packet.Data.FileName
				fileRelativePath := string(nameBytes[:bytes.Index(nameBytes[:], []byte{0})])
				filePath := filepath.Join(arguments.BackendPathArray[0], fileRelativePath)
				hashPath := filepath.Dir(filePath) + "/.hash"
				hashFile, _ := os.OpenFile(hashPath, os.O_RDONLY, 0644)
				buf := make([]byte, 33)
				for {
					_, err := hashFile.Read(buf)
					if err != nil {
						log.Println(err)
						packet.SendAck(ack.NotFound, conn)
						return false
					}
					if string(md5Bytes[:32]) == string(buf[:32]) {
						break
					}
				}
			}
			if !packet.SendAck(ack.OK, conn) {
				return false
			}
			break
		} else {
			buf := make([]byte, packet.Header.Count)
			n, err := dataFile.ReadAt(buf, int64(packet.Header.Offset))
			if err != nil {
				log.Println(err)
				packet.SendAck(ack.NotFound, conn)
				return false
			}
			if !packet.SendBlock(buf[:n], conn) {
				return false
			}
		}
	}
	return true
}
