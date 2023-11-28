package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"time"
)

func test() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v", err)
	}
	defer cli.Close()
	// watch 检测这个key值的变化
	watchCh := cli.Watch(context.Background(), "Cansu")
	for wresp := range watchCh {
		for _, evt := range wresp.Events {
			fmt.Printf("type:%s, key: %s,value: %s\n", evt.Type, evt.Kv.Key, evt.Kv.Value)
		}
	}
	//// put
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//_, err = cli.Put(ctx, "Cansu", "yolo")
	//if err != nil {
	//	fmt.Printf("put to etcd failed, err:%v", err)
	//	return
	//}
	//cancel()
	//// get
	//ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	//gr, err := cli.Get(ctx, "Cansu")
	//if err != nil {
	//	fmt.Printf("get from etcd failed, err:%v", err)
	//	return
	//}
	//for _, ev := range gr.Kvs {
	//	fmt.Printf("key: %s  value:%s\n", ev.Key, ev.Value)
	//}
	//cancel()
}
