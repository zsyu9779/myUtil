package g_rediscache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

type ListOptions struct {
	Options
	// 此处一定要注意，除非要取第一条数据，否则一定要设置Start和Stop
	Start int64
	Stop  int64
}

func ListAop(options *ListOptions, fallback func() ([]interface{}, error)) ([]interface{}, bool, error) {
	err := options.validate()
	if err != nil {
		return nil, false, err
	}
	cacheVs, err := GetRedisClient().LRange(options.Ctx, options.Key, options.Start, options.Stop).Result()
	var result []interface{}
	// 从cache里取到值
	if len(cacheVs) > 0 {
		if len(cacheVs) == 1 && cacheVs[0] == EmptyFlag {
			return result, true, nil
		}
		for _, cacheV := range cacheVs {
			if cacheV == EmptyFlag {
				continue
			}
			rtv := reflect.New(options.Rt)
			rv := rtv.Interface()
			err := json.Unmarshal([]byte(cacheV), rv)
			if err != nil {
				return nil, false, err
			}
			logrus.Info("[Reflect] Value.", rv)
			result = append(result, reflect.ValueOf(rv).Elem().Interface())
			logrus.Info("[Reflect] Result.", result)
		}
		return result, true, nil
	} else {
		if err != nil {
			return nil, false, err
		}
		exists := GetRedisClient().Exists(options.Ctx, options.Key).Val()
		if exists != 0 {
			return result, true, nil
		}
	}

	logrus.Warn("[REDIS][LIST] cant get value from redis cache, maybe load from db!")
	result, err = fallback()
	if err != nil {
		return nil, false, err
	}
	// 回填
	rewriteCount := 0
	if result != nil && len(result) > 0 {
		var cacheVList []string
		for _, item := range result {
			cacheV, isEmpty, err := GetCacheValueItem(item)
			if err != nil {
				logrus.Warn("[REDIS][LIST] GetCacheValueItem error!", err)
				continue
			}
			if !isEmpty {
				rewriteCount++
				cacheVList = append(cacheVList, cacheV)
			}
		}
		UseGLock(options.Ctx, fmt.Sprintf("%s:lock", options.Key)).Then(func() (interface{}, error) {
			GetRedisClient().RPush(options.Ctx, options.Key, cacheVList)
			return nil, nil
		})
		if rewriteCount > 0 {
			if options.Expires == 0 {
				options.Expires = defaultExpire
			}
			GetRedisClient().Expire(options.Ctx, options.Key, options.Expires)
		}
	}

	// 空值回填
	if rewriteCount == 0 && options.EmptyExpires > 0 {
		GetRedisClient().RPush(options.Ctx, options.Key, EmptyFlag)
		GetRedisClient().Expire(options.Ctx, options.Key, options.EmptyExpires)
		logrus.Warn("[REDIS][LIST] cache empty value, key:", options.Key)
	}

	return result, false, nil

}

type ListAopProxy struct {
	options ListOptions
}

func (p *ListAopProxy) WithExpires(expires time.Duration) *ListAopProxy {
	p.options.Expires = expires
	return p
}

func (p *ListAopProxy) WithEmptyExpires(emptyExpires time.Duration) *ListAopProxy {
	p.options.EmptyExpires = emptyExpires
	return p
}

func (p *ListAopProxy) WithStart(start int64) *ListAopProxy {
	p.options.Start = start
	return p
}

func (p *ListAopProxy) WithStop(stop int64) *ListAopProxy {
	p.options.Stop = stop
	return p
}

func (p *ListAopProxy) Then(f func() ([]interface{}, error)) ([]interface{}, bool, error) {
	return ListAop(&p.options, f)
}

func UseListAop(ctx context.Context, key string, rt reflect.Type) *ListAopProxy {
	return &ListAopProxy{ListOptions{Options: Options{Ctx: ctx, Key: key, Rt: rt}}}
}
