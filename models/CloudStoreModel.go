package models

type CloudStore struct {
	Private bool
}

// 创建云存储
func NewCloudStore(private ...bool) (cs *CloudStore) {
	cs = &CloudStore{}
	if len(private) > 0 {
		cs.Private = private[0]
	}
	return
}
