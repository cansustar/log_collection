package test

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"time"
)

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v", err)
		return
	}
	defer cli.Close()

	// put  当使用context.WithTimeout后，最好在紧跟其后defer cancel(),而不直接cancel(),避免在 Put 操作完成后，没有等待 Put 操作的结果就调用了 cancel 函数。这可能导致 Put 操作被取消，尽管它可能仍在执行。
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	str := `[{"path":"D:\\N+_project\\cumt_life_jcry\\logs\\error\\error.log","topic":"cumt_life_jcry_log"},{"path":"D:\\N+_project\\cumt_life_feedback\\logs\\django\\error\\error.log","topic":"cumt_life_feedback_log"}]`
	_, err = cli.Put(ctx, "collect_log_conf", str)
	if err != nil {
		fmt.Printf("put to etcd failed, err:%v", err)
		return
	}
	cancel()
	// get
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	gr, err := cli.Get(ctx, "collect_log_conf")
	if err != nil {
		fmt.Printf("get from etcd failed, err:%v", err)
		return
	}
	for _, ev := range gr.Kvs {
		fmt.Printf("key: %s  value:%s\n", ev.Key, ev.Value)
	}
	cancel()
}
