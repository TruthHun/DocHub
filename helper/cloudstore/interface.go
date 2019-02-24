package CloudStore

// TODO: 获取ObjectURL的时候，如果时间为0，则表示不签名；大于0，则生成签名文件
type CloudStore interface {
	PutObject(string, string, map[string]string) error
	DeleteObjects([]string) error
	GetObjectURL(string, int64) (string, error)
}
