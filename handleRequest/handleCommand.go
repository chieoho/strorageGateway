package handler

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"hash"
	"log"
	"os"
	"path/filepath"

	"../acknowledge"
	"../arguments"
	"../command"
	"../protocol"
)

type UploadHandler struct {
	WriteFds []*os.File
	md5Ctx   hash.Hash
}

func (u *UploadHandler) Handle(packet *protocol.Packet) bool {
	res := u.handleStartUpload(packet)
	if !res {
		return false
	}
	if !u.handleUploadBlock(packet) {
		return false
	}
	return true
}

func (u *UploadHandler) handleStartUpload(packet *protocol.Packet) bool {
	if err := packet.UnmarshalTaskInfo(); err != nil {
		log.Println("failed to Read:", err)
		packet.SendAck(command.UploadReqRet, ackCode.NotFound)
		return false
	}
	nameBytes := packet.Data.FileName
	fileRelativePath := string(nameBytes[:bytes.Index(nameBytes[:], []byte{0})])
	for _, backendPath := range arguments.BackendPathArray {
		filePath := filepath.Join(backendPath, fileRelativePath)
		log.Printf("%s", filePath)
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			log.Printf("create dir err: %s", err)
			packet.SendAck(command.UploadReqRet, ackCode.NotFound)
			return false
		}
		fd, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			packet.SendAck(command.UploadReqRet, ackCode.NotFound)
			return false
		}
		u.WriteFds = append(u.WriteFds, fd)
	}
	if !packet.SendAck(command.UploadReqRet, ackCode.OK) {
		return false
	}
	return true
}

func (u *UploadHandler) handleUploadBlock(packet *protocol.Packet) bool {
	u.md5Ctx = md5.New()
	for {
		res := packet.RecvData()
		if !res {
			packet.SendAck(command.UploadBlockRet, ackCode.NotFound)
			return false
		}
		if packet.Header.Command == command.UploadBlock {
			for _, fd := range u.WriteFds {
				if _, err := fd.Write(packet.DataBytes); err != nil {
					packet.SendAck(command.UploadBlockRet, ackCode.NotFound)
					return false
				}
				if !packet.SendAck(command.UploadBlockRet, ackCode.OK) {
					return false
				}
			}
			u.md5Ctx.Write(packet.DataBytes)
		} else {
			return u.handleUploadBlockEnd(packet)
		}
	}
}

func (u *UploadHandler) handleUploadBlockEnd(packet *protocol.Packet) bool {
	for _, fd := range u.WriteFds {
		if err := fd.Close(); err != nil {
			packet.SendAck(command.UploadBlockEndRet, ackCode.NotFound)
			return false
		}
	}
	cipherStr := u.md5Ctx.Sum(nil)
	md5Str := hex.EncodeToString(cipherStr)
	if err := packet.UnmarshalTaskInfo(); err != nil {
		log.Println("failed to Read:", err)
		packet.SendAck(command.UploadBlockEndRet, ackCode.NotFound)
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
	if !packet.SendAck(command.UploadBlockEndRet, ackCode.OK) {
		return false
	}
	return true
}

type DownloadHandler struct {
	ReadFds []*os.File
}

func (d *DownloadHandler) Handle(packet *protocol.Packet) bool {
	res := d.handleStartDownload(packet)
	if !res {
		return false
	}
	if !d.handleDownloadBlock(packet) {
		return false
	}
	return true
}

func (d *DownloadHandler) handleStartDownload(packet *protocol.Packet) bool {
	if err := packet.UnmarshalTaskInfo(); err != nil {
		log.Println("failed to Read:", err)
		packet.SendAck(command.DownloadReqRet, ackCode.NotFound)
		return false
	}
	nameBytes := packet.Data.FileName
	fileRelativePath := string(nameBytes[:bytes.Index(nameBytes[:], []byte{0})])
	filePath := filepath.Join(arguments.BackendPathArray[0], fileRelativePath)
	log.Printf("%s", filePath)
	fd, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		log.Println(err)
		packet.SendAck(command.DownloadReqRet, ackCode.NotFound)
		return false
	}
	d.ReadFds = append(d.ReadFds, fd)
	fi, _ := fd.Stat()
	packet.Header.Total = uint64(fi.Size())
	if !packet.SendAck(command.DownloadReqRet, ackCode.OK) {
		return false
	}
	return true
}

func (d *DownloadHandler) handleDownloadBlock(packet *protocol.Packet) bool {
	for {
		res := packet.RecvData()
		if !res {
			packet.SendAck(command.DownloadBlockRet, ackCode.NotFound)
			return false
		}

		if packet.Header.Command == command.DownloadBlockEnd {
			return d.handleDownloadBlockEnd(packet)
		} else {
			buf := make([]byte, packet.Header.Count)
			n, err := d.ReadFds[0].ReadAt(buf, int64(packet.Header.Offset))
			if err != nil {
				log.Println(err)
				packet.SendAck(command.DownloadBlockRet, ackCode.NotFound)
				return false
			}
			if !packet.SendBlock(buf[:n]) {
				return false
			}
		}
	}
}

func (d *DownloadHandler) handleDownloadBlockEnd(packet *protocol.Packet) bool {
	if err := d.ReadFds[0].Close(); err != nil {
		packet.SendAck(command.DownloadBlockEndRet, ackCode.NotFound)
		return false
	}
	if err := packet.UnmarshalTaskInfo(); err != nil {
		log.Println("failed to Read:", err)
		packet.SendAck(command.DownloadBlockEndRet, ackCode.NotFound)
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
				packet.SendAck(command.DownloadBlockEndRet, ackCode.NotFound)
				return false
			}
			if string(md5Bytes[:32]) == string(buf[:32]) {
				break
			}
		}
	}
	if !packet.SendAck(command.DownloadBlockEndRet, ackCode.OK) {
		return false
	}
	return true
}
