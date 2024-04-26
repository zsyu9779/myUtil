package go_behind

import (
	"codeup.aliyun.com/aha/social_aha_gotool/gopool"
	"github.com/sirupsen/logrus"
	"testing"
	"time"
)

func TestGoBehind(t *testing.T) {
	pool := gopool.NewPool("test", 100, gopool.NewConfig())
	InitGoBehind(&GoBehindConfig{
		NamespaceMode: "test",
		Logger:        logrus.New(),
		Pool:          pool,
	})
	for i := 0; i < 1; i++ {
		//i1 := i
		//GoBehind(func() {
		//	if i1%5 == 0 {
		//		panic(fmt.Sprintf("panic %d", i1))
		//	}
		//	t.Log("test")
		//})
		GoBehindWithParam(func(param []interface{}) {
			//if i1%5 == 0 {
			//	panic(fmt.Sprintf("panic %d ", i1))
			//}
			aaa := param[1].([]string)
			t.Log("test", aaa)
		}, []interface{}{1, []string{"aaa"}, 3}...)
	}
	select {
	case <-time.After(time.Second * 3):
		t.Log("test success")
		break
	}
}

func TestGoBehindWithParam(t *testing.T) {
	aaa := []interface{}{1, []string{"aaa"}, 3}
	a := aaa[1].([]string)
	t.Log(a)
}
