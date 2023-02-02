package g_es

import (
	"github.com/olivere/elastic/v7"
	"os"
	"time"
)

var esClient *elastic.Client

func InitEsCli(client *elastic.Client) {
	esClient = client
}

func GetEsCli() *elastic.Client {
	return esClient
}

func initEsConnect() *elastic.Client {
	var err error
	cli, err := elastic.NewClient(
		elastic.SetURL("http://172.26.134.3:9200"),
		elastic.SetBasicAuth("", ""),
		// 启用gzip压缩
		elastic.SetGzip(true),
		// 设置监控检查时间间隔
		elastic.SetHealthcheckInterval(10*time.Second),
		// 是否探测非master节点
		elastic.SetSniff(false),
		elastic.SetErrorLog(os.Stdout),
	)
	if err != nil {
		panic(err)
	}
	return cli
}
