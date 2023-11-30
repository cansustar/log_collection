package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log_collection/common"
	"log_collection/tailfile"
	"time"
)

// etcd 相关操作

var (
	client *clientv3.Client
)

func Init(address []string) (err error) {
	client, err = clientv3.New(clientv3.Config{
		Endpoints:   address, // []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v", err)
		return
	}
	return
}

// GetConf 拉去日志收集配置项的函数
func GetConf(key string) (collectEntryList []common.CollectEntry, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	resp, err := client.Get(ctx, key)
	if err != nil {
		logrus.Errorf("get conf from etcd by key: %s failed, err: %v", key, err)
		return
	}
	if len(resp.Kvs) == 0 {
		logrus.Warningf(" get len: 0 conf from etcd by key: %s", key)
		return
	}
	ret := resp.Kvs[0]
	// 解析json字符串。将etcd取出来的ret.Value值，反序列化后，存到collectEntryList中
	err = json.Unmarshal(ret.Value, &collectEntryList)
	if err != nil {
		logrus.Errorf("json unmarshal failed, err:%v", err)
		return
	}
	return
}

// 监控etcd中日志收集项配置变化的函数
func WatchConf(key string) {
	for {
		wCh := client.Watch(context.Background(), key)
		for wresp := range wCh {
			logrus.Info("get a new config")
			for _, evt := range wresp.Events {
				fmt.Printf("config change.", "type:%s, key: %s,value: %s\n", evt.Type, evt.Kv.Key, evt.Kv.Value)
				var newConf []common.CollectEntry // 空的配置切片
				if evt.Type == clientv3.EventTypeDelete {
					// 如果是删除事件，
					logrus.Warning("etcd delete the key!!!")
					tailfile.SendNewConf(newConf) // 直接使用空配置， 删除了这个key的话意味删除所有的path,topic配置
					continue
				}
				err := json.Unmarshal(evt.Kv.Value, &newConf)
				if err != nil {
					logrus.Errorf("json unmarshal new conf failed, err: %s", err)
					continue
				}
				// 如果解析成功，则告诉tailfile模块使用新的配置
				tailfile.SendNewConf(newConf)
			}
		}
	}
}
