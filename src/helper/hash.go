package helper

import (
	"crypto/md5"
	"encoding/hex"
)

func Md5(src string) string {
	md5ctx := md5.New()
	md5ctx.Write([]byte(src))
	return hex.EncodeToString(md5ctx.Sum(nil))
}

func Hash2Int(src string) int {
	result := 0
	for _, v := range src {
		if v >= '0' && v <= '9' {
			result = result*16 + int(v-'0')
		} else if v >= 'a' && v <= 'b' {
			result = result*16 + int(v-'a'+10)
		}
	}

	return result
}
