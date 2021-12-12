package schema

const (
	WaitingStatus = "waiting" // 等待发送
	PendingStatus = "pending" // 已经发送还没打包到区块
	SuccessStatus = "success" // 打包到区块
	FailedStatus  = "failed"  // 失败
)
