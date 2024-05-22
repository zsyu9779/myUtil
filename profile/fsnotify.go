package profile

import (
	"github.com/fsnotify/fsnotify"
	"log"
	"time"
)

// 动态加载配置文件使用fsnotify够轻量（viper暂时不用）
type FsnotifyMonit struct {
	watcher  *fsnotify.Watcher
	lastTime time.Time
}

// 获取fsnotify
// 调用例子参考TestFsnotifyMonit_AddWatcher
func NewFsnotify() (*FsnotifyMonit, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FsnotifyMonit{
		watcher: watcher,
	}, nil
}

// 添加监听观察的路径或者文件
func (fs *FsnotifyMonit) AddWatcher(path string) error {
	return fs.watcher.Add(path)
}

func (fs *FsnotifyMonit) Run(fn func()) {
	for {
		select {
		case event, ok := <-fs.watcher.Events:
			//fmt.Println(event)
			//正常情况
			if ok {
				nowTime := time.Now()
				//防止短时间被频繁刷新
				if nowTime.Sub(fs.lastTime).Seconds() > 2 {
					go fs.handleEvent(event, fn)
				}
			}
		case err, ok := <-fs.watcher.Errors:
			if !ok {
				return
			}
			log.Println("fs watcher error:", err)
		}
	}
}

func (fs *FsnotifyMonit) Close() error {
	return fs.watcher.Close()
}

// 获取事件确认执行
func (fs *FsnotifyMonit) handleEvent(event fsnotify.Event, fn func()) {
	if event.Op == fsnotify.Create || event.Op == fsnotify.Write {
		fs.lastTime = time.Now()
		fn()
	}
}
