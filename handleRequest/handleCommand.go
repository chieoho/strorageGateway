package handler

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"

	"../acknowledge"
	"../arguments"
	"../command"
	"../protocol"
)

func handleUpload(packet *protocol.Packet) bool {
	res := handleStartUpload(packet)
	if !res {
		return false
	}
	if !handleUploadBlock(packet) {
		return false
	}
	return true
}

func handleStartUpload(packet *protocol.Packet) bool {
	if err := packet.UnmarshalTaskInfo(); err != nil {
		log.Println("failed to Read:", err)
		packet.SendAck(ack.NotFound)
		return false
	}
	nameBytes := packet.Data.FileName
	fileRelativePath := string(nameBytes[:bytes.Index(nameBytes[:], []byte{0})])
	for _, backendPath := range arguments.BackendPathArray {
		filePath := filepath.Join(backendPath, fileRelativePath)
		log.Printf("%s", filePath)
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			log.Printf("create dir err: %s", err)
			packet.SendAck(ack.NotFound)
			return false
		}
		fd, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			packet.SendAck(ack.NotFound)
			return false
		}
		packet.Fds = append(packet.Fds, fd)
	}
	if !packet.SendAck(ack.OK) {
		return false
	}
	return true
}

func handleUploadBlock(packet *protocol.Packet) bool {
	md5Ctx := md5.New()
	for {
		res := packet.RecvData()
		if !res {
			packet.SendAck(ack.NotFound)
			return false
		}
		if packet.Header.Command == command.UploadBlock {
			for _, fd := range packet.Fds {
				if _, err := fd.Write(packet.DataBytes); err != nil {
					packet.SendAck(ack.NotFound)
					return false
				}
				if !packet.SendAck(ack.OK) {
					return false
				}
			}
			md5Ctx.Write(packet.DataBytes)
		} else {
			for _, fd := range packet.Fds {
				if err := fd.Close(); err != nil {
					packet.SendAck(ack.NotFound)
					return false
				}
			}
			cipherStr := md5Ctx.Sum(nil)
			md5Str := hex.EncodeToString(cipherStr)
			if err := packet.UnmarshalTaskInfo(); err != nil {
				log.Println("failed to Read:", err)
				packet.SendAck(ack.NotFound)
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
			if !packet.SendAck(ack.OK) {
				return false
			}
			break
		}
	}
	return true
}

func handleDownload(packet *protocol.Packet) bool {
	res := handleStartDownload(packet)
	if !res {
		return false
	}
	if !handleDownloadBlock(packet) {
		return false
	}
	return true
}

func handleStartDownload(packet *protocol.Packet) bool {
	if err := packet.UnmarshalTaskInfo(); err != nil {
		log.Println("failed to Read:", err)
		packet.SendAck(ack.NotFound)
		return false
	}
	nameBytes := packet.Data.FileName
	fileRelativePath := string(nameBytes[:bytes.Index(nameBytes[:], []byte{0})])
	filePath := filepath.Join(arguments.BackendPathArray[0], fileRelativePath)
	log.Printf("%s", filePath)
	fd, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		log.Println(err)
		packet.SendAck(ack.NotFound)
		return false
	}
	packet.Fds = append(packet.Fds, fd)
	fi, _ := fd.Stat()
	packet.Header.Total = uint64(fi.Size())
	if !packet.SendAck(ack.OK) {
		return false
	}
	return true
}

func handleDownloadBlock(packet *protocol.Packet) bool {
	for {
		res := packet.RecvData()
		if !res {
			packet.SendAck(ack.NotFound)
			return false
		}
		if packet.Header.Command == command.DownloadBlockEnd {
			if err := packet.Fds[0].Close(); err != nil {
				packet.SendAck(ack.NotFound)
				return false
			}
			if err := packet.UnmarshalTaskInfo(); err != nil {
				log.Println("failed to Read:", err)
				packet.SendAck(ack.NotFound)
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
						packet.SendAck(ack.NotFound)
						return false
					}
					if string(md5Bytes[:32]) == string(buf[:32]) {
						break
					}
				}
			}
			if !packet.SendAck(ack.OK) {
				return false
			}
			break
		} else {
			buf := make([]byte, packet.Header.Count)
			n, err := packet.Fds[0].ReadAt(buf, int64(packet.Header.Offset))
			if err != nil {
				log.Println(err)
				packet.SendAck(ack.NotFound)
				return false
			}
			if !packet.SendBlock(buf[:n]) {
				return false
			}
		}
	}
	return true
}
