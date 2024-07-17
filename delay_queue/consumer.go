package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

type handleFunc func(msg Message)

func defaultHander(msg Message) {
	fmt.Println(msg)
}

type Consumer struct {
	ctx      context.Context
	duration time.Duration
	ch       chan []string
	handler  handleFunc
	logger   *log.Logger
}

func NewConsumer(ctx context.Context, handler handleFunc) *Consumer {
	return &Consumer{
		ctx:      ctx,
		duration: time.Second,
		ch:       make(chan []string, 1000),
		handler:  handler,
		logger:   log.New(log.Writer(), "consumer: ", log.LstdFlags),
	}
}

func (c *Consumer) listen(redisClient *redis.Client, topic string) {
	// 从 Hashes 中获取数据并处理
	c.goBehind(func() {
		for {
			select {
			case ret := <-c.ch:
				// 批量从hashes中获取数据信息
				key := topic + HashSuffix
				result, err := redisClient.HMGet(c.ctx, key, ret...).Result()
				if err != nil {
					c.logger.Println(err)
				}

				if len(result) > 0 {
					redisClient.HDel(c.ctx, key, ret...)
				}

				msg := Message{}
				for _, v := range result {
					// 由于hashes 和 scoreSet 非事务操作，会出现删除了set但hashes未删除的情况
					if v == nil {
						continue
					}
					str := v.(string)
					json.Unmarshal([]byte(str), &msg)

					// 处理逻辑
					c.goBehind(func() {
						c.handler(msg)
					})
				}

			}
		}
	})
	ticker := time.NewTicker(c.duration)
	defer ticker.Stop()
	for {
		select {
		case <-c.ctx.Done():
			log.Println("consumer quit:", c.ctx.Err())
			return
		case <-ticker.C:
			// read data from redis
			zero := strconv.Itoa(0)
			now := strconv.Itoa(int(time.Now().Unix()))
			opt := &redis.ZRangeBy{
				Min: zero,
				Max: now,
			}

			key := topic + SetSuffix
			result, err := redisClient.ZRangeByScore(c.ctx, key, opt).Result()
			if err != nil {
				log.Fatal(err)
				return
			}
			fmt.Println(result)

			// 获取到数据
			if len(result) > 0 {
				// 从 sorted sets 中移除数据
				redisClient.ZRemRangeByScore(c.ctx, key, zero, now)

				// 写入 chan, 进行hashes处理
				c.ch <- result
			}
		}
	}
}

func (c *Consumer) goBehind(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64<<10)
				buf = buf[:runtime.Stack(buf, false)]
				log := time.Now().Format("2006-01-02 15:04:05.000")
				log += fmt.Sprintf("\t%v\n", r)
				log += " -- stack:" + string(buf)
				function, location := caller()
				c.logger.Printf("function: %s, location: %s, stack: %s", function, location, string(buf))
			}
		}()
		f()
	}()
}

func caller() (string, string) {
	pc, file, line, _ := runtime.Caller(3)
	function := runtime.FuncForPC(pc).Name()                    // 获取函数名
	location := fmt.Sprintf("%s:%d", filepath.Base(file), line) // 获取文件名
	return function, location
}
