package util

import (
	"crypto/md5"
	"encoding/hex"
)

func MD5(pass string) (str string) {
	mh := md5.Sum([]byte(pass))
	return hex.EncodeToString(mh[:])
}
