package command

const (
	UploadReq             = 0x20001
	UploadReqRet          = UploadReq + 1
	UploadBlock           = 0x20003
	UploadBlockRet        = UploadBlock + 1
	UploadBlockEnd        = 0x20005
	UploadBlockEndRet     = UploadBlockEnd + 1
	DownloadReq           = 0x20007
	DownloadReqRet        = DownloadReq + 1
	DownloadBlock         = 0x20009
	DownloadBlockRet      = DownloadBlock + 1
	DownloadBlockEnd      = 0x2000B
	DownloadBlockEndRet   = DownloadBlockEnd + 1
	DeleteReq             = 0x2000D
	DeleteReqRet          = DeleteReq + 1
	GetFileList           = 0x2000F
	GetFileListRet        = GetFileList + 1
	SequentialDownload    = 0x20011
	SequentialDownloadRet = 0x20012
	SequentialUpload      = 0x20013
	SequentialUploadRet   = 0x20014
)
