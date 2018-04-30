package helper

import "github.com/huichen/sego"

const (
	KEY_AES    = "hello wenku.it"
	SALT       = "wenku.it"
	CACHE_CONF = `{"CachePath":"./cache/runtime","FileSuffix":".cache","DirectoryLevel":2,"EmbedExpiry":120}`
)

var Segmenter sego.Segmenter
