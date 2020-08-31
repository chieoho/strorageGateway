package command

const (
	UploadReq           = 0x20001
	UploadReqRet        = UploadReq + 1
	UploadBlock         = 0x20003
	UploadBlockRet      = UploadBlock + 1
	UploadBlockEnd      = 0x20005
	UploadBlockEndRet   = UploadBlockEnd + 1
	DownloadReq         = 0x20007
	DownloadReqRet      = DownloadReq + 1
	DownloadBlock       = 0x20009
	DownloadBlockRet    = DownloadBlock + 1
	DownloadBlockEnd    = 0x2000B
	DownloadBlockEndRet = DownloadBlockEnd + 1

	AamHb = 0x00080001
)
