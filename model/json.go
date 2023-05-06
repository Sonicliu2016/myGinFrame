package model

// 分块上传 : 初始化信息
type BlockFileUploadInfo struct {
	UploadId string `json:"uploadId"`
	FileHash string `json:"fileHash"`
	FileSize int    `json:"fileSize"`
	//每一个分块的大小
	BlockSize  int `json:"blockSize"`
	BlockCount int `json:"blockCount"`
	// 已经上传完成的分块索引列表
	CompletedBlockIdxs []int `json:"completedBlockIdxs"`
}
