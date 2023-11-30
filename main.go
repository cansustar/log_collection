package main

import (
	"fmt"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"log_collection/etcd"
	"log_collection/kafka"
	"log_collection/tailfile"
)

// 日志收集的客户端
// 类似的开源项目还有filebeat
// 收集指定目录下的日志文件，并发送到kafka中

type Config struct {
	KafkaConfig   `ini:"kafka"` // ini标签的值要和config.ini中的节相同
	CollectConfig `ini:"collect"`
	EtcdConfig    `ini:"etcd"`
}

type EtcdConfig struct {
	Address    string `ini:"address"`
	CollectKey string `ini:"collect_key"`
}

type KafkaConfig struct {
	Address  string `ini:"address"`
	Topic    string `ini:"topic"`
	ChanSize int64  `ini:"chan_size"`
}

type CollectConfig struct {
	LogFilePath string `ini:"logfile_path"`
}

func run() {
	select {}
}

func main() {
	// 0. 读配置文件 go-ini， 加载kafka和collect的配置项
	// 下面要用反射给结构体赋值，要改变结构体变量的值，go中函数传参都是拷贝，用new，可以得到一个结构体指针
	var configObj = new(Config)
	err := ini.MapTo(configObj, "./conf/config.ini")
	if err != nil {
		logrus.Error("load config failed, err: %v", err)
		return
	}
	// 打印配置信息
	// "%#v" 这个格式字符串，它用于打印 Go 语言的值的 Go 语法表示形式。具体而言，%#v 会以 Go 语法格式输出变量的值，包括类型信息和字段名（如果是结构体）。
	fmt.Printf("%#v\n", configObj)
	// 1. 初始化（做好准备工作） 连接kafka，检查tail连接是否正常  连接kafka,初始化msgChan,启后台goroutine往kafka发msg
	err = kafka.Init([]string{configObj.KafkaConfig.Address}, configObj.KafkaConfig.ChanSize)
	if err != nil {
		logrus.Error("init kafka failed, err:%v", err)
		return
	}
	logrus.Info("init kafka success!")
	// 优化： 初始化etcd连接，从etcd中拉取要收集日志的配置项
	err = etcd.Init([]string{configObj.EtcdConfig.Address})
	if err != nil {
		logrus.Errorf("init etcd failed, err:%v", err)
		return
	}
	// 从etcd中拉取要收集日志的配置项
	allConf, err := etcd.GetConf(configObj.EtcdConfig.CollectKey)
	if err != nil {
		logrus.Errorf("get conf from etcd failed, err:%v", err)
	}
	fmt.Println(allConf)
	// 派一个小弟去监控etcd中，configObj.EtcdConfig.CollectKey 对应值的变化
	// 启动一个goroutine监控etcd中key的value的变化
	go etcd.WatchConf(configObj.EtcdConfig.CollectKey)
	// 2. 根据配置中的日志路径，创建对应的tailobj，使用tail去收集日志
	// 从结构体中加载对象
	// 将从etcd中获取的配置项
	err = tailfile.Init(allConf)
	if err != nil {
		logrus.Error("init tailfile failed, err:%v", err)
		return
	}
	logrus.Info("init tailfile success!")
	// 3. 把日志通过sarama发往kafka
	// 从tailobj.lines中一行行读日志，包装成kafka msg,丢到MsgChan中
	run()

}
