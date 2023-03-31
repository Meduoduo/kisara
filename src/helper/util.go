package helper

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"strings"
)

const (
	VAL_TYPE_DEFAULT             = 0x0
	VAL_TYPE_INT                 = 0x1
	VAL_TYPE_STRING              = 0x2
	VAL_TYPE_INTERFACE_ARRAY     = 0x3
	VAL_TYPE_STRING_TO_INTERFACE = 0x4
	VAL_TYPE_JSON_NUMBER         = 0x5 //float64
	VAL_TYPE_FLOAT64             = 0x5
)

func Random(min int, max int) int {
	res, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return min + int(res.Int64())
}

func Abs(x int) int {
	if x >= 0 {
		return x
	}
	return -x
}

func RandomStr(len int) string {
	var result = ""
	for i := 0; i < len; i++ {
		result += string(byte(Random(97, 122)))
	}
	return result
}

//根据验证码类型与操作名获取key
func GetCaptchaMethodKey(otype string, method string) string {
	return "captcha_method_" + method + "_" + otype
}

func HandleSensitiveString(src string) string {
	if len(src) <= 4 {
		return "****"
	}
	var dst = []byte(src)
	for i := 2; i < len(dst)-2; i++ {
		dst[i] = '*'
	}
	return string(dst)
}

//根据ip字符串返回ip的int形式，对于ipv6，映射到ipv4上
func ParseIpFromStr(ip string) int {
	index := 0
	ip_len := len(ip)
	ip_int := 0

	if ip[index] == '[' {
		//ipv6

	} else if ip[index] >= '0' && ip[index] <= '9' {
		//ipv4
		pos := 0
		temp_index := 0
		for index < ip_len {
			if ip[index] == '.' {
				ip_part, _ := strconv.Atoi(ip[temp_index:index])
				ip_int += (ip_part << (8 * (3 - pos)))
				temp_index = index + 1
				index++
				pos++
			} else if ip[index] == ':' {
				ip_part, _ := strconv.Atoi(ip[temp_index:index])
				ip_int += ip_part
				break
			}
			index++
		}
	}
	return ip_int
}

func GetInterfaceType(i interface{}) int {
	switch i.(type) {
	case string:
		return VAL_TYPE_STRING
	case float64:
		return VAL_TYPE_JSON_NUMBER
	case int:
		return VAL_TYPE_INT
	case []interface{}:
		return VAL_TYPE_INTERFACE_ARRAY
	case map[string]interface{}:
		return VAL_TYPE_STRING_TO_INTERFACE
	}
	return VAL_TYPE_DEFAULT
}

func StringJoin(strs ...string) string {
	n := 0
	for i := 0; i < len(strs); i++ {
		n += len(strs[i])
	}

	var b strings.Builder
	b.Grow(n)
	for _, s := range strs {
		b.WriteString(s)
	}
	return b.String()
}

// Pointer type is recommended
func ArrayFilter[T any](arr []T, f func(T) bool) []T {
	// allocate a new array with the same capacity as the original
	// to avoid reallocation
	res := make([]T, 0, len(arr))
	for _, v := range arr {
		if f(v) {
			res = append(res, v)
		}
	}
	// shrink the capacity to match the length
	return res[:len(res):len(res)]
}