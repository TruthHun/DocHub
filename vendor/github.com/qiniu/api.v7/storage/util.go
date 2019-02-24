package storage

import (
	"time"
)

// ParsePutTime 提供了将PutTime转换为 time.Time 的功能
func ParsePutTime(putTime int64) (t time.Time) {
	t = time.Unix(0, putTime*100)
	return
}

// IsContextExpired 检查分片上传的ctx是否过期，提前一天让它过期
// 因为我们认为如果断点继续上传的话，最长需要1天时间
func IsContextExpired(blkPut BlkputRet) bool {
	if blkPut.Ctx == "" {
		return false
	}
	target := time.Unix(blkPut.ExpiredAt, 0).AddDate(0, 0, -1)
	now := time.Now()
	return now.After(target)
}
