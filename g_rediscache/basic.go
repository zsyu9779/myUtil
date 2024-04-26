package g_rediscache

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"reflect"
	"time"
)

const (
	// 因为redis不能缓存空值，但是我们又会经常需要缓存空值防止频繁击穿cache
	// 因此用一个标识存储空的值
	EmptyFlag      = "###+--**-+###"
	defaultTimeout = 5 * time.Second
	defaultExpire  = 30 * time.Second
)

func init() {

}

type Options struct {
	Ctx context.Context
	// 非空
	Key string
	// 非空, 如果是集合类型，表示集合里面元素的类型
	Rt           reflect.Type
	Expires      time.Duration
	EmptyExpires time.Duration
}

func (o *Options) validate() error {
	if o.Key == "" {
		return errors.New("Key must not be empty!")
	}
	if o.Rt == nil {
		return errors.New("Rt must not be empty!")
	}
	return nil
}

// 返回的三个参数，依次是: cache值，是否空，错误信息
func GetCacheValueItem(v interface{}) (string, bool, error) {
	cacheV := ""
	cLength := 0
	switch v.(type) {
	// string 类型单独处理
	case string:
		cacheV = v.(string)
		cLength = len(v.(string))
	default:
		jsonB, err := json.Marshal(v)
		if err != nil {
			return "", true, err
		}
		cacheV = string(jsonB)
		cLength = len(cacheV)
	}
	return cacheV, cLength == 0, nil
}

func getRand() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(10000000)
}

func RedisTestSetup() {
	cli := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:6379",
		DialTimeout:  8 * time.Second,
		ReadTimeout:  8 * time.Second,
		WriteTimeout: 8 * time.Second,
		DB:           0,
	})
	InitRedisClient(cli)
}
