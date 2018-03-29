// 公共组件提供的服务列表
// 1. 把interface{}转化为map[string]string
// 2. 把string转化为int型
// 3. 模糊查找服务列表
// 4. 获取随机大小的字符串
// 5. 获取子串
// 6. 两个整数列表，求并集
// 7. 根据url，获取正则表达式字符串   例如： input: /v1/accounts/:id/invalid output: <2, :id>
// 8. 两个整数列表，求差积
package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/1046102779/slicelement"
	"github.com/pkg/errors"
)

type XlsxImportStats struct {
	Count        int `json:"count"`         // xlsx文件记录总数
	FilterCount  int `json:"filter_count"`  // 过滤条数
	SuccessCount int `json:"success_count"` // 成功插入条数
	UpdateCount  int `json:"update_count"`  // 成功更新条数
}

type HeaderParamInfo struct {
	UserId    int
	CompanyId int
}

func Md5String(str string) (md5Str string) {
	h := md5.New()
	io.WriteString(h, str)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func CheckEmailPattern(str string) bool {
	pattern := `^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`
	p := regexp.MustCompile(pattern)
	return p.MatchString(str)
}

func CheckMobilePattern(str string) bool {
	pattern := `^\d{9,}$`
	p := regexp.MustCompile(pattern)
	return p.MatchString(str)
}

// 生成随机字符串
func GetRandomString(size int) string {
	bytes := []byte("0123456789")
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < size; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func ConvertStrToInt(str string) (result int) {
	if "" == strings.TrimSpace(str) {
		result = 0
		return
	}
	resultTemp, _ := strconv.ParseInt(str, 10, 64)
	return int(resultTemp)
}

// interface{} 转化为 map
func ConvertInterfaceToMap(src interface{}) (dest map[string]interface{}, isMap bool) {
	isMap = false
	dest = map[string]interface{}{}
	v := reflect.ValueOf(src)
	if v.Kind() != reflect.Map {
		return
	}
	for _, key := range v.MapKeys() {
		dest[key.String()] = v.MapIndex(key).Interface()
	}
	isMap = true
	return
}

func SubString(str string, begin, length int) (substr string) {
	// 将字符串的转换成[]rune
	rs := []rune(str)
	lth := len(rs)

	// 简单的越界判断
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length
	if end > lth {
		end = lth
	}

	// 返回子串
	return string(rs[begin:end])
}

// 6. 两个整数列表，求并集
func GetUnionByInts(slice1 []int, slice2 []int) (dest []int) {
	if slice1 == nil || len(slice1) <= 0 {
		return slice2
	}
	if slice2 == nil || len(slice2) <= 0 {
		return slice1
	}
	for index := 0; index < len(slice1); index++ {
		if isExist, _ := slicelement.Contains(slice2, slice1[index], ""); !isExist {
			slice2 = append(slice2, slice1[index])
		}
	}
	return slice2
}

// 7. 根据url，获取正则表达式字符串   例如： input: /v1/accounts/:id/invalid output: <2, :id>
func GetRegexpPairByUrl(url string) (position int, regTarget string) {
	if strings.TrimSpace(url) == "" {
		return -1, ""
	}
	fields := strings.Split(url, "/")
	if fields == nil || len(fields) <= 1 {
		return -1, ""
	}
	for index := 1; index < len(fields); index++ {
		if strings.HasPrefix(fields[index], ":") {
			return index - 1, fields[index]
		}
	}
	return -1, ""
}

// 早上凌晨时间
func GetEarliestDate(now *time.Time) (ret time.Time, err error) {
	timeStr := fmt.Sprintf("%s%s", (*now).Format("20060102"), "000000")
	loc, _ := time.LoadLocation("Asia/Shanghai")
	ret, err = time.ParseInLocation("20060102150405", timeStr, loc)
	if err != nil {
		err = errors.Wrap(err, "getEarliestDate")
		return
	}
	return
}
