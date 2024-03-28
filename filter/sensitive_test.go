package util

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestACFilter_Build(t *testing.T) {
	filter, _ := NewACFilter().BuildWithFunc(func() ([]string, error) {
		b, err := ioutil.ReadFile("./key.txt")
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		return strings.Split(string(b), "|"), nil
	})
	fmt.Println(filter.Contains("傻/(^^&%逼哈哈哈"))

}
