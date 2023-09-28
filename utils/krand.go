package utils

import (
	"math/rand"
	"time"
)

const (
	KC_RAND_KIND_NUM   = 0 // 数字
	KC_RAND_KIND_LOWER = 1 // 小写字母
	KC_RAND_KIND_UPPER = 2 // 大写字母
	KC_RAND_KIND_ALL   = 3 // 全部
)

func Krand(size int, kind int) []byte {
	//类型0：十进制数字（ASCII码范围：48-57）
	//类型1：小写字母（ASCII码范围：97-122）
	//类型2：大写字母（ASCII码范围：65-90）
	ikind, kinds, result := kind, [][]int{[]int{10, 48}, []int{26, 97}, []int{26, 65}}, make([]byte, size)
	//如果 kind 不在指定范围内，则默认为生成所有类型的随机字符串。
	isAll := kind > KC_RAND_KIND_UPPER || kind < KC_RAND_KIND_NUM
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		if isAll {
			ikind = rand.Intn(3)
		}
		scope, base := kinds[ikind][0], kinds[ikind][1]
		result[i] = uint8(base + rand.Intn(scope))
	}

	return result
}
