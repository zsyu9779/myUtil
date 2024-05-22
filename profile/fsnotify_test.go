package profile

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"testing"
	"time"
)

func reload() {
	fmt.Println("I am run reload func")
}

func TestFsnotifyMonit_AddWatcher(t *testing.T) {
	fs, _ := NewFsnotify()
	defer fs.Close()
	if err := fs.AddWatcher("./"); err != nil {
		t.Errorf("AddWatcher() error = %v", err)
	}
	fs.Run(reload)

	for {
		time.Sleep(1)
	}
	fmt.Println("ss")
}

func TestFsnotify(t *testing.T) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("NewWatcher failed: ", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		defer close(done)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Printf("%s %s\n", event.Name, event.Op)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("./")
	if err != nil {
		log.Fatal("Add failed:", err)
	}
	<-done
}

func TestChannl(t *testing.T) {
	// 保持程序运行
	done := make(chan bool, 1)
	go func() {
		defer close(done)
	}()
	<-done
}

func TestFor(t *testing.T) {
	done := make(chan bool, 1)
	done <- true
	for {
		select {
		case d := <-done:
			fmt.Println("ss")
			if d {
				return
			}
			t.Error("ss22s")
		default:
			t.Error("ss111s")
		}
		t.Error("for")
	}
	t.Error("ssss")
}
