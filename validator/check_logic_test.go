package validator

import (
	"fmt"
	"regexp"
	"testing"
)

func TestCheckPassword(t *testing.T) {
	//数字字母验证
	res := Check2Password("aaa@aaaaa")
	fmt.Println(res)
}

func TestCheckNickname(t *testing.T) {
	res := CheckNickname("话发顺大法师地方¥!")
	fmt.Println(res)
}

func TestMustComplie(t *testing.T) {
	buf := "$cpuload_useage31231%312 +$cpuloac_ppp /$disk_useage "
	q := regexp.MustCompile("[^$]*\\w")
	x := q.FindAllString(buf, 10)
	for i, s := range x {
		fmt.Println(i, s)
	}
}
